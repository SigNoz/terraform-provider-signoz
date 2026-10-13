package legacyclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	dataRetentionPath   = "api/v1/settings/ttl"
	dataRetentionV2Path = "api/v2/settings/ttl"
)

// dataRetentionResponse maps the GET response envelope.
type dataRetentionResponse struct {
	Status string              `json:"status"`
	Data   model.DataRetention `json:"data"`
}

// perSignalRetentionResponse maps the per-signal GET response (v0.104.0).
type perSignalRetentionResponse struct {
	Status                    string `json:"status"`
	MetricsTTLDurationHrs     int    `json:"metrics_ttl_duration_hrs"`
	MetricsMoveTTLDurationHrs int    `json:"metrics_move_ttl_duration_hrs"`
	LogsTTLDurationHrs        int    `json:"logs_ttl_duration_hrs"`
	LogsMoveTTLDurationHrs    int    `json:"logs_move_ttl_duration_hrs"`
	TracesTTLDurationHrs      int    `json:"traces_ttl_duration_hrs"`
	TracesMoveTTLDurationHrs  int    `json:"traces_move_ttl_duration_hrs"`
}

// GetDataRetention returns the current data retention settings.
// v0.104.0 requires one GET per signal type with ?type= query param.
func (c *Client) GetDataRetention(ctx context.Context) (*model.DataRetention, error) {
	result := &model.DataRetention{}

	for _, signalType := range []string{"metrics", "logs", "traces"} {
		reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionPath)
		if err != nil {
			return nil, err
		}
		parsed, err := url.Parse(reqURL)
		if err != nil {
			return nil, err
		}
		q := parsed.Query()
		q.Set("type", signalType)
		parsed.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(ctx, req)
		if err != nil {
			return nil, err
		}

		var resp perSignalRetentionResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("error parsing %s retention response: %w", signalType, err)
		}

		switch signalType {
		case "metrics":
			result.MetricsTTLDurationHrs = resp.MetricsTTLDurationHrs
			result.MetricsMoveTTLDurationHrs = resp.MetricsMoveTTLDurationHrs
		case "logs":
			result.LogsTTLDurationHrs = resp.LogsTTLDurationHrs
			result.LogsMoveTTLDurationHrs = resp.LogsMoveTTLDurationHrs
		case "traces":
			result.TracesTTLDurationHrs = resp.TracesTTLDurationHrs
			result.TracesMoveTTLDurationHrs = resp.TracesMoveTTLDurationHrs
		}
	}

	tflog.Debug(ctx, "GetDataRetention: data retention settings fetched", map[string]any{"data": result})

	return result, nil
}

// SetDataRetention sets data retention for a single signal type.
// v0.104.0: logs use /api/v2/settings/ttl (JSON body), metrics/traces use /api/v1 (query params).
func (c *Client) SetDataRetention(ctx context.Context, payload *model.DataRetentionUpdate) error {
	var req *http.Request

	if payload.Type == "logs" {
		reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionV2Path)
		if err != nil {
			return err
		}
		v2Body := map[string]interface{}{
			"type":         payload.Type,
			"duration_hrs": payload.DurationHrs,
		}
		if payload.ColdStorageDurationHrs > 0 {
			v2Body["cold_storage_duration_hrs"] = payload.ColdStorageDurationHrs
		}
		rb, err := json.Marshal(v2Body)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(rb)))
		if err != nil {
			return err
		}
	} else {
		reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionPath)
		if err != nil {
			return err
		}
		parsed, err := url.Parse(reqURL)
		if err != nil {
			return err
		}
		q := parsed.Query()
		q.Set("type", payload.Type)
		q.Set("duration", fmt.Sprintf("%dh", payload.DurationHrs))
		q.Set("cold_storage_duration", fmt.Sprintf("%dh", payload.ColdStorageDurationHrs))
		parsed.RawQuery = q.Encode()
		req, err = http.NewRequest(http.MethodPost, parsed.String(), nil)
		if err != nil {
			return err
		}
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	// v0.104.0 returns {"message":"..."} on success, not {"status":"success"}.
	if len(body) > 0 {
		var resp struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Error   string `json:"error"`
			Errors  []struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(body, &resp); err == nil {
			if len(resp.Errors) > 0 {
				return fmt.Errorf("error setting %s retention: %s", payload.Type, resp.Errors[0].Msg)
			}
			if resp.Error != "" {
				return fmt.Errorf("error setting %s retention: %s", payload.Type, resp.Error)
			}
		}
	}

	tflog.Debug(ctx, "SetDataRetention: data retention set", map[string]any{"type": payload.Type})

	return nil
}
