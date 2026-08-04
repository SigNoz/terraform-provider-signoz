package schemas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// pipelineOperatorTypes are the processor types accepted by the pipelines API.
var pipelineOperatorTypes = []string{
	"add",
	"copy",
	"grok_parser",
	"json_parser",
	"move",
	"regex_parser",
	"remove",
	"retain",
	"severity_parser",
	"time_parser",
	"trace_parser",
}

// PipelinesResourceSchema models the org's full ordered list of log pipelines
// as a single resource — the API has no per-pipeline CRUD, a POST replaces the
// whole list. Server-managed mechanics (pipeline/operator ids, orderId, the
// operator output chain) are derived from list position and not exposed.
func PipelinesResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"pipelines": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"alias": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(20),
							},
						},
						"config": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: pipelineOperatorAttributes(),
							},
							Required: true,
						},
						"description": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"filter": schema.SingleNestedAttribute{
							Attributes: pipelineFilterAttributes(),
							Required:   true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
					},
				},
				Required: true,
			},
		},
	}
}

func pipelineFilterAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"key": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"data_type": schema.StringAttribute{
								Required: true,
							},
							"is_column": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"is_json": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"key": schema.StringAttribute{
								Required: true,
							},
							"type": schema.StringAttribute{
								Required: true,
							},
						},
						Required: true,
					},
					"op": schema.StringAttribute{
						Required: true,
					},
					"value": schema.StringAttribute{
						CustomType: jsontypes.NormalizedType{},
						Optional:   true,
					},
				},
			},
			Required: true,
		},
		"op": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("AND"),
			Validators: []validator.String{
				stringvalidator.OneOf("AND", "OR"),
			},
		},
	}
}

func pipelineOperatorAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"default": schema.StringAttribute{
			Optional: true,
		},
		"enable_flattening": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"enable_paths": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"enabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(true),
		},
		"expr": schema.StringAttribute{
			Optional: true,
		},
		"field": schema.StringAttribute{
			Optional: true,
		},
		"fields": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"from": schema.StringAttribute{
			Optional: true,
		},
		"if": schema.StringAttribute{
			Optional: true,
		},
		"layout": schema.StringAttribute{
			Optional: true,
		},
		"layout_type": schema.StringAttribute{
			Optional: true,
		},
		"mapping": schema.MapAttribute{
			ElementType: types.ListType{ElemType: types.StringType},
			Optional:    true,
		},
		"name": schema.StringAttribute{
			Optional: true,
		},
		"on_error": schema.StringAttribute{
			Optional: true,
		},
		"overwrite_text": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"parse_from": schema.StringAttribute{
			Optional: true,
		},
		"parse_to": schema.StringAttribute{
			Optional: true,
		},
		"path_prefix": schema.StringAttribute{
			Optional: true,
		},
		"pattern": schema.StringAttribute{
			Optional: true,
		},
		"regex": schema.StringAttribute{
			Optional: true,
		},
		"span_id": schema.SingleNestedAttribute{
			Attributes: pipelineParseFromAttributes(),
			Optional:   true,
		},
		"to": schema.StringAttribute{
			Optional: true,
		},
		"trace_flags": schema.SingleNestedAttribute{
			Attributes: pipelineParseFromAttributes(),
			Optional:   true,
		},
		"trace_id": schema.SingleNestedAttribute{
			Attributes: pipelineParseFromAttributes(),
			Optional:   true,
		},
		"type": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(pipelineOperatorTypes...),
			},
		},
		"value": schema.StringAttribute{
			Optional: true,
		},
	}
}

func pipelineParseFromAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"parse_from": schema.StringAttribute{
			Required: true,
		},
	}
}

type PipelinesModel struct {
	Id        types.String `tfsdk:"id"`
	Pipelines types.List   `tfsdk:"pipelines"`
}

