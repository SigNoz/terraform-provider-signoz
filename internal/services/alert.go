package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	client *legacyclient.Client
}

// alertResourceModel maps the resource schema data.
type alertResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Alert                types.String `tfsdk:"alert"`
	AlertType            types.String `tfsdk:"alert_type"`
	BroadcastToAll       types.Bool   `tfsdk:"broadcast_to_all"`
	Condition            types.String `tfsdk:"condition"`
	Description          types.String `tfsdk:"description"`
	Disabled             types.Bool   `tfsdk:"disabled"`
	EvalWindow           types.String `tfsdk:"eval_window"`
	Frequency            types.String `tfsdk:"frequency"`
	Labels               types.Map    `tfsdk:"labels"`
	PreferredChannels    types.List   `tfsdk:"preferred_channels"`
	RuleType             types.String `tfsdk:"rule_type"`
	Severity             types.String `tfsdk:"severity"`
	Source               types.String `tfsdk:"source"`
	State                types.String `tfsdk:"state"`
	Summary              types.String `tfsdk:"summary"`
	Version              types.String `tfsdk:"version"`
	SchemaVersion        types.String `tfsdk:"schema_version"`
	NotificationSettings types.Object `tfsdk:"notification_settings"`
	Evaluation           types.String `tfsdk:"evaluation"`
	CreateAt             types.String `tfsdk:"create_at"`
	CreateBy             types.String `tfsdk:"create_by"`
	UpdateAt             types.String `tfsdk:"update_at"`
	UpdateBy             types.String `tfsdk:"update_by"`
}

