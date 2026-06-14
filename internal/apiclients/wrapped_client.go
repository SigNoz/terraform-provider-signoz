package apiclients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
)

const (
	// DefaultEndpoint is the SigNoz API URL used when none is configured.
	DefaultEndpoint = "http://localhost:3301"

	// DefaultTimeout is the per-request HTTP timeout used when none is configured.
	DefaultTimeout = 35 * time.Second

	// DefaultRetryMax is the retry count used when none is configured.
	DefaultRetryMax = 10

	// Header for authentication.
	apiKeyHeader = "SIGNOZ-API-KEY"
)

// WrappedClient wraps the oapi-codegen-generated client with SigNoz auth, a heimdall retrier, and User-Agent stamping, wired once at provider configure
// time. Service code calls the generated methods via `.Gen` (e.g. `r.api.Gen.GetRoutePolicyByIDWithResponse(ctx, id)`).
type WrappedClient struct {
	Gen *ClientWithResponses

	// host and doer back Do, the escape hatch for hand-written endpoints that
	// aren't part of the generated client (the legacy alert/dashboard APIs).
	host *url.URL
	doer HttpRequestDoer
}

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

	return &WrappedClient{Gen: gen, host: host, doer: doer}, nil
}

// BaseURL returns the configured server root without a trailing slash.
func (c *WrappedClient) BaseURL() string {
	return strings.TrimRight(c.host.String(), "/")
}

// Do issues a request to a hand-written endpoint that isn't part of the
// generated client, through the same auth + retry + User-Agent transport. It
// returns the response body on a 2xx and an *APIError otherwise. The legacy
// alert/dashboard resources use it until those endpoints are generated.
func (c *WrappedClient) Do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	endpoint, err := url.JoinPath(c.BaseURL(), path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, ErrorFromResponse(resp, raw)
	}

	return raw, nil
}