type PipelineModel struct {
	Alias       types.String `tfsdk:"alias"`
	Config      types.List   `tfsdk:"config"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Filter      types.Object `tfsdk:"filter"`
	Name        types.String `tfsdk:"name"`
}

type PipelineFilterModel struct {
	Items types.List   `tfsdk:"items"`
	Op    types.String `tfsdk:"op"`
}

type PipelineFilterItemModel struct {
	Key   types.Object         `tfsdk:"key"`
	Op    types.String         `tfsdk:"op"`
	Value jsontypes.Normalized `tfsdk:"value"`
}

type PipelineFilterKeyModel struct {
	DataType types.String `tfsdk:"data_type"`
	IsColumn types.Bool   `tfsdk:"is_column"`
	IsJson   types.Bool   `tfsdk:"is_json"`
	Key      types.String `tfsdk:"key"`
	Type     types.String `tfsdk:"type"`
}

type PipelineOperatorModel struct {
	Default          types.String `tfsdk:"default"`
	EnableFlattening types.Bool   `tfsdk:"enable_flattening"`
	EnablePaths      types.Bool   `tfsdk:"enable_paths"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Expr             types.String `tfsdk:"expr"`
	Field            types.String `tfsdk:"field"`
	Fields           types.List   `tfsdk:"fields"`
	From             types.String `tfsdk:"from"`
	If               types.String `tfsdk:"if"`
	Layout           types.String `tfsdk:"layout"`
	LayoutType       types.String `tfsdk:"layout_type"`
	Mapping          types.Map    `tfsdk:"mapping"`
	Name             types.String `tfsdk:"name"`
	OnError          types.String `tfsdk:"on_error"`
	OverwriteText    types.Bool   `tfsdk:"overwrite_text"`
	ParseFrom        types.String `tfsdk:"parse_from"`
	ParseTo          types.String `tfsdk:"parse_to"`
	PathPrefix       types.String `tfsdk:"path_prefix"`
	Pattern          types.String `tfsdk:"pattern"`
	Regex            types.String `tfsdk:"regex"`
	SpanId           types.Object `tfsdk:"span_id"`
	To               types.String `tfsdk:"to"`
	TraceFlags       types.Object `tfsdk:"trace_flags"`
	TraceId          types.Object `tfsdk:"trace_id"`
	Type             types.String `tfsdk:"type"`
	Value            types.String `tfsdk:"value"`
}

type PipelineParseFromModel struct {
	ParseFrom types.String `tfsdk:"parse_from"`
}

func (PipelineModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"alias":       types.StringType,
		"config":      types.ListType{ElemType: types.ObjectType{AttrTypes: PipelineOperatorModel{}.AttributeTypes()}},
		"description": types.StringType,
		"enabled":     types.BoolType,
		"filter":      types.ObjectType{AttrTypes: PipelineFilterModel{}.AttributeTypes()},
		"name":        types.StringType,
	}
}

func (PipelineFilterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"items": types.ListType{ElemType: types.ObjectType{AttrTypes: PipelineFilterItemModel{}.AttributeTypes()}},
		"op":    types.StringType,
	}
}

func (PipelineFilterItemModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.ObjectType{AttrTypes: PipelineFilterKeyModel{}.AttributeTypes()},
		"op":    types.StringType,
		"value": jsontypes.NormalizedType{},
	}
}

func (PipelineFilterKeyModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data_type": types.StringType,
		"is_column": types.BoolType,
		"is_json":   types.BoolType,
		"key":       types.StringType,
		"type":      types.StringType,
	}
}

func (PipelineOperatorModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"default":           types.StringType,
		"enable_flattening": types.BoolType,
		"enable_paths":      types.BoolType,
		"enabled":           types.BoolType,
		"expr":              types.StringType,
		"field":             types.StringType,
		"fields":            types.ListType{ElemType: types.StringType},
		"from":              types.StringType,
		"if":                types.StringType,
		"layout":            types.StringType,
		"layout_type":       types.StringType,
		"mapping":           types.MapType{ElemType: types.ListType{ElemType: types.StringType}},
		"name":              types.StringType,
		"on_error":          types.StringType,
		"overwrite_text":    types.BoolType,
		"parse_from":        types.StringType,
		"parse_to":          types.StringType,
		"path_prefix":       types.StringType,
		"pattern":           types.StringType,
		"regex":             types.StringType,
		"span_id":           types.ObjectType{AttrTypes: PipelineParseFromModel{}.AttributeTypes()},
		"to":                types.StringType,
		"trace_flags":       types.ObjectType{AttrTypes: PipelineParseFromModel{}.AttributeTypes()},
		"trace_id":          types.ObjectType{AttrTypes: PipelineParseFromModel{}.AttributeTypes()},
		"type":              types.StringType,
		"value":             types.StringType,
	}
}

func (PipelineParseFromModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"parse_from": types.StringType,
	}
}
