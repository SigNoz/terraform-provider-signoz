package model

import (
	"encoding/json"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

const ()

var ()

// Dashboard model.
type Dashboard struct {
	CollapsableRowsMigrated bool                     `json:"collapsableRowsMigrated"`
	Description             string                   `json:"description"`
	Layout                  []map[string]interface{} `json:"layout"`
	Name                    string                   `json:"name"`
	PanelMap                map[string]interface{}   `json:"panelMap"`
	Tags                    []string                 `json:"tags"`
	Title                   string                   `json:"title"`
	UploadedGrafana         bool                     `json:"uploadedGrafana"`
	Variables               map[string]interface{}   `json:"variables"`
	Version                 string                   `json:"version"`
	Widgets                 []map[string]interface{} `json:"widgets"`
	CreatedAt               string                   `json:"createdAt,omitempty"`
	CreatedBy               string                   `json:"createdBy,omitempty"`
	UpdatedAt               string                   `json:"updatedAt,omitempty"`
	UpdatedBy               string                   `json:"updatedBy,omitempty"`
	UUID                    string                   `json:"uuid,omitempty"`
	ID                      int32                    `json:"id"`

	// IsLocked  bool                   `json:"isLocked,omitempty"`
	// Data      map[string]interface{} `json:"data"`
	// data and IsLocked
	// Title       string                 `json:"title"`
	// Description string                 `json:"description"`
	// Tags        []string               `json:"tags"`
	// Layout      []string               `json:"layout"`
	// Widgets     []string               `json:"widgets"`
	// Variables   map[string]interface{} `json:"variables"`
	Source string `json:"source"`
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

// func (d *Dashboard) SetTags(tfTags types.List) {
// 	tags := utils.Map(tfTags.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Tags = tags
// }

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

// func (d *Dashboard) SetWidgets(tfWidgets types.List) {
// 	widgets := utils.Map(tfWidgets.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Widgets = widgets
// }

func (d *Dashboard) SetSourceIfEmpty(hostURL string) {
	d.Source = utils.WithDefault(d.Source, hostURL+"/dashboard")
}

// func (d *Dashboard) SetPanelMap(tfPanelMap types.List) {
// 	widgets := utils.Map(tfPanelMap.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Widgets = widgets
// }

// func (d *Dashboard) SetTags(tfTags types.List) {
// 	widgets := utils.Map(tfTags.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Widgets = widgets
// }

// func (d *Dashboard) SetVariables(tfVariables types.String) {
// 	widgets := utils.Map(tfVariables.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Widgets = widgets
// }

// func (d *Dashboard) SetLayout(tfLayout types.List) {
// 	widgets := utils.Map(tfLayout.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Widgets = widgets
// }

// func (d *Dashboard) SetWidgets(tfWidgets types.List) {
// 	widgets := utils.Map(tfWidgets.Elements(), func(value tfattr.Value) string {
// 		return strings.Trim(value.String(), "\"")
// 	})
// 	d.Widgets = widgets
// }
