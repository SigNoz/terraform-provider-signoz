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
	// routePolicyPath - URL path for route policy APIs.
	routePolicyPath = "api/v1/route_policies"
)

// routePolicyResponse - Maps the response data of GetRoutePolicy and CreateRoutePolicy.
type routePolicyResponse struct {
	Status    string            `json:"status"`
	Error     string            `json:"error"`
	ErrorType string            `json:"errorType"`
	Data      model.RoutePolicy `json:"data"`
}

// GetRoutePolicy - Returns specific route policy.
func (c *Client) GetRoutePolicy(ctx context.Context, id string) (*model.RoutePolicy, error) {
	url, err := url.JoinPath(c.hostURL.String(), routePolicyPath, id)
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

	var bodyObj routePolicyResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetRoutePolicy: error while fetching route policy", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.RoutePolicy{}, fmt.Errorf("error while fetching route policy: %s", bodyObj.Error)
	}

	bodyObj.Data.PopulateFromAPI()
	tflog.Debug(ctx, "GetRoutePolicy: route policy fetched", map[string]any{"routePolicy": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateRoutePolicy - Creates a new route policy.
func (c *Client) CreateRoutePolicy(ctx context.Context, payload *model.RoutePolicy) (*model.RoutePolicy, error) {
	payload.PrepareForAPI()
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), routePolicyPath)
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

	var bodyObj routePolicyResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateRoutePolicy: error while creating route policy", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating route policy: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateRoutePolicy: route policy created", map[string]any{"routePolicy": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateRoutePolicy - Updates an existing route policy.
func (c *Client) UpdateRoutePolicy(ctx context.Context, id string, payload *model.RoutePolicy) error {
	payload.PrepareForAPI()
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), routePolicyPath, id)
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
		tflog.Error(ctx, "UpdateRoutePolicy: error while updating route policy", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating route policy: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateRoutePolicy: route policy updated", map[string]any{"id": id, "bodyData": bodyObj.Data})

	return nil
}

// DeleteRoutePolicy - Deletes an existing route policy.
func (c *Client) DeleteRoutePolicy(ctx context.Context, id string) error {
	url, err := url.JoinPath(c.hostURL.String(), routePolicyPath, id)
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
		tflog.Error(ctx, "DeleteRoutePolicy: error while deleting route policy", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting route policy: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteRoutePolicy: route policy deleted", map[string]any{"id": id, "bodyData": bodyObj.Data})

	return nil
}
