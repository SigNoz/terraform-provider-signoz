package model

import (
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

const ()

var ()

// Dashboard model.
type Dashboard struct {
	Condition map[string]interface{} `json:"condition"`
	Dashboard string                 `json:"dashboard"`
	Source    string                 `json:"source"`
	UUID      string                 `json:"uuid"`
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
