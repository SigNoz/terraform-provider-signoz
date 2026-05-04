// notification_channel is a textbook PreFlatten case from
// `internal/conv/contracts.go`: the GET response carries the channel
// config as a stringified JSON blob in `Data`, not as a typed nested
// object. The Layer 3 flatten here hides that quirk — service code
// passes `*apitypes.AlertmanagertypesChannel` straight in and gets back
// a `*NotificationChannelFlat` with `.ToResource()` / `.ToDataSource()`
// narrowing methods.
//
// Only `slack_configs` is supported in this provider phase. Everything
// else in the model is left as a typed-null list, mirroring what the
// hand-written code did before this refactor.
package conv

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	customtypes "github.com/SigNoz/terraform-provider-signoz/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// supportedNotificationChannels lists the channel-config arrays this
// provider implements end-to-end. Anything else in the model returns an
// error during expand. Phase 3 codegen will mass-generate the rest.
var supportedNotificationChannels = map[string]bool{
	"slack_configs": true,
}

// NotificationChannelFlat is the wide intermediate the codegen-driven
// `services` layer expects: superset of the resource model AND the
// datasource model (CreatedAt / UpdatedAt only live on the DS side).
// `Flatten<R>` returns this; the generated CRUD shells call `.ToResource()`
// or `.ToDataSource()` to project into the right schema.
type NotificationChannelFlat struct {
	Id    types.String
	Name  types.String
	OrgId types.String
	Type  types.String

	CreatedAt types.String
	UpdatedAt types.String

	DiscordConfigs    types.List
	EmailConfigs      types.List
	IncidentioConfigs types.List
	JiraConfigs       types.List
	MattermostConfigs types.List
	MsteamsConfigs    types.List
	Msteamsv2Configs  types.List
	OpsgenieConfigs   types.List
	PagerdutyConfigs  types.List
	PushoverConfigs   types.List
	RocketchatConfigs types.List
	SlackConfigs      types.List
	SnsConfigs        types.List
	TelegramConfigs   types.List
	VictoropsConfigs  types.List
	WebexConfigs      types.List
	WebhookConfigs    types.List
	WechatConfigs     types.List
}

// ToResource narrows the flat shape down to the resource model.
func (f *NotificationChannelFlat) ToResource() *schemas.NotificationChannelModel {
	return &schemas.NotificationChannelModel{
		Id:    f.Id,
		Name:  f.Name,
		OrgId: f.OrgId,
		Type:  f.Type,

		DiscordConfigs:    f.DiscordConfigs,
		EmailConfigs:      f.EmailConfigs,
		IncidentioConfigs: f.IncidentioConfigs,
		JiraConfigs:       f.JiraConfigs,
		MattermostConfigs: f.MattermostConfigs,
		MsteamsConfigs:    f.MsteamsConfigs,
		Msteamsv2Configs:  f.Msteamsv2Configs,
		OpsgenieConfigs:   f.OpsgenieConfigs,
		PagerdutyConfigs:  f.PagerdutyConfigs,
		PushoverConfigs:   f.PushoverConfigs,
		RocketchatConfigs: f.RocketchatConfigs,
		SlackConfigs:      f.SlackConfigs,
		SnsConfigs:        f.SnsConfigs,
		TelegramConfigs:   f.TelegramConfigs,
		VictoropsConfigs:  f.VictoropsConfigs,
		WebexConfigs:      f.WebexConfigs,
		WebhookConfigs:    f.WebhookConfigs,
		WechatConfigs:     f.WechatConfigs,
	}
}

// ToDataSource narrows the flat shape down to the datasource model.
// The DS exposes only the metadata fields (no per-channel-type config
// payload) — the channel-config arrays are dropped here on purpose.
func (f *NotificationChannelFlat) ToDataSource() *schemas.NotificationChannelDataSourceModel {
	return &schemas.NotificationChannelDataSourceModel{
		CreatedAt: f.CreatedAt,
		Id:        f.Id,
		Name:      f.Name,
		OrgId:     f.OrgId,
		Type:      f.Type,
		UpdatedAt: f.UpdatedAt,
	}
}

