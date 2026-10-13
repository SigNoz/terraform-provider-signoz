package services

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &logPipelineDataSource{}
)

// NewLogPipelineDataSource is a helper function to simplify the provider implementation.
func NewLogPipelineDataSource() datasource.DataSource {
	return &logPipelineDataSource{}
}

// logPipelineDataSource is the data source implementation.
type logPipelineDataSource struct {
	client *legacyclient.Client
}

// logPipelineModel maps log pipeline schema data.
type logPipelineModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Alias       types.String `tfsdk:"alias"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	OrderID     types.Int64  `tfsdk:"order_id"`
	Filter      types.String `tfsdk:"filter"`
	Config      types.String `tfsdk:"config"`
	CreatedBy   types.String `tfsdk:"created_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// Configure adds the provider configured client to the data source.
func (d *logPipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			"signoz_log_pipeline",
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
func (d *logPipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_log_pipeline"
}

// Schema defines the schema for the data source.
func (d *logPipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a log pipeline from SigNoz using its ID.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the log pipeline.",
			},
			attr.PipelineName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the log pipeline.",
			},
			attr.Alias: schema.StringAttribute{
				Computed:    true,
				Description: "Alias identifier for the log pipeline.",
			},
			attr.Description: schema.StringAttribute{
				Computed:    true,
				Description: "Description of the log pipeline.",
			},
			attr.PipelineEnabled: schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the log pipeline is enabled.",
			},
			attr.OrderID: schema.Int64Attribute{
				Computed:    true,
				Description: "Order ID of the log pipeline.",
			},
			attr.PipelineFilter: schema.StringAttribute{
				Computed:    true,
				Description: "Filter configuration for the log pipeline as a JSON string.",
			},
			attr.PipelineConfig: schema.StringAttribute{
				Computed:    true,
				Description: "Processor configuration for the log pipeline as a JSON array string.",
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "User who created the log pipeline.",
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the log pipeline was created.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *logPipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data logPipelineModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pipeline, err := d.client.GetLogPipeline(ctx, data.ID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz log pipeline: %s", err.Error()), "signoz_log_pipeline")
		return
	}

	// Set state values from retrieved data.
	data.ID = types.StringValue(pipeline.ID)
	data.Name = types.StringValue(pipeline.Name)
	data.Alias = types.StringValue(pipeline.Alias)
	data.Description = types.StringValue(pipeline.Description)
	data.Enabled = types.BoolValue(pipeline.Enabled)
	data.OrderID = types.Int64Value(int64(pipeline.OrderID))
	data.CreatedBy = types.StringValue(pipeline.CreatedBy)
	data.CreatedAt = types.StringValue(pipeline.CreatedAt)

	data.Filter, err = pipeline.FilterToTerraform()
	if err != nil {
		addReadErr(&resp.Diagnostics, err, "signoz_log_pipeline")
		return
	}

	data.Config, err = pipeline.ConfigToTerraform()
	if err != nil {
		addReadErr(&resp.Diagnostics, err, "signoz_log_pipeline")
		return
	}

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
