package model

import (
	"encoding/json"

	structure "github.com/SigNoz/terraform-provider-signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LogPipeline model.
type LogPipeline struct {
	ID          string                   `json:"id"`
	Alias       string                   `json:"alias,omitempty"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Enabled     bool                     `json:"enabled"`
	OrderID     int                      `json:"order_id"`
	Filter      map[string]interface{}   `json:"filter"`
	Config      []map[string]interface{} `json:"config"`
	CreatedBy   string                   `json:"created_by,omitempty"`
	CreatedAt   string                   `json:"created_at,omitempty"`
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

// SetConfig parses a JSON array string from Terraform and sets the config field.
func (p *LogPipeline) SetConfig(tfConfig types.String) error {
	var config []map[string]interface{}
	err := json.Unmarshal([]byte(tfConfig.ValueString()), &config)
	if err != nil {
		return err
	}

	p.Config = config
	return nil
}
