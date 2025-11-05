package attr

import (
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	Alert                = "alert"
	AlertType            = "alert_type"
	Annotations          = "annotations"
	BroadcastToAll       = "broadcast_to_all"
	Condition            = "condition"
	Disabled             = "disabled"
	EvalWindow           = "eval_window"
	Evaluation           = "evaluation"
	Frequency            = "frequency"
	NotificationSettings = "notification_settings"
	PreferredChannels    = "preferred_channels"
	RuleType             = "rule_type"
	SchemaVersion        = "schema_version"
	Severity             = "severity"
	Source               = "source"
	State                = "state"
	Summary              = "summary"
	Renotify             = "renotify"
	Interval             = "interval"
	AlertStates          = "alert_states"
	Enabled              = "enabled"
	GroupBy              = "group_by"
	UsePolicy            = "use_policy"
)

func NotificationSettingsAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		Renotify:  types.ObjectType{AttrTypes: RenotifyAttrTypes()},
		GroupBy:   types.ListType{ElemType: types.StringType},
		UsePolicy: types.BoolType,
	}
}

func RenotifyAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		Interval:    types.StringType,
		AlertStates: types.ListType{ElemType: types.StringType},
		Enabled:     types.BoolType,
	}
}
