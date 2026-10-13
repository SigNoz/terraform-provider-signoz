// Package plugindash is the public surface for Lacrosse plugin dashboard builders.
// Terraform resources and CLI tools import this package (not internal/model).
package plugindash

import "github.com/SigNoz/terraform-provider-signoz/internal/model"

// Re-export core types and builders used by Terraform and tooling.

const (
	PanelKindWebComponent    = model.PanelKindWebComponent
	DashboardSchemaVersionV2 = model.DashboardSchemaVersionV2
)

type (
	PluginDashboard      = model.PluginDashboard
	PluginPage           = model.PluginPage
	PluginPanel          = model.PluginPanel
	PostableDashboardV2  = model.PostableDashboardV2
	UpdatableDashboardV2 = model.UpdatableDashboardV2
	GettableDashboardV2  = model.GettableDashboardV2
	DevOpsWidget         = model.DevOpsWidget
)

var DevOpsCatalog = model.DevOpsCatalog

func ResolveElement(pluginID string) string { return model.ResolveElement(pluginID) }

func NewStubPluginPanel(title, pluginID string, stub map[string]interface{}, x, y, w, h int) PluginPanel {
	return model.NewStubPluginPanel(title, pluginID, stub, x, y, w, h)
}

func BuildDevOpsStubDashboard(stubs map[string]map[string]interface{}) PluginDashboard {
	return model.BuildDevOpsStubDashboard(stubs)
}
