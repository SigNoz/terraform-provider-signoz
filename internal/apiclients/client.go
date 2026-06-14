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

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
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

// APIError is a non-2xx response from the SigNoz API. It carries the decoded
// `{error, status}` body in Render (nil when the body didn't match the
// envelope) rather than copying its fields, so apitypes.RenderErrorResponse
// stays the single source of truth. Raw is the unparsed body, kept for the
// fallback case.
type APIError struct {
	StatusCode int
	Raw        string
	Render     *apitypes.RenderErrorResponse
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "signoz api: HTTP %d", e.StatusCode)

	if e.Render == nil {
		if e.Raw != "" {
			fmt.Fprintf(&b, ": %s", e.Raw)
		}
		return b.String()
	}

	body := e.Render.Error
	if body.Code != "" {
		fmt.Fprintf(&b, " (%s)", body.Code)
	}
	if body.Message != "" {
		fmt.Fprintf(&b, ": %s", body.Message)
	}
	if body.Errors != nil {
		for _, d := range *body.Errors {
			if d.Message != nil && *d.Message != "" {
				fmt.Fprintf(&b, " | %s", *d.Message)
			}
		}
	}

	return b.String()
}

// decodeError parses SigNoz's `{error, status}` failure body into the same
// `apitypes.RenderErrorResponse` the generated client decodes its JSON4xx/5xx
// fields into. Render stays nil (and Raw carries the body) when it doesn't match.
func decodeError(status int, raw []byte) *APIError {
	out := &APIError{StatusCode: status, Raw: strings.TrimSpace(string(raw))}

	var re apitypes.RenderErrorResponse
	if err := json.Unmarshal(raw, &re); err != nil {
		return out
	}

	if re.Error.Code != "" || re.Error.Message != "" || (re.Error.Errors != nil && len(*re.Error.Errors) > 0) {
		out.Render = &re
	}

	return out
}
