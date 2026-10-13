package services

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &downtimeScheduleDataSource{}
)

// NewDowntimeScheduleDataSource is a helper function to simplify the provider implementation.
func NewDowntimeScheduleDataSource() datasource.DataSource {
	return &downtimeScheduleDataSource{}
}

// downtimeScheduleDataSource is the data source implementation.
type downtimeScheduleDataSource struct {
	client *legacyclient.Client
}

// downtimeScheduleModel maps downtime schedule schema data.
type downtimeScheduleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AlertIDs    types.List   `tfsdk:"alert_ids"`
	AllAlerts   types.Bool   `tfsdk:"all_alerts"`
	Schedule    types.Object `tfsdk:"schedule"`
	CreatedAt   types.String `tfsdk:"created_at"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
}

// Configure adds the provider configured client to the data source.
func (d *downtimeScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			"signoz_downtime_schedule",
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
func (d *downtimeScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_downtime_schedule"
}

// Schema defines the schema for the data source.
func (d *downtimeScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a downtime schedule from SigNoz using its ID.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the downtime schedule.",
			},
			attr.ScheduleName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the downtime schedule.",
			},
			attr.Description: schema.StringAttribute{
				Computed:    true,
				Description: "Description of the downtime schedule.",
			},
			attr.AlertIDs: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of alert IDs suppressed during the downtime window.",
			},
			attr.AllAlerts: schema.BoolAttribute{
				Computed:    true,
				Description: "Whether all alerts are suppressed during the downtime window.",
			},
			attr.Schedule: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Schedule configuration for the downtime window.",
				Attributes: map[string]schema.Attribute{
					attr.StartTime: schema.StringAttribute{
						Computed:    true,
						Description: "Start time of the downtime window in RFC3339 format.",
					},
					attr.EndTime: schema.StringAttribute{
						Computed:    true,
						Description: "End time of the downtime window in RFC3339 format.",
					},
					attr.Timezone: schema.StringAttribute{
						Computed:    true,
						Description: "Timezone of the downtime schedule.",
					},
					attr.Recurrence: schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Recurrence configuration for the downtime schedule.",
						Attributes: map[string]schema.Attribute{
							attr.RecurrenceType: schema.StringAttribute{
								Computed:    true,
								Description: "Type of recurrence (once, daily, weekly).",
							},
							attr.DaysOfWeek: schema.ListAttribute{
								Computed:    true,
								ElementType: types.Int64Type,
								Description: "Days of the week for weekly recurrence (0=Sunday, 6=Saturday).",
							},
							attr.Duration: schema.StringAttribute{
								Computed:    true,
								Description: "Duration of the downtime window for recurring schedules.",
							},
						},
					},
				},
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the downtime schedule.",
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the downtime schedule.",
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the downtime schedule.",
			},
			attr.UpdatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the downtime schedule.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *downtimeScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data downtimeScheduleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetDowntimeSchedule(ctx, data.ID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz downtime schedule: %s", err.Error()), "signoz_downtime_schedule")
		return
	}

	// Set state values from retrieved data.
	data.ID = types.StringValue(result.ID)
	data.Name = types.StringValue(result.Name)
	data.Description = types.StringValue(result.Description)
	data.AllAlerts = types.BoolValue(result.AllAlerts)
	data.CreatedAt = types.StringValue(result.CreatedAt)
	data.CreatedBy = types.StringValue(result.CreatedBy)
	data.UpdatedAt = types.StringValue(result.UpdatedAt)
	data.UpdatedBy = types.StringValue(result.UpdatedBy)

	var diags diag.Diagnostics
	data.AlertIDs, diags = result.AlertIDsToTerraform()
	resp.Diagnostics.Append(diags...)

	data.Schedule, diags = result.ScheduleToTerraform(ctx)
	resp.Diagnostics.Append(diags...)

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
