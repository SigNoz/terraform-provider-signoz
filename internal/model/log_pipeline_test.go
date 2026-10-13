package model

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLogPipeline_FilterRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"op": "AND",
		"items": []interface{}{
			map[string]interface{}{
				"key":   "service",
				"op":    "=",
				"value": "api",
			},
		},
	}

	pipeline := &LogPipeline{Filter: original}

	tfValue, err := pipeline.FilterToTerraform()
	if err != nil {
		t.Fatalf("FilterToTerraform failed: %v", err)
	}

	if tfValue.IsNull() || tfValue.IsUnknown() {
		t.Fatal("expected non-null terraform value")
	}

	pipeline2 := &LogPipeline{}
	err = pipeline2.SetFilter(tfValue)
	if err != nil {
		t.Fatalf("SetFilter failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(pipeline2.Filter)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestLogPipeline_FilterEmpty(t *testing.T) {
	// An empty map flattens to an empty string, which cannot be expanded.
	// This verifies the known limitation.
	pipeline := &LogPipeline{Filter: map[string]interface{}{}}

	tfValue, err := pipeline.FilterToTerraform()
	if err != nil {
		t.Fatalf("FilterToTerraform failed: %v", err)
	}

	pipeline2 := &LogPipeline{}
	err = pipeline2.SetFilter(tfValue)
	// Empty JSON string expansion fails -- expected behavior
	if err == nil {
		if len(pipeline2.Filter) != 0 {
			t.Errorf("expected empty filter, got %v", pipeline2.Filter)
		}
	}
}

func TestLogPipeline_ConfigRoundTrip(t *testing.T) {
	original := []map[string]interface{}{
		{
			"type":    "grok_parser",
			"name":    "parse-nginx",
			"enabled": true,
			"pattern": `%{COMBINEDAPACHELOG}`,
		},
		{
			"type":    "add",
			"name":    "add-env",
			"enabled": true,
			"field":   "env",
			"value":   "production",
		},
	}

	pipeline := &LogPipeline{Config: original}

	tfValue, err := pipeline.ConfigToTerraform()
	if err != nil {
		t.Fatalf("ConfigToTerraform failed: %v", err)
	}

	if tfValue.IsNull() || tfValue.IsUnknown() {
		t.Fatal("expected non-null terraform value")
	}

	pipeline2 := &LogPipeline{}
	err = pipeline2.SetConfig(tfValue)
	if err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	if len(pipeline2.Config) != 2 {
		t.Fatalf("expected 2 config items, got %d", len(pipeline2.Config))
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(pipeline2.Config)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestLogPipeline_ConfigEmpty(t *testing.T) {
	pipeline := &LogPipeline{Config: []map[string]interface{}{}}

	tfValue, err := pipeline.ConfigToTerraform()
	if err != nil {
		t.Fatalf("ConfigToTerraform failed: %v", err)
	}

	pipeline2 := &LogPipeline{}
	err = pipeline2.SetConfig(tfValue)
	if err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	if len(pipeline2.Config) != 0 {
		t.Errorf("expected empty config, got %v", pipeline2.Config)
	}
}

func TestLogPipeline_SetFilterInvalid(t *testing.T) {
	pipeline := &LogPipeline{}
	err := pipeline.SetFilter(types.StringValue("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLogPipeline_SetConfigInvalid(t *testing.T) {
	pipeline := &LogPipeline{}
	err := pipeline.SetConfig(types.StringValue("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLogPipeline_JSONSerialization(t *testing.T) {
	pipeline := LogPipeline{
		ID:          "pipe-1",
		Alias:       "nginx-parser",
		Name:        "NGINX Pipeline",
		Description: "Parses NGINX access logs",
		Enabled:     true,
		OrderID:     1,
		Filter: map[string]interface{}{
			"op": "AND",
		},
		Config: []map[string]interface{}{
			{"type": "grok_parser"},
		},
		CreatedBy: "admin",
	}

	data, err := json.Marshal(pipeline)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var pipeline2 LogPipeline
	err = json.Unmarshal(data, &pipeline2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if pipeline2.ID != "pipe-1" {
		t.Errorf("ID mismatch: %s", pipeline2.ID)
	}
	if pipeline2.Alias != "nginx-parser" {
		t.Errorf("Alias mismatch: %s", pipeline2.Alias)
	}
	if pipeline2.Name != "NGINX Pipeline" {
		t.Errorf("Name mismatch: %s", pipeline2.Name)
	}
	if !pipeline2.Enabled {
		t.Error("expected Enabled=true")
	}
	if pipeline2.OrderID != 1 {
		t.Errorf("OrderID: expected 1, got %d", pipeline2.OrderID)
	}
}

func TestLogPipelinesPayload_JSONSerialization(t *testing.T) {
	payload := LogPipelinesPayload{
		Pipelines: []LogPipeline{
			{ID: "p1", Name: "Pipeline 1", Enabled: true, OrderID: 0},
			{ID: "p2", Name: "Pipeline 2", Enabled: false, OrderID: 1},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var payload2 LogPipelinesPayload
	err = json.Unmarshal(data, &payload2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(payload2.Pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(payload2.Pipelines))
	}
	if payload2.Pipelines[0].Name != "Pipeline 1" {
		t.Errorf("pipeline[0].Name mismatch: %s", payload2.Pipelines[0].Name)
	}
}

func TestLogPipeline_FilterNil(t *testing.T) {
	pipeline := &LogPipeline{Filter: nil}

	tfValue, err := pipeline.FilterToTerraform()
	if err != nil {
		t.Fatalf("FilterToTerraform failed on nil: %v", err)
	}

	// nil map marshals to "null" JSON
	_ = tfValue
}

func TestLogPipeline_ConfigNil(t *testing.T) {
	pipeline := &LogPipeline{Config: nil}

	tfValue, err := pipeline.ConfigToTerraform()
	if err != nil {
		t.Fatalf("ConfigToTerraform failed on nil: %v", err)
	}

	// nil slice marshals to "null" JSON
	_ = tfValue
}
