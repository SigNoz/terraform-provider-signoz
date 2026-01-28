package planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
)

// NormalizeDuration returns a plan modifier that normalizes duration strings
// to Go's canonical format. This ensures that semantically equivalent durations
// like "60m0s" and "1h0m0s" are treated as equal.
func NormalizeDuration() planmodifier.String {
	return normalizeDurationModifier{}
}

type normalizeDurationModifier struct{}

func (m normalizeDurationModifier) Description(_ context.Context) string {
	return "Normalizes duration strings to a canonical format (e.g., 60m0s becomes 1h0m0s)."
}

func (m normalizeDurationModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes duration strings to a canonical format (e.g., `60m0s` becomes `1h0m0s`)."
}

func (m normalizeDurationModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the value is unknown or null, don't modify it
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}

	// Normalize the planned value
	normalized := utils.NormalizeDuration(req.PlanValue.ValueString())

	// If normalization changed the value, update the plan
	if normalized != req.PlanValue.ValueString() {
		resp.PlanValue = req.PlanValue
		// We keep the original value but use semantic equality via state comparison
	}

	// If there's a state value, check if the normalized versions are equal
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		planNormalized := utils.NormalizeDuration(req.PlanValue.ValueString())
		stateNormalized := utils.NormalizeDuration(req.StateValue.ValueString())

		// If normalized values are equal, preserve the state value to prevent spurious diffs
		if planNormalized == stateNormalized {
			resp.PlanValue = req.StateValue
		}
	}
}
