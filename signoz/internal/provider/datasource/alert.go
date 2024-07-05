package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	signozattr "github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/client"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &alertDataSource{}
	// _ datasource.DataSourceWithConfigure = &alertDataSource{}
)

// // alertDataSourceModel maps the data source schema data.
// type alertDataSourceModel struct {
// 	Alert alertModel `tfsdk:"alert"`
// }

// NewAlertDataSource is a helper function to simplify the provider implementation.
func NewAlertDataSource() datasource.DataSource {
	return &alertDataSource{}
}

// alertDataSource is the data source implementation.
type alertDataSource struct {
	client *client.Client
}

// alertModel maps alert schema data.
type alertModel struct {
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
	Source            types.String `tfsdk:"source"`
	State             types.String `tfsdk:"state"`
	Summary           types.String `tfsdk:"summary"`
	Version           types.String `tfsdk:"version"`
}

// Metadata returns the data source type name.
func (d *alertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = SigNozAlert
}

// Schema defines the schema for the data source.
func (d *alertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an alert from Signoz using its ID. The ID can be found in the URL of the alert in the Signoz UI.",
		Attributes: map[string]schema.Attribute{
			signozattr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the alert. The ID can be found in the URL of the alert in the Signoz UI.",
			},
			signozattr.Alert: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the alert.",
			},
			signozattr.AlertType: schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("Type of the alert. Possible values are: %s, %s, %s, and %s.",
					model.AlertTypeMetrics, model.AlertTypeLogs, model.AlertTypeTraces, model.AlertTypeExceptions),
			},
			signozattr.BroadcastToAll: schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to broadcast the alert to all the alert channels.",
			},
			signozattr.Condition: schema.StringAttribute{
				Computed:    true,
				Description: "Condition of the alert.",
			},
			signozattr.Description: schema.StringAttribute{
				Computed:    true,
				Description: "Description of the alert.",
			},
			signozattr.Disabled: schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the alert is disabled.",
			},
			signozattr.EvalWindow: schema.StringAttribute{
				Computed:    true,
				Description: "Evaluation window of the alert.",
			},
			signozattr.Frequency: schema.StringAttribute{
				Computed:    true,
				Description: "Frequency of the alert.",
			},
			signozattr.Labels: schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Labels of the alert. Severity is a required label.",
			},
			signozattr.PreferredChannels: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of preferred channels of the alert. This is a noop if BroadcastToAll is true.",
			},
			signozattr.RuleType: schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("Type of the Alert Rule for threshold. Possible values are: %s and %s.",
					model.AlertRuleTypeThreshold, model.AlertRuleTypeProm),
			},
			signozattr.Source: schema.StringAttribute{
				Computed:    true,
				Description: "Source URL of the alert.",
			},
			signozattr.State: schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("State of the alert. Possible values are: %s, %s, %s, and %s.",
					model.AlertStateInactive, model.AlertStateFiring, model.AlertStatePending, model.AlertStateDisabled),
			},
			signozattr.Summary: schema.StringAttribute{
				Computed:    true,
				Description: "Summary of the alert.",
			},
			signozattr.Version: schema.StringAttribute{
				Computed:    true,
				Description: "Version of the alert.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *alertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data alertModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := d.client.GetAlert(ctx, data.ID.ValueString())
	if err != nil {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unable to read SigNoz alert: %s", err.Error()),
			SigNozAlert,
		)

		return
	}
	data.Condition, err = fetchCondition(alert.Condition)
	if err != nil {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unable to read SigNoz condition: %s", err.Error()),
			SigNozAlert,
		)

		return
	}

	data.Labels, diags = fetchLabels(alert.Labels)
	resp.Diagnostics.Append(diags...)
	data.PreferredChannels, diags = fetchPreferredChannels(alert.PreferredChannels)
	resp.Diagnostics.Append(diags...)

	data.ID = types.StringValue(alert.ID)
	data.Alert = types.StringValue(alert.Alert)
	data.AlertType = types.StringValue(alert.AlertType)
	data.BroadcastToAll = types.BoolValue(alert.BroadcastToAll)
	data.Description = types.StringValue(alert.Annotations.Description)
	data.Disabled = types.BoolValue(alert.Disabled)
	data.EvalWindow = types.StringValue(alert.EvalWindow)
	data.Frequency = types.StringValue(alert.Frequency)
	data.RuleType = types.StringValue(alert.RuleType)
	data.Source = types.StringValue(alert.Source)
	data.State = types.StringValue(alert.State)
	data.Summary = types.StringValue(alert.Annotations.Summary)
	data.Version = types.StringValue(alert.Version)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *alertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		addErr(
			&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *client.Client, got: %T. Please report this issue to the provider developers", req.ProviderData),
			SigNozAlert,
		)

		return
	}

	d.client = client
}
