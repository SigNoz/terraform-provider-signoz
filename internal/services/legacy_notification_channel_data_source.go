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
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &legacyNotificationChannelDataSource{}
)

// NewLegacyNotificationChannelDataSource is a helper function to simplify the provider implementation.
func NewLegacyNotificationChannelDataSource() datasource.DataSource {
	return &legacyNotificationChannelDataSource{}
}

// legacyNotificationChannelDataSource is the data source implementation.
type legacyNotificationChannelDataSource struct {
	client *legacyclient.Client
}

// notificationChannelModel maps notification channel schema data.
type notificationChannelModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	SlackConfigs     types.Object `tfsdk:"slack_configs"`
	WebhookConfigs   types.Object `tfsdk:"webhook_configs"`
	PagerdutyConfigs types.Object `tfsdk:"pagerduty_configs"`
	OpsgenieConfigs  types.Object `tfsdk:"opsgenie_configs"`
	MSTeamsConfigs   types.Object `tfsdk:"msteams_configs"`
	EmailConfigs     types.Object `tfsdk:"email_configs"`
}

// Configure adds the provider configured client to the data source.
func (d *legacyNotificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addReadErr(&resp.Diagnostics,
			fmt.Errorf("unexpected data source configure type. Expected *apiclients.WrappedClient, got: %T. "+
				"Please report this issue to the provider developers", req.ProviderData),
			"signoz_notification_channel",
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
func (d *legacyNotificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "signoz_notification_channel"
}

// Schema defines the schema for the data source.
func (d *legacyNotificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a notification channel from SigNoz using its ID.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "ID of the notification channel.",
			},
			attr.ChannelName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the notification channel.",
			},
			attr.ChannelType: schema.StringAttribute{
				Computed: true,
				Description: fmt.Sprintf("Type of the notification channel. Possible values are: %s, %s, %s, %s, %s, and %s.",
					model.ChannelTypeSlack, model.ChannelTypeWebhook, model.ChannelTypePagerduty,
					model.ChannelTypeOpsgenie, model.ChannelTypeMSTeams, model.ChannelTypeEmail),
			},
			attr.SlackConfigs: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Slack notification configuration.",
				Attributes: map[string]schema.Attribute{
					attr.ApiUrl: schema.StringAttribute{
						Computed:    true,
						Description: "Slack webhook URL.",
					},
					attr.Channel: schema.StringAttribute{
						Computed:    true,
						Description: "Slack channel to send notifications to.",
					},
					attr.SlackTitle: schema.StringAttribute{
						Computed:    true,
						Description: "Title of the Slack message.",
					},
					attr.SlackText: schema.StringAttribute{
						Computed:    true,
						Description: "Text of the Slack message.",
					},
				},
			},
			attr.WebhookConfigs: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Webhook notification configuration.",
				Attributes: map[string]schema.Attribute{
					attr.ApiUrl: schema.StringAttribute{
						Computed:    true,
						Description: "Webhook URL.",
					},
					attr.Username: schema.StringAttribute{
						Computed:    true,
						Description: "Username for webhook authentication.",
					},
					attr.Password: schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "Password for webhook authentication.",
					},
				},
			},
			attr.PagerdutyConfigs: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "PagerDuty notification configuration.",
				Attributes: map[string]schema.Attribute{
					attr.RoutingKey: schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "PagerDuty routing key.",
					},
					attr.ServiceKey: schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "PagerDuty service key.",
					},
					attr.Details: schema.MapAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Additional details for PagerDuty.",
					},
				},
			},
			attr.OpsgenieConfigs: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "OpsGenie notification configuration.",
				Attributes: map[string]schema.Attribute{
					attr.ApiKey: schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "OpsGenie API key.",
					},
					attr.ApiUrl: schema.StringAttribute{
						Computed:    true,
						Description: "OpsGenie API URL.",
					},
				},
			},
			attr.MSTeamsConfigs: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Microsoft Teams notification configuration.",
				Attributes: map[string]schema.Attribute{
					attr.WebhookUrl: schema.StringAttribute{
						Computed:    true,
						Description: "Microsoft Teams webhook URL.",
					},
				},
			},
			attr.EmailConfigs: schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Email notification configuration.",
				Attributes: map[string]schema.Attribute{
					attr.SendResolved: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to send a notification when the alert resolves.",
					},
					attr.To: schema.StringAttribute{
						Computed:    true,
						Description: "Email address to send notifications to.",
					},
					attr.Html: schema.StringAttribute{
						Computed:    true,
						Description: "HTML body of the email notification.",
					},
					attr.Headers: schema.MapAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Additional email headers.",
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *legacyNotificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data notificationChannelModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := d.client.GetChannel(ctx, data.ID.ValueString())
	if err != nil {
		addReadErr(&resp.Diagnostics, fmt.Errorf("unable to read SigNoz notification channel: %s", err.Error()), "signoz_notification_channel")
		return
	}

	// Set state values from retrieved data.
	data.ID = types.StringValue(channel.ID)
	data.Name = types.StringValue(channel.Name)
	data.Type = types.StringValue(channel.Type)

	setDataSourceChannelState(channel, &data)

	// Set state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// setDataSourceChannelState populates the datasource state from the API response.
