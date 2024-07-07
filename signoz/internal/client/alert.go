package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// GetAlert - Returns specific alert.
func (c *Client) GetAlert(ctx context.Context, alertID string) (*model.Alert, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/rules/%s", c.hostURL, alertID), nil)
	if err != nil {
		return &model.Alert{}, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return &model.Alert{}, err
	}

	var bodyObj alertResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return &model.Alert{}, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetAlert: error while fetching alert", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.Alert{}, fmt.Errorf("error while fetching alert: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetAlert: alert fetched", map[string]any{"alert": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateAlert - Creates a new alert.
func (c *Client) CreateAlert(ctx context.Context, alertPayload *model.Alert) (*model.Alert, error) {
	alertPayload.Source = utils.WithDefault(alertPayload.Source, c.hostURL+"/alerts")
	rb, err := json.Marshal(alertPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/rules", c.hostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj alertResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateAlert: error while creating alert", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating alert: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateAlert: alert created", map[string]any{"alert": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateAlert - Updates an existing alert.
func (c *Client) UpdateAlert(ctx context.Context, alertID string, alertPayload *model.Alert) error {
	alertPayload.Source = utils.WithDefault(alertPayload.Source, c.hostURL+"/alerts")
	rb, err := json.Marshal(alertPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/rules/%s", c.hostURL, alertID), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

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
		tflog.Error(ctx, "UpdateAlert: error while updating alert", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating alert: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateAlert: alert updated", map[string]any{"alert": bodyObj.Data})

	return nil
}

// DeleteAlert - Deletes an existing alert.
func (c *Client) DeleteAlert(ctx context.Context, alertID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/rules/%s", c.hostURL, alertID), nil)
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
		tflog.Error(ctx, "DeleteAlert: error while deleting alert", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting alert: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteAlert: alert deleted", map[string]any{"alertID": alertID, "bodyData": bodyObj.Data})

	return nil
}
