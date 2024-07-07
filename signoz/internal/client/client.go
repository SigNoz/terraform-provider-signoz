package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// DefaultHostURL - Default SigNoz URL.
	DefaultHostURL string = "http://localhost:3301"
	// DefaultHTTPTimeout - Default HTTP timeout.
	DefaultHTTPTimeout time.Duration = 10 * time.Second
)

// Client - SigNoz API client.
type Client struct {
	agent      string
	hostURL    string
	token      string
	version    string
	httpClient *httpclient.Client
}

// NewClient - Creates a new client.
func NewClient(host, token string, httpTimeout time.Duration, httpRetryMax int, agent, version string) *Client {
	client := httpclient.NewClient(
		httpclient.WithHTTPClient(
			&http.Client{
				Timeout:   httpTimeout,
				Transport: http.DefaultTransport,
			},
		),
		httpclient.WithHTTPTimeout(httpTimeout),
		httpclient.WithRetrier(
			heimdall.NewRetrier(
				heimdall.NewConstantBackoff(
					5*time.Second,
					1*time.Second,
				),
			),
		),
		httpclient.WithRetryCount(httpRetryMax),
	)

	return &Client{
		agent:      agent,
		hostURL:    host,
		token:      token,
		version:    version,
		httpClient: client,
	}
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	req.Header.Set("SIGNOZ-API-KEY", c.token)

	tflog.Debug(ctx, "Making SigNoz API request", map[string]any{
		"method": req.Method,
		"url":    req.URL.String(),
		"body":   req.Body,
	})

	res, err := c.httpClient.Do(req)
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
