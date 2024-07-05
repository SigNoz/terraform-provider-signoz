package datasource

import (
	"encoding/json"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func fetchCondition(alertCondition map[string]interface{}) (types.String, error) {
	condition, err := json.Marshal(alertCondition)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(string(condition)), nil
}

func fetchLabels(alertLabels map[string]string) (types.Map, diag.Diagnostics) {
	labels := make(map[string]attr.Value, len(alertLabels))
	for key, value := range alertLabels {
		labels[key] = types.StringValue(value)
	}

	return types.MapValue(types.StringType, labels)
}

func fetchPreferredChannels(alertPreferredChannels []string) (types.List, diag.Diagnostics) {
	preferredChannels := utils.Map(alertPreferredChannels, func(value string) attr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, preferredChannels)
}
