package model

import (
	"encoding/json"
	"fmt"

	structure "github.com/SigNoz/terraform-provider-signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LogPipeline model matches server contract: pkg/types/pipelinetypes/pipeline.go
type LogPipeline struct {
	ID          string                   `json:"id"`
	Alias       string                   `json:"alias,omitempty"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Enabled     bool                     `json:"enabled"`
	OrderID     int                      `json:"orderId"`
	Filter      map[string]interface{}   `json:"filter"`
	Config      []map[string]interface{} `json:"config"`
	CreatedBy   string                   `json:"createdBy,omitempty"`
	CreatedAt   string                   `json:"createdAt,omitempty"`
	UpdatedBy   string                   `json:"updatedBy,omitempty"`
	UpdatedAt   string                   `json:"updatedAt,omitempty"`
}

// LogPipelinesPayload is the request/response body for the batch pipeline API.
type LogPipelinesPayload struct {
	Pipelines []LogPipeline `json:"pipelines"`
}

// FilterToTerraform converts the filter map to a normalized JSON string for Terraform state.
func (p LogPipeline) FilterToTerraform() (types.String, error) {
	filter, err := structure.FlattenJsonToString(p.Filter)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(filter), nil
}

// ConfigToTerraform converts the config slice to a JSON array string for Terraform state.
func (p LogPipeline) ConfigToTerraform() (types.String, error) {
	configBytes, err := json.Marshal(p.Config)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(string(configBytes)), nil
}

// SetFilter parses a JSON string from Terraform and sets the filter field.
func (p *LogPipeline) SetFilter(tfFilter types.String) error {
	filter, err := structure.ExpandJsonFromString(tfFilter.ValueString())
	if err != nil {
		return err
	}

	p.Filter = filter
	return nil
}

// SetConfig parses a JSON array string from Terraform and sets the config
// field, filling in the operator fields the server requires and the user may
// omit: an id (generated "op-<position>" when absent) and a non-zero orderId
// (sequence position). Routing rules (output on every operator but the last,
// unique ids) stay the user's responsibility — the server rejects the rest.
func (p *LogPipeline) SetConfig(tfConfig types.String) error {
	config, err := expandPipelineConfig(tfConfig.ValueString())
	if err != nil {
		return err
	}

	p.Config = config
	return nil
}

// expandPipelineConfig parses the TF config JSON and normalizes operators.
func expandPipelineConfig(raw string) ([]map[string]interface{}, error) {
	var config []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return nil, err
	}

	for i, op := range config {
		if op == nil {
			continue
		}

		if id, ok := op["id"].(string); !ok || id == "" {
			op["id"] = fmt.Sprintf("op-%d", i+1)
		}

		if order, ok := op["orderId"].(float64); !ok || order == 0 {
			op["orderId"] = float64(i + 1)
		}
	}

	return config, nil
}

// NormalizePipelineConfigString applies the same operator normalization as
// SetConfig to a config JSON string and returns it compactly marshaled, so
// plan-time comparison can treat user input and server echo as equal.
func NormalizePipelineConfigString(raw string) (string, error) {
	config, err := expandPipelineConfig(raw)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
