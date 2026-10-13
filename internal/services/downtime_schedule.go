package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &downtimeScheduleResource{}
	_ resource.ResourceWithConfigure   = &downtimeScheduleResource{}
	_ resource.ResourceWithImportState = &downtimeScheduleResource{}
)

// NewDowntimeScheduleResource is a helper function to simplify the provider implementation.
func NewDowntimeScheduleResource() resource.Resource {
	return &downtimeScheduleResource{}
}

// downtimeScheduleResource is the resource implementation.
type downtimeScheduleResource struct {
	client *legacyclient.Client
}

// downtimeScheduleResourceModel maps the resource schema data.
type downtimeScheduleResourceModel struct {
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

// Configure adds the provider configured client to the resource.
func (r *downtimeScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, "signoz_downtime_schedule",
		)

		return
	}

	legacyClient, err := legacyclient.NewClient(api)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build legacy SigNoz client", err.Error())
		return
	}

	r.client = legacyClient
}

// Metadata returns the resource type name.
func (r *downtimeScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "signoz_downtime_schedule"
}

// Schema defines the schema for the resource.
func (r *downtimeScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages downtime schedule resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.ScheduleName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the downtime schedule.",
			},
			attr.Description: schema.StringAttribute{
				Optional:    true,
				Description: "Description of the downtime schedule.",
			},
			attr.AlertIDs: schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of alert IDs to suppress during the downtime window.",
			},
			attr.AllAlerts: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to suppress all alerts during the downtime window. By default, it is false.",
				Default:     booldefault.StaticBool(false),
			},
			attr.Schedule: schema.SingleNestedAttribute{
				Required:    true,
				Description: "Schedule configuration for the downtime window.",
				Attributes: map[string]schema.Attribute{
					attr.StartTime: schema.StringAttribute{
						Required:    true,
						Description: "Start time of the downtime window in RFC3339 format.",
					},
					attr.EndTime: schema.StringAttribute{
						Required:    true,
						Description: "End time of the downtime window in RFC3339 format.",
					},
					attr.Timezone: schema.StringAttribute{
						Required:    true,
						Description: "Timezone of the downtime schedule (e.g., UTC, America/New_York).",
					},
					attr.Recurrence: schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Recurrence configuration for the downtime schedule.",
						Attributes: map[string]schema.Attribute{
							attr.RecurrenceType: schema.StringAttribute{
								Required:    true,
								Description: "Type of recurrence. Possible values are: once, daily, weekly.",
								Validators: []validator.String{
									stringvalidator.OneOf(model.RecurrenceTypes...),
								},
							},
							attr.DaysOfWeek: schema.ListAttribute{
								Optional:    true,
								ElementType: types.Int64Type,
								Description: "Days of the week for weekly recurrence (0=Sunday, 6=Saturday).",
							},
							attr.Duration: schema.StringAttribute{
								Optional:    true,
								Description: "Duration of the downtime window for recurring schedules (e.g., 2h, 30m).",
							},
						},
					},
				},
			},
			// computed.
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated unique ID for the downtime schedule.",
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

// Create creates the resource and sets the initial Terraform state.
func (r *downtimeScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan downtimeScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body.
	payload := &model.DowntimeSchedule{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		AllAlerts:   plan.AllAlerts.ValueBool(),
	}

	payload.SetAlertIDs(plan.AlertIDs)

	err := payload.SetSchedule(ctx, plan.Schedule)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, "signoz_downtime_schedule")
		return
	}

	tflog.Debug(ctx, "Creating downtime schedule", map[string]any{"downtimeSchedule": payload})

	// Create new downtime schedule.
	result, err := r.client.CreateDowntimeSchedule(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating downtime schedule",
			"Could not create downtime schedule, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created downtime schedule", map[string]any{"downtimeSchedule": result})

	// Map response to schema and populate Computed attributes.
	plan.ID = types.StringValue(result.ID)
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.AllAlerts = types.BoolValue(result.AllAlerts)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.CreatedBy = types.StringValue(result.CreatedBy)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)
	plan.UpdatedBy = types.StringValue(result.UpdatedBy)

	var diagAlertIDs diag.Diagnostics
	plan.AlertIDs, diagAlertIDs = result.AlertIDsToTerraform()
	resp.Diagnostics.Append(diagAlertIDs...)

	var diagSchedule diag.Diagnostics
	plan.Schedule, diagSchedule = result.ScheduleToTerraform(ctx)
	resp.Diagnostics.Append(diagSchedule...)

	// Set state to populated data.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *downtimeScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state downtimeScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading downtime schedule", map[string]any{"id": state.ID.ValueString()})

	// Get refreshed downtime schedule from SigNoz.
	result, err := r.client.GetDowntimeSchedule(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "Downtime schedule not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err, operationRead, "signoz_downtime_schedule")
		return
	}

	// Overwrite items with refreshed state.
	state.ID = types.StringValue(result.ID)
	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.AllAlerts = types.BoolValue(result.AllAlerts)
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.CreatedBy = types.StringValue(result.CreatedBy)
	state.UpdatedAt = types.StringValue(result.UpdatedAt)
	state.UpdatedBy = types.StringValue(result.UpdatedBy)

	var diagAlertIDs diag.Diagnostics
	state.AlertIDs, diagAlertIDs = result.AlertIDsToTerraform()
	resp.Diagnostics.Append(diagAlertIDs...)

	var diagSchedule diag.Diagnostics
	state.Schedule, diagSchedule = result.ScheduleToTerraform(ctx)
	resp.Diagnostics.Append(diagSchedule...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *downtimeScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan, state downtimeScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	payload := &model.DowntimeSchedule{
		ID:          state.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		AllAlerts:   plan.AllAlerts.ValueBool(),
	}

	payload.SetAlertIDs(plan.AlertIDs)

	err := payload.SetSchedule(ctx, plan.Schedule)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_downtime_schedule")
		return
	}

	tflog.Debug(ctx, "Updating downtime schedule", map[string]any{"downtimeSchedule": payload})

	// Update existing downtime schedule.
	err = r.client.UpdateDowntimeSchedule(ctx, state.ID.ValueString(), payload)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_downtime_schedule")
		return
	}

	// Fetch updated downtime schedule.
	result, err := r.client.GetDowntimeSchedule(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_downtime_schedule")
		return
	}

	// Preserve plan values — v0.104.0 API may omit fields with omitempty.
	plan.ID = types.StringValue(result.ID)
	if result.Name != "" {
		plan.Name = types.StringValue(result.Name)
	}
	// Description, AllAlerts: keep plan values (API may omit empty description or return false for bool)
	if result.CreatedAt != "" {
		plan.CreatedAt = types.StringValue(result.CreatedAt)
	} else {
		plan.CreatedAt = state.CreatedAt
	}
	if result.CreatedBy != "" {
		plan.CreatedBy = types.StringValue(result.CreatedBy)
	} else {
		plan.CreatedBy = state.CreatedBy
	}
	if result.UpdatedAt != "" {
		plan.UpdatedAt = types.StringValue(result.UpdatedAt)
	} else {
		plan.UpdatedAt = state.UpdatedAt
	}
	if result.UpdatedBy != "" {
		plan.UpdatedBy = types.StringValue(result.UpdatedBy)
	} else {
		plan.UpdatedBy = state.UpdatedBy
	}
	// AlertIDs and Schedule: keep plan values (API may omit empty lists)

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *downtimeScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state downtimeScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing downtime schedule.
	err := r.client.DeleteDowntimeSchedule(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, "signoz_downtime_schedule")
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *downtimeScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
