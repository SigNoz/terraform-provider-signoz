// Package legacyclient holds the fork's hand-written SigNoz endpoints — the
// legacy alert/dashboard APIs plus the fork-only resources. Transport
// (auth header, retries, User-Agent) is delegated to the upstream
// WrappedClient so generated and hand-written calls share one config path.
package legacyclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
)

// Client - SigNoz API client for hand-written endpoints.
type Client struct {
	hostURL *url.URL
	api     *apiclients.WrappedClient
}

// NewClient creates a legacy client bound to the WrappedClient's transport.
func NewClient(api *apiclients.WrappedClient) (*Client, error) {
	host, err := url.Parse(api.BaseURL())
	if err != nil {
		return nil, err
	}

	return &Client{hostURL: host, api: api}, nil
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")

	return c.api.DoRequest(ctx, req)
}
