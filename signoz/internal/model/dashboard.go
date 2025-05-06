package model

import (
	"encoding/json"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

// Dashboard model.
type Dashboard struct {
	CollapsableRowsMigrated bool `json:"collapsableRowsMigrated"`
	// CreatedAt               string                   `json:"createdAt,omitempty"` // CreatedAt, CreatedBy, UpdatedAt, UpdatedBy, ID, UUID are not really present in Dashboard reponse. Comment out these things
	// CreatedBy               string                   `json:"createdBy,omitempty"`
	Description string `json:"description"`
	// ID              int32                    `json:"id,omitempty"` // if uuid,omitempty, then id should also be that
	Layout   []map[string]interface{} `json:"layout"`
	Name     string                   `json:"name"`
	PanelMap map[string]interface{}   `json:"panelMap,omitempty"`
	Source   string                   `json:"source"`
	Tags     []string                 `json:"tags"`
	Title    string                   `json:"title"`
	// UpdatedAt       string                   `json:"updatedAt,omitempty"`
	// UpdatedBy       string                   `json:"updatedBy,omitempty"`
	UploadedGrafana bool `json:"uploadedGrafana"`
	// UUID            string                   `json:"uuid,omitempty"`
	Variables map[string]interface{} `json:"variables"`
	// Version                 string                   `json:"version,omitempty"`
	Widgets []map[string]interface{} `json:"widgets"`
}

func (d Dashboard) PanelMapToTerraform() (types.String, error) {
	if d.PanelMap == nil {
		return types.StringNull(), nil
	}
	panelMap, err := structure.FlattenJsonToString(d.PanelMap)
	if err != nil {
		return types.StringNull(), err
	}

	return types.StringValue(panelMap), nil
}

func (d Dashboard) VariablesToTerraform() (types.String, error) {
	variables, err := structure.FlattenJsonToString(d.Variables)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(variables), nil
}

func (d Dashboard) TagsToTerraform() (types.List, diag.Diagnostics) {
	tags := utils.Map(d.Tags, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, tags)
}

func (d Dashboard) LayoutToTerraform() (types.String, error) {
	b, err := json.Marshal(d.Layout)
	if err != nil {
		return types.StringValue(""), err
	}
	return types.StringValue(string(b)), nil
}

func (d Dashboard) WidgetsToTerraform() (types.String, error) {
	b, err := json.Marshal(d.Widgets)
	if err != nil {
		return types.StringValue(""), err
	}
	return types.StringValue(string(b)), nil
}

func (d *Dashboard) SetVariables(tfVariables types.String) error {
	variables, err := structure.ExpandJsonFromString(tfVariables.ValueString())
	if err != nil {
		return err
	}
	d.Variables = variables
	return nil
}

func (d *Dashboard) SetPanelMap(tfPanelMap types.String) error {
	if tfPanelMap.ValueString() == "" {
		d.PanelMap = make(map[string]interface{})
		return nil
	}
	panelMap, err := structure.ExpandJsonFromString(tfPanelMap.ValueString())
	if err != nil {
		return err
	}
	d.PanelMap = panelMap
	return nil
}

func (d *Dashboard) SetTags(tfTags types.List) {
	tags := utils.Map(tfTags.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	d.Tags = tags
}

func (d *Dashboard) SetLayout(tfLayout types.String) error {
	var layout []map[string]interface{}
	err := json.Unmarshal([]byte(tfLayout.ValueString()), &layout)
	if err != nil {
		return err
	}
	d.Layout = layout
	return nil
}

func (d *Dashboard) SetWidgets(tfWidgets types.String) error {
	var widgets []map[string]interface{}
	err := json.Unmarshal([]byte(tfWidgets.ValueString()), &widgets)
	if err != nil {
		return err
	}
	d.Widgets = widgets
	return nil
}

func (d *Dashboard) SetSourceIfEmpty(hostURL string) {
	d.Source = utils.WithDefault(d.Source, hostURL+"/dashboard")
}
