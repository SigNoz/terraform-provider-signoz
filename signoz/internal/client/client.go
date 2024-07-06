package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DefaultHostURL - Default SigNoz URL.
const DefaultHostURL string = "http://localhost:3301"

// Client - SigNoz API client.
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient - Creates a new client.
func NewClient(host, token *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// Default SigNoz URL
		HostURL: DefaultHostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	if token != nil {
		c.Token = *token
	}

	return &c, nil
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	req.Header.Set("SIGNOZ-API-KEY", c.Token)

	tflog.Debug(ctx, "Making SigNoz API request", map[string]any{"method": req.Method, "url": req.URL.String()})

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
