package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func loadStub(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	// Walk up from this test file to monorepo terraform/signoz-dashboards/stub-data
	_, file, _, _ := runtime.Caller(0)
	// .../repos/terraform-provider-signoz/internal/model/devops_catalog_test.go
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", ".."))
	path := filepath.Join(root, "terraform", "signoz-dashboards", "stub-data", name)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Fixture lives in a sibling signoz-dashboards checkout whose
			// placement depends on the enclosing environment, not this repo.
			t.Skipf("stub fixture not found at %s — skipping", path)
		}
		t.Fatalf("read stub %s: %v (path=%s)", name, err, path)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json %s: %v", name, err)
	}
	return m
}

func TestBuildDevOpsStubDashboard_EmbedsCatalogAndStubs(t *testing.T) {
	stubs := map[string]map[string]interface{}{
		"fleet":   loadStub(t, "fleet-nodes.json"),
		"status":  loadStub(t, "system-status.json"),
		"workers": loadStub(t, "worker-tally.json"),
		"otel":    loadStub(t, "otel-endpoints.json"),
		"pg":      loadStub(t, "pg-activity.json"),
	}

	pd := BuildDevOpsStubDashboard(stubs)
	if pd.Title != "Lacrosse DevOps Stubs" {
		t.Fatalf("title: %s", pd.Title)
	}
	if len(pd.Pages) != 2 {
		t.Fatalf("pages: %d", len(pd.Pages))
	}
	// 3 + 2 panels
	n := 0
	for _, p := range pd.Pages {
		n += len(p.Panels)
	}
	if n != 5 {
		t.Fatalf("expected 5 panels, got %d", n)
	}

	v2 := pd.ToDashboardV2()
	panels, ok := v2.Data["panels"].(map[string]interface{})
	if !ok || len(panels) != 5 {
		t.Fatalf("panels map size: %#v", v2.Data["panels"])
	}

	// Every panel must be WebComponent with stub mode and catalog element
	foundFleet := false
	for _, raw := range panels {
		p := raw.(map[string]interface{})
		spec := p["spec"].(map[string]interface{})
		plugin := spec["plugin"].(map[string]interface{})
		if plugin["kind"] != PanelKindWebComponent {
			t.Fatalf("kind: %#v", plugin["kind"])
		}
		ps := plugin["spec"].(map[string]interface{})
		pid, _ := ps["pluginId"].(string)
		el, _ := ps["element"].(string)
		if want := ResolveElement(pid); want != "" && el != want {
			t.Fatalf("plugin %s element got %s want %s", pid, el, want)
		}
		props, _ := ps["properties"].(map[string]interface{})
		if props["mode"] != "stub" {
			t.Fatalf("mode not stub for %s: %#v", pid, props)
		}
		if props["stub"] == nil {
			t.Fatalf("missing stub payload for %s", pid)
		}
		if pid == "fleet-node-grid" {
			foundFleet = true
			stub := props["stub"].(map[string]interface{})
			nodes, ok := stub["nodes"].([]interface{})
			if !ok || len(nodes) < 1 {
				t.Fatalf("fleet stub nodes: %#v", stub["nodes"])
			}
		}
		// queries empty
		if q, ok := spec["queries"].([]interface{}); !ok || len(q) != 0 {
			t.Fatalf("queries must be empty for WC: %#v", spec["queries"])
		}
	}
	if !foundFleet {
		t.Fatal("fleet-node-grid panel missing")
	}

	// pages order
	pages, ok := v2.Data["pages"].([]string)
	if !ok || len(pages) != 2 || pages[0] != "Fleet" {
		// JSON may retype as []interface{}
		if pi, ok := v2.Data["pages"].([]interface{}); ok && len(pi) == 2 {
			if pi[0] != "Fleet" {
				t.Fatalf("pages: %#v", pi)
			}
		} else {
			t.Fatalf("pages: %#v", v2.Data["pages"])
		}
	}

	doc, err := v2.ToJSON()
	if err != nil || len(doc) < 200 {
		t.Fatalf("document: err=%v len=%d", err, len(doc))
	}
}

func TestResolveElement_Catalog(t *testing.T) {
	if ResolveElement("pg-activity") != "lax-pg-activity" {
		t.Fatal(ResolveElement("pg-activity"))
	}
	if ResolveElement("unknown") != "" {
		t.Fatal("unknown should be empty")
	}
}

func TestBuildDevOpsStubDashboard_ToDashboardV1_WidgetsCarryPluginStubs(t *testing.T) {
	stubs := map[string]map[string]interface{}{
		"fleet":   loadStub(t, "fleet-nodes.json"),
		"status":  loadStub(t, "system-status.json"),
		"workers": loadStub(t, "worker-tally.json"),
		"otel":    loadStub(t, "otel-endpoints.json"),
		"pg":      loadStub(t, "pg-activity.json"),
	}
	pd := BuildDevOpsStubDashboard(stubs)
	v1 := pd.ToDashboardV1()
	if v1.Title != "Lacrosse DevOps Stubs" {
		t.Fatalf("title %s", v1.Title)
	}
	if len(v1.Widgets) != 5 {
		t.Fatalf("widgets %d", len(v1.Widgets))
	}
	if len(v1.Layout) != 5 {
		t.Fatalf("layout %d", len(v1.Layout))
	}
	// Each widget description embeds lacrossePlugin WebComponent + stubMode
	for _, w := range v1.Widgets {
		desc, _ := w["description"].(string)
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(desc), &meta); err != nil {
			t.Fatalf("desc json: %v %s", err, desc)
		}
		if meta["stubMode"] != true {
			t.Fatalf("stubMode: %#v", meta)
		}
		plugin, ok := meta["lacrossePlugin"].(map[string]interface{})
		if !ok || plugin["kind"] != PanelKindWebComponent {
			t.Fatalf("plugin: %#v", meta["lacrossePlugin"])
		}
		spec := plugin["spec"].(map[string]interface{})
		props := spec["properties"].(map[string]interface{})
		if props["mode"] != "stub" || props["stub"] == nil {
			t.Fatalf("props: %#v", props)
		}
	}
}
