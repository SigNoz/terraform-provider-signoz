package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// dashboardPath - URL path for dashboard APIs.
	dashboardPath = "api/v1/dashboards"
)

// GetDashboard - Returns specific dashboard.
func (c *Client) GetDashboard(ctx context.Context, dashboardUUID string) (*dashboardData, error) {
	url, err := url.JoinPath(c.hostURL.String(), dashboardPath, dashboardUUID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj dashboardResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetDashboard: error while fetching dashboard", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})

		return &dashboardData{}, fmt.Errorf("error while fetching dashboard: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetDashboard: dashboard fetched", map[string]any{"dashboard": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateDashboard - Creates a new dashboard.
func (c *Client) CreateDashboard(ctx context.Context, dashboardPayload *model.Dashboard) (*dashboardData, error) {
	dashboardPayload.SetSourceIfEmpty(c.hostURL.String())
	rb, err := json.Marshal(dashboardPayload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), dashboardPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj dashboardResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateDashboard: error while creating dashboard", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating dashboard: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateDashboard: dashboard created", map[string]any{"dashboard": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateDashboard - Updates an existing dashboard.
func (c *Client) UpdateDashboard(ctx context.Context, dashboardUUID string, dashboardPayload *model.Dashboard) error {
	dashboardPayload.SetSourceIfEmpty(c.hostURL.String())
	rb, err := json.Marshal(dashboardPayload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), dashboardPath, dashboardUUID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	var bodyObj signozResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "UpdateDashboard: error while updating dashboard", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating dashboard: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateDashboard: dashboard updated", map[string]any{"dashboard": bodyObj.Data})

	return nil
}

// DeleteDashboard - Deletes an existing dashboard.
func (c *Client) DeleteDashboard(ctx context.Context, dashboardUUID string) error {
	url, err := url.JoinPath(c.hostURL.String(), dashboardPath, dashboardUUID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	var bodyObj signozResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "DeleteDashboard: error while deleting dashboard", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting dashboard: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteDashboard: dashboard deleted", map[string]any{"dashboardUUID": dashboardUUID, "bodyData": bodyObj.Data})

	return nil
}
