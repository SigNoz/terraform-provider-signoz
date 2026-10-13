package services

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
)

func TestRefreshKeyModel_PreservesValueWhenReadOmits(t *testing.T) {
	state := ingestionKeyResourceModel{
		ID:    types.StringValue("key-1"),
		Name:  types.StringValue("old-name"),
		Value: types.StringValue("signoz-secret"),
	}

	// Server masks the value on read: the stored secret must survive.
	key := &model.IngestionKey{ID: "key-1", Name: "masked", Value: ""}

	if diags := refreshKeyModel(&state, key); diags.HasError() {
		t.Fatalf("refresh: %v", diags)
	}

	if state.Value.ValueString() != "signoz-secret" {
		t.Fatalf("value was clobbered: %q", state.Value.ValueString())
	}

	if state.Name.ValueString() != "masked" {
		t.Fatalf("name not refreshed: %q", state.Name.ValueString())
	}

	// Server returns the value: state follows the server.
	key = &model.IngestionKey{ID: "key-1", Name: "open", Value: "fresh-value"}

	if diags := refreshKeyModel(&state, key); diags.HasError() {
		t.Fatalf("refresh: %v", diags)
	}

	if state.Value.ValueString() != "fresh-value" {
		t.Fatalf("value not updated: %q", state.Value.ValueString())
	}
}

func TestRefreshKeyModel_ExpiryZeroMapsToNull(t *testing.T) {
	state := ingestionKeyResourceModel{
		ID:        types.StringValue("key-1"),
		ExpiresAt: types.StringValue("2027-03-01T12:00:00Z"),
	}

	key := &model.IngestionKey{ID: "key-1", Name: "n", ExpiresAt: time.Time{}.UTC()}

	if diags := refreshKeyModel(&state, key); diags.HasError() {
		t.Fatalf("refresh: %v", diags)
	}

	if !state.ExpiresAt.IsNull() {
		t.Fatalf("unexpiring key should read back null, got %q", state.ExpiresAt.ValueString())
	}
}

func TestRefreshKeyModel_TagsNullStaysNull(t *testing.T) {
	state := ingestionKeyResourceModel{ID: types.StringValue("key-1")}

	key := &model.IngestionKey{ID: "key-1", Name: "n", Tags: nil}

	if diags := refreshKeyModel(&state, key); diags.HasError() {
		t.Fatalf("refresh: %v", diags)
	}

	if !state.Tags.IsNull() {
		t.Fatalf("nil server tags should stay null, got %v", state.Tags)
	}
}

func testInt64(v int64) *int64 { return &v }

func TestRefreshLimitModel_PopulatesFromServer(t *testing.T) {
	created := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	limit := &model.Limit{
		ID:        "limit-1",
		KeyID:     "key-9",
		Signal:    model.SignalTraces,
		CreatedAt: created,
		UpdatedAt: created,
		Config:    model.LimitConfig{Day: &model.LimitValue{Count: testInt64(5)}},
		Metric:    model.LimitMetric{Day: &model.LimitMetricValue{Count: 5}},
		Tags:      []string{"tier:gold"},
	}

	m := ingestionKeyLimitResourceModel{ID: types.StringValue("limit-1")}

	if diags := refreshLimitModel(&m, limit); diags.HasError() {
		t.Fatalf("refresh: %v", diags)
	}

	if m.KeyID.ValueString() != "key-9" || m.Signal.ValueString() != model.SignalTraces {
		t.Fatalf("identity: %s %s", m.KeyID, m.Signal)
	}

	if m.CreatedAt.ValueString() != created.Format(time.RFC3339Nano) {
		t.Fatalf("created_at: %s", m.CreatedAt)
	}

	config, diags := limitConfigFromTerraform(m.Config)
	if diags.HasError() {
		t.Fatalf("config expand: %v", diags)
	}

	if config.Day == nil || config.Day.Count == nil || *config.Day.Count != 5 || config.Second != nil {
		t.Fatalf("config round trip: %+v", config)
	}

	if len(m.Tags.Elements()) != 1 {
		t.Fatalf("tags: %v", m.Tags)
	}
}

func TestLimitConfigRoundTrip(t *testing.T) {
	day := testInt64(100)
	secondSize := testInt64(2048)

	config := model.LimitConfig{
		Day:    &model.LimitValue{Size: day},
		Second: &model.LimitValue{Size: secondSize},
	}

	value, diags := limitConfigToTerraform(config)
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}

	expanded, diags := limitConfigFromTerraform(value)
	if diags.HasError() {
		t.Fatalf("expand: %v", diags)
	}

	if expanded.Day == nil || expanded.Day.Size == nil || *expanded.Day.Size != 100 || expanded.Day.Count != nil {
		t.Fatalf("day: %+v", expanded.Day)
	}

	if expanded.Second == nil || expanded.Second.Size == nil || *expanded.Second.Size != 2048 || expanded.Second.Count != nil {
		t.Fatalf("second: %+v", expanded.Second)
	}

	// Nil arms stay nil across the round trip.
	empty, diags := limitConfigToTerraform(model.LimitConfig{})
	if diags.HasError() {
		t.Fatalf("flatten empty: %v", diags)
	}

	expanded, diags = limitConfigFromTerraform(empty)
	if diags.HasError() || expanded.Day != nil || expanded.Second != nil {
		t.Fatalf("empty round trip: %+v %v", expanded, diags)
	}
}

func TestLimitsToTerraform_NullAndPopulated(t *testing.T) {
	list, diags := limitsToTerraform(nil)
	if diags.HasError() || !list.IsNull() {
		t.Fatalf("nil limits -> %v %v", list, diags)
	}

	list, diags = limitsToTerraform([]model.Limit{{
		ID: "limit-1", KeyID: "key-1", Signal: model.SignalLogs,
		Config: model.LimitConfig{Second: &model.LimitValue{Count: testInt64(10)}},
		Metric: model.LimitMetric{Second: &model.LimitMetricValue{Count: 10}},
	}})
	if diags.HasError() || list.IsNull() || len(list.Elements()) != 1 {
		t.Fatalf("limits -> %v %v", list, diags)
	}
}

func TestIngestionKeyFlowsClassifyNotFound(t *testing.T) {
	wrapped := fmt.Errorf("get: %w", &apiclients.APIError{StatusCode: http.StatusNotFound, Raw: `{"status":"error"}`})

	if !isNotFound(wrapped) {
		t.Fatal("wrapped APIError 404 must classify as not found")
	}

	if isNotFound(errNotFound{}) {
		t.Fatal("bare error must not classify as 404")
	}

	if isNotFound(&apiclients.APIError{StatusCode: http.StatusBadRequest}) {
		t.Fatal("non-404 APIError must not classify as not found")
	}
}

type errNotFound struct{}

func (errNotFound) Error() string { return "not found" }
