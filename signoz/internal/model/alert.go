package model

import "github.com/SigNoz/terraform-provider-signoz/signoz/internal/attr"

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
)

//nolint:gochecknoglobals
var (
	AlertTypes      = []string{AlertTypeMetrics, AlertTypeLogs, AlertTypeTraces, AlertTypeExceptions}
	AlertRuleTypes  = []string{AlertRuleTypeThreshold, AlertRuleTypeProm}
	AlertSeverities = []string{AlertSeverityCritical, AlertSeverityError, AlertSeverityWarning, AlertSeverityInfo}
	AlertStates     = []string{AlertStateInactive, AlertStatePending, AlertStateFiring, AlertStateDisabled}
)

// Alert model
type Alert struct {
	ID                string                 `json:"id"`
	Alert             string                 `json:"alert"`
	AlertType         string                 `json:"alertType"`
	Annotations       AlertAnnotations       `json:"annotations"`
	BroadcastToAll    bool                   `json:"broadcastToAll"`
	Condition         map[string]interface{} `json:"condition"`
	Disabled          bool                   `json:"disabled,omitempty"`
	EvalWindow        string                 `json:"evalWindow"`
	Frequency         string                 `json:"frequency"`
	Labels            map[string]string      `json:"labels"`
	PreferredChannels []string               `json:"preferredChannels"`
	RuleType          string                 `json:"ruleType"`
	Source            string                 `json:"source"`
	State             string                 `json:"state,omitempty"`
	Version           string                 `json:"version"`
	CreateAt          string                 `json:"createAt,omitempty"`
	CreateBy          string                 `json:"createBy,omitempty"`
	UpdateAt          string                 `json:"updateAt,omitempty"`
	UpdateBy          string                 `json:"updateBy,omitempty"`
}

// Alert Annotations model
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

func (a Alert) GetState() string {
	return AlertStateInactive
}
