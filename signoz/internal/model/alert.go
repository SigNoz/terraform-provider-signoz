package model

import (
	"context"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

const (
	AlertTypeMetrics    = "METRIC_BASED_ALERT"
	AlertTypeLogs       = "LOGS_BASED_ALERT"
	AlertTypeTraces     = "TRACES_BASED_ALERT"
	AlertTypeExceptions = "EXCEPTIONS_BASED_ALERT"

	AlertRuleTypeThreshold = "threshold_rule"
	AlertRuleTypeProm      = "promql_rule"

	AlertSeverityCritical = "critical"
	AlertSeverityError    = "error"
	AlertSeverityWarning  = "warning"
	AlertSeverityInfo     = "info"

	AlertStateInactive = "inactive"
	AlertStatePending  = "pending"
	AlertStateFiring   = "firing"
	AlertStateDisabled = "disabled"
	AlertStateNoData   = "nodata"

	AlertTerraformLabel = "managedBy:terraform"
)

//nolint:gochecknoglobals
var (
	AlertTypes      = []string{AlertTypeMetrics, AlertTypeLogs, AlertTypeTraces, AlertTypeExceptions}
	AlertRuleTypes  = []string{AlertRuleTypeThreshold, AlertRuleTypeProm}
	AlertSeverities = []string{AlertSeverityCritical, AlertSeverityError, AlertSeverityWarning, AlertSeverityInfo}
	AlertStates     = []string{AlertStateInactive, AlertStatePending, AlertStateFiring, AlertStateDisabled, AlertStateNoData}
)

// Alert model.
type Alert struct {
	ID                   string                 `json:"id"`
	Alert                string                 `json:"alert"`
	AlertType            string                 `json:"alertType"`
	Annotations          AlertAnnotations       `json:"annotations"`
	BroadcastToAll       bool                   `json:"broadcastToAll"`
	Condition            map[string]interface{} `json:"condition"`
	Disabled             bool                   `json:"disabled,omitempty"`
	EvalWindow           string                 `json:"evalWindow"`
	Frequency            string                 `json:"frequency"`
	Labels               map[string]string      `json:"labels"`
	PreferredChannels    []string               `json:"preferredChannels"`
	RuleType             string                 `json:"ruleType"`
	Source               string                 `json:"source"`
	State                string                 `json:"state,omitempty"`
	Version              string                 `json:"version"`
	SchemaVersion        string                 `json:"schemaVersion,omitempty"`
	NotificationSettings NotificationSettings   `json:"notificationSettings,omitempty"`
	Evaluation           map[string]interface{} `json:"evaluation,omitempty"`
	CreateAt             string                 `json:"createAt,omitempty"`
	CreateBy             string                 `json:"createBy,omitempty"`
	UpdateAt             string                 `json:"updateAt,omitempty"`
	UpdateBy             string                 `json:"updateBy,omitempty"`
}

type NotificationSettings struct {
	GroupBy   []string `json:"groupBy,omitempty"`
	Renotify  Renotify `json:"renotify,omitempty"`
	UsePolicy bool     `json:"usePolicy,omitempty"`
}

type Renotify struct {
	Enabled          bool     `json:"enabled"`
	ReNotifyInterval string   `json:"interval,omitempty"`
	AlertStates      []string `json:"alertStates,omitempty"`
}

// Alert Annotations model.
type AlertAnnotations struct {
	Description string `json:"description"`
	Summary     string `json:"summary"`
}

func (a Alert) GetID() string {
	return a.ID
}

func (a Alert) GetName() string {
	return a.Alert
}

func (a Alert) GetType() string {
	return a.AlertType
}

func (a Alert) ConditionToTerraform() (types.String, error) {
	condition, err := structure.FlattenJsonToString(a.Condition)
	if err != nil {
		return types.StringValue(""), err
	}

	return types.StringValue(condition), nil
}

func (a Alert) NotificationSettingsToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	renotifyAttrTypes := attr.RenotifyAttrTypes()

	notificationSettingsAttrTypes := attr.NotificationSettingsAttrTypes()

	ns := a.NotificationSettings

	// Build Renotify object
	alertStatesList, diagsStates := types.ListValueFrom(ctx, types.StringType, ns.Renotify.AlertStates)
	diags.Append(diagsStates...)
	if diags.HasError() {
		return types.ObjectNull(notificationSettingsAttrTypes), diags
	}

	renotifyObj, diagsRenotify := types.ObjectValue(
		renotifyAttrTypes,
		map[string]tfattr.Value{
			attr.Interval:    types.StringValue(ns.Renotify.ReNotifyInterval),
			attr.AlertStates: alertStatesList,
			attr.Enabled:     types.BoolValue(ns.Renotify.Enabled),
		},
	)
	diags.Append(diagsRenotify...)
	if diags.HasError() {
		return types.ObjectNull(notificationSettingsAttrTypes), diags
	}

	var groupByList basetypes.ListValue
	if ns.GroupBy == nil {
		groupByList = types.ListValueMust(types.StringType, []tfattr.Value{})
	} else {
		var diagsGroup diag.Diagnostics
		groupByList, diagsGroup = types.ListValueFrom(ctx, types.StringType, ns.GroupBy)
		diags.Append(diagsGroup...)
		if diags.HasError() {
			return types.ObjectNull(notificationSettingsAttrTypes), diags
		}
	}

	objVal, diagsObj := types.ObjectValue(
		notificationSettingsAttrTypes,
		map[string]tfattr.Value{
			attr.Renotify:  renotifyObj,
			attr.GroupBy:   groupByList,
			attr.UsePolicy: types.BoolValue(ns.UsePolicy),
		},
	)
	diags.Append(diagsObj...)

	return objVal, diags
}

