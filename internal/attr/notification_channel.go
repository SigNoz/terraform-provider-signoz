package attr

import (
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ChannelName = "name"
	ChannelType = "type"
	// Slack config.
	SlackConfigs = "slack_configs"
	ApiUrl       = "api_url"
	Channel      = "channel"
	SlackTitle   = "title"
	SlackText    = "text"
	// Webhook config.
	WebhookConfigs = "webhook_configs"
	Username       = "username"
	Password       = "password"
	// PagerDuty config.
	PagerdutyConfigs = "pagerduty_configs"
	RoutingKey       = "routing_key"
	ServiceKey       = "service_key"
	Details          = "details"
	// OpsGenie config.
	OpsgenieConfigs = "opsgenie_configs"
	ApiKey          = "api_key"
	// MS Teams config.
	MSTeamsConfigs = "msteams_configs"
	WebhookUrl     = "webhook_url"
	// Email config.
	EmailConfigs = "email_configs"
	SendResolved = "send_resolved"
	To           = "to"
	Html         = "html"
	Headers      = "headers"
)

func SlackConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		ApiUrl:     types.StringType,
		Channel:    types.StringType,
		SlackTitle: types.StringType,
		SlackText:  types.StringType,
	}
}

func WebhookConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		ApiUrl:   types.StringType,
		Username: types.StringType,
		Password: types.StringType,
	}
}

func PagerdutyConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		RoutingKey: types.StringType,
		ServiceKey: types.StringType,
		Details:    types.MapType{ElemType: types.StringType},
	}
}

func OpsgenieConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		ApiKey: types.StringType,
		ApiUrl: types.StringType,
	}
}

func MSTeamsConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		WebhookUrl: types.StringType,
	}
}

func EmailConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		SendResolved: types.BoolType,
		To:           types.StringType,
		Html:         types.StringType,
		Headers:      types.MapType{ElemType: types.StringType},
	}
}
