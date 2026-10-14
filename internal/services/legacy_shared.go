// Shared constants and diagnostic helpers for the legacy (fork) services.
// addErr and addReadErr keep the distinct signatures the original
// resource/datasource packages used.
package services

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	SigNozAlert           = "signoz_alert"
	SigNozDashboard       = "signoz_dashboard"
	SigNozPluginDashboard = "signoz_plugin_dashboard"

	operationCreate = "create"
	operationRead   = "read"
	operationUpdate = "update"
	operationDelete = "delete"

	operationConfigure = "configure"

	alertDefaultEvalWindow   = "5m0s"
	alertDefaultDescription  = "This alert is fired when the defined metric (current value: {{$value}}) crosses the threshold ({{$threshold}})"
	alertDefaultFrequency    = "1m0s"
	alertDefaultSummary      = "The rule threshold is set to {{$threshold}}, and the observed metric value is {{$value}}"
	alertDefaultSourceSuffix = "alerts"
	// The server's v5 rules API rejects every other version ("only v5 is
	// supported"), so a v4 default made creates fail and downgraded v5 rules
	// on refresh plans.
	alertDefaultVersion = "v5"
)

// addErr adds an error to the diagnostics (resource operations).
func addErr(diagnostics *diag.Diagnostics, err error, operation string, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operation, resource),
		err.Error(),
	)
}

// addReadErr adds an error to the diagnostics (data source reads).
func addReadErr(diagnostics *diag.Diagnostics, err error, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operationRead, resource),
		err.Error(),
	)
}
