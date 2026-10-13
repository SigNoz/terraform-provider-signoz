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

const dashboardV2Path = "api/v2/dashboards"

// dashboardV2Envelope is the {status,data} wrapper every /api/v2/dashboards
// response uses.
type dashboardV2Envelope struct {
	Status string                    `json:"status"`
	Data   model.GettableDashboardV2 `json:"data"`
}

func decodeDashboardV2Envelope(body []byte, op string) (*model.GettableDashboardV2, error) {
	var env dashboardV2Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("%s: decode response: %w", op, err)
	}
	if env.Status != "success" {
		return nil, fmt.Errorf("%s: unexpected status %q in response", op, env.Status)
	}
	if env.Data.ID == "" {
		return nil, fmt.Errorf("%s: response carries no dashboard id", op)
	}
	return &env.Data, nil
}

// GetDashboardV2 - GET /api/v2/dashboards/{id}
func (c *Client) GetDashboardV2(ctx context.Context, id string) (*model.GettableDashboardV2, error) {
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

	got, err := decodeDashboardV2Envelope(body, "get dashboard v2")
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "GetDashboardV2", map[string]any{"id": got.ID})
	return got, nil
}

// CreateDashboardV2 - POST /api/v2/dashboards. The server answers 201 with the
// created dashboard (generated name included).
func (c *Client) CreateDashboardV2(ctx context.Context, payload model.PostableDashboardV2) (*model.GettableDashboardV2, error) {
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

	created, err := decodeDashboardV2Envelope(body, "create dashboard v2")
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "CreateDashboardV2", map[string]any{"id": created.ID, "name": created.Name})
	return created, nil
}

// UpdateDashboardV2 - PUT /api/v2/dashboards/{id}
func (c *Client) UpdateDashboardV2(ctx context.Context, id string, payload model.UpdatableDashboardV2) (*model.GettableDashboardV2, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	u, err := url.JoinPath(c.hostURL.String(), dashboardV2Path, id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, u, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	updated, err := decodeDashboardV2Envelope(body, "update dashboard v2")
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "UpdateDashboardV2", map[string]any{"id": id})
	return updated, nil
}

// DeleteDashboardV2 - DELETE /api/v2/dashboards/{id} (204, no body)
func (c *Client) DeleteDashboardV2(ctx context.Context, id string) error {
	u, err := url.JoinPath(c.hostURL.String(), dashboardV2Path, id)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	if _, err := c.doRequest(ctx, req); err != nil {
		return err
	}
	tflog.Debug(ctx, "DeleteDashboardV2", map[string]any{"id": id})
	return nil
}
