package model

// DevOpsWidget is a known Lacrosse devops-dashboard WebComponent catalog entry.
// tagNames align with docs/devops-dashboard/COMPONENT-CATALOG.md.
type DevOpsWidget struct {
	PluginID string
	Element  string
	Category string
}

// DevOpsCatalog maps plugin_id → element/category for validation and defaults.
var DevOpsCatalog = map[string]DevOpsWidget{
	"fleet-node-grid":      {PluginID: "fleet-node-grid", Element: "lax-fleet-node-grid", Category: "fleet"},
	"system-status-board":  {PluginID: "system-status-board", Element: "lax-system-status-board", Category: "fleet"},
	"worker-tally-chips":   {PluginID: "worker-tally-chips", Element: "lax-worker-tally-chips", Category: "workers"},
	"worker-inspector":     {PluginID: "worker-inspector", Element: "lax-worker-inspector", Category: "workers"},
	"otel-endpoints-card":  {PluginID: "otel-endpoints-card", Element: "lax-otel-endpoints-card", Category: "observability"},
	"pg-activity":          {PluginID: "pg-activity", Element: "lax-pg-activity", Category: "data"},
	"mnesia-report":        {PluginID: "mnesia-report", Element: "lax-mnesia-report", Category: "data"},
	"genai-model-board":    {PluginID: "genai-model-board", Element: "lax-genai-model-board", Category: "ai"},
	"deploy-status-board":  {PluginID: "deploy-status-board", Element: "lax-deploy-status-board", Category: "deploy"},
	"topology-role-matrix": {PluginID: "topology-role-matrix", Element: "lax-role-matrix", Category: "topology"},
	"hello":                {PluginID: "hello", Element: "lax-hello-wc", Category: "demo"},
}

// ResolveElement returns catalog element for pluginID, or empty if unknown.
func ResolveElement(pluginID string) string {
	if w, ok := DevOpsCatalog[pluginID]; ok {
		return w.Element
	}
	return ""
}

// NewStubPluginPanel builds a panel carrying the Lacrosse stub-data convention:
// properties {mode: "stub", stub: {...}} feed demo data to the widget element.
// This is a data mode of the deployed plugins, not a stubbed API path.
func NewStubPluginPanel(title, pluginID string, stub map[string]interface{}, x, y, w, h int) PluginPanel {
	el := ResolveElement(pluginID)
	if el == "" {
		el = "lax-hello-wc"
	}
	if stub == nil {
		stub = map[string]interface{}{}
	}
	props := map[string]interface{}{
		"mode": "stub",
		"stub": stub,
	}
	return PluginPanel{
		Title:    title,
		PluginID: pluginID,
		Element:  el,
		Attributes: map[string]string{
			"data-source": "stub",
		},
		Properties: props,
		ConfigKeys: []string{},
		X:          x,
		Y:          y,
		W:          w,
		H:          h,
	}
}

// BuildDevOpsStubDashboard constructs the canonical Lacrosse devops stub dashboard
// used by terraform/signoz-dashboards. stubs keys: fleet, status, workers, otel, pg.
func BuildDevOpsStubDashboard(stubs map[string]map[string]interface{}) PluginDashboard {
	get := func(k string) map[string]interface{} {
		if stubs == nil {
			return map[string]interface{}{}
		}
		if v, ok := stubs[k]; ok && v != nil {
			return v
		}
		return map[string]interface{}{}
	}
	return PluginDashboard{
		Title:       "Lacrosse DevOps Stubs",
		Description: "Terraform-managed dashboard hosting devops WebComponent widgets with embedded stub data.",
		Tags:        []string{"lacrosse", "devops", "plugins", "stub"},
		Pages: []PluginPage{
			{
				Name: "Fleet",
				Panels: []PluginPanel{
					NewStubPluginPanel("Fleet node grid", "fleet-node-grid", get("fleet"), 0, 0, 8, 6),
					NewStubPluginPanel("System status", "system-status-board", get("status"), 8, 0, 4, 6),
					NewStubPluginPanel("Worker tally", "worker-tally-chips", get("workers"), 0, 6, 6, 4),
				},
			},
			{
				Name: "Data & OTel",
				Panels: []PluginPanel{
					NewStubPluginPanel("OTel endpoints", "otel-endpoints-card", get("otel"), 0, 0, 6, 4),
					NewStubPluginPanel("PG activity", "pg-activity", get("pg"), 6, 0, 6, 4),
				},
			},
		},
	}
}
