// HAND-WRITTEN: re-introduced after `skaff convertors` cascade-skipped this
// schema. ConfigSlackConfig has an `*ConfigHTTPClientConfig` field whose
// transitive descendants (ConfigOAuth2.Claims is `*map[string]interface{}`)
// the generator can't model without a FieldExpander/PreExpander hook.
// Until those hooks land, we hand-write this file and reject `http_config`
// at expand-time. Actions and Fields delegate to the generated `*List`
// helpers, so the surface stays consistent with the rest of the layer.
package conv

import (
	"context"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	customtypes "github.com/SigNoz/terraform-provider-signoz/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExpandConfigSlackConfig(ctx context.Context, v customtypes.ConfigSlackConfigValue) (*apitypes.ConfigSlackConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v.IsNull() || v.IsUnknown() {
		return nil, diags
	}
	if !v.HttpConfig.IsNull() && !v.HttpConfig.IsUnknown() {
		diags.AddError("Unsupported slack field",
			"slack_configs[].http_config is not yet supported in this provider phase (cascading skip on ConfigOAuth2.Claims map[string]interface{})")
		return nil, diags
	}

	actions, d := ExpandConfigSlackActionList(ctx, v.Actions)
	diags.Append(d...)
	fields, d := ExpandConfigSlackFieldList(ctx, v.Fields)
	diags.Append(d...)
	mrkdwn, d := StringPointerSliceFromList(ctx, v.MrkdwnIn)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &apitypes.ConfigSlackConfig{
		Actions:      actions,
		ApiUrl:       StringPointer(v.ApiUrl),
		ApiUrlFile:   StringPointer(v.ApiUrlFile),
		AppToken:     StringPointer(v.AppToken),
		AppTokenFile: StringPointer(v.AppTokenFile),
		AppUrl:       StringPointer(v.AppUrl),
		CallbackId:   StringPointer(v.CallbackId),
		Channel:      StringPointer(v.Channel),
		Color:        StringPointer(v.Color),
		Fallback:     StringPointer(v.Fallback),
		Fields:       fields,
		Footer:       StringPointer(v.Footer),
		IconEmoji:    StringPointer(v.IconEmoji),
		IconUrl:      StringPointer(v.IconUrl),
		ImageUrl:     StringPointer(v.ImageUrl),
		LinkNames:    BoolPointer(v.LinkNames),
		MessageText:  StringPointer(v.MessageText),
		MrkdwnIn:     mrkdwn,
		Pretext:      StringPointer(v.Pretext),
		SendResolved: BoolPointer(v.SendResolved),
		ShortFields:  BoolPointer(v.ShortFields),
		Text:         StringPointer(v.Text),
		ThumbUrl:     StringPointer(v.ThumbUrl),
		Timeout:      Int64Pointer(v.Timeout),
		Title:        StringPointer(v.Title),
		TitleLink:    StringPointer(v.TitleLink),
		Username:     StringPointer(v.Username),
	}, diags
}

func FlattenConfigSlackConfig(ctx context.Context, in *apitypes.ConfigSlackConfig) (customtypes.ConfigSlackConfigValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if in == nil {
		return customtypes.NewConfigSlackConfigValueNull(), diags
	}

	actions, d := FlattenConfigSlackActionList(ctx, in.Actions)
	diags.Append(d...)
	fields, d := FlattenConfigSlackFieldList(ctx, in.Fields)
	diags.Append(d...)
	mrkdwn, d := ListFromStringPointerSlice(ctx, in.MrkdwnIn)
	diags.Append(d...)
	if diags.HasError() {
		return customtypes.NewConfigSlackConfigValueUnknown(), diags
	}

	httpType := customtypes.ConfigHttpclientConfigValue{}.AttributeTypes(ctx)
	rv, d := customtypes.NewConfigSlackConfigValue(
		customtypes.ConfigSlackConfigValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"actions":        actions,
			"api_url":        StringFromPointer(in.ApiUrl),
			"api_url_file":   StringFromPointer(in.ApiUrlFile),
			"app_token":      StringFromPointer(in.AppToken),
			"app_token_file": StringFromPointer(in.AppTokenFile),
			"app_url":        StringFromPointer(in.AppUrl),
			"callback_id":    StringFromPointer(in.CallbackId),
			"channel":        StringFromPointer(in.Channel),
			"color":          StringFromPointer(in.Color),
			"fallback":       StringFromPointer(in.Fallback),
			"fields":         fields,
			"footer":         StringFromPointer(in.Footer),
			"http_config":    types.ObjectNull(httpType),
			"icon_emoji":     StringFromPointer(in.IconEmoji),
			"icon_url":       StringFromPointer(in.IconUrl),
			"image_url":      StringFromPointer(in.ImageUrl),
			"link_names":     BoolFromPointer(in.LinkNames),
			"message_text":   StringFromPointer(in.MessageText),
			"mrkdwn_in":      mrkdwn,
			"pretext":        StringFromPointer(in.Pretext),
			"send_resolved":  BoolFromPointer(in.SendResolved),
			"short_fields":   BoolFromPointer(in.ShortFields),
			"text":           StringFromPointer(in.Text),
			"thumb_url":      StringFromPointer(in.ThumbUrl),
			"timeout":        Int64FromPointer(in.Timeout),
			"title":          StringFromPointer(in.Title),
			"title_link":     StringFromPointer(in.TitleLink),
			"username":       StringFromPointer(in.Username),
		},
	)
	diags.Append(d...)
	return rv, diags
}

// configSlackConfigValueFromObject + ExpandConfigSlackConfigList +
// FlattenConfigSlackConfigList are emitted by `skaff convertors` into
// `zz_generated_config_slack_config.go` — the schema-level hook
// dispatch (HasUserHook=true) skips emitting Expand/Flatten core but
// still emits the supplemental helpers, so we don't repeat them here.
