package legacyclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

func TestSetDataRetention_MetricsUsesV1QueryParams(t *testing.T) {
	var rawQuery, path, method string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method, rawQuery = r.URL.Path, r.Method, r.URL.RawQuery
		writeJSON(w, http.StatusOK, `{"message":"ttl updated"}`)
	})

	client := newTestClient(t, handler)

	err := client.SetDataRetention(context.Background(), &model.DataRetentionUpdate{
		Type:               "metrics",
		DurationHrs:        720,
		ColdStorageVolume:  "s3-volume",
		ColdStorageMoveHrs: 168,
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	if method != http.MethodPost || path != "/api/v1/settings/ttl" {
		t.Fatalf("request: %s %s", method, path)
	}

	// Server contract: type, duration; coldStorage + toColdDuration (the old
	// cold_storage_duration param does not exist).
	for _, want := range []string{"type=metrics", "duration=720h", "coldStorage=s3-volume", "toColdDuration=168h"} {
		if !strings.Contains(rawQuery, want) {
			t.Fatalf("query %q missing %q", rawQuery, want)
		}
	}
}

func TestSetDataRetention_V1WithoutVolumeOmitsColdParams(t *testing.T) {
	var rawQuery string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		writeJSON(w, http.StatusOK, `{"message":"ttl updated"}`)
	})

	client := newTestClient(t, handler)

	err := client.SetDataRetention(context.Background(), &model.DataRetentionUpdate{
		Type:        "traces",
		DurationHrs: 168,
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	if strings.Contains(rawQuery, "coldStorage") || strings.Contains(rawQuery, "toColdDuration") {
		t.Fatalf("cold params must be omitted without a volume: %q", rawQuery)
	}
}

func TestSetDataRetention_LogsUsesV2DaysBody(t *testing.T) {
	var path, method, body string
	var sawGet bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method

		if r.Method == http.MethodGet {
			sawGet = true
			writeJSON(w, http.StatusOK, `{"version":"v2","status":"success","default_ttl_days":15,"ttl_conditions":[{"conditions":[{"key":"k8s.namespace.name","values":["prod"]}],"ttlDays":90}]}`)
			return
		}

		b, _ := io.ReadAll(r.Body)
		body = string(b)

		writeJSON(w, http.StatusOK, `{"message":"custom retention updated"}`)
	})

	client := newTestClient(t, handler)

	err := client.SetDataRetention(context.Background(), &model.DataRetentionUpdate{
		Type:               "logs",
		DurationHrs:        360,
		ColdStorageVolume:  "s3-volume",
		ColdStorageMoveHrs: 168,
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	if !sawGet {
		t.Fatalf("update must read the live rule set before POSTing")
	}
	if method != http.MethodPost || path != "/api/v2/settings/ttl" {
		t.Fatalf("request: %s %s", method, path)
	}

	var payload model.CustomRetentionTTLParams
	strictDecode(t, []byte(body), &payload)

	if payload.Type != "logs" {
		t.Fatalf("type: %s", payload.Type)
	}
	if payload.DefaultTTLDays != 15 {
		t.Fatalf("defaultTTLDays: %d (360h must map to 15 days)", payload.DefaultTTLDays)
	}

	// Server contract: SetTTLV2 replaces the whole rule set, so the live
	// conditions must be carried into the body instead of an empty list.
	if len(payload.TTLConditions) != 1 ||
		len(payload.TTLConditions[0].Filters) != 1 ||
		payload.TTLConditions[0].Filters[0].Key != "k8s.namespace.name" ||
		payload.TTLConditions[0].TTLDays != 90 {
		t.Fatalf("ttlConditions must carry the live rules: %+v", payload.TTLConditions)
	}

	if payload.ColdStorageVolume != "s3-volume" || payload.ColdStorageDurationDays != 7 {
		t.Fatalf("cold storage: %+v", payload)
	}
}

func TestSetDataRetention_LogsV2WithoutMoveOmitsColdStorageDays(t *testing.T) {
	var body string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, `{"version":"v2","status":"success","default_ttl_days":10}`)
			return
		}

		b, _ := io.ReadAll(r.Body)
		body = string(b)
		writeJSON(w, http.StatusOK, `{"message":"custom retention updated"}`)
	})

	client := newTestClient(t, handler)

	err := client.SetDataRetention(context.Background(), &model.DataRetentionUpdate{
		Type:        "logs",
		DurationHrs: 240,
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	if strings.Contains(body, "coldStorage") {
		t.Fatalf("cold storage keys must be omitted: %s", body)
	}
	if !strings.Contains(body, `"defaultTTLDays":10`) {
		t.Fatalf("240h must map to 10 days: %s", body)
	}
	if !strings.Contains(body, `"ttlConditions":[]`) {
		t.Fatalf("no live rules must yield an empty (non-null) list: %s", body)
	}
}

func TestSetDataRetention_LogsV2UnavailableStillPosts(t *testing.T) {
	var sawPost bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notFound(w)
			return
		}

		sawPost = true
		writeJSON(w, http.StatusOK, `{"message":"custom retention updated"}`)
	})

	client := newTestClient(t, handler)

	err := client.SetDataRetention(context.Background(), &model.DataRetentionUpdate{
		Type:        "logs",
		DurationHrs: 240,
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	if !sawPost {
		t.Fatalf("a missing v2 GET (older server) must not block the update")
	}
}

func TestGetDataRetention_MergesV1AndV2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/settings/ttl" && r.URL.Query().Get("type") == "metrics":
			writeJSON(w, http.StatusOK, `{"expected_metrics_ttl_duration_hrs":720,"expected_metrics_move_ttl_duration_hrs":168}`)

		case r.URL.Path == "/api/v1/settings/ttl" && r.URL.Query().Get("type") == "logs":
			writeJSON(w, http.StatusOK, `{"expected_logs_ttl_duration_hrs":999,"expected_logs_move_ttl_duration_hrs":99}`)

		case r.URL.Path == "/api/v1/settings/ttl" && r.URL.Query().Get("type") == "traces":
			writeJSON(w, http.StatusOK, `{"expected_traces_ttl_duration_hrs":336,"expected_traces_move_ttl_duration_hrs":0}`)

		case r.URL.Path == "/api/v2/settings/ttl":
			writeJSON(w, http.StatusOK, `{"version":"v2","status":"success","default_ttl_days":15,"cold_storage_volume":"s3-volume","cold_storage_ttl_days":3}`)

		default:
			http.NotFound(w, r)
		}
	})

	client := newTestClient(t, handler)

	got, err := client.GetDataRetention(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.MetricsTTLDurationHrs != 720 || got.MetricsMoveTTLDurationHrs != 168 {
		t.Fatalf("metrics: %+v", got)
	}
	if got.TracesTTLDurationHrs != 336 {
		t.Fatalf("traces: %+v", got)
	}

	// Logs must come from v2 (days→hours), not the stale v1 values.
	if got.LogsTTLDurationHrs != 360 || got.LogsMoveTTLDurationHrs != 72 {
		t.Fatalf("logs v2 override: %+v", got)
	}
}

func TestGetDataRetention_V2UnavailableKeepsV1Logs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/settings/ttl" {
			notFound(w)
			return
		}
		writeJSON(w, http.StatusOK, `{"expected_`+r.URL.Query().Get("type")+`_ttl_duration_hrs":48,"expected_`+r.URL.Query().Get("type")+`_move_ttl_duration_hrs":0}`)
	})

	client := newTestClient(t, handler)

	got, err := client.GetDataRetention(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.LogsTTLDurationHrs != 48 {
		t.Fatalf("logs should keep v1 values when v2 is unavailable: %+v", got)
	}
}
