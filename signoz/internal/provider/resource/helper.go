package resource

import (
	"fmt"
	"strings"

	signozattr "github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// addErr adds an error to the diagnostics.
func addErr(diagnostics *diag.Diagnostics, err error, operation string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operation, SigNozAlert),
		err.Error(),
	)
}

// createLabels creates labels for an alert.
func createLabels(planLabels types.Map, planSeverity types.String) map[string]string {
	labels := make(map[string]string)
	for key, value := range planLabels.Elements() {
		labels[key] = strings.Trim(value.String(), "\"")
	}
	terraformLabel := strings.Split(alertTerraformLabel, ":")
	labels[strings.TrimSpace(terraformLabel[0])] = strings.TrimSpace(terraformLabel[1])

	if planSeverity.ValueString() != "" {
		labels[signozattr.Severity] = planSeverity.ValueString()
	}

	return labels
}

// createRuleType creates rule type for an alert.
func createRuleType(planRuleType types.String) (string, error) {
	ruleType := planRuleType.ValueString()
	if ruleType == "" || (ruleType != model.AlertRuleTypeProm && ruleType != model.AlertRuleTypeThreshold) {
		return "", fmt.Errorf("invalid rule type %s", ruleType)
	}
	return ruleType, nil
}

// fetchLabels fetches labels for an alert.
func fetchLabels(labels map[string]string) (types.Map, diag.Diagnostics) {
	elements := map[string]attr.Value{}
	terraformLabels := strings.Split(alertTerraformLabel, ":")
	for key, value := range labels {
		if key == signozattr.Severity || key == terraformLabels[0] {
			continue
		}
		elements[key] = types.StringValue(value)
	}
	return types.MapValue(types.StringType, elements)
}

// fetchPreferredChannels fetches preferred channels for an alert.
func fetchPreferredChannels(preferredChannels []string) (types.List, diag.Diagnostics) {
	elements := utils.Map(preferredChannels, func(value string) attr.Value {
		return types.StringValue(value)
	})
	return types.ListValue(types.StringType, elements)
}
