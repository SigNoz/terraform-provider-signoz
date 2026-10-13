package legacyclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Gateway endpoints (pkg/apiserver/signozapiserver/gateway.go in the SigNoz
// server). The deprecated POST /ingestion_keys/{keyId}/limits and the
// deprecated PATCH/DELETE under /ingestion_keys/limits/{limitId} are
// deliberately unused.
const (
	ingestionKeysPath   = "api/v2/gateway/ingestion_keys"
	ingestionLimitsPath = "api/v2/gateway/ingestion_limits"
)

// gatewayResponse is the {"status","data"} envelope render.Success emits.
type gatewayResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

// CreateIngestionKey creates a key; the response carries the one-time value.
func (c *Client) CreateIngestionKey(ctx context.Context, key model.PostableIngestionKey) (*model.CreatedIngestionKey, error) {
	var created model.CreatedIngestionKey

	req, err := c.gatewayRequest(ctx, http.MethodPost, ingestionKeysPath, key)
	if err != nil {
		return nil, err
	}

	if err := c.doGateway(ctx, req, &created); err != nil {
		return nil, fmt.Errorf("error while creating ingestion key: %w", err)
	}

	tflog.Debug(ctx, "CreateIngestionKey: ingestion key created", map[string]any{"id": created.ID})

	return &created, nil
}

// GetIngestionKey returns a single key, limits included.
func (c *Client) GetIngestionKey(ctx context.Context, keyID string) (*model.IngestionKey, error) {
	var key model.IngestionKey

	req, err := c.gatewayRequest(ctx, http.MethodGet, ingestionKeysPath+"/"+url.PathEscape(keyID), nil)
	if err != nil {
		return nil, err
	}

	if err := c.doGateway(ctx, req, &key); err != nil {
		return nil, fmt.Errorf("error while fetching ingestion key: %w", err)
	}

	tflog.Debug(ctx, "GetIngestionKey: ingestion key fetched", map[string]any{"id": key.ID})

	return &key, nil
}

// UpdateIngestionKey PATCHes name, tags and expires_at.
func (c *Client) UpdateIngestionKey(ctx context.Context, keyID string, key model.PostableIngestionKey) error {
	req, err := c.gatewayRequest(ctx, http.MethodPatch, ingestionKeysPath+"/"+url.PathEscape(keyID), key)
	if err != nil {
		return err
	}

	if err := c.doGateway(ctx, req, nil); err != nil {
		return fmt.Errorf("error while updating ingestion key: %w", err)
	}

	tflog.Debug(ctx, "UpdateIngestionKey: ingestion key updated", map[string]any{"keyID": keyID})

	return nil
}

// DeleteIngestionKey deletes a key; the server cascades its limits.
func (c *Client) DeleteIngestionKey(ctx context.Context, keyID string) error {
	req, err := c.gatewayRequest(ctx, http.MethodDelete, ingestionKeysPath+"/"+url.PathEscape(keyID), nil)
	if err != nil {
		return err
	}

	if err := c.doGateway(ctx, req, nil); err != nil {
		return fmt.Errorf("error while deleting ingestion key: %w", err)
	}

	tflog.Debug(ctx, "DeleteIngestionKey: ingestion key deleted", map[string]any{"keyID": keyID})

	return nil
}

// CreateIngestionKeyLimit attaches a limit to a key via the non-deprecated
// POST /api/v2/gateway/ingestion_limits (keyId travels in the body).
func (c *Client) CreateIngestionKeyLimit(ctx context.Context, limit model.PostableIngestionKeyLimit) (*model.CreatedIngestionKeyLimit, error) {
	var created model.CreatedIngestionKeyLimit

	req, err := c.gatewayRequest(ctx, http.MethodPost, ingestionLimitsPath, limit)
	if err != nil {
		return nil, err
	}

	if err := c.doGateway(ctx, req, &created); err != nil {
		return nil, fmt.Errorf("error while creating ingestion key limit: %w", err)
	}

	tflog.Debug(ctx, "CreateIngestionKeyLimit: limit created", map[string]any{"id": created.ID})

	return &created, nil
}

// GetIngestionKeyLimit returns a single limit.
func (c *Client) GetIngestionKeyLimit(ctx context.Context, limitID string) (*model.Limit, error) {
	var limit model.Limit

	req, err := c.gatewayRequest(ctx, http.MethodGet, ingestionLimitsPath+"/"+url.PathEscape(limitID), nil)
	if err != nil {
		return nil, err
	}

	if err := c.doGateway(ctx, req, &limit); err != nil {
		return nil, fmt.Errorf("error while fetching ingestion key limit: %w", err)
	}

	tflog.Debug(ctx, "GetIngestionKeyLimit: limit fetched", map[string]any{"id": limit.ID})

	return &limit, nil
}

// GetIngestionKeyLimits returns all limits attached to a key; the data payload
// is a bare JSON array.
func (c *Client) GetIngestionKeyLimits(ctx context.Context, keyID string) ([]model.Limit, error) {
	var limits []model.Limit

	req, err := c.gatewayRequest(ctx, http.MethodGet, ingestionKeysPath+"/"+url.PathEscape(keyID)+"/limits", nil)
	if err != nil {
		return nil, err
	}

	if err := c.doGateway(ctx, req, &limits); err != nil {
		return nil, fmt.Errorf("error while fetching ingestion key limits: %w", err)
	}

	tflog.Debug(ctx, "GetIngestionKeyLimits: limits fetched", map[string]any{"keyID": keyID, "count": len(limits)})

	return limits, nil
}

// UpdateIngestionKeyLimit PATCHes config and tags.
func (c *Client) UpdateIngestionKeyLimit(ctx context.Context, limitID string, limit model.UpdatableIngestionKeyLimit) error {
	req, err := c.gatewayRequest(ctx, http.MethodPatch, ingestionLimitsPath+"/"+url.PathEscape(limitID), limit)
	if err != nil {
		return err
	}

	if err := c.doGateway(ctx, req, nil); err != nil {
		return fmt.Errorf("error while updating ingestion key limit: %w", err)
	}

	tflog.Debug(ctx, "UpdateIngestionKeyLimit: limit updated", map[string]any{"limitID": limitID})

	return nil
}

// DeleteIngestionKeyLimit deletes a limit.
func (c *Client) DeleteIngestionKeyLimit(ctx context.Context, limitID string) error {
	req, err := c.gatewayRequest(ctx, http.MethodDelete, ingestionLimitsPath+"/"+url.PathEscape(limitID), nil)
	if err != nil {
		return err
	}

	if err := c.doGateway(ctx, req, nil); err != nil {
		return fmt.Errorf("error while deleting ingestion key limit: %w", err)
	}

	tflog.Debug(ctx, "DeleteIngestionKeyLimit: limit deleted", map[string]any{"limitID": limitID})

	return nil
}

// gatewayRequest builds a JSON request against the gateway API. A nil payload
// produces a bodyless request.
func (c *Client) gatewayRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	target, err := url.JoinPath(c.hostURL.String(), path)
	if err != nil {
		return nil, err
	}

	var body *bytes.Reader

	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(encoded)
	}

	return http.NewRequestWithContext(ctx, method, target, body)
}

// doGateway executes the request and decodes the success envelope into out
// (nil out skips decoding — the 204 paths still carry a {"status"} body).
// Non-2xx responses surface as *apiclients.APIError so callers can branch on
// the status code.
func (c *Client) doGateway(ctx context.Context, req *http.Request, out any) error {
	raw, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	var envelope gatewayResponse[json.RawMessage]

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}

	if envelope.Status != "success" {
		return fmt.Errorf("unexpected response status %q", envelope.Status)
	}

	if len(envelope.Data) == 0 || bytes.Equal(envelope.Data, []byte("null")) {
		return nil
	}

	return json.Unmarshal(envelope.Data, out)
}
