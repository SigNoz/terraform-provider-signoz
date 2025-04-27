package resource

import (
	"context"
	"fmt"
	"regexp"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &alertResource{}
	_ resource.ResourceWithConfigure   = &alertResource{}
	_ resource.ResourceWithImportState = &alertResource{}
)

// NewAlertResource is a helper function to simplify the provider implementation.
func NewAlertResource() resource.Resource {
	return &alertResource{}
}

// alertResource is the resource implementation.
type alertResource struct {
	client *client.Client
}

// alertResourceModel maps the resource schema data.
type alertResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Alert             types.String `tfsdk:"alert"`
	AlertType         types.String `tfsdk:"alert_type"`
	BroadcastToAll    types.Bool   `tfsdk:"broadcast_to_all"`
	Condition         types.String `tfsdk:"condition"`
	Description       types.String `tfsdk:"description"`
	Disabled          types.Bool   `tfsdk:"disabled"`
	EvalWindow        types.String `tfsdk:"eval_window"`
	Frequency         types.String `tfsdk:"frequency"`
	Labels            types.Map    `tfsdk:"labels"`
	PreferredChannels types.List   `tfsdk:"preferred_channels"`
	RuleType          types.String `tfsdk:"rule_type"`
	Severity          types.String `tfsdk:"severity"`
	Source            types.String `tfsdk:"source"`
	State             types.String `tfsdk:"state"`
	Summary           types.String `tfsdk:"summary"`
	Version           types.String `tfsdk:"version"`
	CreateAt          types.String `tfsdk:"create_at"`
	CreateBy          types.String `tfsdk:"create_by"`
	UpdateAt          types.String `tfsdk:"update_at"`
	UpdateBy          types.String `tfsdk:"update_by"`
}

// Configure adds the provider configured client to the resource.
func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			SigNozAlert,
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *alertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = SigNozAlert
}

