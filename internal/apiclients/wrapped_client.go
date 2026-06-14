package apiclients

import (
	"fmt"
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

	return &WrappedClient{Gen: gen}, nil
}
