// Package legacyclient carries the fork's hand-written SigNoz API endpoints:
// the legacy alert/dashboard APIs plus the fork-added plugin-dashboard-v2,
// data_retention, downtime_schedule, ingestion_key, log_pipeline,
// notification_channel, route_policy and saved_view endpoints.
//
// Endpoint methods, request shapes, and response parsing are unchanged from
// the original fork client; transport (SIGNOZ-API-KEY auth header, heimdall
// retries, User-Agent) is delegated to the upstream WrappedClient so that
// generated and hand-written calls share one configuration path.
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