// Schema defines the schema for the resource.
func (r *alertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages alert resources in SigNoz.",
		Attributes: map[string]schema.Attribute{
			attr.Alert: schema.StringAttribute{
				Required:    true,
				Description: "Name of the alert.",
			},
			attr.AlertType: schema.StringAttribute{
				Required: true,
				Description: fmt.Sprintf("Type of the alert. Possible values are: %s, %s, %s, and %s.",
					model.AlertTypeMetrics, model.AlertTypeLogs, model.AlertTypeTraces, model.AlertTypeExceptions),
				Validators: []validator.String{
					stringvalidator.OneOf(model.AlertTypes...),
				},
			},
			attr.BroadcastToAll: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether to broadcast the alert to all the alerting channels. " +
					"By default, the alert is only sent to the preferred channels.",
			},
			attr.Condition: schema.StringAttribute{
				Required:    true,
				Description: "Condition of the alert.",
			},
			attr.Description: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the alert.",
				Default:     stringdefault.StaticString(alertDefaultDescription),
			},
			attr.Disabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the alert is disabled.",
				Default:     booldefault.StaticBool(false),
			},
			attr.EvalWindow: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The evaluation window of the alert. By default, it is 5m0s.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^([0-9]+h)?([0-9]+m)?([0-9]+s)?$`), "invalid alert evaluation window. It should be in format of 5m0s or 15m30s"),
				},
				Default: stringdefault.StaticString(alertDefaultEvalWindow),
			},
			attr.Frequency: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The frequency of the alert. By default, it is 1m0s.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^([0-9]+h)?([0-9]+m)?([0-9]+s)?$`), "invalid alert frequency. It should be in format of 1m0s or 10m30s"),
				},
				Default: stringdefault.StaticString(alertDefaultFrequency),
			},
			attr.Labels: schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Labels of the alert. Severity is a required label.",
			},
			attr.PreferredChannels: schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Preferred channels of the alert. By default, it is empty.",
			},
			attr.RuleType: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: fmt.Sprintf("Type of the alert. Possible values are: %s and %s.",
					model.AlertRuleTypeThreshold, model.AlertRuleTypeProm),
				Validators: []validator.String{
					stringvalidator.OneOf(model.AlertRuleTypes...),
				},
			},
			attr.Severity: schema.StringAttribute{
				Required: true,
				Description: fmt.Sprintf("Severity of the alert. Possible values are: %s, %s, %s, and %s.",
					model.AlertSeverityInfo, model.AlertSeverityWarning, model.AlertSeverityError, model.AlertSeverityCritical),
				Validators: []validator.String{
					stringvalidator.OneOf(model.AlertSeverities...),
				},
			},
			attr.Source: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Source of the alert. By default, it is <SIGNOZ_ENDPOINT>/alerts.",
			},
			attr.Summary: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Summary of the alert.",
				Default:     stringdefault.StaticString(alertDefaultSummary),
			},
			attr.Version: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Version of the alert. By default, it is v4.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`v\d+`), "alert version should be of the form v3, v4, etc."),
				},
				Default: stringdefault.StaticString(alertDefaultVersion),
			},
			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated unique ID for the alert.",
			},
			attr.State: schema.StringAttribute{
				Computed:    true,
				Description: "State of the alert.",
			},
			attr.CreateAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the alert.",
			},
			attr.CreateBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the alert.",
			},
			attr.UpdateAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the alert.",
			},
			attr.UpdateBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the alert.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body
	alertPayload := &model.Alert{
		Alert:     plan.Alert.ValueString(),
		AlertType: plan.AlertType.ValueString(),
		Annotations: model.AlertAnnotations{
			Description: plan.Description.ValueString(),
			Summary:     plan.Summary.ValueString(),
		},
		BroadcastToAll: plan.BroadcastToAll.ValueBool(),
		EvalWindow:     plan.EvalWindow.ValueString(),
		Frequency:      plan.Frequency.ValueString(),
		RuleType:       plan.RuleType.ValueString(),
		Source:         plan.Source.ValueString(),
		Version:        plan.Version.ValueString(),
	}

	err := alertPayload.SetCondition(plan.Condition)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate)
		return
	}

	alertPayload.SetLabels(plan.Labels, plan.Severity)
	alertPayload.SetPreferredChannels(plan.PreferredChannels)

	tflog.Debug(ctx, "Creating alert", map[string]any{"alert": alertPayload})

	// Create new alert
	alert, err := r.client.CreateAlert(ctx, alertPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating alert",
			"Could not create alert, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created alert", map[string]any{"alert": alert})

	// Map response to schema and populate Computed attributes
	plan.ID = types.StringValue(alert.ID)
	plan.Disabled = types.BoolValue(alert.Disabled)
	plan.Source = types.StringValue(alert.Source)
	plan.State = types.StringValue(alert.State)
	plan.CreateAt = types.StringValue(alert.CreateAt)
	plan.CreateBy = types.StringValue(alert.CreateBy)
	plan.UpdateAt = types.StringValue(alert.UpdateAt)
	plan.UpdateBy = types.StringValue(alert.UpdateBy)

	// Set state to populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state alertResourceModel
	var diag diag.Diagnostics
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading alert", map[string]any{"alert": state.ID.ValueString()})

	// Get refreshed alert from SigNoz
	alert, err := r.client.GetAlert(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead)
		return
	}

	// Overwrite items with refreshed state
	state.Alert = types.StringValue(alert.Alert)
	state.AlertType = types.StringValue(alert.AlertType)
	state.BroadcastToAll = types.BoolValue(alert.BroadcastToAll)
	state.Description = types.StringValue(alert.Annotations.Description)
	state.Disabled = types.BoolValue(alert.Disabled)
	state.EvalWindow = types.StringValue(alert.EvalWindow)
	state.Frequency = types.StringValue(alert.Frequency)
	state.RuleType = types.StringValue(alert.RuleType)
	state.Severity = types.StringValue(alert.Labels[attr.Severity])
	state.Source = types.StringValue(alert.Source)
	state.State = types.StringValue(alert.State)
	state.Summary = types.StringValue(alert.Annotations.Summary)
	state.Version = types.StringValue(alert.Version)
	state.CreateAt = types.StringValue(alert.CreateAt)
	state.CreateBy = types.StringValue(alert.CreateBy)
	state.UpdateAt = types.StringValue(alert.UpdateAt)
	state.UpdateBy = types.StringValue(alert.UpdateBy)

	state.Condition, err = alert.ConditionToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead)
		return
	}

	state.Labels, diag = alert.LabelsToTerraform()
	resp.Diagnostics.Append(diag...)

	state.PreferredChannels, diag = alert.PreferredChannelsToTerraform()
	resp.Diagnostics.Append(diag...)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan, state alertResourceModel
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
	alertUpdate := &model.Alert{
		ID:        state.ID.ValueString(),
		Alert:     plan.Alert.ValueString(),
		AlertType: plan.AlertType.ValueString(),
		Annotations: model.AlertAnnotations{
			Description: plan.Description.ValueString(),
			Summary:     plan.Summary.ValueString(),
		},
		BroadcastToAll: plan.BroadcastToAll.ValueBool(),
		Disabled:       plan.Disabled.ValueBool(),
		EvalWindow:     plan.EvalWindow.ValueString(),
		Frequency:      plan.Frequency.ValueString(),
		RuleType:       plan.RuleType.ValueString(),
		Source:         plan.Source.ValueString(),
		State:          state.State.ValueString(),
		Version:        plan.Version.ValueString(),
		CreateAt:       state.CreateAt.ValueString(),
		CreateBy:       state.CreateBy.ValueString(),
		UpdateAt:       state.UpdateAt.ValueString(),
		UpdateBy:       state.UpdateBy.ValueString(),
	}

	err = alertUpdate.SetCondition(plan.Condition)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	alertUpdate.SetLabels(plan.Labels, plan.Severity)
	alertUpdate.SetPreferredChannels(plan.PreferredChannels)

	// Update existing alert
	err = r.client.UpdateAlert(ctx, state.ID.ValueString(), alertUpdate)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	// Fetch updated alert
	alert, err := r.client.GetAlert(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	// Overwrite items with refreshed state
	plan.ID = types.StringValue(alert.ID)
	plan.Alert = types.StringValue(alert.Alert)
	plan.AlertType = types.StringValue(alert.AlertType)
	plan.BroadcastToAll = types.BoolValue(alert.BroadcastToAll)
	plan.Description = types.StringValue(alert.Annotations.Description)
	plan.Disabled = types.BoolValue(alert.Disabled)
	plan.EvalWindow = types.StringValue(alert.EvalWindow)
	plan.Frequency = types.StringValue(alert.Frequency)
	plan.RuleType = types.StringValue(alert.RuleType)
	plan.Severity = types.StringValue(alert.Labels[attr.Severity])
	plan.Source = types.StringValue(alert.Source)
	plan.State = types.StringValue(alert.State)
	plan.Summary = types.StringValue(alert.Annotations.Summary)
	plan.Version = types.StringValue(alert.Version)
	plan.CreateAt = types.StringValue(alert.CreateAt)
	plan.CreateBy = types.StringValue(alert.CreateBy)
	plan.UpdateAt = types.StringValue(alert.UpdateAt)
	plan.UpdateBy = types.StringValue(alert.UpdateBy)

	plan.Condition, err = alert.ConditionToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate)
		return
	}

	plan.Labels, diag = alert.LabelsToTerraform()
	resp.Diagnostics.Append(diag...)

	plan.PreferredChannels, diag = alert.PreferredChannelsToTerraform()
	resp.Diagnostics.Append(diag...)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing alert
	err := r.client.DeleteAlert(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete)
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
