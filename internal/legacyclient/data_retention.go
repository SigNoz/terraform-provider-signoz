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

		var resp model.DataRetention
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

	if err := c.overrideLogsWithCustomRetention(ctx, result); err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "GetDataRetention: data retention settings fetched", map[string]any{"data": result})

	return result, nil
}

// getCustomRetentionTTL GETs /api/v2/settings/ttl. A nil result with no error
// means the endpoint is unavailable (4xx — older server without custom
// retention); callers fall back to the v1 values.
func (c *Client) getCustomRetentionTTL(ctx context.Context) (*model.CustomRetentionTTLResponseData, error) {
	reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionV2Path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		var apiErr *apiclients.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
			tflog.Debug(ctx, "v2 custom retention unavailable", map[string]any{
				"status_code": apiErr.StatusCode,
			})
			return nil, nil
		}
		return nil, err
	}

	var resp model.CustomRetentionTTLResponseData
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("error parsing logs custom retention response: %w", err)
	}

	return &resp, nil
}

// overrideLogsWithCustomRetention replaces the logs hours with the v2 custom
// retention values (days → hours).
func (c *Client) overrideLogsWithCustomRetention(ctx context.Context, result *model.DataRetention) error {
	resp, err := c.getCustomRetentionTTL(ctx)
	if err != nil || resp == nil {
		return err
	}

	moveHrs := 0

	if resp.ColdStorageTTLDays > 0 {
		moveHrs = int(resp.ColdStorageTTLDays) * 24
	}

	result.LogsTTLDurationHrs = int(resp.DefaultTTLDays) * 24
	result.LogsMoveTTLDurationHrs = moveHrs

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

		// The server's SetTTLV2 replaces the whole rule set: an empty or
		// omitted ttlConditions wipes every existing per-condition rule
		// (validateTTLConditions accepts an empty list). The provider does not
		// manage rules, so carry the live set back into the body.
		live, err := c.getCustomRetentionTTL(ctx)
		if err != nil {
			return err
		}

		conditions := []model.CustomRetentionRule{}
		if live != nil && len(live.TTLConditions) > 0 {
			conditions = live.TTLConditions
		}

		v2Body := model.CustomRetentionTTLParams{
			Type:           "logs",
			DefaultTTLDays: payload.DurationHrs / 24,
			TTLConditions:  conditions,
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
