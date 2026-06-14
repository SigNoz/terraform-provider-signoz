package client

import "github.com/SigNoz/terraform-provider-signoz/internal/apiclients"

// Client adapts the shared wrapped API client to the legacy alert and dashboard
// endpoints, which are not yet part of the generated client. All transport
// (auth, retries, User-Agent) lives in apiclients.WrappedClient.
type Client struct {
	wc *apiclients.WrappedClient
}

// New wraps the shared API client for the legacy alert and dashboard resources.
func New(wc *apiclients.WrappedClient) *Client {
	return &Client{wc: wc}
}
