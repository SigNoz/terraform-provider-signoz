package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &legacyRoutePolicyResource{}
	_ resource.ResourceWithConfigure   = &legacyRoutePolicyResource{}
	_ resource.ResourceWithImportState = &legacyRoutePolicyResource{}
)

// NewLegacyRoutePolicyResource is a helper function to simplify the provider implementation.
func NewLegacyRoutePolicyResource() resource.Resource {
	return &legacyRoutePolicyResource{}
}

// legacyRoutePolicyResource is the resource implementation.
type legacyRoutePolicyResource struct {
	client *legacyclient.Client
}

// legacyRoutePolicyResourceModel maps the resource schema data.
type legacyRoutePolicyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	RouteName        types.String `tfsdk:"route_name"`
	Description      types.String `tfsdk:"description"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Matchers         types.List   `tfsdk:"matchers"`
	ChannelIDs       types.List   `tfsdk:"channel_ids"`
	ContinueMatching types.Bool   `tfsdk:"continue_matching"`
	GroupWait        types.String `tfsdk:"group_wait"`
	GroupInterval    types.String `tfsdk:"group_interval"`
	RepeatInterval   types.String `tfsdk:"repeat_interval"`
	CreatedAt        types.String `tfsdk:"created_at"`
	CreatedBy        types.String `tfsdk:"created_by"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	UpdatedBy        types.String `tfsdk:"updated_by"`
}

