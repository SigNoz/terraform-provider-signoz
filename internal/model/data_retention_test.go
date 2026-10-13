package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDataRetention_JSONSerialization(t *testing.T) {
	dr := DataRetention{
		MetricsTTLDurationHrs:     720,
		MetricsMoveTTLDurationHrs: 360,
		LogsTTLDurationHrs:        168,
		LogsMoveTTLDurationHrs:    72,
		TracesTTLDurationHrs:      168,
		TracesMoveTTLDurationHrs:  72,
	}

	data, err := json.Marshal(dr)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var dr2 DataRetention
	err = json.Unmarshal(data, &dr2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if dr2.MetricsTTLDurationHrs != 720 {
		t.Errorf("MetricsTTLDurationHrs: expected 720, got %d", dr2.MetricsTTLDurationHrs)
	}
	if dr2.MetricsMoveTTLDurationHrs != 360 {
		t.Errorf("MetricsMoveTTLDurationHrs: expected 360, got %d", dr2.MetricsMoveTTLDurationHrs)
	}
	if dr2.LogsTTLDurationHrs != 168 {
		t.Errorf("LogsTTLDurationHrs: expected 168, got %d", dr2.LogsTTLDurationHrs)
	}
	if dr2.LogsMoveTTLDurationHrs != 72 {
		t.Errorf("LogsMoveTTLDurationHrs: expected 72, got %d", dr2.LogsMoveTTLDurationHrs)
	}
	if dr2.TracesTTLDurationHrs != 168 {
		t.Errorf("TracesTTLDurationHrs: expected 168, got %d", dr2.TracesTTLDurationHrs)
	}
	if dr2.TracesMoveTTLDurationHrs != 72 {
		t.Errorf("TracesMoveTTLDurationHrs: expected 72, got %d", dr2.TracesMoveTTLDurationHrs)
	}
}

func TestDataRetention_JSONFieldNames(t *testing.T) {
	dr := DataRetention{
		MetricsTTLDurationHrs:     100,
		MetricsMoveTTLDurationHrs: 50,
		LogsTTLDurationHrs:        200,
		LogsMoveTTLDurationHrs:    100,
		TracesTTLDurationHrs:      300,
		TracesMoveTTLDurationHrs:  150,
	}

	data, err := json.Marshal(dr)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	expectedFields := []string{
		"expected_metrics_ttl_duration_hrs",
		"expected_metrics_move_ttl_duration_hrs",
		"expected_logs_ttl_duration_hrs",
		"expected_logs_move_ttl_duration_hrs",
		"expected_traces_ttl_duration_hrs",
		"expected_traces_move_ttl_duration_hrs",
	}

	for _, field := range expectedFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("expected JSON field %q not found", field)
		}
	}
}

func TestDataRetention_ZeroValues(t *testing.T) {
	dr := DataRetention{}

	data, err := json.Marshal(dr)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var dr2 DataRetention
	err = json.Unmarshal(data, &dr2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if dr2.MetricsTTLDurationHrs != 0 {
		t.Errorf("expected 0, got %d", dr2.MetricsTTLDurationHrs)
	}
}

func TestDataRetention_UnmarshalFromAPI(t *testing.T) {
	// Simulate an API response with expected_* fields
	apiJSON := `{
		"expected_metrics_ttl_duration_hrs": 2160,
		"expected_metrics_move_ttl_duration_hrs": 720,
		"expected_logs_ttl_duration_hrs": 720,
		"expected_logs_move_ttl_duration_hrs": 360,
		"expected_traces_ttl_duration_hrs": 720,
		"expected_traces_move_ttl_duration_hrs": 360
	}`

	var dr DataRetention
	err := json.Unmarshal([]byte(apiJSON), &dr)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if dr.MetricsTTLDurationHrs != 2160 {
		t.Errorf("MetricsTTLDurationHrs: expected 2160, got %d", dr.MetricsTTLDurationHrs)
	}
	if dr.MetricsMoveTTLDurationHrs != 720 {
		t.Errorf("MetricsMoveTTLDurationHrs: expected 720, got %d", dr.MetricsMoveTTLDurationHrs)
	}
}

func TestCustomRetentionTTLParams_JSONFieldNames(t *testing.T) {
	params := CustomRetentionTTLParams{
		Type:                    "logs",
		DefaultTTLDays:          15,
		TTLConditions:           []CustomRetentionRule{},
		ColdStorageVolume:       "s3-bucket",
		ColdStorageDurationDays: 7,
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	for _, field := range []string{"type", "defaultTTLDays", "ttlConditions", "coldStorageVolume", "coldStorageDurationDays"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("expected JSON field %q not found", field)
		}
	}
}

func TestCustomRetentionTTLParams_OmitEmptyColdStorage(t *testing.T) {
	params := CustomRetentionTTLParams{
		Type:           "logs",
		DefaultTTLDays: 15,
		TTLConditions:  []CustomRetentionRule{},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if strings.Contains(string(data), "coldStorage") {
		t.Fatalf("cold storage keys must be omitted when unset: %s", data)
	}
}

func TestCustomRetentionTTLResponseData_Decode(t *testing.T) {
	apiJSON := `{
		"version": "v2",
		"status": "success",
		"expected_logs_ttl_duration_hrs": 360,
		"expected_logs_move_ttl_duration_hrs": 72,
		"default_ttl_days": 15,
		"cold_storage_volume": "s3-bucket",
		"cold_storage_ttl_days": 3
	}`

	var resp CustomRetentionTTLResponseData
	if err := json.Unmarshal([]byte(apiJSON), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.DefaultTTLDays != 15 || resp.ColdStorageTTLDays != 3 || resp.ColdStorageVolume != "s3-bucket" {
		t.Fatalf("decoded: %#v", resp)
	}
}
