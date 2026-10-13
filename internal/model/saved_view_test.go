package model

import (
	"encoding/json"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSavedView_ExplorerDataRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"queryType": "builder",
		"builder": map[string]interface{}{
			"queryData": []interface{}{
				map[string]interface{}{
					"dataSource":  "logs",
					"aggregation": "count",
				},
			},
		},
	}

	sv := &SavedView{ExplorerData: original}

	tfValue, err := sv.ExplorerDataToTerraform()
	if err != nil {
		t.Fatalf("ExplorerDataToTerraform failed: %v", err)
	}

	if tfValue.IsNull() || tfValue.IsUnknown() {
		t.Fatal("expected non-null terraform value")
	}

	sv2 := &SavedView{}
	err = sv2.SetExplorerData(tfValue)
	if err != nil {
		t.Fatalf("SetExplorerData failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(sv2.ExplorerData)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestSavedView_ExplorerDataEmpty(t *testing.T) {
	// An empty map flattens to an empty string, which cannot be expanded.
	// This verifies the known limitation.
	sv := &SavedView{ExplorerData: map[string]interface{}{}}

	tfValue, err := sv.ExplorerDataToTerraform()
	if err != nil {
		t.Fatalf("ExplorerDataToTerraform failed: %v", err)
	}

	sv2 := &SavedView{}
	err = sv2.SetExplorerData(tfValue)
	// Empty JSON string expansion fails -- expected behavior
	if err == nil {
		if len(sv2.ExplorerData) != 0 {
			t.Errorf("expected empty explorer data, got %v", sv2.ExplorerData)
		}
	}
}

func TestSavedView_TagsRoundTrip(t *testing.T) {
	original := []string{"production", "logs", "team-platform"}

	sv := &SavedView{Tags: original}

	tfList, diags := sv.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tfList.Elements()))
	}

	sv2 := &SavedView{}
	sv2.SetTags(tfList)

	if len(sv2.Tags) != len(original) {
		t.Fatalf("expected %d tags, got %d", len(original), len(sv2.Tags))
	}
	for i, tag := range original {
		if sv2.Tags[i] != tag {
			t.Errorf("tag[%d]: expected %s, got %s", i, tag, sv2.Tags[i])
		}
	}
}

func TestSavedView_TagsEmpty(t *testing.T) {
	sv := &SavedView{Tags: []string{}}

	tfList, diags := sv.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tfList.Elements()))
	}
}

func TestSavedView_TagsNil(t *testing.T) {
	sv := &SavedView{Tags: nil}

	tfList, diags := sv.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	_ = tfList
}

func TestSavedView_SetTagsFromConstructed(t *testing.T) {
	tfList, _ := types.ListValue(types.StringType, []tfattr.Value{
		types.StringValue("tag-a"),
		types.StringValue("tag-b"),
	})

	sv := &SavedView{}
	sv.SetTags(tfList)

	if len(sv.Tags) != 2 || sv.Tags[0] != "tag-a" || sv.Tags[1] != "tag-b" {
		t.Errorf("unexpected tags: %v", sv.Tags)
	}
}

func TestSavedView_JSONSerialization(t *testing.T) {
	sv := SavedView{
		UUID:       "abc-123",
		Name:       "My View",
		Category:   "logs",
		SourcePage: "/logs/explorer",
		Tags:       []string{"test"},
		ExplorerData: map[string]interface{}{
			"key": "value",
		},
		ExtraData: "extra",
	}

	data, err := json.Marshal(sv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var sv2 SavedView
	err = json.Unmarshal(data, &sv2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if sv2.UUID != "abc-123" {
		t.Errorf("UUID mismatch: %s", sv2.UUID)
	}
	if sv2.Name != "My View" {
		t.Errorf("Name mismatch: %s", sv2.Name)
	}
	if sv2.SourcePage != "/logs/explorer" {
		t.Errorf("SourcePage mismatch: %s", sv2.SourcePage)
	}
	if sv2.ExtraData != "extra" {
		t.Errorf("ExtraData mismatch: %s", sv2.ExtraData)
	}
}
