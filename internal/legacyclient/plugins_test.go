package legacyclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

func TestGetPlugin_DefaultAndVersion(t *testing.T) {
	var seenPath, seenQuery string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery

		manifest := model.PluginManifest{
			ID: "hello", Version: "0.3.0", DisplayName: "Hello", Element: "lax-hello-wc", Source: "bundled",
		}
		b, _ := json.Marshal(struct {
			Status string               `json:"status"`
			Data   model.PluginManifest `json:"data"`
		}{Status: "success", Data: manifest})

		writeJSON(w, http.StatusOK, string(b))
	})

	client := newTestClient(t, handler)
	ctx := context.Background()

	got, err := client.GetPlugin(ctx, "hello")
	if err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	if seenPath != "/api/v1/plugins/hello" || seenQuery != "" {
		t.Fatalf("default manifest request: %s?%s", seenPath, seenQuery)
	}
	if got.Element != "lax-hello-wc" || got.Version != "0.3.0" {
		t.Fatalf("manifest: %+v", got)
	}

	got, err = client.GetPluginVersion(ctx, "hello", "0.1.0")
	if err != nil {
		t.Fatalf("get plugin version: %v", err)
	}
	if seenPath != "/api/v1/plugins/hello" || seenQuery != "version=0.1.0" {
		t.Fatalf("versioned request: %s?%s", seenPath, seenQuery)
	}
	if got.Version != "0.3.0" {
		t.Fatalf("manifest: %+v", got)
	}
}

func TestGetPlugin_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"status":"error","error":{"code":404,"msg":"not found"}}`)
	})

	client := newTestClient(t, handler)

	_, err := client.GetPlugin(context.Background(), "missing")
	assert404(t, err)
}

func TestListPlugins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/plugins" {
			http.NotFound(w, r)
			return
		}

		b, _ := json.Marshal(struct {
			Status string                 `json:"status"`
			Data   []model.PluginManifest `json:"data"`
		}{Status: "success", Data: []model.PluginManifest{
			{ID: "hello", Version: "0.3.0", Element: "lax-hello-wc"},
			{ID: "hello", Version: "0.2.0", Element: "lax-hello-wc"},
		}})

		writeJSON(w, http.StatusOK, string(b))
	})

	client := newTestClient(t, handler)

	got, err := client.ListPlugins(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 || got[0].ID != "hello" || got[1].Version != "0.2.0" {
		t.Fatalf("plugins: %+v", got)
	}
}

func TestPluginCatalogServer_NotFoundServes404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"status":"error","error":{"code":404,"msg":"not found"}}`)
	}))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("fake 404 helper: %d", res.StatusCode)
	}
}
