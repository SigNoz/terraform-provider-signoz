package legacyclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	api, err := apiclients.New(srv.URL, "", "test", "", 5*time.Second, 1)
	if err != nil {
		t.Fatalf("build wrapped client: %v", err)
	}

	client, err := NewClient(api)
	if err != nil {
		t.Fatalf("build legacy client: %v", err)
	}

	return client
}

func strictDecode(t *testing.T, body []byte, target any) {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		t.Errorf("strict decode: %v (body: %s)", err, body)
	}
}

// samplePostable is a contract-valid POST body (mirrors ToPostable output).
func samplePostable() model.PostableDashboardV2 {
	pd := model.PluginDashboard{
		Title:       "Fleet DevOps",
		Description: "Managed board",
		Tags:        []string{"lacrosse", "team:platform"},
		Pages: []model.PluginPage{{
			Name: "Fleet",
			Panels: []model.PluginPanel{
				{ID: "panel-a", Title: "Nodes", PluginID: "fleet-node-grid", PluginVersion: "0.2.0", Element: "lax-fleet-node-grid", X: 0, Y: 0, W: 8, H: 6},
				{ID: "panel-b", Title: "Status", PluginID: "system-status-board", PluginVersion: "0.1.0", Element: "lax-system-status-board", X: 8, Y: 0, W: 4, H: 6},
			},
		}},
	}

	return pd.ToPostable()
}

// v2Wire* types mirror the server's perses dashboard structs field-for-field;
// decoding request bodies with DisallowUnknownFields proves no stray fields
// are ever sent.

