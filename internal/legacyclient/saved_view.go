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
	// savedViewPath - URL path for saved view APIs.
	savedViewPath = "api/v1/explorer/views"
)

// savedViewResponse - Maps the response data of GetSavedView and CreateSavedView.
type savedViewResponse struct {
	Status    string          `json:"status"`
	Error     string          `json:"error"`
	ErrorType string          `json:"errorType"`
	Data      model.SavedView `json:"data"`
}

// GetSavedView - Returns specific saved view.
func (c *Client) GetSavedView(ctx context.Context, viewID string) (*model.SavedView, error) {
	url, err := url.JoinPath(c.hostURL.String(), savedViewPath, viewID)
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

	var bodyObj savedViewResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetSavedView: error while fetching saved view", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.SavedView{}, fmt.Errorf("error while fetching saved view: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetSavedView: saved view fetched", map[string]any{"savedView": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateSavedView - Creates a new saved view.
func (c *Client) CreateSavedView(ctx context.Context, payload *model.SavedView) (*model.SavedView, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), savedViewPath)
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

	var bodyObj savedViewResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateSavedView: error while creating saved view", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating saved view: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateSavedView: saved view created", map[string]any{"savedView": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateSavedView - Updates an existing saved view.
func (c *Client) UpdateSavedView(ctx context.Context, viewID string, payload *model.SavedView) error {
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), savedViewPath, viewID)
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
		tflog.Error(ctx, "UpdateSavedView: error while updating saved view", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating saved view: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateSavedView: saved view updated", map[string]any{"savedView": bodyObj.Data})

	return nil
}

// DeleteSavedView - Deletes an existing saved view.
func (c *Client) DeleteSavedView(ctx context.Context, viewID string) error {
	url, err := url.JoinPath(c.hostURL.String(), savedViewPath, viewID)
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
		tflog.Error(ctx, "DeleteSavedView: error while deleting saved view", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting saved view: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteSavedView: saved view deleted", map[string]any{"viewID": viewID, "bodyData": bodyObj.Data})

	return nil
}
