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
	// dashboardV2Path - SigNoz V2 dashboards (Perses-style / plugin panels).
	dashboardV2Path = "api/v2/dashboards"
)

type dashboardV2Response struct {
	Status    string                `json:"status"`
	Error     string                `json:"error,omitempty"`
	ErrorType string                `json:"errorType,omitempty"`
	Data      model.DashboardV2Data `json:"data"`
}

// GetDashboardV2 - GET /api/v2/dashboards/{id}
func (c *Client) GetDashboardV2(ctx context.Context, id string) (*model.DashboardV2Data, error) {
	u, err := url.JoinPath(c.hostURL.String(), dashboardV2Path, id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp dashboardV2Response
	if err := json.Unmarshal(body, &resp); err != nil {
		// Some builds return the dashboard object at the top level.
		var raw model.DashboardV2Data
		if err2 := json.Unmarshal(body, &raw); err2 == nil && raw.ID != "" {
			return &raw, nil
		}
		return nil, err
	}
	if resp.Status != "" && resp.Status != "success" {
		return nil, fmt.Errorf("get dashboard v2: %s", resp.Error)
	}
	if resp.Data.ID == "" && resp.Error != "" {
		return nil, fmt.Errorf("get dashboard v2: %s", resp.Error)
	}
	// Normalize id if nested only under data map
	if resp.Data.ID == "" {
		var envelope map[string]interface{}
		if err := json.Unmarshal(body, &envelope); err == nil {
			if d, ok := envelope["data"].(map[string]interface{}); ok {
				if id, ok := d["id"].(string); ok {
					resp.Data.ID = id
				}
				if t, ok := d["title"].(string); ok {
					resp.Data.Title = t
				}
				if desc, ok := d["description"].(string); ok {
					resp.Data.Description = desc
				}
				if data, ok := d["data"].(map[string]interface{}); ok {
					resp.Data.Data = data
				}
			}
		}
	}
	tflog.Debug(ctx, "GetDashboardV2", map[string]any{"id": resp.Data.ID})
	return &resp.Data, nil
}

// CreateDashboardV2 - POST /api/v2/dashboards
func (c *Client) CreateDashboardV2(ctx context.Context, payload model.DashboardV2) (*model.DashboardV2Data, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	u, err := url.JoinPath(c.hostURL.String(), dashboardV2Path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp dashboardV2Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "" && resp.Status != "success" {
		return nil, fmt.Errorf("create dashboard v2: %s", resp.Error)
	}
	if resp.Data.ID == "" {
		// try alternate shapes
		var alt struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &alt)
		if alt.Data.ID != "" {
			resp.Data.ID = alt.Data.ID
		} else if alt.ID != "" {
			resp.Data.ID = alt.ID
		}
	}
	if resp.Data.ID == "" {
		return nil, fmt.Errorf("create dashboard v2: empty id in response: %s", string(body))
	}
	tflog.Debug(ctx, "CreateDashboardV2", map[string]any{"id": resp.Data.ID})
	return &resp.Data, nil
}

// UpdateDashboardV2 - PUT /api/v2/dashboards/{id}
func (c *Client) UpdateDashboardV2(ctx context.Context, id string, payload model.DashboardV2) error {
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u, err := url.JoinPath(c.hostURL.String(), dashboardV2Path, id)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, u, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	var resp signozResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		// some servers return 200 with empty body
		if len(strings.TrimSpace(string(body))) == 0 {
			return nil
		}
		return err
	}
	if resp.Status != "" && resp.Status != "success" {
		return fmt.Errorf("update dashboard v2: %s", resp.Error)
	}
	tflog.Debug(ctx, "UpdateDashboardV2", map[string]any{"id": id})
	return nil
}

// DeleteDashboardV2 - DELETE /api/v2/dashboards/{id}
func (c *Client) DeleteDashboardV2(ctx context.Context, id string) error {
	u, err := url.JoinPath(c.hostURL.String(), dashboardV2Path, id)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	tflog.Debug(ctx, "DeleteDashboardV2", map[string]any{"id": id})
	return nil
}