func setDataSourceChannelState(channel *model.NotificationChannel, state *notificationChannelModel) {
	if len(channel.SlackConfigs) > 0 {
		cfg := channel.SlackConfigs[0]
		state.SlackConfigs, _ = types.ObjectValue(
			attr.SlackConfigAttrTypes(),
			map[string]tfattr.Value{
				attr.ApiUrl:     types.StringValue(cfg.ApiUrl),
				attr.Channel:    types.StringValue(cfg.Channel),
				attr.SlackTitle: types.StringValue(cfg.Title),
				attr.SlackText:  types.StringValue(cfg.Text),
			},
		)
	} else {
		state.SlackConfigs = types.ObjectNull(attr.SlackConfigAttrTypes())
	}

	if len(channel.WebhookConfigs) > 0 {
		cfg := channel.WebhookConfigs[0]
		state.WebhookConfigs, _ = types.ObjectValue(
			attr.WebhookConfigAttrTypes(),
			map[string]tfattr.Value{
				attr.ApiUrl:   types.StringValue(cfg.ApiUrl),
				attr.Username: types.StringValue(cfg.Username),
				attr.Password: types.StringValue(cfg.Password),
			},
		)
	} else {
		state.WebhookConfigs = types.ObjectNull(attr.WebhookConfigAttrTypes())
	}

	if len(channel.PagerdutyConfigs) > 0 {
		cfg := channel.PagerdutyConfigs[0]
		var detailsVal types.Map
		if len(cfg.Details) > 0 {
			m := make(map[string]tfattr.Value)
			for k, v := range cfg.Details {
				m[k] = types.StringValue(v)
			}
			detailsVal, _ = types.MapValue(types.StringType, m)
		} else {
			detailsVal = types.MapNull(types.StringType)
		}
		state.PagerdutyConfigs, _ = types.ObjectValue(
			attr.PagerdutyConfigAttrTypes(),
			map[string]tfattr.Value{
				attr.RoutingKey: types.StringValue(cfg.RoutingKey),
				attr.ServiceKey: types.StringValue(cfg.ServiceKey),
				attr.Details:    detailsVal,
			},
		)
	} else {
		state.PagerdutyConfigs = types.ObjectNull(attr.PagerdutyConfigAttrTypes())
	}

	if len(channel.OpsgenieConfigs) > 0 {
		cfg := channel.OpsgenieConfigs[0]
		state.OpsgenieConfigs, _ = types.ObjectValue(
			attr.OpsgenieConfigAttrTypes(),
			map[string]tfattr.Value{
				attr.ApiKey: types.StringValue(cfg.ApiKey),
				attr.ApiUrl: types.StringValue(cfg.ApiUrl),
			},
		)
	} else {
		state.OpsgenieConfigs = types.ObjectNull(attr.OpsgenieConfigAttrTypes())
	}

	if len(channel.MSTeamsConfigs) > 0 {
		cfg := channel.MSTeamsConfigs[0]
		state.MSTeamsConfigs, _ = types.ObjectValue(
			attr.MSTeamsConfigAttrTypes(),
			map[string]tfattr.Value{
				attr.WebhookUrl: types.StringValue(cfg.WebhookUrl),
			},
		)
	} else {
		state.MSTeamsConfigs = types.ObjectNull(attr.MSTeamsConfigAttrTypes())
	}

	if len(channel.EmailConfigs) > 0 {
		cfg := channel.EmailConfigs[0]
		var headersVal types.Map
		if len(cfg.Headers) > 0 {
			m := make(map[string]tfattr.Value)
			for k, v := range cfg.Headers {
				m[k] = types.StringValue(v)
			}
			headersVal, _ = types.MapValue(types.StringType, m)
		} else {
			headersVal = types.MapNull(types.StringType)
		}
		state.EmailConfigs, _ = types.ObjectValue(
			attr.EmailConfigAttrTypes(),
			map[string]tfattr.Value{
				attr.SendResolved: types.BoolValue(cfg.SendResolved),
				attr.To:           types.StringValue(cfg.To),
				attr.Html:         types.StringValue(cfg.Html),
				attr.Headers:      headersVal,
			},
		)
	} else {
		state.EmailConfigs = types.ObjectNull(attr.EmailConfigAttrTypes())
	}
}
