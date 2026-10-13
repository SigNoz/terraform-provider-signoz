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
	// downtimeSchedulePath - URL path for downtime schedule APIs.
	downtimeSchedulePath = "api/v1/downtime_schedules"
)

// downtimeScheduleResponse - Maps the response data of GetDowntimeSchedule and CreateDowntimeSchedule.
type downtimeScheduleResponse struct {
	Status    string                 `json:"status"`
	Error     string                 `json:"error"`
	ErrorType string                 `json:"errorType"`
	Data      model.DowntimeSchedule `json:"data"`
}

// GetDowntimeSchedule - Returns a specific downtime schedule.
func (c *Client) GetDowntimeSchedule(ctx context.Context, id string) (*model.DowntimeSchedule, error) {
	url, err := url.JoinPath(c.hostURL.String(), downtimeSchedulePath, id)
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

	var bodyObj downtimeScheduleResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetDowntimeSchedule: error while fetching downtime schedule", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.DowntimeSchedule{}, fmt.Errorf("error while fetching downtime schedule: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetDowntimeSchedule: downtime schedule fetched", map[string]any{"downtimeSchedule": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateDowntimeSchedule - Creates a new downtime schedule.
func (c *Client) CreateDowntimeSchedule(ctx context.Context, payload *model.DowntimeSchedule) (*model.DowntimeSchedule, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), downtimeSchedulePath)
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

	var bodyObj downtimeScheduleResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateDowntimeSchedule: error while creating downtime schedule", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating downtime schedule: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateDowntimeSchedule: downtime schedule created", map[string]any{"downtimeSchedule": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateDowntimeSchedule - Updates an existing downtime schedule.
func (c *Client) UpdateDowntimeSchedule(ctx context.Context, id string, payload *model.DowntimeSchedule) error {
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), downtimeSchedulePath, id)
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
		tflog.Error(ctx, "UpdateDowntimeSchedule: error while updating downtime schedule", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating downtime schedule: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateDowntimeSchedule: downtime schedule updated", map[string]any{"downtimeSchedule": bodyObj.Data})

	return nil
}

// DeleteDowntimeSchedule - Deletes an existing downtime schedule.
func (c *Client) DeleteDowntimeSchedule(ctx context.Context, id string) error {
	url, err := url.JoinPath(c.hostURL.String(), downtimeSchedulePath, id)
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
		tflog.Error(ctx, "DeleteDowntimeSchedule: error while deleting downtime schedule", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting downtime schedule: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteDowntimeSchedule: downtime schedule deleted", map[string]any{"id": id, "bodyData": bodyObj.Data})

	return nil
}
