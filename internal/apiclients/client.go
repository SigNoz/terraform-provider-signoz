// Package apiclients wraps the oapi-codegen-generated SigNoz client with the
// transport concerns it doesn't cover: a heimdall retrier, SigNoz auth + UA
// headers, and decoding of the `{error, status}` failure envelope into a typed
// `*APIError`.
package apiclients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// DefaultEndpoint is the SigNoz API URL used when none is configured.
	DefaultEndpoint = "http://localhost:3301"
	// DefaultTimeout is the per-request HTTP timeout used when none is configured.
	DefaultTimeout = 35 * time.Second
	// DefaultRetryMax is the retry count used when none is configured.
	DefaultRetryMax = 10

	apiKeyHeader = "SIGNOZ-API-KEY"
)

// WrappedClient wraps the oapi-codegen-generated client with SigNoz auth, a
// heimdall retrier, and User-Agent stamping, wired once at provider configure
// time. Service code calls the generated methods via `.Gen`
// (e.g. `r.api.Gen.GetRoutePolicyByIDWithResponse(ctx, id)`).
type WrappedClient struct {
	Gen *ClientWithResponses
}

// New constructs a WrappedClient. endpoint, timeout, and retryMax fall back to
// their defaults when zero-valued.
func New(endpoint, token, agent, version string, timeout time.Duration, retryMax int) (*WrappedClient, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if retryMax == 0 {
		retryMax = DefaultRetryMax
	}

	host, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
	}

	netc := &http.Client{
		Timeout:   timeout,
		Transport: http.DefaultTransport,
	}

	httpc := httpclient.NewClient(
		httpclient.WithHTTPClient(netc),
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(heimdall.NewRetrier(
			heimdall.NewConstantBackoff(2*time.Second, 100*time.Millisecond),
		)),
		httpclient.WithRetryCount(retryMax),
	)

	ua := strings.TrimSpace(agent)
	if version != "" {
		if ua != "" {
			ua += "/"
		}
		ua += version
	}

	doer := &authDoer{inner: httpc, token: token, userAgent: ua}
	gen, err := NewClientWithResponses(
		strings.TrimRight(host.String(), "/"),
		WithHTTPClient(doer),
	)
	if err != nil {
		return nil, fmt.Errorf("build generated client: %w", err)
	}

	return &WrappedClient{Gen: gen}, nil
}

// authDoer implements the generated client's HttpRequestDoer by delegating to
// heimdall and stamping the SigNoz auth + UA + Accept headers on every request.
type authDoer struct {
	inner     *httpclient.Client
	token     string
	userAgent string
}

func (d *authDoer) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if d.token != "" {
		req.Header.Set(apiKeyHeader, d.token)
	}
	if d.userAgent != "" {
		req.Header.Set("User-Agent", d.userAgent)
	}

	tflog.Debug(req.Context(), "signoz api request", map[string]any{"method": req.Method, "url": req.URL.String()})

	resp, err := d.inner.Do(req)
	if resp != nil {
		tflog.Debug(req.Context(), "signoz api response", map[string]any{
			"method":      req.Method,
			"url":         req.URL.String(),
			"status_code": resp.StatusCode,
		})
	}
	return resp, err
}

// ErrorFromResponse converts a non-2xx HTTP response from the generated client
// (raw `*http.Response` plus its body bytes) into a typed `*APIError`.
func ErrorFromResponse(resp *http.Response, body []byte) *APIError {
	if resp == nil {
		return &APIError{Raw: strings.TrimSpace(string(body))}
	}
	return decodeError(resp.StatusCode, body)
}

type errorEnvelope struct {
	Error  errorBody `json:"error"`
	Status string    `json:"status"`
}

type errorBody struct {
	Code    string                `json:"code"`
	Errors  []errorBodyAdditional `json:"errors"`
	Message string                `json:"message"`
	URL     string                `json:"url"`
}

type errorBodyAdditional struct {
	Message string `json:"message"`
}

// APIError is a non-2xx response from the SigNoz API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	URL        string
	Details    []string
	Raw        string
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "signoz api: HTTP %d", e.StatusCode)
	if e.Code != "" {
		fmt.Fprintf(&b, " (%s)", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	for _, d := range e.Details {
		fmt.Fprintf(&b, " | %s", d)
	}
	if e.Message == "" && len(e.Details) == 0 && e.Raw != "" {
		fmt.Fprintf(&b, ": %s", e.Raw)
	}
	return b.String()
}

func decodeError(status int, raw []byte) *APIError {
	out := &APIError{StatusCode: status, Raw: strings.TrimSpace(string(raw))}
	var env errorEnvelope
	if err := json.Unmarshal(raw, &env); err == nil && (env.Error.Message != "" || env.Error.Code != "" || len(env.Error.Errors) > 0) {
		out.Code = env.Error.Code
		out.Message = env.Error.Message
		out.URL = env.Error.URL
		for _, e := range env.Error.Errors {
			if e.Message != "" {
				out.Details = append(out.Details, e.Message)
			}
		}
	}
	return out
}
