package model_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

// Strict wire mirrors (server decodes with DisallowUnknownFields at every
// level) to prove the document the builder emits is exactly the v2 contract.
type wireSpec struct {
	Display struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"display"`
	Variables []map[string]any     `json:"variables"`
	Panels    map[string]wirePanel `json:"panels"`
	Layouts   []wireLayout         `json:"layouts"`
}

type wirePanel struct {
	Kind string `json:"kind"`
	Spec struct {
		Display struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"display"`
		Plugin struct {
			Kind string `json:"kind"`
			Spec struct {
				PluginID      string            `json:"pluginId"`
				PluginVersion string            `json:"pluginVersion"`
				Element       string            `json:"element"`
				Attributes    map[string]string `json:"attributes"`
				Properties    map[string]any    `json:"properties"`
				ConfigKeys    []string          `json:"configKeys"`
			} `json:"spec"`
		} `json:"plugin"`
		Queries []map[string]any `json:"queries"`
		Links   []map[string]any `json:"links"`
	} `json:"spec"`
}

type wireLayout struct {
	Kind string `json:"kind"`
	Spec struct {
		Display struct {
			Title string `json:"title"`
		} `json:"display"`
		Items []struct {
			X       int `json:"x"`
			Y       int `json:"y"`
			Width   int `json:"width"`
			Height  int `json:"height"`
			Content struct {
				Ref string `json:"$ref"`
			} `json:"content"`
		} `json:"items"`
	} `json:"spec"`
}

type wirePostable struct {
	SchemaVersion string `json:"schemaVersion"`
	Image         string `json:"image"`
	Name          string `json:"name"`
	GenerateName  bool   `json:"generateName"`
	Tags          []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"tags"`
	Spec wireSpec `json:"spec"`
}

func TestDocument_PostableWire_MultiPageDevopsStubs(t *testing.T) {
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

	doc, err := pd.DocumentJSON()
	if err != nil {
		t.Fatal(err)
	}

	dec := json.NewDecoder(strings.NewReader(doc))
	dec.DisallowUnknownFields()
	var post wirePostable
	if err := dec.Decode(&post); err != nil {
		t.Fatalf("document is not strict v2 wire shape: %v", err)
	}

	if post.SchemaVersion != "v6" || !post.GenerateName || post.Name != "" {
		t.Fatalf("envelope: %+v", post)
	}
	if len(post.Spec.Panels) != 4 || len(post.Spec.Layouts) != 2 {
		t.Fatalf("panels %d layouts %d", len(post.Spec.Panels), len(post.Spec.Layouts))
	}

	fleet := post.Spec.Panels["panel-fleet"]
	if fleet.Kind != model.PanelEnvelopeKind || fleet.Spec.Plugin.Kind != model.PanelKindWebComponent {
		t.Fatalf("kinds: %q/%q", fleet.Kind, fleet.Spec.Plugin.Kind)
	}
	if fleet.Spec.Plugin.Spec.PluginID != "fleet-node-grid" || fleet.Spec.Plugin.Spec.Element != "lax-fleet-node-grid" {
		t.Fatalf("plugin spec: %+v", fleet.Spec.Plugin.Spec)
	}
	if len(fleet.Spec.Queries) != 0 {
		t.Fatalf("queries must be empty for WebComponent panels: %v", fleet.Spec.Queries)
	}
	if fleet.Spec.Plugin.Spec.Properties["mode"] != "stub" {
		t.Fatalf("expected stub mode in properties: %v", fleet.Spec.Plugin.Spec.Properties)
	}

	if post.Spec.Layouts[0].Spec.Display.Title != "Fleet" || post.Spec.Layouts[1].Spec.Display.Title != "Data" {
		t.Fatalf("layout titles: %q/%q", post.Spec.Layouts[0].Spec.Display.Title, post.Spec.Layouts[1].Spec.Display.Title)
	}
	if len(post.Spec.Layouts[0].Spec.Items) != 3 || len(post.Spec.Layouts[1].Spec.Items) != 1 {
		t.Fatalf("layout item counts: %d/%d", len(post.Spec.Layouts[0].Spec.Items), len(post.Spec.Layouts[1].Spec.Items))
	}

	if !bytes.Contains([]byte(doc), []byte("fleet-node-grid")) || !bytes.Contains([]byte(doc), []byte("Lacrosse DevOps Stubs")) {
		t.Fatal("document missing expected inputs")
	}

	if !bytes.Contains([]byte(doc), []byte(`"image": ""`)) {
		t.Fatal("image must always serialize (the server contract has no omitempty on image)")
	}

	if bytes.Contains([]byte(doc), []byte(`"links"`)) {
		t.Fatal("panels without links must not emit a links field")
	}
}