func (a Alert) EvaluationToTerraform() (types.String, error) {
	evaluation, err := structure.FlattenJsonToString(a.Evaluation)
	if err != nil {
		return types.StringValue(""), err
	}
	return types.StringValue(evaluation), nil
}

func (a Alert) LabelsToTerraform() (types.Map, diag.Diagnostics) {
	elements := map[string]tfattr.Value{}
	terraformLabels := strings.Split(AlertTerraformLabel, ":")
	for key, value := range a.Labels {
		if key == attr.Severity || key == terraformLabels[0] {
			continue
		}
		elements[key] = types.StringValue(value)
	}
	return types.MapValue(types.StringType, elements)
}

func (a Alert) PreferredChannelsToTerraform() (types.List, diag.Diagnostics) {
	preferredChannels := utils.Map(a.PreferredChannels, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, preferredChannels)
}

func (a Alert) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:                a.ID,
		attr.Alert:             a.Alert,
		attr.AlertType:         a.AlertType,
		attr.Annotations:       a.Annotations,
		attr.BroadcastToAll:    a.BroadcastToAll,
		attr.Condition:         a.Condition,
		attr.Disabled:          a.Disabled,
		attr.EvalWindow:        a.EvalWindow,
		attr.Frequency:         a.Frequency,
		attr.Labels:            a.Labels,
		attr.PreferredChannels: a.PreferredChannels,
		attr.RuleType:          a.RuleType,
		attr.Source:            a.Source,
		attr.State:             a.State,
		attr.Version:           a.Version,
		attr.CreateAt:          a.CreateAt,
		attr.CreateBy:          a.CreateBy,
		attr.UpdateAt:          a.UpdateAt,
		attr.UpdateBy:          a.UpdateBy,
		// attr.Description:       a.Description,
		// attr.Summary:           a.Summary,
		// attr.Severity:          a.Severity,
	}
}

func (a *Alert) SetCondition(tfCondition types.String) error {
	condition, err := structure.ExpandJsonFromString(tfCondition.ValueString())
	if err != nil {
		return err
	}

	a.Condition = condition
	return nil
}

func (a *Alert) SetEvaluation(tfEvaluation types.String) error {
	evaluation, err := structure.ExpandJsonFromString(tfEvaluation.ValueString())
	if err != nil {
		return err
	}

	a.Evaluation = evaluation
	return nil
}

func (a *Alert) SetNotificationSettings(ctx context.Context, tfNotification types.Object) error {
	if utils.IsNullOrUnknown(tfNotification) {
		return nil
	}

	attrs := tfNotification.Attributes()

	var renotify Renotify
	if renotifyAttr, ok := attrs[attr.Renotify]; ok && !utils.IsNullOrUnknown(renotifyAttr) {
		renotifyObj := renotifyAttr.(types.Object)
		renotifyAttrs := renotifyObj.Attributes()

		if v, ok := renotifyAttrs[attr.Enabled]; ok {
			if b, ok2 := v.(types.Bool); ok2 && !utils.IsNullOrUnknown(b) {
				renotify.Enabled = b.ValueBool()
			}
		}
		if v, ok := renotifyAttrs[attr.Interval]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				renotify.ReNotifyInterval = s.ValueString()
			}
		}
		if v, ok := renotifyAttrs[attr.AlertStates]; ok {
			if list, ok2 := v.(types.List); ok2 && !utils.IsNullOrUnknown(list) {
				var alertStates []string
				list.ElementsAs(ctx, &alertStates, false)
				renotify.AlertStates = alertStates
			}
		}
	}

	var ns NotificationSettings
	if v, ok := attrs[attr.GroupBy]; ok {
		if list, ok2 := v.(types.List); ok2 && !utils.IsNullOrUnknown(list) {
			var groupBy []string
			list.ElementsAs(ctx, &groupBy, false)
			ns.GroupBy = groupBy
		}
	}

	if v, ok := attrs[attr.UsePolicy]; ok {
		if b, ok2 := v.(types.Bool); ok2 && !utils.IsNullOrUnknown(b) {
			ns.UsePolicy = b.ValueBool()
		}
	}

	ns.Renotify = renotify
	a.NotificationSettings = ns

	return nil
}

func (a *Alert) SetLabels(tfLabels types.Map, tfSeverity types.String) {
	labels := make(map[string]string)

	for key, value := range tfLabels.Elements() {
		labels[key] = strings.Trim(value.String(), "\"")
	}

	terraformLabel := strings.Split(AlertTerraformLabel, ":")
	labels[strings.TrimSpace(terraformLabel[0])] = strings.TrimSpace(terraformLabel[1])

	if tfSeverity.ValueString() != "" {
		labels[attr.Severity] = tfSeverity.ValueString()
	}

	a.Labels = labels
}

func (a *Alert) SetPreferredChannels(tfPreferredChannels types.List) {
	preferredChannels := utils.Map(tfPreferredChannels.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	a.PreferredChannels = preferredChannels
}

func (a *Alert) SetSourceIfEmpty(hostURL string) {
	a.Source = utils.WithDefault(a.Source, hostURL+"/alerts")
}
