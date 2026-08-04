package conv

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/convtypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// integrationPipelineAliasPrefix marks pipelines owned by installed SigNoz
// integrations (alias `integration--<id>--<name>`). The server merges them
// into every read with a fresh random id, and silently drops user pipelines
// whose alias collides with the prefix — so they are excluded here on both
// sides: never posted, never flattened into state.
const integrationPipelineAliasPrefix = "integration"

func ExpandPipelinetypesPostablePipelines(ctx context.Context, m schemas.PipelinesModel) (*apitypes.PipelinetypesPostablePipelines, diag.Diagnostics) {
	var diags diag.Diagnostics

	var pipelines []schemas.PipelineModel
	diags.Append(m.Pipelines.ElementsAs(ctx, &pipelines, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := &apitypes.PipelinetypesPostablePipelines{
		Pipelines: make([]apitypes.PipelinetypesPostablePipeline, 0, len(pipelines)),
	}
	for i, p := range pipelines {
		if strings.HasPrefix(p.Alias.ValueString(), integrationPipelineAliasPrefix) {
			diags.AddError(
				"Reserved pipeline alias",
				fmt.Sprintf("Alias %q: the %q prefix is reserved for SigNoz integration-managed pipelines.", p.Alias.ValueString(), integrationPipelineAliasPrefix),
			)
			return nil, diags
		}

		expanded, d := expandPipeline(ctx, p, i+1)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out.Pipelines = append(out.Pipelines, *expanded)
	}

	return out, diags
}

// expandPipeline converts one planned pipeline. orderID comes from list
// position — the API renumbers positionally on every save, so position is the
// only stable order.
func expandPipeline(ctx context.Context, m schemas.PipelineModel, orderID int) (*apitypes.PipelinetypesPostablePipeline, diag.Diagnostics) {
	var diags diag.Diagnostics

	var filterModel schemas.PipelineFilterModel
	diags.Append(m.Filter.As(ctx, &filterModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	filter, d := expandPipelineFilter(ctx, filterModel)
	diags.Append(d...)

	var operators []schemas.PipelineOperatorModel
	diags.Append(m.Config.ElementsAs(ctx, &operators, false)...)
	if diags.HasError() {
		return nil, diags
	}

	config := make([]apitypes.PipelinetypesPipelineOperator, 0, len(operators))
	for i, op := range operators {
		expanded, d := expandPipelineOperator(ctx, op, i, len(operators))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		config = append(config, *expanded)
	}

	return &apitypes.PipelinetypesPostablePipeline{
		OrderID:     orderID,
		Name:        m.Name.ValueString(),
		Alias:       m.Alias.ValueString(),
		Description: convtypes.StringPointer(m.Description),
		Enabled:     m.Enabled.ValueBool(),
		Filter:      filter,
		Config:      config,
	}, diags
}

func expandPipelineFilter(ctx context.Context, m schemas.PipelineFilterModel) (*apitypes.V3FilterSet, diag.Diagnostics) {
	var diags diag.Diagnostics

	var items []schemas.PipelineFilterItemModel
	diags.Append(m.Items.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := &apitypes.V3FilterSet{
		Operator: m.Op.ValueString(),
		Items:    make([]apitypes.V3FilterItem, 0, len(items)),
	}
	for _, item := range items {
		var key schemas.PipelineFilterKeyModel
		diags.Append(item.Key.As(ctx, &key, basetypes.ObjectAsOptions{})...)

		var value interface{}
		if !item.Value.IsNull() && !item.Value.IsUnknown() {
			if err := json.Unmarshal([]byte(item.Value.ValueString()), &value); err != nil {
				diags.AddError("Invalid pipeline filter value", err.Error())
			}
		}
		if diags.HasError() {
			return nil, diags
		}

		out.Items = append(out.Items, apitypes.V3FilterItem{
			Key: apitypes.V3AttributeKey{
				Key:      key.Key.ValueString(),
				DataType: key.DataType.ValueString(),
				Type:     key.Type.ValueString(),
				IsColumn: key.IsColumn.ValueBool(),
				IsJSON:   key.IsJson.ValueBool(),
			},
			Value:    value,
			Operator: item.Op.ValueString(),
		})
	}

	return out, diags
}

// expandPipelineOperator converts one planned operator. Operator ids and the
// output chain are positional: operator i is "op-<i+1>", every operator but
// the last outputs to the next one — the linear chain the API validates.
func expandPipelineOperator(ctx context.Context, m schemas.PipelineOperatorModel, index, total int) (*apitypes.PipelinetypesPipelineOperator, diag.Diagnostics) {
	var diags diag.Diagnostics

	fields, d := convtypes.StringPointerSliceFromList(ctx, m.Fields)
	diags.Append(d...)

	var mapping *map[string][]string
	if !m.Mapping.IsNull() && !m.Mapping.IsUnknown() {
		mm := map[string][]string{}
		diags.Append(m.Mapping.ElementsAs(ctx, &mm, false)...)
		mapping = &mm
	}

	traceID, d := expandPipelineParseFrom(ctx, m.TraceId)
	diags.Append(d...)
	spanID, d := expandPipelineParseFrom(ctx, m.SpanId)
	diags.Append(d...)
	traceFlags, d := expandPipelineParseFrom(ctx, m.TraceFlags)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	output := ""
	if index < total-1 {
		output = pipelineOperatorID(index + 1)
	}

	return &apitypes.PipelinetypesPipelineOperator{
		Type:                  m.Type.ValueString(),
		ID:                    pipelineOperatorID(index),
		Output:                output,
		OnError:               convtypes.StringPointer(m.OnError),
		If:                    convtypes.StringPointer(m.If),
		OrderID:               index + 1,
		Enabled:               m.Enabled.ValueBool(),
		Name:                  convtypes.StringPointer(m.Name),
		ParseTo:               convtypes.StringPointer(m.ParseTo),
		Pattern:               convtypes.StringPointer(m.Pattern),
		Regex:                 convtypes.StringPointer(m.Regex),
		ParseFrom:             convtypes.StringPointer(m.ParseFrom),
		TraceId:               traceID,
		SpanId:                spanID,
		TraceFlags:            traceFlags,
		Field:                 convtypes.StringPointer(m.Field),
		Value:                 convtypes.StringPointer(m.Value),
		From:                  convtypes.StringPointer(m.From),
		To:                    convtypes.StringPointer(m.To),
		Expr:                  convtypes.StringPointer(m.Expr),
		Fields:                fields,
		Default:               convtypes.StringPointer(m.Default),
		Layout:                convtypes.StringPointer(m.Layout),
		LayoutType:            convtypes.StringPointer(m.LayoutType),
		EnableFlattening:      m.EnableFlattening.ValueBool(),
		EnablePaths:           m.EnablePaths.ValueBool(),
		PathPrefix:            convtypes.StringPointer(m.PathPrefix),
		Mapping:               mapping,
		OverwriteSeverityText: m.OverwriteText.ValueBool(),
	}, diags
}

func expandPipelineParseFrom(ctx context.Context, o types.Object) (*apitypes.PipelinetypesParseFrom, diag.Diagnostics) {
	var diags diag.Diagnostics

	if o.IsNull() || o.IsUnknown() {
		return nil, diags
	}

	var m schemas.PipelineParseFromModel
	diags.Append(o.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	return &apitypes.PipelinetypesParseFrom{ParseFrom: m.ParseFrom.ValueString()}, diags
}

func pipelineOperatorID(index int) string {
	return fmt.Sprintf("op-%d", index+1)
}

func FlattenLogparsingpipelinePipelinesResponse(ctx context.Context, g *apitypes.LogparsingpipelinePipelinesResponse) (*schemas.PipelinesDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if g == nil {
		return nil, diags
	}

	models := make([]schemas.PipelineModel, 0, len(g.Pipelines))
	for _, p := range g.Pipelines {
		if strings.HasPrefix(p.Alias, integrationPipelineAliasPrefix) {
			continue
		}

		model, d := flattenPipeline(ctx, p)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		models = append(models, *model)
	}

	pipelinesFlat, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: schemas.PipelineModel{}.AttributeTypes()}, models)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &schemas.PipelinesDataSourceModel{
		Pipelines: pipelinesFlat,
		Version:   convtypes.IntFromPointer(g.Version),
	}, diags
}

func flattenPipeline(ctx context.Context, g apitypes.PipelinetypesGettablePipeline) (*schemas.PipelineModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	filterFlat, d := flattenPipelineFilter(ctx, g.Filter)
	diags.Append(d...)

	operatorModels := make([]schemas.PipelineOperatorModel, 0, len(g.Config))
	for _, op := range g.Config {
		model, d := flattenPipelineOperator(ctx, op)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		operatorModels = append(operatorModels, *model)
	}

	configFlat, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: schemas.PipelineOperatorModel{}.AttributeTypes()}, operatorModels)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &schemas.PipelineModel{
		Alias:       types.StringValue(g.Alias),
		Config:      configFlat,
		Description: convtypes.StringFromPointer(g.Description),
		Enabled:     types.BoolValue(g.Enabled),
		Filter:      filterFlat,
		Name:        types.StringValue(g.Name),
	}, diags
}

func flattenPipelineFilter(ctx context.Context, g *apitypes.V3FilterSet) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	filterAttrTypes := schemas.PipelineFilterModel{}.AttributeTypes()
	if g == nil {
		return types.ObjectNull(filterAttrTypes), diags
	}

	itemModels := make([]schemas.PipelineFilterItemModel, 0, len(g.Items))
	for _, item := range g.Items {
		keyFlat, d := types.ObjectValueFrom(ctx, schemas.PipelineFilterKeyModel{}.AttributeTypes(), schemas.PipelineFilterKeyModel{
			DataType: types.StringValue(item.Key.DataType),
			IsColumn: types.BoolValue(item.Key.IsColumn),
			IsJson:   types.BoolValue(item.Key.IsJSON),
			Key:      types.StringValue(item.Key.Key),
			Type:     types.StringValue(item.Key.Type),
		})
		diags.Append(d...)

		valueFlat := jsontypes.NewNormalizedNull()
		if item.Value != nil {
			raw, err := json.Marshal(item.Value)
			if err != nil {
				diags.AddError("Flatten pipeline filter value", err.Error())
			} else {
				valueFlat = jsontypes.NewNormalizedValue(string(raw))
			}
		}
		if diags.HasError() {
			return types.ObjectNull(filterAttrTypes), diags
		}

		itemModels = append(itemModels, schemas.PipelineFilterItemModel{
			Key:   keyFlat,
			Op:    types.StringValue(item.Operator),
			Value: valueFlat,
		})
	}

	itemsFlat, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: schemas.PipelineFilterItemModel{}.AttributeTypes()}, itemModels)
	diags.Append(d...)

	filterFlat, d := types.ObjectValueFrom(ctx, filterAttrTypes, schemas.PipelineFilterModel{
		Items: itemsFlat,
		Op:    types.StringValue(g.Operator),
	})
	diags.Append(d...)

	return filterFlat, diags
}

