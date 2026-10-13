package model

import (
	"encoding/json"
	"testing"
)

func TestPluginDashboard_ToDashboardV2_WebComponent(t *testing.T) {
	pd := PluginDashboard{
		Title:       "Lacrosse Plugins",
		Description: "plugin host dashboards",
		Tags:        []string{"lacrosse", "plugins"},
		Pages: []PluginPage{
			{
				Name: "Overview",
				Panels: []PluginPanel{
					{
						Title:    "Hello",
						PluginID: "hello",
						Element:  "lax-hello-wc",
						X:        0, Y: 0, W: 6, H: 4,
						Attributes: map[string]string{"data-env": "stage"},
					},
				},
			},
		},
	}

	v2 := pd.ToDashboardV2()
	if v2.Title != "Lacrosse Plugins" {
		t.Fatalf("title: %s", v2.Title)
	}
	panels, ok := v2.Data["panels"].(map[string]interface{})
	if !ok || len(panels) != 1 {
		t.Fatalf("expected 1 panel, got %#v", v2.Data["panels"])
	}
	for _, raw := range panels {
		p := raw.(map[string]interface{})
		spec := p["spec"].(map[string]interface{})
		plugin := spec["plugin"].(map[string]interface{})
		if plugin["kind"] != PanelKindWebComponent {
			t.Fatalf("plugin kind: %#v", plugin["kind"])
		}
		pspec := plugin["spec"].(map[string]interface{})
		if pspec["pluginId"] != "hello" || pspec["element"] != "lax-hello-wc" {
			t.Fatalf("plugin spec: %#v", pspec)
		}
		q := spec["queries"].([]interface{})
		if len(q) != 0 {
			t.Fatalf("queries should be empty for WC panels")
		}
	}

	// Ensure JSON serializes
	b, err := json.Marshal(v2)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 50 {
		t.Fatalf("short json: %s", b)
	}
}
