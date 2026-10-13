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
	// logPipelinePath - POST /api/v1/logs/pipelines (batch replace).
	logPipelinePath = "api/v1/logs/pipelines"
	// logPipelineLatestPath - GET /api/v1/logs/pipelines/{version}; the server
	// only registers versioned GET paths ("latest" resolves to the newest
	// pipeline version). A bare GET on logPipelinePath 404s.
	logPipelineLatestPath = "api/v1/logs/pipelines/latest"
)

// logPipelinesResponse - Maps the response data of log pipeline APIs.
type logPipelinesResponse struct {
	Status    string           `json:"status"`
	Error     string           `json:"error,omitempty"`
	ErrorType string           `json:"errorType,omitempty"`
	Data      logPipelinesData `json:"data"`
}

type logPipelinesData struct {
	Pipelines []model.LogPipeline `json:"pipelines"`
}

// GetLogPipelines - Returns all log pipelines for the latest version.
func (c *Client) GetLogPipelines(ctx context.Context) ([]model.LogPipeline, error) {
	reqURL, err := url.JoinPath(c.hostURL.String(), logPipelineLatestPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj logPipelinesResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetLogPipelines: error while fetching log pipelines", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
		})

		return nil, fmt.Errorf("error while fetching log pipelines: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetLogPipelines: log pipelines fetched", map[string]any{
		"count": len(bodyObj.Data.Pipelines),
	})

	return bodyObj.Data.Pipelines, nil
}

// GetLogPipeline - Returns a specific log pipeline by ID.
func (c *Client) GetLogPipeline(ctx context.Context, id string) (*model.LogPipeline, error) {
	pipelines, err := c.GetLogPipelines(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range pipelines {
		if p.ID == id {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("log pipeline with ID %s not found", id)
}

// SaveLogPipelines - POSTs the full pipeline array (batch-replace).
func (c *Client) SaveLogPipelines(ctx context.Context, pipelines []model.LogPipeline) ([]model.LogPipeline, error) {
	payload := model.LogPipelinesPayload{
		Pipelines: pipelines,
	}

	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	reqURL, err := url.JoinPath(c.hostURL.String(), logPipelinePath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj logPipelinesResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "SaveLogPipelines: error while saving log pipelines", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
		})
		return nil, fmt.Errorf("error while saving log pipelines: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "SaveLogPipelines: log pipelines saved", map[string]any{
		"count": len(bodyObj.Data.Pipelines),
	})

	return bodyObj.Data.Pipelines, nil
}

// CreateLogPipeline - Creates a new log pipeline by appending to the existing array and POSTing.
// Returns the newly created pipeline (matched by alias or name).
func (c *Client) CreateLogPipeline(ctx context.Context, pipeline *model.LogPipeline) (*model.LogPipeline, error) {
	existing, err := c.GetLogPipelines(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching existing pipelines for create: %w", err)
	}

	updated := append(existing, *pipeline)

	saved, err := c.SaveLogPipelines(ctx, updated)
	if err != nil {
		return nil, err
	}

	// Find the newly created pipeline by alias (preferred) or name.
	for _, p := range saved {
		if pipeline.Alias != "" && p.Alias == pipeline.Alias {
			return &p, nil
		}
		if p.Name == pipeline.Name {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("created log pipeline not found in response (alias=%s, name=%s)", pipeline.Alias, pipeline.Name)
}

// UpdateLogPipeline - Updates an existing log pipeline by replacing it in the array and POSTing.
func (c *Client) UpdateLogPipeline(ctx context.Context, id string, pipeline *model.LogPipeline) (*model.LogPipeline, error) {
	existing, err := c.GetLogPipelines(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching existing pipelines for update: %w", err)
	}

	found := false
	for i, p := range existing {
		if p.ID == id {
			pipeline.ID = id
			existing[i] = *pipeline
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("log pipeline with ID %s not found for update", id)
	}

	saved, err := c.SaveLogPipelines(ctx, existing)
	if err != nil {
		return nil, err
	}

	for _, p := range saved {
		if p.ID == id {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("updated log pipeline with ID %s not found in response", id)
}

// DeleteLogPipeline - Deletes a log pipeline by removing it from the array and POSTing.
func (c *Client) DeleteLogPipeline(ctx context.Context, id string) error {
	existing, err := c.GetLogPipelines(ctx)
	if err != nil {
		return fmt.Errorf("error fetching existing pipelines for delete: %w", err)
	}

	updated := make([]model.LogPipeline, 0, len(existing))
	found := false
	for _, p := range existing {
		if p.ID == id {
			found = true
			continue
		}
		updated = append(updated, p)
	}

	if !found {
		return fmt.Errorf("log pipeline with ID %s not found for delete", id)
	}

	_, err = c.SaveLogPipelines(ctx, updated)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, "DeleteLogPipeline: log pipeline deleted", map[string]any{"pipelineID": id})

	return nil
}
