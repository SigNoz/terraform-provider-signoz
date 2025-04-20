package resource

const (
	SigNozAlert     = "signoz_alert"
	SigNozDashboard = "signoz_dashboard"

	operationCreate = "create"
	operationRead   = "read"
	operationUpdate = "update"
	operationDelete = "delete"

	alertDefaultEvalWindow   = "5m0s"
	alertDefaultDescription  = "This alert is fired when the defined metric (current value: {{$value}}) crosses the threshold ({{$threshold}})"
	alertDefaultFrequency    = "1m0s"
	alertDefaultSummary      = "The rule threshold is set to {{$threshold}}, and the observed metric value is {{$value}}"
	alertDefaultSourceSuffix = "alerts"
	alertDefaultVersion      = "v4"
)