// Configure adds the provider configured client to the resource.
func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			operationConfigure, SigNozAlert,
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
				DeprecationMessage: "This field is no longer needed and will be ignored",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
				Description: "Labels of the alert. Do not set a `severity` key here: the provider writes the top-level `severity` attribute into the alert's labels (overwriting any `severity` key given here) and reads it back from there.",
			},
			attr.PreferredChannels: schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Preferred channels of the alert. By default, it is empty. These are ignored when `broadcast_to_all` is true.",
			},
			attr.RuleType: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: fmt.Sprintf("Type of the alert. Possible values are: %s and %s.",
					model.AlertRuleTypeThreshold, model.AlertRuleTypeProm),
				Validators: []validator.String{
					stringvalidator.OneOf(model.AlertRuleTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Description: "Version of the alert. The server only accepts v5, which is the default.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`v\d+`), "alert version should be of the form v3, v4, etc."),
				},
				Default: stringdefault.StaticString(alertDefaultVersion),
			},
			attr.SchemaVersion: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Schema version of the alert. By default, it is v1. For v2+ schemas, additional fields like evaluation and notification_settings are supported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.NotificationSettings: schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Notification settings for the alert. Only used when schema_version is v2 or higher.",
				Attributes: map[string]schema.Attribute{
					attr.Renotify: schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The alert re-notification settings.",
						Attributes: map[string]schema.Attribute{
							attr.Interval: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(regexp.MustCompile(`^([0-9]+h)?([0-9]+m)?([0-9]+s)?$`), "invalid alert evaluation window. It should be in format of 5m0s or 15m30s"),
								},
								Description: "Re-notify interval of the alert.",
							},
							attr.AlertStates: schema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
								Validators: []validator.List{
									listvalidator.ValueStringsAre([]validator.String{
										stringvalidator.OneOf(model.AlertStates...),
									}...),
								},
								Description: "For which AlertStates of the alert want to get renotified.",
							},
							attr.Enabled: schema.BoolAttribute{
								Computed:    true,
								Optional:    true,
								Default:     booldefault.StaticBool(true),
								Description: "To enable the re-notification of the alert.",
							},
						},
					},
					attr.GroupBy: schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "Group by the alert. By default, it is empty, use '__all__' to get different notification for each unique parameters.",
					},
					attr.UsePolicy: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Use policy in the alert. By default, it is false.",
					},
				},
			},
			attr.Evaluation: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Evaluation settings for the alert (JSON). Only used when schema_version is v2 or higher.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// computed.
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated unique ID for the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.State: schema.StringAttribute{
				Computed:    true,
				Description: "State of the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.CreateAt: schema.StringAttribute{
				Computed:    true,
				Description: "Creation time of the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.CreateBy: schema.StringAttribute{
				Computed:    true,
				Description: "Creator of the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.UpdateAt: schema.StringAttribute{
				Computed:    true,
				Description: "Last update time of the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.UpdateBy: schema.StringAttribute{
				Computed:    true,
				Description: "Last updater of the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body.
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
		SchemaVersion:  plan.SchemaVersion.ValueString(),
	}

	err := alertPayload.SetCondition(plan.Condition)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozAlert)
		return
	}

	if !utils.IsNullOrUnknown(plan.NotificationSettings) {
		err := alertPayload.SetNotificationSettings(ctx, plan.NotificationSettings)
		if err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, SigNozAlert)
			return
		}
	}

	if !plan.Evaluation.IsNull() && plan.Evaluation.ValueString() != "" {
		err := alertPayload.SetEvaluation(plan.Evaluation)
		if err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, SigNozAlert)
			return
		}
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

	// Map response to schema and populate Computed attributes.
	plan.ID = types.StringValue(alert.ID)
	plan.BroadcastToAll = types.BoolValue(alert.BroadcastToAll)
	plan.Disabled = types.BoolValue(alert.Disabled)
	plan.Source = types.StringValue(alert.Source)
	plan.State = types.StringValue(alert.State)
	plan.SchemaVersion = types.StringValue(alert.SchemaVersion)
	plan.CreateAt = types.StringValue(alert.CreateAt)
	plan.CreateBy = types.StringValue(alert.CreateBy)
	plan.UpdateAt = types.StringValue(alert.UpdateAt)
	plan.UpdateBy = types.StringValue(alert.UpdateBy)

	// rule_type is Optional+Computed without a Default: with it unset the plan
	// value stays unknown after apply unless Create fills it from the response
	// ("Provider returned invalid result object"). The other server-supplied
	// fields (eval_window, frequency, version) have static plan-time defaults,
	// so overwriting them here would fight the planned values.
	if alert.RuleType != "" {
		plan.RuleType = types.StringValue(alert.RuleType)
	}

	//As condition is JSON string, updated response contains extra keys
	plan.Condition, err = alertPayload.ConditionToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, SigNozAlert)
		return
	}

	var diagLabels diag.Diagnostics
	plan.Labels, diagLabels = alert.LabelsToTerraform()
	resp.Diagnostics.Append(diagLabels...)

	if alert.SchemaVersion != "" && alert.SchemaVersion != "v1" {
		plan.NotificationSettings, diagLabels = alert.NotificationSettingsToTerraform(ctx)
		resp.Diagnostics.Append(diagLabels...)

		var evalErr error
		plan.Evaluation, evalErr = alert.EvaluationToTerraform()
		if evalErr != nil {
			addErr(&resp.Diagnostics, evalErr, operationCreate, SigNozAlert)
		}
	} else {
		if utils.IsNullOrUnknown(plan.NotificationSettings) {
			plan.NotificationSettings = types.ObjectNull(
				attr.NotificationSettingsAttrTypes())
		}
		plan.Evaluation = types.StringNull()
	}

	// Set state to populated data.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading alert", map[string]any{"alert": state.ID.ValueString()})

	// Get refreshed alert from SigNoz.
	alert, err := r.client.GetAlert(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "Alert not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err, operationRead, SigNozAlert)
		return
	}

	// Overwrite items with refreshed state.
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
	state.SchemaVersion = types.StringValue(alert.SchemaVersion)

	// Preserve state condition — API response includes extra default keys
	// (IsAnomaly, isColumn, isJSON, ShiftBy, etc.) that cause perpetual diffs.
	// Only update on import (when state condition is empty/null).
	if state.Condition.IsNull() || state.Condition.ValueString() == "" {
		state.Condition, err = alert.ConditionToTerraform()
		if err != nil {
			addErr(&resp.Diagnostics, err, operationRead, SigNozAlert)
			return
		}
	}

	var diagLabelsRead diag.Diagnostics
	state.Labels, diagLabelsRead = alert.LabelsToTerraform()
	resp.Diagnostics.Append(diagLabelsRead...)

	var diagPreferredChannels diag.Diagnostics
	state.PreferredChannels, diagPreferredChannels = alert.PreferredChannelsToTerraform()
	resp.Diagnostics.Append(diagPreferredChannels...)

	if alert.SchemaVersion != "" && alert.SchemaVersion != "v1" {
		var diagNotifSettings diag.Diagnostics
		state.NotificationSettings, diagNotifSettings = alert.NotificationSettingsToTerraform(ctx)
		resp.Diagnostics.Append(diagNotifSettings...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Evaluation, err = alert.EvaluationToTerraform()
		if err != nil {
			addErr(&resp.Diagnostics, err, operationRead, SigNozAlert)
			return
		}
	} else {
		if utils.IsNullOrUnknown(state.NotificationSettings) {
			state.NotificationSettings = types.ObjectNull(
				attr.NotificationSettingsAttrTypes())
		}
		state.Evaluation = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan, state alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
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
		SchemaVersion:  plan.SchemaVersion.ValueString(),
		CreateAt:       state.CreateAt.ValueString(),
		CreateBy:       state.CreateBy.ValueString(),
		UpdateAt:       state.UpdateAt.ValueString(),
		UpdateBy:       state.UpdateBy.ValueString(),
	}

	err = alertUpdate.SetCondition(plan.Condition)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozAlert)
		return
	}

	tflog.Debug(ctx, "alert update", map[string]interface{}{"alertUpdate": alertUpdate})

	if !utils.IsNullOrUnknown(plan.NotificationSettings) {
		err := alertUpdate.SetNotificationSettings(ctx, plan.NotificationSettings)
		if err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, SigNozAlert)
			return
		}
	}

	if !plan.Evaluation.IsNull() && plan.Evaluation.ValueString() != "" {
		err := alertUpdate.SetEvaluation(plan.Evaluation)
		if err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, SigNozAlert)
			return
		}
	}

	alertUpdate.SetLabels(plan.Labels, plan.Severity)
	alertUpdate.SetPreferredChannels(plan.PreferredChannels)

	// Update existing alert.
	err = r.client.UpdateAlert(ctx, state.ID.ValueString(), alertUpdate)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozAlert)
		return
	}

	// Fetch updated alert.
	alert, err := r.client.GetAlert(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozAlert)
		return
	}

	// Preserve plan values — v0.104.0 API response may omit fields with
	// omitempty or return zero-values for bools/strings.
	plan.ID = types.StringValue(alert.ID)
	if alert.Alert != "" {
		plan.Alert = types.StringValue(alert.Alert)
	}
	if alert.AlertType != "" {
		plan.AlertType = types.StringValue(alert.AlertType)
	}
	// BroadcastToAll: always set to a concrete value (API omits with omitempty)
	plan.BroadcastToAll = types.BoolValue(alert.BroadcastToAll)
	// Disabled: keep plan value — API returns false for both disabled and enabled alerts
	if alert.Annotations.Description != "" {
		plan.Description = types.StringValue(alert.Annotations.Description)
	}
	if alert.EvalWindow != "" {
		plan.EvalWindow = types.StringValue(alert.EvalWindow)
	}
	if alert.Frequency != "" {
		plan.Frequency = types.StringValue(alert.Frequency)
	}
	if alert.RuleType != "" {
		plan.RuleType = types.StringValue(alert.RuleType)
	} else if plan.RuleType.IsUnknown() || plan.RuleType.IsNull() {
		plan.RuleType = types.StringValue("")
	}
	if alert.Labels[attr.Severity] != "" {
		plan.Severity = types.StringValue(alert.Labels[attr.Severity])
	}
	// Source: Optional+Computed without Default — must resolve to concrete
	if alert.Source != "" {
		plan.Source = types.StringValue(alert.Source)
	} else if plan.Source.IsUnknown() || plan.Source.IsNull() {
		plan.Source = types.StringValue("")
	}
	// State: Computed-only — must always resolve
	if alert.State != "" {
		plan.State = types.StringValue(alert.State)
	} else if plan.State.IsUnknown() || plan.State.IsNull() {
		plan.State = types.StringValue("inactive")
	}
	if alert.Annotations.Summary != "" {
		plan.Summary = types.StringValue(alert.Annotations.Summary)
	}
	if alert.Version != "" {
		plan.Version = types.StringValue(alert.Version)
	}
	// SchemaVersion: must resolve to concrete value (Computed without Default)
	if alert.SchemaVersion != "" {
		plan.SchemaVersion = types.StringValue(alert.SchemaVersion)
	} else if plan.SchemaVersion.IsUnknown() || plan.SchemaVersion.IsNull() {
		plan.SchemaVersion = types.StringValue("")
	}
	// Server timestamps churn on every write; the plan pins the pre-update
	// values (UseStateForUnknown), so writing the bumped ones here would fail
	// the apply with an inconsistent-result error. The next plan's Read
	// refresh converges them.
	plan.CreateAt = state.CreateAt
	plan.CreateBy = state.CreateBy
	plan.UpdateAt = state.UpdateAt
	plan.UpdateBy = state.UpdateBy

	// Condition: use the payload to avoid extra keys from API response
	plan.Condition, err = alertUpdate.ConditionToTerraform()
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, SigNozAlert)
		return
	}

	// Labels/channels: use API values if present, otherwise resolve unknowns
	if len(alert.Labels) > 0 {
		var diagLabelsUpdate diag.Diagnostics
		plan.Labels, diagLabelsUpdate = alert.LabelsToTerraform()
		resp.Diagnostics.Append(diagLabelsUpdate...)
	} else if plan.Labels.IsUnknown() {
		plan.Labels = types.MapNull(types.StringType)
	}

	if len(alert.PreferredChannels) > 0 {
		var diagPreferredChannelsUpdate diag.Diagnostics
		plan.PreferredChannels, diagPreferredChannelsUpdate = alert.PreferredChannelsToTerraform()
		resp.Diagnostics.Append(diagPreferredChannelsUpdate...)
	} else if plan.PreferredChannels.IsUnknown() {
		plan.PreferredChannels = types.ListNull(types.StringType)
	}

	// Notification settings / evaluation: only update from API if schema version indicates v2
	schemaVer := alert.SchemaVersion
	if schemaVer == "" {
		schemaVer = plan.SchemaVersion.ValueString()
	}
	if schemaVer != "" && schemaVer != "v1" {
		if len(alert.NotificationSettings.GroupBy) > 0 || alert.NotificationSettings.Renotify.ReNotifyInterval != "" {
			var diagNotif diag.Diagnostics
			plan.NotificationSettings, diagNotif = alert.NotificationSettingsToTerraform(ctx)
			resp.Diagnostics.Append(diagNotif...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else if utils.IsNullOrUnknown(plan.NotificationSettings) {
			plan.NotificationSettings = types.ObjectNull(
				attr.NotificationSettingsAttrTypes())
		}

		if alert.Evaluation != nil {
			plan.Evaluation, err = alert.EvaluationToTerraform()
			if err != nil {
				addErr(&resp.Diagnostics, err, operationUpdate, SigNozAlert)
				return
			}
		} else if plan.Evaluation.IsUnknown() {
			plan.Evaluation = types.StringNull()
		}
	} else {
		if utils.IsNullOrUnknown(plan.NotificationSettings) {
			plan.NotificationSettings = types.ObjectNull(
				attr.NotificationSettingsAttrTypes())
		}
		plan.Evaluation = types.StringNull()
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing alert.
	err := r.client.DeleteAlert(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, SigNozAlert)
		return
	}
}

// ImportState imports Terraform state into the resource.
func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
