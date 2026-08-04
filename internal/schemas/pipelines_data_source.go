package schemas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PipelinesDataSourceSchema reads a log pipelines config version — the one
// given by `version`, or the latest when unset.
func PipelinesDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pipelines": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"alias": schema.StringAttribute{
							Computed: true,
						},
						"config": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: pipelineOperatorDataSourceAttributes(),
							},
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"filter": schema.SingleNestedAttribute{
							Attributes: pipelineFilterDataSourceAttributes(),
							Computed:   true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"version": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func pipelineFilterDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"key": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"data_type": schema.StringAttribute{
								Computed: true,
							},
							"is_column": schema.BoolAttribute{
								Computed: true,
							},
							"is_json": schema.BoolAttribute{
								Computed: true,
							},
							"key": schema.StringAttribute{
								Computed: true,
							},
							"type": schema.StringAttribute{
								Computed: true,
							},
						},
						Computed: true,
					},
					"op": schema.StringAttribute{
						Computed: true,
					},
					"value": schema.StringAttribute{
						CustomType: jsontypes.NormalizedType{},
						Computed:   true,
					},
				},
			},
			Computed: true,
		},
		"op": schema.StringAttribute{
			Computed: true,
		},
	}
}

func pipelineOperatorDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"default": schema.StringAttribute{
			Computed: true,
		},
		"enable_flattening": schema.BoolAttribute{
			Computed: true,
		},
		"enable_paths": schema.BoolAttribute{
			Computed: true,
		},
		"enabled": schema.BoolAttribute{
			Computed: true,
		},
		"expr": schema.StringAttribute{
			Computed: true,
		},
		"field": schema.StringAttribute{
			Computed: true,
		},
		"fields": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"from": schema.StringAttribute{
			Computed: true,
		},
		"if": schema.StringAttribute{
			Computed: true,
		},
		"layout": schema.StringAttribute{
			Computed: true,
		},
		"layout_type": schema.StringAttribute{
			Computed: true,
		},
		"mapping": schema.MapAttribute{
			ElementType: types.ListType{ElemType: types.StringType},
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"on_error": schema.StringAttribute{
			Computed: true,
		},
		"overwrite_text": schema.BoolAttribute{
			Computed: true,
		},
		"parse_from": schema.StringAttribute{
			Computed: true,
		},
		"parse_to": schema.StringAttribute{
			Computed: true,
		},
		"path_prefix": schema.StringAttribute{
			Computed: true,
		},
		"pattern": schema.StringAttribute{
			Computed: true,
		},
		"regex": schema.StringAttribute{
			Computed: true,
		},
		"span_id": schema.SingleNestedAttribute{
			Attributes: pipelineParseFromDataSourceAttributes(),
			Computed:   true,
		},
		"to": schema.StringAttribute{
			Computed: true,
		},
		"trace_flags": schema.SingleNestedAttribute{
			Attributes: pipelineParseFromDataSourceAttributes(),
			Computed:   true,
		},
		"trace_id": schema.SingleNestedAttribute{
			Attributes: pipelineParseFromDataSourceAttributes(),
			Computed:   true,
		},
		"type": schema.StringAttribute{
			Computed: true,
		},
		"value": schema.StringAttribute{
			Computed: true,
		},
	}
}

func pipelineParseFromDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"parse_from": schema.StringAttribute{
			Computed: true,
		},
	}
}

type PipelinesDataSourceModel struct {
	Pipelines types.List  `tfsdk:"pipelines"`
	Version   types.Int64 `tfsdk:"version"`
}
