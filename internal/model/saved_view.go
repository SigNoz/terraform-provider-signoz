package model

import (
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	structure "github.com/SigNoz/terraform-provider-signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SavedView model.
type SavedView struct {
	UUID         string                 `json:"uuid"`
	Name         string                 `json:"name"`
	Category     string                 `json:"category,omitempty"`
	CreatedBy    string                 `json:"created_by,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
	UpdatedAt    string                 `json:"updated_at,omitempty"`
	SourcePage   string                 `json:"source_page"`
	Tags         []string               `json:"tags,omitempty"`
	ExplorerData map[string]interface{} `json:"explorer_data"`
	ExtraData    string                 `json:"extra_data,omitempty"`
}

func (v SavedView) ExplorerDataToTerraform() (types.String, error) {
	explorerData, err := structure.FlattenJsonToString(v.ExplorerData)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(explorerData), nil
}

func (v SavedView) TagsToTerraform() (types.List, diag.Diagnostics) {
	tags := utils.Map(v.Tags, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, tags)
}

func (v *SavedView) SetExplorerData(tfExplorerData types.String) error {
	explorerData, err := structure.ExpandJsonFromString(tfExplorerData.ValueString())
	if err != nil {
		return err
	}

	v.ExplorerData = explorerData
	return nil
}

func (v *SavedView) SetTags(tfTags types.List) {
	tags := utils.Map(tfTags.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	v.Tags = tags
}
