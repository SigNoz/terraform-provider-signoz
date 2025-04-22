package model

import (
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

const ()

var ()

// Dashboard model.
type Dashboard struct {
	Condition map[string]interface{} `json:"condition"`
	Dashboard string                 `json:"dashboard"`
	Labels    map[string]string      `json:"labels"`
	Source    string                 `json:"source"`
	UUID      string                 `json:"uuid"`
	CreateAt  string                 `json:"createAt,omitempty"`
	CreateBy  string                 `json:"createBy,omitempty"`
	UpdateAt  string                 `json:"updateAt,omitempty"`
	UpdateBy  string                 `json:"updateBy,omitempty"`
}

func (d *Dashboard) SetCondition(tfCondition types.String) error {
	condition, err := structure.ExpandJsonFromString(tfCondition.ValueString())
	if err != nil {
		return err
	}

	d.Condition = condition
	return nil
}

func (d *Dashboard) SetSourceIfEmpty(hostURL string) {
	d.Source = utils.WithDefault(d.Source, hostURL+"/dashboards")
}

func (d *Dashboard) ConditionToTerraform() (types.String, error) {
	condition, err := structure.FlattenJsonToString(d.Condition)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(condition), nil
}

func (d *Dashboard) LabelsToTerraform() (types.Map, diag.Diagnostics) {
	elements := map[string]tfattr.Value{}
	terraformLabels := strings.Split(AlertTerraformLabel, ":")
	for key, value := range d.Labels {
		if key == attr.Severity || key == terraformLabels[0] {
			continue
		}
		elements[key] = types.StringValue(value)
	}
	return types.MapValue(types.StringType, elements)
}