// ExpandAlertmanagertypesPostableChannel converts the framework resource model into the
// POST body. The wire shape uses `apitypes.AlertmanagertypesPostableChannel`
// (Name required); same shape as `ConfigReceiver` modulo Name's pointerness.
func ExpandAlertmanagertypesPostableChannel(ctx context.Context, m schemas.NotificationChannelModel) (*apitypes.AlertmanagertypesPostableChannel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err := rejectUnsupportedNotificationChannels(m); err != nil {
		diags.AddError("Unsupported channel type", err.Error())
		return nil, diags
	}

	out := &apitypes.AlertmanagertypesPostableChannel{Name: m.Name.ValueString()}

	slacks, d := ExpandConfigSlackConfigList(ctx, m.SlackConfigs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	out.SlackConfigs = slacks
	return out, diags
}

// ExpandConfigReceiver is the PUT-shape variant. ConfigReceiver and
// AlertmanagertypesPostableChannel differ only in Name (pointer here).
func ExpandConfigReceiver(ctx context.Context, m schemas.NotificationChannelModel) (*apitypes.ConfigReceiver, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err := rejectUnsupportedNotificationChannels(m); err != nil {
		diags.AddError("Unsupported channel type", err.Error())
		return nil, diags
	}

	name := m.Name.ValueString()
	out := &apitypes.ConfigReceiver{Name: &name}

	slacks, d := ExpandConfigSlackConfigList(ctx, m.SlackConfigs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	out.SlackConfigs = slacks
	return out, diags
}

// FlattenAlertmanagertypesChannel converts the GET response into the
// wide flat. The `Data` field is a stringified JSON receiver — we
// PreFlatten it into a typed `ConfigReceiver` here, then dispatch on
// the outer `Type` field to populate the relevant channel-config list.
func FlattenAlertmanagertypesChannel(ctx context.Context, g *apitypes.AlertmanagertypesChannel) (*NotificationChannelFlat, diag.Diagnostics) {
	var diags diag.Diagnostics
	if g == nil {
		return nil, diags
	}

	out := &NotificationChannelFlat{
		Id:        types.StringValue(g.Id),
		Name:      types.StringValue(g.Name),
		OrgId:     types.StringValue(g.OrgId),
		Type:      types.StringValue(g.Type),
		CreatedAt: TimeStringFromPointer(g.CreatedAt),
		UpdatedAt: TimeStringFromPointer(g.UpdatedAt),
	}
	setNullChannelListsOnFlat(ctx, out)

	if g.Data == "" {
		return out, diags
	}

	var rec apitypes.ConfigReceiver
	if err := json.Unmarshal([]byte(g.Data), &rec); err != nil {
		diags.AddError("Decode notification_channel data",
			fmt.Sprintf("data field is not valid receiver JSON: %v", err))
		return nil, diags
	}
	switch g.Type {
	case "slack":
		slackList, d := FlattenConfigSlackConfigList(ctx, rec.SlackConfigs)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out.SlackConfigs = slackList
	default:
		diags.AddWarning("Channel type not yet flattened",
			fmt.Sprintf("notification_channel %q has type %q which is not yet flattened by terraform-provider-signoz; only slack is implemented in this phase", g.Id, g.Type))
	}
	return out, diags
}

// rejectUnsupportedNotificationChannels iterates the 17 not-yet-implemented
// channel arrays on the model. If any is non-null and non-unknown, returns
// an error pointing at that field.
func rejectUnsupportedNotificationChannels(m schemas.NotificationChannelModel) error {
	checks := map[string]types.List{
		"discord_configs":    m.DiscordConfigs,
		"email_configs":      m.EmailConfigs,
		"incidentio_configs": m.IncidentioConfigs,
		"jira_configs":       m.JiraConfigs,
		"mattermost_configs": m.MattermostConfigs,
		"msteams_configs":    m.MsteamsConfigs,
		"msteamsv2_configs":  m.Msteamsv2Configs,
		"opsgenie_configs":   m.OpsgenieConfigs,
		"pagerduty_configs":  m.PagerdutyConfigs,
		"pushover_configs":   m.PushoverConfigs,
		"rocketchat_configs": m.RocketchatConfigs,
		"sns_configs":        m.SnsConfigs,
		"telegram_configs":   m.TelegramConfigs,
		"victorops_configs":  m.VictoropsConfigs,
		"webex_configs":      m.WebexConfigs,
		"webhook_configs":    m.WebhookConfigs,
		"wechat_configs":     m.WechatConfigs,
	}
	for name, l := range checks {
		if !l.IsNull() && !l.IsUnknown() && len(l.Elements()) > 0 {
			return fmt.Errorf("%s is not yet supported by terraform-provider-signoz; only slack_configs is implemented in this phase", name)
		}
	}
	return nil
}

// setNullChannelListsOnFlat assigns a typed null List to every
// channel-config attribute on the flat — required so the framework
// doesn't see a value-of-the-wrong-type when the resource branch later
// serialises state via `ToResource()`.
func setNullChannelListsOnFlat(ctx context.Context, f *NotificationChannelFlat) {
	objType := func(at map[string]attr.Type) types.ObjectType {
		return types.ObjectType{AttrTypes: at}
	}
	f.DiscordConfigs = types.ListNull(objType(customtypes.ConfigDiscordConfigValue{}.AttributeTypes(ctx)))
	f.EmailConfigs = types.ListNull(objType(customtypes.ConfigEmailConfigValue{}.AttributeTypes(ctx)))
	f.IncidentioConfigs = types.ListNull(objType(customtypes.ConfigIncidentioConfigValue{}.AttributeTypes(ctx)))
	f.JiraConfigs = types.ListNull(objType(customtypes.ConfigJiraConfigValue{}.AttributeTypes(ctx)))
	f.MattermostConfigs = types.ListNull(objType(customtypes.ConfigMattermostConfigValue{}.AttributeTypes(ctx)))
	f.MsteamsConfigs = types.ListNull(objType(customtypes.ConfigMsteamsConfigValue{}.AttributeTypes(ctx)))
	f.Msteamsv2Configs = types.ListNull(objType(customtypes.ConfigMsteamsV2ConfigValue{}.AttributeTypes(ctx)))
	f.OpsgenieConfigs = types.ListNull(objType(customtypes.ConfigOpsGenieConfigValue{}.AttributeTypes(ctx)))
	f.PagerdutyConfigs = types.ListNull(objType(customtypes.ConfigPagerdutyConfigValue{}.AttributeTypes(ctx)))
	f.PushoverConfigs = types.ListNull(objType(customtypes.ConfigPushoverConfigValue{}.AttributeTypes(ctx)))
	f.RocketchatConfigs = types.ListNull(objType(customtypes.ConfigRocketchatConfigValue{}.AttributeTypes(ctx)))
	f.SlackConfigs = types.ListNull(objType(customtypes.ConfigSlackConfigValue{}.AttributeTypes(ctx)))
	f.SnsConfigs = types.ListNull(objType(customtypes.ConfigSnsconfigValue{}.AttributeTypes(ctx)))
	f.TelegramConfigs = types.ListNull(objType(customtypes.ConfigTelegramConfigValue{}.AttributeTypes(ctx)))
	f.VictoropsConfigs = types.ListNull(objType(customtypes.ConfigVictorOpsConfigValue{}.AttributeTypes(ctx)))
	f.WebexConfigs = types.ListNull(objType(customtypes.ConfigWebexConfigValue{}.AttributeTypes(ctx)))
	f.WebhookConfigs = types.ListNull(objType(customtypes.ConfigWebhookConfigValue{}.AttributeTypes(ctx)))
	f.WechatConfigs = types.ListNull(objType(customtypes.ConfigWechatConfigValue{}.AttributeTypes(ctx)))
}
