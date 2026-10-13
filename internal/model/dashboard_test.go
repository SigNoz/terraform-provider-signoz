package model

import (
	"encoding/json"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDashboard_VariablesRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"query": map[string]interface{}{
			"type":     "query",
			"dataType": "string",
		},
	}

	dash := &Dashboard{Variables: original}

	tfValue, err := dash.VariablesToTerraform()
	if err != nil {
		t.Fatalf("VariablesToTerraform failed: %v", err)
	}

	dash2 := &Dashboard{}
	err = dash2.SetVariables(tfValue)
	if err != nil {
		t.Fatalf("SetVariables failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(dash2.Variables)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestDashboard_PanelMapRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"panel1": map[string]interface{}{
			"title": "CPU Usage",
			"type":  "graph",
		},
	}

	dash := &Dashboard{PanelMap: original}

	tfValue, err := dash.PanelMapToTerraform()
	if err != nil {
		t.Fatalf("PanelMapToTerraform failed: %v", err)
	}

	if tfValue.IsNull() {
		t.Fatal("expected non-null panel map")
	}

	dash2 := &Dashboard{}
	err = dash2.SetPanelMap(tfValue)
	if err != nil {
		t.Fatalf("SetPanelMap failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(dash2.PanelMap)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestDashboard_PanelMapNil(t *testing.T) {
	dash := &Dashboard{PanelMap: nil}

	tfValue, err := dash.PanelMapToTerraform()
	if err != nil {
		t.Fatalf("PanelMapToTerraform failed: %v", err)
	}

	if !tfValue.IsNull() {
		t.Error("expected null for nil PanelMap")
	}
}

func TestDashboard_SetPanelMapEmpty(t *testing.T) {
	dash := &Dashboard{}
	err := dash.SetPanelMap(types.StringValue(""))
	if err != nil {
		t.Fatalf("SetPanelMap failed on empty: %v", err)
	}

	if dash.PanelMap == nil {
		t.Error("expected non-nil empty map")
	}
	if len(dash.PanelMap) != 0 {
		t.Errorf("expected empty map, got %d entries", len(dash.PanelMap))
	}
}

func TestDashboard_TagsRoundTrip(t *testing.T) {
	original := []string{"production", "metrics", "team-a"}

	dash := &Dashboard{Tags: original}

	tfList, diags := dash.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tfList.Elements()))
	}

	dash2 := &Dashboard{}
	dash2.SetTags(tfList)

	if len(dash2.Tags) != len(original) {
		t.Fatalf("expected %d tags, got %d", len(original), len(dash2.Tags))
	}
	for i, tag := range original {
		if dash2.Tags[i] != tag {
			t.Errorf("tag[%d]: expected %s, got %s", i, tag, dash2.Tags[i])
		}
	}
}

func TestDashboard_TagsEmpty(t *testing.T) {
	dash := &Dashboard{Tags: []string{}}

	tfList, diags := dash.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tfList.Elements()))
	}
}

func TestDashboard_LayoutRoundTrip(t *testing.T) {
	original := []map[string]interface{}{
		{"i": "widget1", "x": float64(0), "y": float64(0), "w": float64(6), "h": float64(4)},
		{"i": "widget2", "x": float64(6), "y": float64(0), "w": float64(6), "h": float64(4)},
	}

	dash := &Dashboard{Layout: original}

	tfValue, err := dash.LayoutToTerraform()
	if err != nil {
		t.Fatalf("LayoutToTerraform failed: %v", err)
	}

	dash2 := &Dashboard{}
	err = dash2.SetLayout(tfValue)
	if err != nil {
		t.Fatalf("SetLayout failed: %v", err)
	}

	if len(dash2.Layout) != 2 {
		t.Fatalf("expected 2 layout items, got %d", len(dash2.Layout))
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(dash2.Layout)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestDashboard_WidgetsRoundTrip(t *testing.T) {
	original := []map[string]interface{}{
		{
			"id":    "widget1",
			"title": "Request Rate",
			"type":  "timeseries",
		},
	}

	dash := &Dashboard{Widgets: original}

	tfValue, err := dash.WidgetsToTerraform()
	if err != nil {
		t.Fatalf("WidgetsToTerraform failed: %v", err)
	}

	dash2 := &Dashboard{}
	err = dash2.SetWidgets(tfValue)
	if err != nil {
		t.Fatalf("SetWidgets failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(dash2.Widgets)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestDashboard_SetSourceIfEmpty(t *testing.T) {
	dash := &Dashboard{}
	dash.SetSourceIfEmpty("https://signoz.example.com")
	if dash.Source != "https://signoz.example.com/dashboard" {
		t.Errorf("expected source with /dashboard suffix, got %s", dash.Source)
	}

	dash2 := &Dashboard{Source: "https://custom.example.com/dash"}
	dash2.SetSourceIfEmpty("https://signoz.example.com")
	if dash2.Source != "https://custom.example.com/dash" {
		t.Errorf("expected existing source preserved, got %s", dash2.Source)
	}
}

func TestDashboard_JSONSerialization(t *testing.T) {
	dash := Dashboard{
		Name:                    "Test Dashboard",
		Title:                   "Test Title",
		Description:             "A test dashboard",
		Source:                  "https://signoz.example.com/dashboard",
		CollapsableRowsMigrated: true,
		UploadedGrafana:         false,
		Tags:                    []string{"test"},
		Variables:               map[string]interface{}{"key": "value"},
		Layout:                  []map[string]interface{}{{"x": float64(0)}},
		Widgets:                 []map[string]interface{}{{"id": "w1"}},
	}

	data, err := json.Marshal(dash)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var dash2 Dashboard
	err = json.Unmarshal(data, &dash2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if dash2.Name != dash.Name {
		t.Errorf("Name mismatch: %s != %s", dash2.Name, dash.Name)
	}
	if dash2.Title != dash.Title {
		t.Errorf("Title mismatch: %s != %s", dash2.Title, dash.Title)
	}
	if !dash2.CollapsableRowsMigrated {
		t.Error("expected CollapsableRowsMigrated=true")
	}
}

func TestDashboard_SetLayoutInvalid(t *testing.T) {
	dash := &Dashboard{}
	err := dash.SetLayout(types.StringValue("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDashboard_SetWidgetsInvalid(t *testing.T) {
	dash := &Dashboard{}
	err := dash.SetWidgets(types.StringValue("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDashboard_TagsNil(t *testing.T) {
	dash := &Dashboard{Tags: nil}

	tfList, diags := dash.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	// nil tags should produce empty list (utils.Map returns empty slice)
	_ = tfList
}

func TestDashboard_SetTagsFromList(t *testing.T) {
	tfList, _ := types.ListValue(types.StringType, []tfattr.Value{
		types.StringValue("a"),
		types.StringValue("b"),
	})

	dash := &Dashboard{}
	dash.SetTags(tfList)

	if len(dash.Tags) != 2 || dash.Tags[0] != "a" || dash.Tags[1] != "b" {
		t.Errorf("unexpected tags: %v", dash.Tags)
	}
}
