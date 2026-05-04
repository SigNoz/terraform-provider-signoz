// Layer 3 expand/flatten for route_policy. Pattern A — `Flatten`
// returns `*schemas.RoutePolicyDataSourceModel` directly; the codegen
// `services` template auto-emits the narrowing `routePolicyResourceFromDS`.
package conv

import (
	"context"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ExpandAlertmanagertypesPostableRoutePolicy converts the framework resource model into
// the wire-format POST/PUT body.
func ExpandAlertmanagertypesPostableRoutePolicy(ctx context.Context, m schemas.RoutePolicyModel) (*apitypes.AlertmanagertypesPostableRoutePolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	channels, d := StringPointerSliceFromList(ctx, m.Channels)
	diags.Append(d...)

	tags, d := StringPointerSliceFromList(ctx, m.Tags)
	diags.Append(d...)

	if diags.HasError() {
		return nil, diags
	}

	out := &apitypes.AlertmanagertypesPostableRoutePolicy{
		Channels:    channels,
		Description: StringPointer(m.Description),
		Expression:  m.Expression.ValueString(),
		Name:        m.Name.ValueString(),
		Tags:        tags,
	}
	if !m.Kind.IsNull() && !m.Kind.IsUnknown() {
		k := apitypes.AlertmanagertypesExpressionKind(m.Kind.ValueString())
		out.Kind = &k
	}
	return out, diags
}

// FlattenAlertmanagertypesGettableRoutePolicy converts a server response into the wide datasource
// model. The resource layer narrows.
func FlattenAlertmanagertypesGettableRoutePolicy(ctx context.Context, g *apitypes.AlertmanagertypesGettableRoutePolicy) (*schemas.RoutePolicyDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if g == nil {
		return nil, diags
	}

	channels, d := ListFromStringPointerSlice(ctx, g.Channels)
	diags.Append(d...)

	tags, d := ListFromStringPointerSlice(ctx, g.Tags)
	diags.Append(d...)

	if diags.HasError() {
		return nil, diags
	}

	var kind types.String
	if g.Kind != nil {
		kind = types.StringValue(string(*g.Kind))
	} else {
		kind = types.StringNull()
	}

	return &schemas.RoutePolicyDataSourceModel{
		Channels:    channels,
		CreatedAt:   TimeStringFromValue(g.CreatedAt),
		CreatedBy:   StringFromPointer(g.CreatedBy),
		Description: StringFromPointer(g.Description),
		Expression:  types.StringValue(g.Expression),
		Id:          types.StringValue(g.Id),
		Kind:        kind,
		Name:        types.StringValue(g.Name),
		Tags:        tags,
		UpdatedAt:   TimeStringFromValue(g.UpdatedAt),
		UpdatedBy:   StringFromPointer(g.UpdatedBy),
	}, diags
}
