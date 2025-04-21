package resource

import (
	"context"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
)

// NewDashboardResource is a helper function to simplify the provider implementation.
func NewDashboardResource() resource.Resource {
	return &dashboardResource{}
}

// dashboardResource is the resource implementation.
type dashboardResource struct {
	client *client.Client
}

// dashboardResourceModel maps the resource schema data.
type dashboardResourceModel struct {
	Condition         types.String `tfsdk:"condition"`
	UUID              types.String `tfsdk:"uuid"`
}

// Configure adds the provider configured client to the resource.
func (r *dashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			// todo: check the below message, looks a bit off. should have been "unexpected resource configure type..."
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			SigNozDashboard,
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = SigNozDashboard
}

// Schema defines the schema for the resource.
func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages dashboard resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.Dashboard: schema.StringAttribute{
				Required: true,
				Description: "Name of the dashboard.",
			},
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body
	dashboardPayload := &model.Dashboard{
		Dashboard: plan.Dashboard.ValueString(),
	}

	err := dashboardPayload.SetCondition(plan.Condition)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate)
		return
	}

	tflog.Debug(ctx, "Creating dashboard", map[string]any{"dashboard": dashboardPayload})

	// Create new dashboard
	dashboard, err := r.client.CreateDashboard(ctx, dashboardPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating dashboard",
			"Could not create dashboard, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created dashboard", map[string]any{"dashboard": dashboard})

	// Map response to schema and populate Computed attributes
	//todo:

	// Set state to populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state dashboardResourceModel
	var diag diag.Diagnostics
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading dashboard", map[string]any{"alert": state.UUID.ValueString()})

	// Get refreshed dashboard from SigNoz
	dashboard, err := r.client.GetDashboard(ctx, state.UUID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead)
		return
	}

	// Overwrite items with refreshed state
	// todo:
	state.CreateAt = types.StringValue(dashboard.CreateAt)
	state.CreateBy = types.StringValue(dashboard.CreateBy)
	state.UpdateAt = types.StringValue(dashboard.UpdateAt)
	state.UpdateBy = types.StringValue(dashboard.UpdateBy)

	state.Condition, err = dashboard.ConditionToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead)
		return
	}

	state.Labels, diag = dashboard.LabelsToTerraform()
	resp.Diagnostics.Append(diag...)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan, state dashboardResourceModel
	var diag diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var err error
	dashboardUpdate := &model.Dashboard{
		UUID:        state.UUID.ValueString(),
		// todo:
	}

	err = dashboardUpdate.SetCondition(plan.Condition)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	// Update existing dashboard
	err = r.client.UpdateDashboard(ctx, state.UUID.ValueString(), dashboardUpdate)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	// Fetch updated dashboard
	dashboard, err := r.client.GetDashboard(ctx, state.UUID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	// Overwrite items with refreshed state
	plan.UUID = types.StringValue(dashboard.UUID)
	// todo:
	plan.CreateAt = types.StringValue(dashboard.CreateAt)
	plan.CreateBy = types.StringValue(dashboard.CreateBy)
	plan.UpdateAt = types.StringValue(dashboard.UpdateAt)
	plan.UpdateBy = types.StringValue(dashboard.UpdateBy)

	plan.Condition, err = dashboard.ConditionToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	plan.Labels, diag = dashboard.LabelsToTerraform()
	resp.Diagnostics.Append(diag...)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r* dashboardResource) Delete(ctx context.Context, req resource.DeleteResource, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing dashboard
	err := r.client.DeleteDashboard(ctx, state.UUID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete)
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}