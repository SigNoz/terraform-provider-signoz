package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// alertResponse - Maps the response data.
type alertResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	ErrorType string      `json:"errorType,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// GetAlert - Returns specific alert.
func (c *Client) GetAlert(ctx context.Context, alertID string) (*model.Alert, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/rules/%s", c.HostURL, alertID), nil)
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
		return &model.Alert{}, fmt.Errorf("error: %s, type: %s, body: %s", bodyObj.Error, bodyObj.ErrorType, string(body))
	}

	alertByteArr, err := json.Marshal(bodyObj.Data)
	if err != nil {
		return &model.Alert{}, err
	}

	tflog.Debug(ctx, "GetAlert", map[string]any{"alert": string(alertByteArr)})

	var alert *model.Alert
	err = json.Unmarshal(alertByteArr, &alert)
	if err != nil {
		return &model.Alert{}, err
	}

	return alert, nil
}

// CreateAlert - Creates a new alert.
func (c *Client) CreateAlert(ctx context.Context, alertPayload *model.Alert) (*model.Alert, error) {
	rb, err := json.Marshal(alertPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/rules", c.HostURL), strings.NewReader(string(rb)))
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
		return nil, fmt.Errorf("error: %s, type: %s, body: %s", bodyObj.Error, bodyObj.ErrorType, string(body))
	}

	alertByteArr, err := json.Marshal(bodyObj.Data)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "Created alert", map[string]any{"alert": string(alertByteArr)})

	var alert *model.Alert
	err = json.Unmarshal(alertByteArr, &alert)
	if err != nil {
		return nil, err
	}

	return alert, nil
}

// UpdateAlert - Updates an existing alert.
func (c *Client) UpdateAlert(ctx context.Context, alertID string, alert *model.Alert) (*model.Alert, error) {
	rb, err := json.Marshal(alert)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/rules/%s", c.HostURL, alertID), strings.NewReader(string(rb)))
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
		return nil, fmt.Errorf("error: %s, type: %s, body: %s", bodyObj.Error, bodyObj.ErrorType, string(body))
	}

	alertByteArr, err := json.Marshal(bodyObj.Data)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("UpdateAlert: alertID: %s, responseData: %s", alertID, string(alertByteArr)))

	var alertObj *model.Alert
	err = json.Unmarshal(alertByteArr, &alertObj)
	if err != nil {
		return nil, err
	}

	return alertObj, nil
}

// DeleteAlert - Deletes an existing alert.
func (c *Client) DeleteAlert(ctx context.Context, alertID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/rules/%s", c.HostURL, alertID), nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	var bodyObj alertResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		return fmt.Errorf("error: %s, type: %s, body: %s", bodyObj.Error, bodyObj.ErrorType, string(body))
	}

	responseData, ok := bodyObj.Data.(string)
	if !ok {
		return fmt.Errorf("error: invalid data type: %T", bodyObj.Data)
	}

	tflog.Debug(ctx, fmt.Sprintf("DeleteAlert: alertID: %s, responseData: %s", alertID, responseData))

	return nil
}
