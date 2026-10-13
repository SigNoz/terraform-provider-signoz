package legacyclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	dataRetentionPath   = "api/v1/settings/ttl"
	dataRetentionV2Path = "api/v2/settings/ttl"
)

// dataRetentionResponse maps the GET response envelope:
// {"status":"success","data":{...expected_* fields...}}.
type dataRetentionResponse struct {
	Status string              `json:"status"`
	Data   model.DataRetention `json:"data"`
}

// customRetentionResponse maps the GET /api/v2/settings/ttl envelope
// (logs-only custom retention). Days on the wire; converted to hours here.
type customRetentionResponse struct {
	Status string                               `json:"status"`
	Data   model.CustomRetentionTTLResponseData `json:"data"`
}

// GetDataRetention returns the current data retention settings: one v1 GET per
// signal type, with logs overridden by the v2 custom-retention values when the
// server answers v2. A v2 4xx (older server, no custom retention) keeps the
// v1 logs values.
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

		var resp dataRetentionResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("error parsing %s retention response: %w", signalType, err)
		}

		switch signalType {
		case "metrics":
			result.MetricsTTLDurationHrs = resp.Data.MetricsTTLDurationHrs
			result.MetricsMoveTTLDurationHrs = resp.Data.MetricsMoveTTLDurationHrs
		case "logs":
			result.LogsTTLDurationHrs = resp.Data.LogsTTLDurationHrs
			result.LogsMoveTTLDurationHrs = resp.Data.LogsMoveTTLDurationHrs
		case "traces":
			result.TracesTTLDurationHrs = resp.Data.TracesTTLDurationHrs
			result.TracesMoveTTLDurationHrs = resp.Data.TracesMoveTTLDurationHrs
		}
	}

	if err := c.overrideLogsWithCustomRetention(ctx, result); err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "GetDataRetention: data retention settings fetched", map[string]any{"data": result})

	return result, nil
}

// overrideLogsWithCustomRetention replaces the logs hours with the v2 custom
// retention values (days → hours).
func (c *Client) overrideLogsWithCustomRetention(ctx context.Context, result *model.DataRetention) error {
	reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionV2Path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		var apiErr *apiclients.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
			tflog.Debug(ctx, "GetDataRetention: v2 custom retention unavailable, keeping v1 logs values", map[string]any{
				"status_code": apiErr.StatusCode,
			})
			return nil
		}
		return err
	}

	var resp customRetentionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("error parsing logs custom retention response: %w", err)
	}

	if resp.Status == "success" {
		moveHrs := 0

		if resp.Data.ColdStorageTTLDays > 0 {
			moveHrs = int(resp.Data.ColdStorageTTLDays) * 24
		}

		result.LogsTTLDurationHrs = int(resp.Data.DefaultTTLDays) * 24
		result.LogsMoveTTLDurationHrs = moveHrs
	}

	return nil
}

// SetDataRetention sets data retention for a single signal type.
// logs use /api/v2/settings/ttl with a days-based JSON body; metrics/traces
// use /api/v1 query params (coldStorage + toColdDuration; the server only
// honors the move when a volume is set).
func (c *Client) SetDataRetention(ctx context.Context, payload *model.DataRetentionUpdate) error {
	var req *http.Request

	if payload.Type == "logs" {
		reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionV2Path)
		if err != nil {
			return err
		}

		v2Body := model.CustomRetentionTTLParams{
			Type:           "logs",
			DefaultTTLDays: payload.DurationHrs / 24,
			TTLConditions:  []model.CustomRetentionRule{},
		}

		if payload.ColdStorageVolume != "" {
			v2Body.ColdStorageVolume = payload.ColdStorageVolume

			if payload.ColdStorageMoveHrs > 0 {
				v2Body.ColdStorageDurationDays = payload.ColdStorageMoveHrs / 24
			}
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

		if payload.ColdStorageVolume != "" {
			q.Set("coldStorage", payload.ColdStorageVolume)
			q.Set("toColdDuration", fmt.Sprintf("%dh", payload.ColdStorageMoveHrs))
		}

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
