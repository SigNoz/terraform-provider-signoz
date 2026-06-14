package apiclients

import (
	"net/http"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ HttpRequestDoer = (*authDoer)(nil)

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
