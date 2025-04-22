package model

import (
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
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Layout      []string               `json:"layout"`
	Widgets     []string               `json:"widgets"`
	Variables   map[string]interface{} `json:"variables"`
	CreateAt    string                 `json:"createAt,omitempty"`
	CreateBy    string                 `json:"createBy,omitempty"`
	UpdateAt    string                 `json:"updateAt,omitempty"`
	UpdateBy    string                 `json:"updateBy,omitempty"`
	UUID        string                 `json:"uuid,omitempty"`
}

func (d *Dashboard) SetVariables(tfVariables types.String) error {
	variables, err := structure.ExpandJsonFromString(tfVariables.ValueString())
	if err != nil {
		return err
	}

	d.Variables = variables
	return nil
}

func (d *Dashboard) SetTags(tfTags types.List) {
	tags := utils.Map(tfTags.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	d.Tags = tags
}

func (d *Dashboard) SetLayout(tfLayout types.List) {
	layout := utils.Map(tfLayout.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	d.Layout = layout
}

func (d *Dashboard) SetWidgets(tfWidgets types.List) {
	widgets := utils.Map(tfWidgets.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	d.Widgets = widgets
}
