package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func sampleDashboard() PluginDashboard {
	return PluginDashboard{
		Title:       "Fleet DevOps",
		Description: "Managed board",
		Tags:        []string{"lacrosse", "team:platform"},
		Pages: []PluginPage{
			{
				Name: "Fleet",
				Panels: []PluginPanel{
					{
						Title: "Nodes", PluginID: "fleet-node-grid", PluginVersion: "0.2.0",
						Element:    "lax-fleet-node-grid",
						Attributes: map[string]string{"data-source": "stub"},
						Properties: map[string]interface{}{"mode": "stub"},
						ConfigKeys: []string{"endpoints"},
						X:          0, Y: 0, W: 8, H: 6,
					},
					{ID: "fixed01", Title: "Status", PluginID: "system-status-board", Element: "lax-system-status-board", X: 8, Y: 0, W: 4, H: 6},
				},
			},
		},
	}
}

func TestPluginDashboard_ToPostableShape(t *testing.T) {
	post := sampleDashboard().ToPostable()

	if post.SchemaVersion != "v6" {
		t.Fatalf("schemaVersion: %s", post.SchemaVersion)
	}
	if !post.GenerateName || post.Name != "" {
		t.Fatalf("expected generateName=true with empty name, got %v/%q", post.GenerateName, post.Name)
	}
	if len(post.Tags) != 2 || post.Tags[0].Key != "tag" || post.Tags[0].Value != "lacrosse" ||
		post.Tags[1].Key != "team" || post.Tags[1].Value != "platform" {
		t.Fatalf("tags: %#v", post.Tags)
	}

	spec := post.Spec
	if spec.Display.Name != "Fleet DevOps" || spec.Display.Description != "Managed board" {
		t.Fatalf("display: %#v", spec.Display)
	}
	if spec.Variables == nil || len(spec.Variables) != 0 {
		t.Fatalf("variables must be non-null empty, got %#v", spec.Variables)
	}
	if len(spec.Panels) != 2 || len(spec.Layouts) != 1 {
		t.Fatalf("panels %d layouts %d", len(spec.Panels), len(spec.Layouts))
	}

	layout := spec.Layouts[0]
	if layout.Kind != "Grid" || layout.Spec.Display == nil || layout.Spec.Display.Title != "Fleet" {
		t.Fatalf("layout: %#v", layout)
	}
	if len(layout.Spec.Items) != 2 {
		t.Fatalf("items: %d", len(layout.Spec.Items))
	}

	for _, item := range layout.Spec.Items {
		id := panelIDFromRef(item.Content.Ref)
		panel, ok := spec.Panels[id]
		if !ok {
			t.Fatalf("item ref %q has no panel", item.Content.Ref)
		}
		if panel.Kind != PanelEnvelopeKind {
			t.Fatalf("panel kind: %s", panel.Kind)
		}
		if panel.Spec.Queries == nil || len(panel.Spec.Queries) != 0 {
			t.Fatalf("queries must be non-null empty, got %#v", panel.Spec.Queries)
		}
		if panel.Spec.Plugin.Kind != PanelKindWebComponent {
			t.Fatalf("plugin kind: %s", panel.Spec.Plugin.Kind)
		}
	}

	first := spec.Panels[panelIDFromRef(layout.Spec.Items[0].Content.Ref)]
	if first.Spec.Plugin.Spec.PluginID != "fleet-node-grid" ||
		first.Spec.Plugin.Spec.PluginVersion != "0.2.0" ||
		first.Spec.Plugin.Spec.Element != "lax-fleet-node-grid" ||
		first.Spec.Display.Name != "Nodes" {
		t.Fatalf("first panel: %#v", first)
	}
	if first.Spec.Plugin.Spec.Attributes["data-source"] != "stub" {
		t.Fatalf("attributes: %#v", first.Spec.Plugin.Spec.Attributes)
	}

	second := spec.Panels["fixed01"]
	if second.Spec.Display.Name != "Status" {
		t.Fatalf("second panel: %#v", second)
	}
	if layout.Spec.Items[1].X != 8 || layout.Spec.Items[1].Width != 4 || layout.Spec.Items[1].Height != 6 {
		t.Fatalf("second item geometry: %#v", layout.Spec.Items[1])
	}
}

func TestPluginDashboard_GeneratedIDsArePanelKeys(t *testing.T) {
	pd := PluginDashboard{
		Title: "T",
		Pages: []PluginPage{{Name: "P", Panels: []PluginPanel{
			{Title: "a", PluginID: "hello", Element: "lax-hello-wc", W: 0, H: 0},
		}}},
	}

	post := pd.ToPostable()
	if len(post.Spec.Panels) != 1 {
		t.Fatalf("panels: %d", len(post.Spec.Panels))
	}

	for id := range post.Spec.Panels {
		if len(id) != 32 {
			t.Fatalf("generated id %q is not 32 hex chars", id)
		}
		if !strings.Contains(post.Spec.Layouts[0].Spec.Items[0].Content.Ref, id) {
			t.Fatalf("item ref missing id %q", id)
		}
		// zero geometry defaults to 6x4
		if post.Spec.Layouts[0].Spec.Items[0].Width != 6 || post.Spec.Layouts[0].Spec.Items[0].Height != 4 {
			t.Fatalf("default geometry: %#v", post.Spec.Layouts[0].Spec.Items[0])
		}
	}
}