// Configure adds the provider configured client to the resource.
func (r *legacyRoutePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, "signoz_route_policy",
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
func (r *legacyRoutePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "signoz_route_policy"
}

// Schema defines the schema for the resource.
func (r *legacyRoutePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages route policy resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.RouteName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the route policy.",
			},
			attr.Description: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the route policy.",
				Default:     stringdefault.StaticString(""),
			},
			attr.Enabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the route policy is enabled. By default, it is true.",
				Default:     booldefault.StaticBool(true),
			},
			attr.Matchers: schema.ListNestedAttribute{
				Required:    true,
				Description: "List of matchers for the route policy.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.MatchName: schema.StringAttribute{
							Required:    true,
							Description: "Name of the matcher field (e.g., severity, service).",
						},
						attr.MatchValue: schema.StringAttribute{
							Required:    true,
							Description: "Value to match against.",
						},
						attr.MatchType: schema.StringAttribute{
							Required:    true,
							Description: "Type of match to perform (e.g., equals, regex).",
						},
					},
				},
			},
			attr.ChannelIDs: schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of notification channel IDs to route alerts to.",
			},
			attr.ContinueMatching: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to continue matching subsequent route policies after this one matches. By default, it is false.",
				Default:     booldefault.StaticBool(false),
			},
			attr.GroupWait: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait before sending a notification about new alerts that are added to a group. By default, it is 30s.",
				Default:     stringdefault.StaticString("30s"),
			},
			attr.GroupInterval: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait before sending a notification about new alerts added to an already firing group. By default, it is 5m.",
				Default:     stringdefault.StaticString("5m"),
			},
			attr.RepeatInterval: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait before re-sending a notification. By default, it is 4h.",
				Default:     stringdefault.StaticString("4h"),
			},
			// computed.
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated unique ID for the route policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.CreatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the route policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.CreatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the route policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.UpdatedAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the route policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.UpdatedBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the route policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *legacyRoutePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan legacyRoutePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body.
	payload := &model.RoutePolicy{
		Name:           plan.RouteName.ValueString(),
		Description:    plan.Description.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
		Continue:       plan.ContinueMatching.ValueBool(),
		GroupWait:      plan.GroupWait.ValueString(),
		GroupInterval:  plan.GroupInterval.ValueString(),
		RepeatInterval: plan.RepeatInterval.ValueString(),
	}

	payload.SetMatchers(plan.Matchers)
	payload.SetChannelIDs(plan.ChannelIDs)

	tflog.Debug(ctx, "Creating route policy", map[string]any{"routePolicy": payload})

	// Create new route policy.
	routePolicy, err := r.client.CreateRoutePolicy(ctx, payload)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, "signoz_route_policy")
		return
	}

	tflog.Debug(ctx, "Created route policy", map[string]any{"routePolicy": routePolicy})

	// Preserve plan values — v0.104.0 API response doesn't include all fields
	// (matchers, group_wait, group_interval, repeat_interval, enabled, channel_ids).
	plan.ID = types.StringValue(routePolicy.ID)
	plan.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	plan.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	plan.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	plan.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)

	// Set state to populated data.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *legacyRoutePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state legacyRoutePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading route policy", map[string]any{"id": state.ID.ValueString()})

	// Get refreshed route policy from SigNoz.
	routePolicy, err := r.client.GetRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "Route policy not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err, operationRead, "signoz_route_policy")
		return
	}

	// Preserve state values when API returns empty — v0.104.0 omits many fields.
	state.ID = types.StringValue(routePolicy.ID)
	if routePolicy.Name != "" {
		state.RouteName = types.StringValue(routePolicy.Name)
	}
	if routePolicy.Description != "" {
		state.Description = types.StringValue(routePolicy.Description)
	}
	// Enabled, ContinueMatching: API returns false for both — keep state
	if routePolicy.GroupWait != "" {
		state.GroupWait = types.StringValue(routePolicy.GroupWait)
	}
	if routePolicy.GroupInterval != "" {
		state.GroupInterval = types.StringValue(routePolicy.GroupInterval)
	}
	if routePolicy.RepeatInterval != "" {
		state.RepeatInterval = types.StringValue(routePolicy.RepeatInterval)
	}
	if routePolicy.CreatedAt != "" {
		state.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	}
	if routePolicy.CreatedBy != "" {
		state.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	}
	if routePolicy.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	}
	if routePolicy.UpdatedBy != "" {
		state.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)
	}

	if len(routePolicy.Matchers) > 0 {
		var diagMatchers diag.Diagnostics
		state.Matchers, diagMatchers = routePolicy.MatchersToTerraform()
		resp.Diagnostics.Append(diagMatchers...)
	}

	if len(routePolicy.ChannelIDs) > 0 {
		var diagChannelIDs diag.Diagnostics
		state.ChannelIDs, diagChannelIDs = routePolicy.ChannelIDsToTerraform()
		resp.Diagnostics.Append(diagChannelIDs...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *legacyRoutePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan, state legacyRoutePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	payload := &model.RoutePolicy{
		ID:             state.ID.ValueString(),
		Name:           plan.RouteName.ValueString(),
		Description:    plan.Description.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
		Continue:       plan.ContinueMatching.ValueBool(),
		GroupWait:      plan.GroupWait.ValueString(),
		GroupInterval:  plan.GroupInterval.ValueString(),
		RepeatInterval: plan.RepeatInterval.ValueString(),
		CreatedAt:      state.CreatedAt.ValueString(),
		CreatedBy:      state.CreatedBy.ValueString(),
		UpdatedAt:      state.UpdatedAt.ValueString(),
		UpdatedBy:      state.UpdatedBy.ValueString(),
	}

	payload.SetMatchers(plan.Matchers)
	payload.SetChannelIDs(plan.ChannelIDs)

	tflog.Debug(ctx, "Updating route policy", map[string]any{"routePolicy": payload})

	// Update existing route policy.
	err := r.client.UpdateRoutePolicy(ctx, state.ID.ValueString(), payload)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_route_policy")
		return
	}

	// Fetch updated route policy.
	routePolicy, err := r.client.GetRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, "signoz_route_policy")
		return
	}

	// Preserve plan values — v0.104.0 API response doesn't reliably return
	// all fields (group_wait, group_interval, repeat_interval, enabled, continue_matching).
	plan.ID = types.StringValue(routePolicy.ID)
	if routePolicy.Name != "" {
		plan.RouteName = types.StringValue(routePolicy.Name)
	}
	if routePolicy.CreatedAt != "" {
		plan.CreatedAt = types.StringValue(routePolicy.CreatedAt)
	} else {
		plan.CreatedAt = state.CreatedAt
	}
	if routePolicy.CreatedBy != "" {
		plan.CreatedBy = types.StringValue(routePolicy.CreatedBy)
	} else {
		plan.CreatedBy = state.CreatedBy
	}
	if routePolicy.UpdatedAt != "" {
		plan.UpdatedAt = types.StringValue(routePolicy.UpdatedAt)
	} else {
		plan.UpdatedAt = state.UpdatedAt
	}
	if routePolicy.UpdatedBy != "" {
		plan.UpdatedBy = types.StringValue(routePolicy.UpdatedBy)
	} else {
		plan.UpdatedBy = state.UpdatedBy
	}
	// These fields are unreliably returned by the API — keep plan values.

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *legacyRoutePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state legacyRoutePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing route policy.
	err := r.client.DeleteRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, "signoz_route_policy")
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *legacyRoutePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
