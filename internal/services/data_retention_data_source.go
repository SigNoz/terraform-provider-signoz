package services

import (
	"context"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	dataRetentionDatasourceSyntheticID = "settings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &dataRetentionDataSource{}
)

// NewDataRetentionDataSource is a helper function to simplify the provider implementation.
func NewDataRetentionDataSource() datasource.DataSource {
	return &dataRetentionDataSource{}
}

// dataRetentionDataSource is the data source implementation.
type dataRetentionDataSource struct {
	client *legacyclient.Client
}

// dataRetentionModel maps data retention schema data.
type dataRetentionModel struct {
	ID                        types.String `tfsdk:"id"`
	MetricsTTLDurationHrs     types.Int64  `tfsdk:"metrics_ttl_duration_hrs"`
	MetricsMoveTTLDurationHrs types.Int64  `tfsdk:"metrics_move_ttl_duration_hrs"`
	LogsTTLDurationHrs        types.Int64  `tfsdk:"logs_ttl_duration_hrs"`
	LogsMoveTTLDurationHrs    types.Int64  `tfsdk:"logs_move_ttl_duration_hrs"`
	TracesTTLDurationHrs      types.Int64  `tfsdk:"traces_ttl_duration_hrs"`
	TracesMoveTTLDurationHrs  types.Int64  `tfsdk:"traces_move_ttl_duration_hrs"`
}

// Configure adds the provider configured client to the data source.
func (d *dataRetentionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			SigNozDataRetention,
		)

		return
	}

	legacyClient, err := legacyclient.NewClient(api)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build legacy SigNoz client", err.Error())
		return
	}

	d.client = legacyClient
}

// Metadata returns the data source type name.
func (d *dataRetentionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_data_retention"
}

// Schema defines the schema for the data source.
func (d *dataRetentionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the current data retention settings from SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Synthetic ID for the data retention settings. Always \"settings\".",
			},
			attr.MetricsTTLDuration: schema.Int64Attribute{
				Computed:    true,
				Description: "Hours to retain metrics data.",
			},
			attr.MetricsMoveTTLDuration: schema.Int64Attribute{
				Computed:    true,
				Description: "Hours before moving metrics to cold storage. 0 means disabled.",
			},
			attr.LogsTTLDuration: schema.Int64Attribute{
				Computed:    true,
				Description: "Hours to retain logs data.",
			},
			attr.LogsMoveTTLDuration: schema.Int64Attribute{
				Computed:    true,
				Description: "Hours before moving logs to cold storage. 0 means disabled.",
			},
			attr.TracesTTLDuration: schema.Int64Attribute{
				Computed:    true,
				Description: "Hours to retain traces data.",
			},
			attr.TracesMoveTTLDuration: schema.Int64Attribute{
				Computed:    true,
				Description: "Hours before moving traces to cold storage. 0 means disabled.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataRetentionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataRetentionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	retention, err := d.client.GetDataRetention(ctx)
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz data retention settings: %s", err.Error()), SigNozDataRetention)
		return
	}

	data.ID = types.StringValue(dataRetentionDatasourceSyntheticID)
	data.MetricsTTLDurationHrs = types.Int64Value(int64(retention.MetricsTTLDurationHrs))
	data.MetricsMoveTTLDurationHrs = types.Int64Value(int64(retention.MetricsMoveTTLDurationHrs))
	data.LogsTTLDurationHrs = types.Int64Value(int64(retention.LogsTTLDurationHrs))
	data.LogsMoveTTLDurationHrs = types.Int64Value(int64(retention.LogsMoveTTLDurationHrs))
	data.TracesTTLDurationHrs = types.Int64Value(int64(retention.TracesTTLDurationHrs))
	data.TracesMoveTTLDurationHrs = types.Int64Value(int64(retention.TracesMoveTTLDurationHrs))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