func TestPluginDashboard_RoundTripFromGettable(t *testing.T) {
	pd := sampleDashboard()

	post := pd.ToPostable()
	got := GettableDashboardV2{
		ID:   "dash-1",
		Name: "fleet-devops",
	}

	// Convert postable tags to gettable shape.
	got.Tags = make([]GettableTag, 0, len(post.Tags))
	for _, tag := range post.Tags {
		got.Tags = append(got.Tags, GettableTag{Key: tag.Key, Value: tag.Value})
	}
	got.Spec = post.Spec

	round := FromGettable(got)
	if round.Title != pd.Title || round.Description != pd.Description {
		t.Fatalf("title/description: %q/%q", round.Title, round.Description)
	}
	if strings.Join(round.Tags, ",") != strings.Join(pd.Tags, ",") {
		t.Fatalf("tags: %#v", round.Tags)
	}
	if len(round.Pages) != 1 || round.Pages[0].Name != "Fleet" {
		t.Fatalf("pages: %#v", round.Pages)
	}

	orig := pd.Pages[0].Panels
	back := round.Pages[0].Panels
	if len(back) != len(orig) {
		t.Fatalf("panels: %d", len(back))
	}
	for i := range back {
		if back[i].Title != orig[i].Title ||
			back[i].PluginID != orig[i].PluginID ||
			back[i].PluginVersion != orig[i].PluginVersion ||
			back[i].Element != orig[i].Element ||
			back[i].X != orig[i].X || back[i].Y != orig[i].Y ||
			back[i].W != orig[i].W || back[i].H != orig[i].H {
			t.Fatalf("panel %d mismatch: %#v vs %#v", i, back[i], orig[i])
		}
	}
	if back[1].ID != "fixed01" {
		t.Fatalf("explicit id lost: %q", back[1].ID)
	}
}

func TestTagsRoundTrip(t *testing.T) {
	tags := TagsToPostable([]string{"lacrosse", "team:platform", "a:b:c"})
	if tags[2].Key != "a" || tags[2].Value != "b:c" {
		t.Fatalf("split at first colon: %#v", tags[2])
	}

	got := make([]GettableTag, 0, len(tags))
	for _, tag := range tags {
		got = append(got, GettableTag{Key: tag.Key, Value: tag.Value})
	}

	back := TagsFromGettable(got)
	if strings.Join(back, "|") != "lacrosse|team:platform|a:b:c" {
		t.Fatalf("round trip: %#v", back)
	}
}

func TestTagsFromGettable_PlainKeyCollision(t *testing.T) {
	// A server tag {tag: foo} reads back as plain "foo" — the same string a
	// user would configure. "tag:foo" as user input maps to {tag, foo} and
	// therefore converges to "foo" after one apply. Known, inherent mapping.
	back := TagsFromGettable([]GettableTag{{Key: "tag", Value: "foo"}})
	if len(back) != 1 || back[0] != "foo" {
		t.Fatalf("plain key: %#v", back)
	}
}

func TestPluginDashboard_ToUpdatable(t *testing.T) {
	up := sampleDashboard().ToUpdatable("fleet-devops")

	if up.Name != "fleet-devops" {
		t.Fatalf("name: %s", up.Name)
	}
	if up.SchemaVersion != "v6" {
		t.Fatalf("schemaVersion: %s", up.SchemaVersion)
	}
	if len(up.Spec.Panels) != 2 || len(up.Tags) != 2 {
		t.Fatalf("updatable payload: %#v", up)
	}
}

func TestPluginDashboard_DocumentJSONDeterministic(t *testing.T) {
	pd := sampleDashboard()
	// Pin panel ids: generated ids are per-ToPostable random by design.
	for pi := range pd.Pages {
		for fi := range pd.Pages[pi].Panels {
			pd.Pages[pi].Panels[fi].ID = "panel" + string(rune('a'+fi))
		}
	}

	a, err := pd.DocumentJSON()
	if err != nil {
		t.Fatal(err)
	}

	b, err := pd.DocumentJSON()
	if err != nil {
		t.Fatal(err)
	}

	if a != b {
		t.Fatal("document JSON is not deterministic")
	}
	if !strings.Contains(a, `"generateName": true`) || !strings.Contains(a, `"#/spec/panels/panela"`) {
		t.Fatalf("unexpected document shape: %s", a)
	}
}

func TestPropertiesJSON(t *testing.T) {
	if PropertiesJSON(nil) != "" || PropertiesJSON(map[string]interface{}{}) != "" {
		t.Fatal("empty properties must render empty string")
	}

	out := PropertiesJSON(map[string]interface{}{"b": 1, "a": true})
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("not json: %v", err)
	}
	if out != `{"a":true,"b":1}` {
		t.Fatalf("keys not sorted/compact: %s", out)
	}
}
