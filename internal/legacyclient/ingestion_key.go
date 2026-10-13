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
	// ingestionKeyPath - URL path for ingestion key APIs.
	ingestionKeyPath = "api/v2/gateway/ingestion_keys"
)

// ingestionKeyResponse - Maps the response data of GetIngestionKey and CreateIngestionKey.
type ingestionKeyResponse struct {
	Status    string             `json:"status"`
	Error     string             `json:"error"`
	ErrorType string             `json:"errorType"`
	Data      model.IngestionKey `json:"data"`
}

// GetIngestionKey - Returns a specific ingestion key.
func (c *Client) GetIngestionKey(ctx context.Context, keyID string) (*model.IngestionKey, error) {
	url, err := url.JoinPath(c.hostURL.String(), ingestionKeyPath, keyID)
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

	var bodyObj ingestionKeyResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetIngestionKey: error while fetching ingestion key", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.IngestionKey{}, fmt.Errorf("error while fetching ingestion key: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetIngestionKey: ingestion key fetched", map[string]any{"ingestionKey": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateIngestionKey - Creates a new ingestion key.
func (c *Client) CreateIngestionKey(ctx context.Context, payload *model.IngestionKey) (*model.IngestionKey, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), ingestionKeyPath)
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

	var bodyObj ingestionKeyResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateIngestionKey: error while creating ingestion key", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating ingestion key: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateIngestionKey: ingestion key created", map[string]any{"ingestionKey": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateIngestionKey - Updates an existing ingestion key using PATCH.
func (c *Client) UpdateIngestionKey(ctx context.Context, keyID string, payload *model.IngestionKey) error {
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), ingestionKeyPath, keyID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(string(rb)))
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
		tflog.Error(ctx, "UpdateIngestionKey: error while updating ingestion key", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating ingestion key: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateIngestionKey: ingestion key updated", map[string]any{"keyID": keyID})

	return nil
}

// DeleteIngestionKey - Deletes an existing ingestion key.
func (c *Client) DeleteIngestionKey(ctx context.Context, keyID string) error {
	url, err := url.JoinPath(c.hostURL.String(), ingestionKeyPath, keyID)
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
		tflog.Error(ctx, "DeleteIngestionKey: error while deleting ingestion key", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting ingestion key: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteIngestionKey: ingestion key deleted", map[string]any{"keyID": keyID})

	return nil
}

// SetIngestionKeyLimits - Sets rate limits for an ingestion key.
func (c *Client) SetIngestionKeyLimits(ctx context.Context, keyID string, limits *model.IngestionRateLimit) error {
	rb, err := json.Marshal(limits)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), ingestionKeyPath, keyID, "limits")
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rb)))
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
		tflog.Error(ctx, "SetIngestionKeyLimits: error while setting rate limits", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while setting ingestion key rate limits: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "SetIngestionKeyLimits: rate limits set", map[string]any{"keyID": keyID})

	return nil
}
