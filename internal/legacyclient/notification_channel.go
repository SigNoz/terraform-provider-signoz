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
	// channelPath - URL path for notification channel APIs.
	channelPath = "api/v1/channels"
)

// channelListResponse - Maps the response data of ListChannels (array).
type channelListResponse struct {
	Status    string                      `json:"status"`
	Error     string                      `json:"error"`
	ErrorType string                      `json:"errorType"`
	Data      []model.NotificationChannel `json:"data"`
}

// channelResponse - Maps the response data of GetChannel and CreateChannel.
type channelResponse struct {
	Status    string                    `json:"status"`
	Error     string                    `json:"error"`
	ErrorType string                    `json:"errorType"`
	Data      model.NotificationChannel `json:"data"`
}

// GetChannel - Returns specific notification channel.
func (c *Client) GetChannel(ctx context.Context, channelID string) (*model.NotificationChannel, error) {
	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
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

	var bodyObj channelResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetChannel: error while fetching channel", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.NotificationChannel{}, fmt.Errorf("error while fetching channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetChannel: channel fetched", map[string]any{"channel": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateChannel - Creates a new notification channel.
func (c *Client) CreateChannel(ctx context.Context, channelPayload *model.NotificationChannel) (*model.NotificationChannel, error) {
	rb, err := json.Marshal(channelPayload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath)
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

	// v0.104.0 returns 204 No Content on success — empty body.
	if len(body) == 0 {
		tflog.Debug(ctx, "CreateChannel: empty response, fetching channel by name", map[string]any{"name": channelPayload.Name})
		return c.findChannelByName(ctx, channelPayload.Name)
	}

	var bodyObj channelResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateChannel: error while creating channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateChannel: channel created", map[string]any{"channel": bodyObj.Data})

	return &bodyObj.Data, nil
}

// findChannelByName lists all channels and returns the one matching the given name.
func (c *Client) findChannelByName(ctx context.Context, name string) (*model.NotificationChannel, error) {
	listURL, err := url.JoinPath(c.hostURL.String(), channelPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, listURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var listObj channelListResponse
	err = json.Unmarshal(body, &listObj)
	if err != nil {
		return nil, err
	}

	for i := range listObj.Data {
		if listObj.Data[i].Name == name {
			return &listObj.Data[i], nil
		}
	}

	return nil, fmt.Errorf("channel %q created but not found in list", name)
}

// UpdateChannel - Updates an existing notification channel.
func (c *Client) UpdateChannel(ctx context.Context, channelID string, channelPayload *model.NotificationChannel) error {
	rb, err := json.Marshal(channelPayload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
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

	// v0.104.0 returns 204 No Content on success.
	if len(body) == 0 {
		tflog.Debug(ctx, "UpdateChannel: channel updated (empty response)", map[string]any{"channelID": channelID})
		return nil
	}

	var bodyObj signozResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "UpdateChannel: error while updating channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateChannel: channel updated", map[string]any{"channelID": channelID})

	return nil
}

// DeleteChannel - Deletes an existing notification channel.
func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
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
		tflog.Error(ctx, "DeleteChannel: error while deleting channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteChannel: channel deleted", map[string]any{"channelID": channelID})

	return nil
}
