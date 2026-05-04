// Layer 3 expand/flatten for planned_maintenance. Pattern A — Postable
// expand consumes the resource model, Flatten returns the wide
// `*schemas.PlannedMaintenanceDataSourceModel`; the codegen `services`
// template auto-emits `plannedMaintenanceResourceFromDS` to narrow on
// the resource side.
package conv

import (
	"context"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ExpandRuletypesPostablePlannedMaintenance converts a
// `schemas.PlannedMaintenanceModel` (resource plan) into the wire-format
// POST/PUT body.
func ExpandRuletypesPostablePlannedMaintenance(ctx context.Context, m schemas.PlannedMaintenanceModel) (*apitypes.RuletypesPostablePlannedMaintenance, diag.Diagnostics) {
	var diags diag.Diagnostics

	alertIds, d := StringPointerSliceFromList(ctx, m.AlertIds)
	diags.Append(d...)

	sched, d := ExpandRuletypesSchedule(ctx, m.Schedule)
	diags.Append(d...)

	if diags.HasError() {
		return nil, diags
	}

	out := &apitypes.RuletypesPostablePlannedMaintenance{
		AlertIds:    alertIds,
		Description: StringPointer(m.Description),
		Name:        m.Name.ValueString(),
	}
	if sched != nil {
		out.Schedule = *sched
	}
	return out, diags
}

// FlattenRuletypesPlannedMaintenance converts a server response into the wide
// datasource model. The resource layer narrows via
// `plannedMaintenanceResourceFromDS`.
func FlattenRuletypesPlannedMaintenance(ctx context.Context, g *apitypes.RuletypesPlannedMaintenance) (*schemas.PlannedMaintenanceDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if g == nil {
		return nil, diags
	}

	alertIds, d := ListFromStringPointerSlice(ctx, g.AlertIds)
	diags.Append(d...)

	sched, d := FlattenRuletypesSchedule(ctx, &g.Schedule)
	diags.Append(d...)

	if diags.HasError() {
		return nil, diags
	}

	return &schemas.PlannedMaintenanceDataSourceModel{
		AlertIds:    alertIds,
		CreatedAt:   TimeStringFromPointer(g.CreatedAt),
		CreatedBy:   StringFromPointer(g.CreatedBy),
		Description: StringFromPointer(g.Description),
		Id:          types.StringValue(g.Id),
		Kind:        types.StringValue(string(g.Kind)),
		Name:        types.StringValue(g.Name),
		Schedule:    sched,
		UpdatedAt:   TimeStringFromPointer(g.UpdatedAt),
		UpdatedBy:   StringFromPointer(g.UpdatedBy),
	}, diags
}