func flattenPipelineOperator(ctx context.Context, g apitypes.PipelinetypesPipelineOperator) (*schemas.PipelineOperatorModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	fieldsFlat, d := convtypes.ListFromStringPointerSlice(ctx, g.Fields)
	diags.Append(d...)

	mappingFlat := types.MapNull(types.ListType{ElemType: types.StringType})
	if g.Mapping != nil {
		mappingFlat, d = types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, *g.Mapping)
		diags.Append(d...)
	}

	traceIDFlat, d := flattenPipelineParseFrom(ctx, g.TraceId)
	diags.Append(d...)
	spanIDFlat, d := flattenPipelineParseFrom(ctx, g.SpanId)
	diags.Append(d...)
	traceFlagsFlat, d := flattenPipelineParseFrom(ctx, g.TraceFlags)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &schemas.PipelineOperatorModel{
		Default:          convtypes.StringFromPointer(g.Default),
		EnableFlattening: types.BoolValue(g.EnableFlattening),
		EnablePaths:      types.BoolValue(g.EnablePaths),
		Enabled:          types.BoolValue(g.Enabled),
		Expr:             convtypes.StringFromPointer(g.Expr),
		Field:            convtypes.StringFromPointer(g.Field),
		Fields:           fieldsFlat,
		From:             convtypes.StringFromPointer(g.From),
		If:               convtypes.StringFromPointer(g.If),
		Layout:           convtypes.StringFromPointer(g.Layout),
		LayoutType:       convtypes.StringFromPointer(g.LayoutType),
		Mapping:          mappingFlat,
		Name:             convtypes.StringFromPointer(g.Name),
		OnError:          convtypes.StringFromPointer(g.OnError),
		OverwriteText:    types.BoolValue(g.OverwriteSeverityText),
		ParseFrom:        convtypes.StringFromPointer(g.ParseFrom),
		ParseTo:          convtypes.StringFromPointer(g.ParseTo),
		PathPrefix:       convtypes.StringFromPointer(g.PathPrefix),
		Pattern:          convtypes.StringFromPointer(g.Pattern),
		Regex:            convtypes.StringFromPointer(g.Regex),
		SpanId:           spanIDFlat,
		To:               convtypes.StringFromPointer(g.To),
		TraceFlags:       traceFlagsFlat,
		TraceId:          traceIDFlat,
		Type:             types.StringValue(g.Type),
		Value:            convtypes.StringFromPointer(g.Value),
	}, diags
}

func flattenPipelineParseFrom(ctx context.Context, g *apitypes.PipelinetypesParseFrom) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrTypes := schemas.PipelineParseFromModel{}.AttributeTypes()
	if g == nil {
		return types.ObjectNull(attrTypes), diags
	}

	flat, d := types.ObjectValueFrom(ctx, attrTypes, schemas.PipelineParseFromModel{
		ParseFrom: types.StringValue(g.ParseFrom),
	})
	diags.Append(d...)

	return flat, diags
}
