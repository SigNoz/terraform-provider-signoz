package legacyclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

// The fake below mirrors the SigNoz gateway endpoints (pkg/gateway +
// pkg/apiserver/signozapiserver/gateway.go). Request bodies are decoded with
// DisallowUnknownFields against server-shaped types, so any field the provider
// sends that the server does not declare fails the round trip, and the
// create-limit route is the non-deprecated POST /api/v2/gateway/ingestion_limits.

type fakeLimitConfig struct {
	Day    *fakeLimitValue `json:"day,omitempty"`
	Second *fakeLimitValue `json:"second,omitempty"`
}

type fakeLimitValue struct {
	Size  *int64 `json:"size,omitempty"`
	Count *int64 `json:"count,omitempty"`
}

type fakeKeyPostBody struct {
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	ExpiresAt time.Time `json:"expires_at"`
}

type fakeLimitPostBody struct {
	KeyID  string          `json:"keyId"`
	Signal string          `json:"signal"`
	Config fakeLimitConfig `json:"config"`
	Tags   []string        `json:"tags"`
}

type fakeLimitPatchBody struct {
	Config fakeLimitConfig `json:"config"`
	Tags   []string        `json:"tags"`
}

type fakeStoredKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type fakeStoredLimit struct {
	ID        string          `json:"id"`
	Signal    string          `json:"signal"`
	Config    fakeLimitConfig `json:"config"`
	Tags      []string        `json:"tags"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	KeyID     string          `json:"key_id"`
	Metric    fakeMetric      `json:"metric"`
}

type fakeMetric struct {
	Day    *fakeMetricValue `json:"day,omitempty"`
	Second *fakeMetricValue `json:"second,omitempty"`
}

type fakeMetricValue struct {
	Count int64 `json:"count"`
	Size  int64 `json:"size"`
}

type gatewayFake struct {
	srv      *httptest.Server
	keys     map[string]*fakeStoredKey
	limits   map[string]*fakeStoredLimit
	maskRead bool

	lastKeyPost   *fakeKeyPostBody
	lastLimitPost *fakeLimitPostBody
	lastKeyPatch  *fakeKeyPostBody
}

func newGatewayFake(t *testing.T) *gatewayFake {
	t.Helper()

	fake := &gatewayFake{
		keys:   map[string]*fakeStoredKey{},
		limits: map[string]*fakeStoredLimit{},
	}

	fake.srv = httptest.NewServer(http.HandlerFunc(fake.route))
	t.Cleanup(fake.srv.Close)

	return fake
}

func (f *gatewayFake) client(t *testing.T) *Client {
	t.Helper()

	api, err := apiclients.New(f.srv.URL, "", "test", "", 5*time.Second, 1)
	if err != nil {
		t.Fatalf("build wrapped client: %v", err)
	}

	client, err := NewClient(api)
	if err != nil {
		t.Fatalf("build legacy client: %v", err)
	}

	return client
}

func (f *gatewayFake) route(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case r.Method == http.MethodPost && path == "/api/v2/gateway/ingestion_keys":
		var body fakeKeyPostBody

		if !f.decode(w, r, &body) {
			return
		}

		f.lastKeyPost = &body

		now := time.Now().UTC()
		id := fmt.Sprintf("key-%d", len(f.keys)+1)
		f.keys[id] = &fakeStoredKey{
			ID: id, Name: body.Name, Value: fmt.Sprintf("%064x", len(f.keys)+1),
			ExpiresAt: body.ExpiresAt, Tags: body.Tags, CreatedAt: now, UpdatedAt: now,
		}

		f.writeData(w, http.StatusCreated, map[string]string{"id": id, "value": f.keys[id].Value})

	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v2/gateway/ingestion_keys/") && strings.HasSuffix(path, "/limits"):
		keyID := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v2/gateway/ingestion_keys/"), "/limits")

		key, ok := f.keys[keyID]
		if !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_not_found")
			return
		}

		limits := make([]*fakeStoredLimit, 0)

		for _, l := range f.limits {
			if l.KeyID == key.ID {
				limits = append(limits, l)
			}
		}

		f.writeData(w, http.StatusOK, limits)

	case r.Method == http.MethodPatch && strings.HasPrefix(path, "/api/v2/gateway/ingestion_keys/"):
		key, ok := f.keys[strings.TrimPrefix(path, "/api/v2/gateway/ingestion_keys/")]
		if !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_not_found")
			return
		}

		var body fakeKeyPostBody

		if !f.decode(w, r, &body) {
			return
		}

		f.lastKeyPatch = &body
		key.Name, key.Tags, key.ExpiresAt, key.UpdatedAt = body.Name, body.Tags, body.ExpiresAt, time.Now().UTC()

		f.writeData(w, http.StatusNoContent, nil)

	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/v2/gateway/ingestion_keys/"):
		keyID := strings.TrimPrefix(path, "/api/v2/gateway/ingestion_keys/")

		if _, ok := f.keys[keyID]; !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_not_found")
			return
		}

		delete(f.keys, keyID)

		for id, l := range f.limits {
			if l.KeyID == keyID {
				delete(f.limits, id)
			}
		}

		f.writeData(w, http.StatusNoContent, nil)

	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v2/gateway/ingestion_keys/"):
		key, ok := f.keys[strings.TrimPrefix(path, "/api/v2/gateway/ingestion_keys/")]
		if !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_not_found")
			return
		}

		shown := *key

		if f.maskRead {
			shown.Value = ""
		}

		f.writeData(w, http.StatusOK, shown)

	case r.Method == http.MethodPost && path == "/api/v2/gateway/ingestion_limits":
		var body fakeLimitPostBody

		if !f.decode(w, r, &body) {
			return
		}

		f.lastLimitPost = &body

		if _, ok := f.keys[body.KeyID]; !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_not_found")
			return
		}

		now := time.Now().UTC()
		id := fmt.Sprintf("limit-%d", len(f.limits)+1)
		f.limits[id] = &fakeStoredLimit{
			ID: id, Signal: body.Signal, Config: body.Config, Tags: body.Tags,
			CreatedAt: now, UpdatedAt: now, KeyID: body.KeyID,
			Metric: f.metricFor(body.Config),
		}

		f.writeData(w, http.StatusCreated, map[string]string{"id": id})

	case r.Method == http.MethodPatch && strings.HasPrefix(path, "/api/v2/gateway/ingestion_limits/"):
		limit, ok := f.limits[strings.TrimPrefix(path, "/api/v2/gateway/ingestion_limits/")]
		if !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_limit_not_found")
			return
		}

		var body fakeLimitPatchBody

		if !f.decode(w, r, &body) {
			return
		}

		limit.Config, limit.Tags, limit.UpdatedAt = body.Config, body.Tags, time.Now().UTC()
		limit.Metric = f.metricFor(body.Config)

		f.writeData(w, http.StatusNoContent, nil)

	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/v2/gateway/ingestion_limits/"):
		limitID := strings.TrimPrefix(path, "/api/v2/gateway/ingestion_limits/")

		if _, ok := f.limits[limitID]; !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_limit_not_found")
			return
		}

		delete(f.limits, limitID)

		f.writeData(w, http.StatusNoContent, nil)

	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v2/gateway/ingestion_limits/"):
		limit, ok := f.limits[strings.TrimPrefix(path, "/api/v2/gateway/ingestion_limits/")]
		if !ok {
			f.writeError(w, http.StatusNotFound, "ingestion_key_limit_not_found")
			return
		}

		f.writeData(w, http.StatusOK, limit)

	default:
		f.writeError(w, http.StatusNotFound, "no_route")
	}
}

// metricFor mirrors the server: the metric snapshot dereferences the config's
// optional values to zero.
func (f *gatewayFake) metricFor(c fakeLimitConfig) fakeMetric {
	m := fakeMetric{}

	arm := func(v *fakeLimitValue) *fakeMetricValue {
		if v == nil {
			return nil
		}

		out := &fakeMetricValue{}
		if v.Count != nil {
			out.Count = *v.Count
		}
		if v.Size != nil {
			out.Size = *v.Size
		}

		return out
	}

	m.Day, m.Second = arm(c.Day), arm(c.Second)

	return m
}

// decode is the server's strict body binding: unknown fields are rejected.
func (f *gatewayFake) decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		f.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid body: %v", err))
		return false
	}

	return true
}

func (f *gatewayFake) writeData(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": data})
}

func (f *gatewayFake) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "error",
		"error":  map[string]any{"code": code, "message": message},
	})
}

func int64Ptr(v int64) *int64 { return &v }

func TestIngestionKeyClient_CRUDRoundTrip(t *testing.T) {
	fake := newGatewayFake(t)
	client := fake.client(t)
	ctx := context.Background()

	// Create with no expiry: the wire body must carry the zero-time sentinel.
	created, err := client.CreateIngestionKey(ctx, model.PostableIngestionKey{
		Name: "prod-key",
		Tags: []string{"env:prod"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.ID == "" || created.Value == "" {
		t.Fatalf("create response missing id/value: %+v", created)
	}

	if fake.lastKeyPost == nil || !fake.lastKeyPost.ExpiresAt.IsZero() {
		t.Fatalf("expected zero-time expires_at on create wire body, got %+v", fake.lastKeyPost)
	}

	// Read back: the server returns the value and the zero-time expiry.
	key, err := client.GetIngestionKey(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if key.Name != "prod-key" || key.Value != created.Value || !model.NoExpiry(key.ExpiresAt) {
		t.Fatalf("round trip mismatch: %+v", key)
	}

	if len(key.Tags) != 1 || key.Tags[0] != "env:prod" {
		t.Fatalf("tags: %v", key.Tags)
	}

	// Update: new name, tags and a real expiry.
	expiry := time.Date(2027, 3, 1, 12, 0, 0, 0, time.UTC)

	if err := client.UpdateIngestionKey(ctx, created.ID, model.PostableIngestionKey{
		Name:      "prod-key-renamed",
		Tags:      []string{"env:prod", "team:platform"},
		ExpiresAt: expiry,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	if fake.lastKeyPatch == nil || !fake.lastKeyPatch.ExpiresAt.Equal(expiry) {
		t.Fatalf("patch body expiry: %+v", fake.lastKeyPatch)
	}

	key, err = client.GetIngestionKey(ctx, created.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}

	if key.Name != "prod-key-renamed" || len(key.Tags) != 2 || !key.ExpiresAt.Equal(expiry) {
		t.Fatalf("post-update key: %+v", key)
	}

	// Delete, then confirm the typed 404.
	if err := client.DeleteIngestionKey(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = client.GetIngestionKey(ctx, created.ID)

	var apiErr *apiclients.APIError

	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected APIError 404 after delete, got %v", err)
	}
}

func TestIngestionKeyClient_ValueOmittedOnRead(t *testing.T) {
	fake := newGatewayFake(t)
	fake.maskRead = true

	client := fake.client(t)
	ctx := context.Background()

	created, err := client.CreateIngestionKey(ctx, model.PostableIngestionKey{Name: "masked"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	key, err := client.GetIngestionKey(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if key.Value != "" {
		t.Fatalf("expected masked read to omit value, got %q", key.Value)
	}
}

func TestIngestionKeyLimitClient_CRUDRoundTrip(t *testing.T) {
	fake := newGatewayFake(t)
	client := fake.client(t)
	ctx := context.Background()

	created, err := client.CreateIngestionKey(ctx, model.PostableIngestionKey{Name: "limited"})
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	// Create a limit via the non-deprecated endpoint; the wire body must use
	// camelCase keyId.
	postable := model.PostableIngestionKeyLimit{
		KeyID:  created.ID,
		Signal: model.SignalLogs,
		Config: model.LimitConfig{
			Day:    &model.LimitValue{Count: int64Ptr(1000)},
			Second: &model.LimitValue{Size: int64Ptr(2048)},
		},
		Tags: []string{"tier:gold"},
	}

	limitCreated, err := client.CreateIngestionKeyLimit(ctx, postable)
	if err != nil {
		t.Fatalf("create limit: %v", err)
	}

	if limitCreated.ID == "" {
		t.Fatal("empty limit id")
	}

	if fake.lastLimitPost == nil || fake.lastLimitPost.KeyID != created.ID {
		t.Fatalf("create-limit wire body: %+v", fake.lastLimitPost)
	}

	// Read: config arms round-trip including which dimension was left out.
	limit, err := client.GetIngestionKeyLimit(ctx, limitCreated.ID)
	if err != nil {
		t.Fatalf("get limit: %v", err)
	}

	if limit.KeyID != created.ID || limit.Signal != model.SignalLogs {
		t.Fatalf("limit identity: %+v", limit)
	}

	if limit.Config.Day == nil || limit.Config.Day.Count == nil || *limit.Config.Day.Count != 1000 || limit.Config.Day.Size != nil {
		t.Fatalf("config.day: %+v", limit.Config.Day)
	}

	if limit.Config.Second == nil || limit.Config.Second.Size == nil || *limit.Config.Second.Size != 2048 || limit.Config.Second.Count != nil {
		t.Fatalf("config.second: %+v", limit.Config.Second)
	}

	if limit.Metric.Day == nil || limit.Metric.Day.Count != 1000 || limit.Metric.Day.Size != 0 {
		t.Fatalf("metric.day: %+v", limit.Metric.Day)
	}

	// Update: config + tags only.
	if err := client.UpdateIngestionKeyLimit(ctx, limitCreated.ID, model.UpdatableIngestionKeyLimit{
		Config: model.LimitConfig{Day: &model.LimitValue{Count: int64Ptr(2000)}},
		Tags:   nil,
	}); err != nil {
		t.Fatalf("update limit: %v", err)
	}

	limit, err = client.GetIngestionKeyLimit(ctx, limitCreated.ID)
	if err != nil {
		t.Fatalf("get limit after update: %v", err)
	}

	if limit.Config.Day == nil || limit.Config.Day.Count == nil || *limit.Config.Day.Count != 2000 || limit.Config.Second != nil {
		t.Fatalf("post-update config: %+v", limit.Config)
	}

	// List via the key-scoped endpoint.
	limits, err := client.GetIngestionKeyLimits(ctx, created.ID)
	if err != nil {
		t.Fatalf("list limits: %v", err)
	}

	if len(limits) != 1 || limits[0].ID != limitCreated.ID {
		t.Fatalf("list: %+v", limits)
	}

	// Delete, then confirm the typed 404.
	if err := client.DeleteIngestionKeyLimit(ctx, limitCreated.ID); err != nil {
		t.Fatalf("delete limit: %v", err)
	}

	_, err = client.GetIngestionKeyLimit(ctx, limitCreated.ID)

	var apiErr *apiclients.APIError

	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected APIError 404 after limit delete, got %v", err)
	}
}

func TestIngestionKeyLimitClient_CreateOnMissingKey(t *testing.T) {
	fake := newGatewayFake(t)
	client := fake.client(t)
	ctx := context.Background()

	_, err := client.CreateIngestionKeyLimit(ctx, model.PostableIngestionKeyLimit{
		KeyID:  "no-such-key",
		Signal: model.SignalTraces,
	})

	var apiErr *apiclients.APIError

	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected APIError 404 for missing parent key, got %v", err)
	}
}

func TestIngestionKeyClient_StrictServerRejectsUnknownFields(t *testing.T) {
	fake := newGatewayFake(t)
	client := fake.client(t)
	ctx := context.Background()

	// A body the server does not declare (old fork shape) must fail loudly.
	if _, err := client.CreateIngestionKey(ctx, model.PostableIngestionKey{Name: "x"}); err != nil {
		t.Fatalf("baseline create failed: %v", err)
	}

	// Direct request with an unknown field: the fake answers 400.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fake.srv.URL+"/api/v2/gateway/ingestion_keys",
		strings.NewReader(`{"name":"x","day_limit":10}`))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.doRequest(ctx, req)

	var apiErr *apiclients.APIError

	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected strict-decode 400 for unknown field, got %v", err)
	}
}
