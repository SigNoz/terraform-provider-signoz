package model

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func loadStub(t *testing.T, name string) map[string]interface{} {
	t.Helper()

	// Stub data lives in the sibling signoz-dashboards checkout (its
	// placement depends on the enclosing environment, not this repo), so a
	// missing fixture skips rather than fails.
	paths := []string{
		"../../../../../terraform/signoz-dashboards/stub-data/" + name,
		"../../../../../../terraform/signoz-dashboards/stub-data/" + name,
		"../../../../../../../terraform/signoz-dashboards/stub-data/" + name,
	}

	var b []byte
	var err error
	for _, p := range paths {
		b, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Skipf("stub fixture %s not found — skipping", name)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json %s: %v", name, err)
	}

	return m
}

func devOpsStubs(t *testing.T) map[string]map[string]interface{} {
	t.Helper()

	return map[string]map[string]interface{}{
		"fleet":   loadStub(t, "fleet-nodes.json"),
		"status":  loadStub(t, "system-status.json"),
		"workers": loadStub(t, "worker-tally.json"),
		"otel":    loadStub(t, "otel-endpoints.json"),
		"pg":      loadStub(t, "pg-activity.json"),
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

func TestBuildDevOpsStubDashboard_Shape(t *testing.T) {
	pd := BuildDevOpsStubDashboard(devOpsStubs(t))
	if pd.Title != "Lacrosse DevOps Stubs" {
		t.Fatalf("title: %s", pd.Title)
	}
	if len(pd.Pages) != 2 {
		t.Fatalf("pages: %d", len(pd.Pages))
	}

	n := 0
	for _, p := range pd.Pages {
		n += len(p.Panels)
	}
	if n != 5 {
		t.Fatalf("expected 5 panels, got %d", n)
	}

	if pd.Pages[0].Name != "Fleet" || pd.Pages[1].Name != "Data & OTel" {
		t.Fatalf("page names: %q, %q", pd.Pages[0].Name, pd.Pages[1].Name)
	}
}

func TestBuildDevOpsStubDashboard_ToPostable_PanelsCarryCatalogAndStubs(t *testing.T) {
	pd := BuildDevOpsStubDashboard(devOpsStubs(t))
	post := pd.ToPostable()

	if len(post.Spec.Panels) != 5 {
		t.Fatalf("panels map size: %d", len(post.Spec.Panels))
	}

	foundFleet := false
	for _, panel := range post.Spec.Panels {
		plugin := panel.Spec.Plugin
		if plugin.Kind != PanelKindWebComponent {
			t.Fatalf("kind: %s", plugin.Kind)
		}

		spec := plugin.Spec
		if want := ResolveElement(spec.PluginID); want != "" && spec.Element != want {
			t.Fatalf("plugin %s element got %s want %s", spec.PluginID, spec.Element, want)
		}

		if spec.Properties["mode"] != "stub" {
			t.Fatalf("mode not stub for %s: %#v", spec.PluginID, spec.Properties)
		}
		if spec.Properties["stub"] == nil {
			t.Fatalf("missing stub payload for %s", spec.PluginID)
		}

		if spec.PluginID == "fleet-node-grid" {
			foundFleet = true
			stub := spec.Properties["stub"].(map[string]interface{})
			nodes, ok := stub["nodes"].([]interface{})
			if !ok || len(nodes) < 1 {
				t.Fatalf("fleet stub nodes: %#v", stub["nodes"])
			}
		}

		if len(panel.Spec.Queries) != 0 {
			t.Fatalf("queries must be empty for WebComponent panels: %#v", panel.Spec.Queries)
		}
	}
	if !foundFleet {
		t.Fatal("fleet-node-grid panel missing")
	}

	for _, layout := range post.Spec.Layouts {
		for _, item := range layout.Spec.Items {
			if !strings.HasPrefix(item.Content.Ref, "#/spec/panels/") {
				t.Fatalf("item ref: %q", item.Content.Ref)
			}
			if _, ok := post.Spec.Panels[panelIDFromRef(item.Content.Ref)]; !ok {
				t.Fatalf("item ref %q does not resolve", item.Content.Ref)
			}
		}
	}

	doc, err := pd.DocumentJSON()
	if err != nil || len(doc) < 200 {
		t.Fatalf("document: err=%v len=%d", err, len(doc))
	}
}

func TestNewStubPluginPanel_DataModeConvention(t *testing.T) {
	// The {mode: "stub", stub: {...}} properties are the Lacrosse stub-data
	// convention of the deployed plugins — a data mode, not a stubbed API path.
	panel := NewStubPluginPanel("Demo", "hello", map[string]interface{}{"x": 1}, 0, 0, 6, 4)

	if panel.Element != "lax-hello-wc" {
		t.Fatalf("element: %s", panel.Element)
	}
	if panel.Properties["mode"] != "stub" {
		t.Fatalf("properties: %#v", panel.Properties)
	}
	if panel.Attributes["data-source"] != "stub" {
		t.Fatalf("attributes: %#v", panel.Attributes)
	}

	fallback := NewStubPluginPanel("Demo", "not-in-catalog", nil, 0, 0, 6, 4)
	if fallback.Element != "lax-hello-wc" {
		t.Fatalf("fallback element: %s", fallback.Element)
	}
}
