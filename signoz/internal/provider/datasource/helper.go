package datasource

import (
	"fmt"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func addErr(diagnostics *diag.Diagnostics, err error, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operationRead, resource),
		err.Error(),
	)
}

func fetchLabels(labels map[string]string) (types.Map, diag.Diagnostics) {
	elements := map[string]tfattr.Value{}
	terraformLabels := strings.Split(alertTerraformLabel, ":")
	for key, value := range labels {
		if key == attr.Severity || key == terraformLabels[0] {
			continue
		}
		elements[key] = types.StringValue(value)
	}
	return types.MapValue(types.StringType, elements)
}

func fetchPreferredChannels(alertPreferredChannels []string) (types.List, diag.Diagnostics) {
	preferredChannels := utils.Map(alertPreferredChannels, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, preferredChannels)
}
