package services

import (
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform mappings between the gateway limit wire types (model) and the
// framework object values used by the limit resource and the key data source.

func limitValueToTerraform(v *model.LimitValue) (types.Object, diag.Diagnostics) {
	if v == nil {
		return types.ObjectNull(attr.LimitValueAttrTypes()), nil
	}

	value, d := types.ObjectValue(attr.LimitValueAttrTypes(), map[string]tfattr.Value{
		attr.Size:  int64PtrToTerraform(v.Size),
		attr.Count: int64PtrToTerraform(v.Count),
	})

	return value, d
}

func limitValueFromTerraform(object types.Object) (*model.LimitValue, diag.Diagnostics) {
	if object.IsNull() || object.IsUnknown() {
		return nil, nil
	}

	attrs := object.Attributes()

	size, d := int64PtrFromTerraform(attrs[attr.Size])

	diags := diag.Diagnostics{}
	diags.Append(d...)

	count, d := int64PtrFromTerraform(attrs[attr.Count])

	diags.Append(d...)

	if diags.HasError() {
		return nil, diags
	}

	return &model.LimitValue{Size: size, Count: count}, diags
}

func limitConfigToTerraform(c model.LimitConfig) (types.Object, diag.Diagnostics) {
	day, d := limitValueToTerraform(c.Day)

	diags := diag.Diagnostics{}
	diags.Append(d...)

	second, d := limitValueToTerraform(c.Second)

	diags.Append(d...)

	if diags.HasError() {
		return types.ObjectNull(attr.LimitConfigAttrTypes()), diags
	}

	value, d := types.ObjectValue(attr.LimitConfigAttrTypes(), map[string]tfattr.Value{
		attr.Day:    day,
		attr.Second: second,
	})

	return value, d
}

func limitConfigFromTerraform(object types.Object) (model.LimitConfig, diag.Diagnostics) {
	if object.IsNull() || object.IsUnknown() {
		return model.LimitConfig{}, nil
	}

	attrs := object.Attributes()

	day, d := limitValueFromTerraform(asObject(attrs[attr.Day]))

	diags := diag.Diagnostics{}
	diags.Append(d...)

	second, d := limitValueFromTerraform(asObject(attrs[attr.Second]))

	diags.Append(d...)

	if diags.HasError() {
		return model.LimitConfig{}, diags
	}

	return model.LimitConfig{Day: day, Second: second}, diags
}

func limitMetricToTerraform(m model.LimitMetric) (types.Object, diag.Diagnostics) {
	metricAttrs := map[string]tfattr.Type{attr.Count: types.Int64Type, attr.Size: types.Int64Type}

	arm := func(v *model.LimitMetricValue) (types.Object, diag.Diagnostics) {
		if v == nil {
			return types.ObjectNull(metricAttrs), nil
		}

		return types.ObjectValue(metricAttrs, map[string]tfattr.Value{
			attr.Count: types.Int64Value(v.Count),
			attr.Size:  types.Int64Value(v.Size),
		})
	}

	day, d := arm(m.Day)

	diags := diag.Diagnostics{}
	diags.Append(d...)

	second, d := arm(m.Second)

	diags.Append(d...)

	if diags.HasError() {
		return types.ObjectNull(attr.LimitMetricAttrTypes()), diags
	}

	value, d := types.ObjectValue(attr.LimitMetricAttrTypes(), map[string]tfattr.Value{
		attr.Day:    day,
		attr.Second: second,
	})

	return value, d
}

// limitToTerraformValue renders a full server limit as the element value of
// the key data source's limits list.
func limitToTerraformValue(l model.Limit) (tfattr.Value, diag.Diagnostics) {
	elementTypes := limitElementAttrTypes()

	config, d := limitConfigToTerraform(l.Config)

	diags := diag.Diagnostics{}
	diags.Append(d...)

	metric, d := limitMetricToTerraform(l.Metric)

	diags.Append(d...)

	tags := types.ListNull(types.StringType)

	if l.Tags != nil {
		tagsValue, d := types.ListValue(types.StringType, utils.Map(l.Tags, func(tag string) tfattr.Value {
			return types.StringValue(tag)
		}))

		diags.Append(d...)
		tags = tagsValue
	}

	if diags.HasError() {
		return types.ObjectNull(elementTypes), diags
	}

	value, d := types.ObjectValue(elementTypes, map[string]tfattr.Value{
		attr.ID:        types.StringValue(l.ID),
		attr.Signal:    types.StringValue(l.Signal),
		attr.Config:    config,
		attr.Metric:    metric,
		attr.KeyTags:   tags,
		attr.CreatedAt: types.StringValue(l.CreatedAt.Format(time.RFC3339Nano)),
		attr.UpdatedAt: types.StringValue(l.UpdatedAt.Format(time.RFC3339Nano)),
		attr.KeyID:     types.StringValue(l.KeyID),
	})

	return value, d
}

func limitElementAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.ID:        types.StringType,
		attr.Signal:    types.StringType,
		attr.Config:    types.ObjectType{AttrTypes: attr.LimitConfigAttrTypes()},
		attr.Metric:    types.ObjectType{AttrTypes: attr.LimitMetricAttrTypes()},
		attr.KeyTags:   types.ListType{ElemType: types.StringType},
		attr.CreatedAt: types.StringType,
		attr.UpdatedAt: types.StringType,
		attr.KeyID:     types.StringType,
	}
}

func int64PtrToTerraform(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(*v)
}

func int64PtrFromTerraform(v tfattr.Value) (*int64, diag.Diagnostics) {
	i, ok := v.(types.Int64)
	if !ok {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid limit value", "expected an integer attribute")}
	}

	if i.IsNull() || i.IsUnknown() {
		return nil, nil
	}

	value := i.ValueInt64()

	return &value, nil
}

func asObject(v tfattr.Value) types.Object {
	if object, ok := v.(types.Object); ok {
		return object
	}

	return types.ObjectNull(attr.LimitValueAttrTypes())
}
