// Package client is the SigNoz HTTP API client used by the provider's
// resources and datasources. Requests are retried with heimdall's constant
// backoff retrier; responses are unwrapped from SigNoz's `{data, status}`
// envelope (and `{error, status}` on failure) into a typed `*APIError`.
package apiclients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// Client is the heimdall-backed SigNoz API client. The generated
// `*apiclients.ClientWithResponses` is the primary surface; the
// hand-written `Do` method below stays around for resources that haven't
// migrated yet.
type WrappedClient struct {
	httpClient *httpclient.Client
	endpoint   *url.URL
	token      string
	userAgent  string

	// Gen is the oapi-codegen-generated client. Service code calls
	// `r.client.Gen.GetDowntimeScheduleByIDWithResponse(ctx, id)` etc.
	// directly. Constructed in New so heimdall + auth headers are wired
	// once at provider configure time.
	Gen *ClientWithResponses
}

// New constructs a Client. endpoint, timeout, and retryMax fall back to their
// defaults when zero-valued.
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

	c := &WrappedClient{
		httpClient: httpc,
		endpoint:   host,
		token:      token,
		userAgent:  ua,
	}

	doer := &authDoer{inner: httpc, token: token, userAgent: ua}
	gen, err := NewClientWithResponses(
		strings.TrimRight(host.String(), "/"),
		WithHTTPClient(doer),
	)
	if err != nil {
		return nil, fmt.Errorf("build generated client: %w", err)
	}
	c.Gen = gen
	return c, nil
}

// authDoer implements `apiclients.HttpRequestDoer` by delegating to
// heimdall and stamping the SigNoz auth + UA headers + Accept on every
// request. Lives here (not in apiclients) because apiclients is generated
// and we want all auth handling in one place. tflog is used for parity
// with the hand-written `Do` so debug streams stay consistent.
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

// ErrorFromResponse converts a non-2xx HTTP response from the generated
// client (raw `*http.Response` plus its body bytes) into a typed
// `*APIError`. Use from service code when the generated `*Response`
// wrapper's typed `JSON4xx` field is nil — kept compatible with the
// existing `IsNotFound`-based plumbing.
func ErrorFromResponse(resp *http.Response, body []byte) *APIError {
	if resp == nil {
		return &APIError{Raw: strings.TrimSpace(string(body))}
	}
	return decodeError(resp.StatusCode, body)
}

// Do executes an HTTP request against the configured endpoint. body may be
// nil. out may be nil (e.g. 204 responses). Successful responses are decoded
// as `{data, status}` and the `data` field is unmarshaled into out (when
// non-nil). Non-2xx responses are decoded as `{error, status}` and returned
// as *APIError.
func (c *WrappedClient) Do(ctx context.Context, method, path string, body, out any) error {
	full := strings.TrimRight(c.endpoint.String(), "/") + path

	var reqBody io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, full, reqBody)
	if err != nil {
		return fmt.Errorf("build %s %s: %w", method, full, err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set(apiKeyHeader, c.token)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	tflog.Debug(ctx, "signoz api request", map[string]any{"method": method, "url": full})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, full, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	tflog.Debug(ctx, "signoz api response", map[string]any{
		"method":      method,
		"url":         full,
		"status_code": resp.StatusCode,
		"body_len":    len(raw),
	})

	if resp.StatusCode/100 != 2 {
		return decodeError(resp.StatusCode, raw)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent || len(raw) == 0 {
		return nil
	}

	var env successEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode success envelope: %w", err)
	}
	if len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode response data: %w", err)
	}
	return nil
}

type successEnvelope struct {
	Data   json.RawMessage `json:"data"`
	Status string          `json:"status"`
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

// IsNotFound reports whether err is a 404 from the SigNoz API.
func IsNotFound(err error) bool {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.StatusCode == http.StatusNotFound
	}
	return false
}