type v2WireSpec struct {
	Display struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"display"`
	Variables []map[string]any       `json:"variables"`
	Panels    map[string]v2WirePanel `json:"panels"`
	Layouts   []v2WireLayout         `json:"layouts"`
}

type v2WirePanel struct {
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
	} `json:"spec"`
}

type v2WireLayout struct {
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

type v2WireTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type v2WirePostable struct {
	SchemaVersion string      `json:"schemaVersion"`
	Name          string      `json:"name"`
	GenerateName  bool        `json:"generateName"`
	Tags          []v2WireTag `json:"tags"`
	Spec          v2WireSpec  `json:"spec"`
}

type v2WireUpdatable struct {
	SchemaVersion string      `json:"schemaVersion"`
	Name          string      `json:"name"`
	Tags          []v2WireTag `json:"tags"`
	Spec          v2WireSpec  `json:"spec"`
}

// validateDashboardV2Body enforces the server-side invariants the perses API
// applies on decode. Violations fail the test via t.
func validateDashboardV2Body(t *testing.T, tags []v2WireTag, spec v2WireSpec) {
	t.Helper()

	if len(tags) > 10 {
		t.Errorf("MaxTagsPerDashboard exceeded: %d", len(tags))
	}

	if spec.Variables == nil {
		t.Error("variables must be non-null")
	}
	if spec.Panels == nil {
		t.Error("panels must be non-null")
	}
	if spec.Layouts == nil {
		t.Error("layouts must be non-null")
	}

	for id, panel := range spec.Panels {
		if panel.Kind != "Panel" {
			t.Errorf("panel %s kind: %s", id, panel.Kind)
		}
		if panel.Spec.Plugin.Kind != "signoz/WebComponentPanel" {
			t.Errorf("panel %s plugin kind: %s", id, panel.Spec.Plugin.Kind)
		}
		if panel.Spec.Queries == nil || len(panel.Spec.Queries) != 0 {
			t.Errorf("panel %s: WebComponent panels require empty non-null queries", id)
		}
		if panel.Spec.Plugin.Spec.Element == "" || !strings.Contains(panel.Spec.Plugin.Spec.Element, "-") {
			t.Errorf("panel %s element: %q", id, panel.Spec.Plugin.Spec.Element)
		}
	}

	for li, layout := range spec.Layouts {
		if layout.Kind != "Grid" {
			t.Errorf("layout %d kind: %s", li, layout.Kind)
		}
		for _, item := range layout.Spec.Items {
			if item.Width < 1 || item.Height < 1 || item.X < 0 || item.Y < 0 || item.X+item.Width > 12 {
				t.Errorf("layout %d item geometry: %+v", li, item)
			}
			id := strings.TrimPrefix(item.Content.Ref, "#/spec/panels/")
			if _, ok := spec.Panels[id]; !ok {
				t.Errorf("layout %d item ref %q unresolved", li, item.Content.Ref)
			}
		}
	}
}

// dashboardV2Store is a minimal in-memory v2 API with the same envelope,
// immutability, and 201/404 semantics as the server. Handlers report body
// violations through t (FailNow from handler goroutines is not allowed, so
// t.Errorf + a 500 response is used).
type dashboardV2Store struct {
	mu      sync.Mutex
	next    int
	dash    map[string]*model.GettableDashboardV2
	created int
	updated int
	deleted int
}

func writeJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, `{"status":"error","error":{"code":404,"msg":"dashboard not found"}}`)
}

func (s *dashboardV2Store) writeEnvelope(t *testing.T, d *model.GettableDashboardV2, code int, w http.ResponseWriter) {
	t.Helper()

	b, err := json.Marshal(struct {
		Status string                    `json:"status"`
		Data   model.GettableDashboardV2 `json:"data"`
	}{Status: "success", Data: *d})
	if err != nil {
		t.Errorf("marshal envelope: %v", err)
		writeJSON(w, http.StatusInternalServerError, "{}")
		return
	}

	writeJSON(w, code, string(b))
}

func tagsToWire(tags []v2WireTag) []model.GettableTag {
	out := make([]model.GettableTag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, model.GettableTag{Key: tag.Key, Value: tag.Value})
	}
	return out
}

// specFromWire converts the strict test mirror back into the model shape the
// store echoes in responses.
func specFromWire(w v2WireSpec) model.DashboardSpecV2 {
	spec := model.DashboardSpecV2{
		Display:   model.DisplayV2{Name: w.Display.Name, Description: w.Display.Description},
		Variables: w.Variables,
		Panels:    map[string]model.PanelV2{},
		Layouts:   []model.LayoutV2{},
	}

	for id, p := range w.Panels {
		spec.Panels[id] = model.PanelV2{
			Kind: p.Kind,
			Spec: model.PanelSpecV2{
				Display: model.DisplayV2{Name: p.Spec.Display.Name, Description: p.Spec.Display.Description},
				Plugin: model.PanelPluginV2{
					Kind: p.Spec.Plugin.Kind,
					Spec: model.WebComponentPanelSpec{
						PluginID:      p.Spec.Plugin.Spec.PluginID,
						PluginVersion: p.Spec.Plugin.Spec.PluginVersion,
						Element:       p.Spec.Plugin.Spec.Element,
						Attributes:    p.Spec.Plugin.Spec.Attributes,
						Properties:    p.Spec.Plugin.Spec.Properties,
						ConfigKeys:    p.Spec.Plugin.Spec.ConfigKeys,
					},
				},
				Queries: p.Spec.Queries,
			},
		}
	}

	for _, l := range w.Layouts {
		layout := model.LayoutV2{Kind: l.Kind, Spec: model.GridLayoutSpecV2{Items: []model.GridItemV2{}}}
		if l.Spec.Display.Title != "" {
			layout.Spec.Display = &model.GridLayoutDisplayV2{Title: l.Spec.Display.Title}
		}
		for _, it := range l.Spec.Items {
			layout.Spec.Items = append(layout.Spec.Items, model.GridItemV2{
				X: it.X, Y: it.Y, Width: it.Width, Height: it.Height,
				Content: model.JSONPointerV2{Ref: it.Content.Ref},
			})
		}
		spec.Layouts = append(spec.Layouts, layout)
	}

	return spec
}

func (s *dashboardV2Store) Handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v2/dashboards") {
			http.NotFound(w, r)
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		id := strings.TrimPrefix(r.URL.Path, "/api/v2/dashboards/")

		switch r.Method {
		case http.MethodPost:
			var post v2WirePostable
			strictDecode(t, mustReadAll(r), &post)

			if post.SchemaVersion != "v6" {
				t.Errorf("schemaVersion: %s", post.SchemaVersion)
			}
			if !post.GenerateName || post.Name != "" {
				t.Errorf("create must use generateName with empty name: %+v", post)
			}
			validateDashboardV2Body(t, post.Tags, post.Spec)

			s.next++
			created := &model.GettableDashboardV2{
				ID:            fmt.Sprintf("%08d", s.next),
				Name:          fmt.Sprintf("dash-%08d", s.next),
				SchemaVersion: "v6",
				Tags:          tagsToWire(post.Tags),
				Spec:          specFromWire(post.Spec),
			}
			s.dash[created.ID] = created
			s.created++

			s.writeEnvelope(t, created, http.StatusCreated, w)

		case http.MethodGet:
			d, ok := s.dash[id]
			if !ok {
				notFound(w)
				return
			}
			s.writeEnvelope(t, d, http.StatusOK, w)

		case http.MethodPut:
			existing, ok := s.dash[id]
			if !ok {
				notFound(w)
				return
			}

			var up v2WireUpdatable
			strictDecode(t, mustReadAll(r), &up)

			if up.Name != existing.Name {
				t.Errorf("dashboard name is immutable: got %q want %q", up.Name, existing.Name)
			}
			validateDashboardV2Body(t, up.Tags, up.Spec)

			existing.Tags = tagsToWire(up.Tags)
			existing.Spec = specFromWire(up.Spec)
			s.updated++

			s.writeEnvelope(t, existing, http.StatusOK, w)

		case http.MethodDelete:
			if _, ok := s.dash[id]; !ok {
				notFound(w)
				return
			}
			delete(s.dash, id)
			s.deleted++
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func mustReadAll(r *http.Request) []byte {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	return buf.Bytes()
}

func TestDashboardV2_CRUDRoundTrip(t *testing.T) {
	store := &dashboardV2Store{dash: map[string]*model.GettableDashboardV2{}}
	client := newTestClient(t, store.Handler(t))

	ctx := context.Background()

	created, err := client.CreateDashboardV2(ctx, samplePostable())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.Name == "" || created.SchemaVersion != "v6" {
		t.Fatalf("created: %+v", created)
	}
	if store.created != 1 {
		t.Fatalf("create count: %d", store.created)
	}

	got, err := client.GetDashboardV2(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Fatalf("get mismatch: %+v vs %+v", got, created)
	}

	// Update: change title + tags through the updatable shape.
	pd := model.FromGettable(*got)
	pd.Title = "Fleet DevOps v2"
	pd.Tags = append(pd.Tags, "extra")

	updated, err := client.UpdateDashboardV2(ctx, created.ID, pd.ToUpdatable(created.Name))
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Spec.Display.Name != "Fleet DevOps v2" {
		t.Fatalf("update title: %+v", updated.Spec.Display)
	}
	if len(updated.Tags) != 3 {
		t.Fatalf("update tags: %+v", updated.Tags)
	}
	if store.updated != 1 {
		t.Fatalf("update count: %d", store.updated)
	}

	if err := client.DeleteDashboardV2(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if store.deleted != 1 {
		t.Fatalf("delete count: %d", store.deleted)
	}

	_, err = client.GetDashboardV2(ctx, created.ID)

	var apiErr *apiclients.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 APIError after delete, got %v", err)
	}
}

func TestDashboardV2_DeleteNotFoundIsClean404(t *testing.T) {
	store := &dashboardV2Store{dash: map[string]*model.GettableDashboardV2{}}
	client := newTestClient(t, store.Handler(t))

	err := client.DeleteDashboardV2(context.Background(), "nope")

	var apiErr *apiclients.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 APIError, got %v", err)
	}
}

func TestDashboardV2_SampleBodyPassesStrictFake(t *testing.T) {
	store := &dashboardV2Store{dash: map[string]*model.GettableDashboardV2{}}
	client := newTestClient(t, store.Handler(t))

	// The store decodes every body with DisallowUnknownFields and validates
	// the perses invariants — a green create proves the payload shape.
	if _, err := client.CreateDashboardV2(context.Background(), samplePostable()); err != nil {
		t.Fatalf("sample body rejected by strict fake server: %v", err)
	}
}
