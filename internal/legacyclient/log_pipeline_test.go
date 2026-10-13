package legacyclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

func assert404(t *testing.T, err error) {
	t.Helper()

	var apiErr *apiclients.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 APIError, got %v", err)
	}
}

func TestGetLogPipelines_UsesLatestVersionPath(t *testing.T) {
	var path, method string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		method = r.Method

		b, _ := json.Marshal(struct {
			Status string `json:"status"`
			Data   struct {
				Pipelines []model.LogPipeline `json:"pipelines"`
			} `json:"data"`
		}{Status: "success", Data: struct {
			Pipelines []model.LogPipeline `json:"pipelines"`
		}{Pipelines: []model.LogPipeline{{ID: "p1", Alias: "k8s", Name: "k8s"}}}})

		writeJSON(w, http.StatusOK, string(b))
	})

	client := newTestClient(t, handler)

	pipelines, err := client.GetLogPipelines(context.Background())
	if err != nil {
		t.Fatalf("get pipelines: %v", err)
	}

	// The server only registers GET /api/v1/logs/pipelines/{version}; a bare
	// GET on /api/v1/logs/pipelines 404s (http_handler.go registers no route).
	if method != http.MethodGet || path != "/api/v1/logs/pipelines/latest" {
		t.Fatalf("list request: %s %s", method, path)
	}
	if len(pipelines) != 1 || pipelines[0].ID != "p1" {
		t.Fatalf("pipelines: %+v", pipelines)
	}
}

func TestSaveLogPipelines_PostsBatch(t *testing.T) {
	var path, method, body string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		method = r.Method
		b, _ := io.ReadAll(r.Body)
		body = string(b)

		writeJSON(w, http.StatusOK, `{"status":"success","data":{"pipelines":[{"id":"p1","alias":"k8s","name":"k8s"}]}}`)
	})

	client := newTestClient(t, handler)

	saved, err := client.SaveLogPipelines(context.Background(), []model.LogPipeline{{ID: "p1", Alias: "k8s", Name: "k8s"}})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if method != http.MethodPost || path != "/api/v1/logs/pipelines" {
		t.Fatalf("save request: %s %s", method, path)
	}
	var payload model.LogPipelinesPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("save body: %v (%s)", err, body)
	}
	if len(payload.Pipelines) != 1 || payload.Pipelines[0].Alias != "k8s" {
		t.Fatalf("payload: %+v", payload)
	}
	if len(saved) != 1 || saved[0].ID != "p1" {
		t.Fatalf("saved: %+v", saved)
	}
}
