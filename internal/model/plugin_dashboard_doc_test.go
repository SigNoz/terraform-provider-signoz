package model_test

import (
	"encoding/json"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

// Isolated package so pre-existing broken tests in package model do not block this suite.

func TestToDashboardV2_MultiPageDevopsStubs(t *testing.T) {
	pd := model.PluginDashboard{
		Title:       "Lacrosse DevOps Stubs",
		Description: "stub widgets",
		Tags:        []string{"lacrosse", "devops", "stub"},
		Pages: []model.PluginPage{
			{
				Name: "Fleet",
				Panels: []model.PluginPanel{
					{
						ID: "panel-fleet", Title: "Fleet node grid",
						PluginID: "fleet-node-grid", Element: "lax-fleet-node-grid",
						X: 0, Y: 0, W: 8, H: 6,
						Properties: map[string]interface{}{"mode": "stub", "stub": map[string]interface{}{"nodes": []interface{}{}}},
					},
					{
						ID: "panel-status", Title: "System status",
						PluginID: "system-status-board", Element: "lax-system-status-board",
						X: 8, Y: 0, W: 4, H: 6,
					},
					{
						ID: "panel-workers", Title: "Worker tally",
						PluginID: "worker-tally-chips", Element: "lax-worker-tally-chips",
						X: 0, Y: 6, W: 6, H: 4,
					},
				},
			},
			{
				Name: "Data",
				Panels: []model.PluginPanel{
					{
						ID: "panel-otel", Title: "OTel endpoints",
						PluginID: "otel-endpoints-card", Element: "lax-otel-endpoints-card",
						X: 0, Y: 0, W: 6, H: 4,
					},
				},
			},
		},
	}

	v2 := pd.ToDashboardV2()
	if v2.Title != "Lacrosse DevOps Stubs" {
		t.Fatalf("title=%q", v2.Title)
	}
	panels, ok := v2.Data["panels"].(map[string]interface{})
	if !ok {
		t.Fatalf("panels type %T", v2.Data["panels"])
	}
	if len(panels) != 4 {
		t.Fatalf("want 4 panels, got %d", len(panels))
	}

	// Spot-check fleet panel kind + empty queries + stub mode property
	fleet := panels["panel-fleet"].(map[string]interface{})
	spec := fleet["spec"].(map[string]interface{})
	plugin := spec["plugin"].(map[string]interface{})
	if plugin["kind"] != model.PanelKindWebComponent {
		t.Fatalf("kind=%v", plugin["kind"])
	}
	pspec := plugin["spec"].(map[string]interface{})
	if pspec["pluginId"] != "fleet-node-grid" || pspec["element"] != "lax-fleet-node-grid" {
		t.Fatalf("plugin spec=%v", pspec)
	}
	if q := spec["queries"].([]interface{}); len(q) != 0 {
		t.Fatalf("queries must be empty for WC: %v", q)
	}
	props := pspec["properties"].(map[string]interface{})
	if props["mode"] != "stub" {
		t.Fatalf("expected stub mode in properties, got %v", props)
	}

	pages, ok := v2.Data["pages"].([]string)
	if !ok || len(pages) != 2 || pages[0] != "Fleet" {
		t.Fatalf("pages=%v", v2.Data["pages"])
	}

	// Multi-page layouts map (values are []map[string]interface{} before JSON)
	layouts, ok := v2.Data["layouts"].(map[string]interface{})
	if !ok {
		t.Fatalf("layouts type %T (multi-page should be map)", v2.Data["layouts"])
	}
	switch fleetLayout := layouts["Fleet"].(type) {
	case []map[string]interface{}:
		if len(fleetLayout) != 3 {
			t.Fatalf("fleet layout len=%d", len(fleetLayout))
		}
	case []interface{}:
		if len(fleetLayout) != 3 {
			t.Fatalf("fleet layout len=%d", len(fleetLayout))
		}
	default:
		t.Fatalf("fleet layout type %T", layouts["Fleet"])
	}

	b, err := json.Marshal(v2)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(b) {
		t.Fatal("invalid json")
	}
	// Input-sensitive: title must appear in document
	if !containsBytes(b, []byte("fleet-node-grid")) || !containsBytes(b, []byte("Lacrosse DevOps Stubs")) {
		t.Fatalf("document missing expected inputs: %s", b[:min(200, len(b))])
	}
}

func containsBytes(hay, needle []byte) bool {
	return len(hay) >= len(needle) && (string(hay) != "" && (len(needle) == 0 || indexOf(hay, needle) >= 0))
}

func indexOf(hay, needle []byte) int {
	// simple search
outer:
	for i := 0; i+len(needle) <= len(hay); i++ {
		for j := 0; j < len(needle); j++ {
			if hay[i+j] != needle[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
