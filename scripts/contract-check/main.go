//go:build ignore

// Command contract-check feeds the provider's emitted wire JSON into the
// SigNoz server's strict request types, and the server's response types back
// into the provider's client decode shapes. It runs inside a throwaway module
// (created by scripts/contract-check/run.sh) so the provider's go.mod never
// depends on the server. See scripts/contract-check/README.md.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	v3 "github.com/SigNoz/signoz/pkg/query-service/model/v3"
	"github.com/SigNoz/signoz/pkg/types"
	"github.com/SigNoz/signoz/pkg/types/dashboardtypes"
	"github.com/SigNoz/signoz/pkg/types/gatewaytypes"
	"github.com/SigNoz/signoz/pkg/types/pipelinetypes"
	"github.com/SigNoz/signoz/pkg/types/retentiontypes"
	"github.com/SigNoz/signoz/pkg/valuer"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var failures int

func main() {
	check("log_pipeline: provider POST body decodes into server PostablePipelines and passes IsValid", checkPipelinePost)
	check("log_pipeline: server GET body decodes into the provider client shape", checkPipelineGet)
	check("data_retention: provider POST body decodes into server retentiontypes.CustomRetentionTTLParams", checkRetentionPost)
	check("data_retention: server GET response decodes into the provider client shape", checkRetentionGet)
	check("plugin_dashboard: provider POST body decodes into server PostableDashboardV2 and passes Validate", checkDashboardPost)
	check("plugin_dashboard: carried PUT body decodes into server UpdatableDashboardV2 with server-owned fields intact", checkDashboardUpdate)
	check("plugin_dashboard: server GET body (incl. spec/panel links) decodes into the provider client shape", checkDashboardGet)
	check("ingestion_key: provider POST/PATCH bodies decode into server gatewaytypes", checkIngestionKeyPost)
	check("ingestion_key: server key/limit responses (enveloped) decode into the provider client shapes", checkIngestionKeyGet)

	if failures > 0 {
		fmt.Printf("\n%d check(s) FAILED\n", failures)
		os.Exit(1)
	}

	fmt.Println("\nall contract checks passed")
}

func check(name string, fn func() error) {
	if err := fn(); err != nil {
		failures++
		fmt.Printf("FAIL %s\n  %v\n", name, err)

		return
	}

	fmt.Printf("PASS %s\n", name)
}

func decodeStrict(data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	return dec.Decode(out)
}

func envelope(status string, data any) ([]byte, error) {
	return json.Marshal(map[string]any{"status": status, "data": data})
}

func int64Ptr(v int64) *int64 { return &v }

// sampleFilter is a canonical v3.FilterSet as the provider's filter JSON
// string expands to.
func sampleFilter() map[string]any {
	return map[string]any{
		"op": "AND",
		"items": []any{map[string]any{
			"key": map[string]any{
				"key": "service.name", "dataType": "string", "type": "tag", "isColumn": true, "isJSON": false,
			},
			"value": "api",
			"op":    "=",
		}},
	}
}

func sampleDashboard() model.PluginDashboard {
	return model.PluginDashboard{
		Title:       "DevOps Status",
		Description: "Real-time infrastructure and application status dashboard",
		Tags:        []string{"devops", "team:platform"},
		Pages: []model.PluginPage{{
			Name: "Overview",
			Panels: []model.PluginPanel{{
				ID: "abc123def456789a", Title: "System Status",
				PluginID: "devops-status", PluginVersion: "1.0.0", Element: "lax-devops-status",
				Attributes: map[string]string{"env": "prod"},
				Properties: map[string]any{"stub": true},
				ConfigKeys: []string{"env"},
				X:          0, Y: 0, W: 6, H: 4,
			}},
		}},
	}
}

func checkPipelinePost() error {
	// The provider builds the payload from TF config through SetConfig, which
	// fills in the operator id/orderId the server requires — go through it so
	// the emitted JSON is what the server actually receives.
	pipeline := model.LogPipeline{
		OrderID: 1, Name: "parse json", Alias: "parsejson", Description: "provider-managed pipeline", Enabled: true,
	}

	if err := pipeline.SetFilter(tftypes.StringValue(`{"op":"AND","items":[{"key":{"key":"service.name","dataType":"string","type":"tag","isColumn":true,"isJSON":false},"value":"api","op":"="}]}`)); err != nil {
		return fmt.Errorf("SetFilter: %w", err)
	}

	// Deliberately no operator id/orderId: SetConfig must inject them.
	if err := pipeline.SetConfig(tftypes.StringValue(`[{"type":"json_parser","enabled":true,"parse_to":"attributes"}]`)); err != nil {
		return fmt.Errorf("SetConfig: %w", err)
	}

	payload := model.LogPipelinesPayload{Pipelines: []model.LogPipeline{pipeline}}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var postable pipelinetypes.PostablePipelines
	if err := decodeStrict(raw, &postable); err != nil {
		return fmt.Errorf("server strict decode: %w", err)
	}

	if len(postable.Pipelines) != 1 {
		return fmt.Errorf("decoded %d pipelines, want 1", len(postable.Pipelines))
	}

	if err := postable.Pipelines[0].IsValid(); err != nil {
		return fmt.Errorf("server IsValid: %w", err)
	}

	return nil
}

func checkPipelineGet() error {
	gettable := pipelinetypes.GettablePipeline{
		StoreablePipeline: pipelinetypes.StoreablePipeline{
			Identifiable: types.Identifiable{ID: valuer.MustNewUUID("0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0001")},
			OrderID:      1,
			Enabled:      true,
			Name:         "parse json",
			Alias:        "parsejson",
			Description:  "provider-managed pipeline",
		},
		Filter: &v3.FilterSet{
			Operator: "AND",
			Items: []v3.FilterItem{{
				Key:      v3.AttributeKey{Key: "service.name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag, IsColumn: true},
				Value:    "api",
				Operator: v3.FilterOperatorEqual,
			}},
		},
		Config: []pipelinetypes.PipelineOperator{{Type: "json_parser", OrderId: 1, Enabled: true, ParseTo: "attributes"}},
	}

	raw, err := envelope("success", map[string]any{"pipelines": []pipelinetypes.GettablePipeline{gettable}})
	if err != nil {
		return err
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Pipelines []model.LogPipeline `json:"pipelines"`
		} `json:"data"`
	}
	if err := decodeStrict(raw, &resp); err != nil {
		return fmt.Errorf("provider strict decode: %w", err)
	}

	if resp.Status != "success" || len(resp.Data.Pipelines) != 1 {
		return fmt.Errorf("envelope/pipelines: %+v", resp)
	}

	got := resp.Data.Pipelines[0]
	if got.ID == "" || got.Name != "parse json" || got.Alias != "parsejson" || got.OrderID != 1 || !got.Enabled {
		return fmt.Errorf("pipeline fields: %+v", got)
	}

	if got.Filter == nil || got.Filter["op"] != "AND" {
		return fmt.Errorf("filter: %#v", got.Filter)
	}

	if len(got.Config) != 1 || got.Config[0]["type"] != "json_parser" || got.Config[0]["orderId"] != float64(1) {
		return fmt.Errorf("config: %#v", got.Config)
	}

	return nil
}

func checkRetentionPost() error {
	params := model.CustomRetentionTTLParams{
		Type:           "logs",
		DefaultTTLDays: 15,
		TTLConditions: []model.CustomRetentionRule{{
			Filters: []model.FilterCondition{{Key: "k8s.namespace.name", Values: []string{"prod"}}},
			TTLDays: 90,
		}},
		ColdStorageVolume:       "s3-volume",
		ColdStorageDurationDays: 7,
	}

	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}

	var server retentiontypes.CustomRetentionTTLParams
	if err := decodeStrict(raw, &server); err != nil {
		return fmt.Errorf("server strict decode: %w", err)
	}

	if server.Type != "logs" || server.DefaultTTLDays != 15 {
		return fmt.Errorf("type/days: %+v", server)
	}

	if server.ColdStorageVolume != "s3-volume" || server.ToColdStorageDurationDays != 7 {
		return fmt.Errorf("cold storage: %+v", server)
	}

	if len(server.TTLConditions) != 1 ||
		len(server.TTLConditions[0].Filters) != 1 ||
		server.TTLConditions[0].Filters[0].Key != "k8s.namespace.name" ||
		len(server.TTLConditions[0].Filters[0].Values) != 1 ||
		server.TTLConditions[0].Filters[0].Values[0] != "prod" ||
		server.TTLConditions[0].TTLDays != 90 {
		return fmt.Errorf("conditions: %+v", server.TTLConditions)
	}

	return nil
}

func checkRetentionGet() error {
	server := retentiontypes.GetCustomRetentionTTLResponse{
		Version:        "v2",
		Status:         "success",
		DefaultTTLDays: 15,
		TTLConditions: []retentiontypes.CustomRetentionRule{{
			Filters: []retentiontypes.FilterCondition{{Key: "k8s.namespace.name", Values: []string{"prod"}}},
			TTLDays: 90,
		}},
		ColdStorageVolume:  "s3-volume",
		ColdStorageTTLDays: 3,
	}

	raw, err := json.Marshal(server)
	if err != nil {
		return err
	}

	var provider model.CustomRetentionTTLResponseData
	if err := decodeStrict(raw, &provider); err != nil {
		return fmt.Errorf("provider strict decode: %w", err)
	}

	if provider.DefaultTTLDays != 15 || provider.ColdStorageTTLDays != 3 || provider.ColdStorageVolume != "s3-volume" {
		return fmt.Errorf("scalars: %+v", provider)
	}

	if len(provider.TTLConditions) != 1 ||
		len(provider.TTLConditions[0].Filters) != 1 ||
		provider.TTLConditions[0].Filters[0].Key != "k8s.namespace.name" ||
		provider.TTLConditions[0].TTLDays != 90 {
		return fmt.Errorf("conditions: %+v", provider.TTLConditions)
	}

	return nil
}

func checkDashboardPost() error {
	pd := sampleDashboard()

	raw, err := json.Marshal(pd.ToPostable())
	if err != nil {
		return err
	}

	// PostableDashboardV2.UnmarshalJSON decodes with DisallowUnknownFields and
	// runs Validate — an error here is the server rejecting the body.
	var postable dashboardtypes.PostableDashboardV2
	if err := json.Unmarshal(raw, &postable); err != nil {
		return fmt.Errorf("server decode/validate: %w", err)
	}

	return nil
}

func checkDashboardUpdate() error {
	serverGettable := serverDashboardLiteral()

	var server dashboardtypes.GettableDashboardV2
	if err := json.Unmarshal([]byte(serverGettable), &server); err != nil {
		return fmt.Errorf("server literal decode: %w", err)
	}

	serverRaw, err := json.Marshal(server)
	if err != nil {
		return err
	}

	var live model.GettableDashboardV2
	if err := decodeStrict(serverRaw, &live); err != nil {
		return fmt.Errorf("provider strict decode: %w", err)
	}

	pd := sampleDashboard()
	pd.CarryServerPreserved(model.FromGettable(live))

	raw, err := json.Marshal(pd.ToUpdatable("signoz-devops-status"))
	if err != nil {
		return err
	}

	// UpdatableDashboardV2.UnmarshalJSON decodes with DisallowUnknownFields and
	// runs Validate; the server-owned image and links must survive the trip.
	var updatable dashboardtypes.UpdatableDashboardV2
	if err := json.Unmarshal(raw, &updatable); err != nil {
		return fmt.Errorf("server decode/validate: %w", err)
	}

	if updatable.Image != "" {
		return fmt.Errorf("carried image: %q", updatable.Image)
	}

	if len(updatable.Spec.Links) != 1 || updatable.Spec.Links[0].URL != "https://docs.example" {
		return fmt.Errorf("carried spec links: %+v", updatable.Spec.Links)
	}

	panel := updatable.Spec.Panels["abc123def456789a"]
	if panel == nil || len(panel.Spec.Links) != 1 || panel.Spec.Links[0].URL != "https://runbook.example" {
		return fmt.Errorf("carried panel links: %+v", panel)
	}

	return nil
}

func checkDashboardGet() error {
	raw := serverDashboardLiteral()

	var server dashboardtypes.GettableDashboardV2
	if err := json.Unmarshal([]byte(raw), &server); err != nil {
		return fmt.Errorf("server literal decode: %w", err)
	}

	serverRaw, err := json.Marshal(server)
	if err != nil {
		return err
	}

	var providerGettable model.GettableDashboardV2
	if err := decodeStrict(serverRaw, &providerGettable); err != nil {
		return fmt.Errorf("provider strict decode: %w", err)
	}

	got := model.FromGettable(providerGettable)

	if got.Title != "DevOps Status" || got.Image != "" {
		return fmt.Errorf("display/image: %q / %q", got.Title, got.Image)
	}

	if len(got.Tags) != 2 || got.Tags[0] != "devops" || got.Tags[1] != "team:platform" {
		return fmt.Errorf("tags: %#v", got.Tags)
	}

	if len(got.SpecLinks) != 1 || got.SpecLinks[0]["url"] != "https://docs.example" {
		return fmt.Errorf("spec links: %#v", got.SpecLinks)
	}

	panel := got.Pages[0].Panels[0]
	if panel.ID != "abc123def456789a" || panel.Element != "lax-devops-status" {
		return fmt.Errorf("panel: %+v", panel)
	}

	if len(panel.Links) != 1 || panel.Links[0]["targetBlank"] != true {
		return fmt.Errorf("panel links: %#v", panel.Links)
	}

	return nil
}

// serverDashboardLiteral is the JSON the server's GettableDashboardV2 emits
// for a WebComponent-panel dashboard (tags transcribed from
// dashboardtypes/perses_dashboard.go + perses_dashboard_data.go +
// perses_replicas.go). It carries spec-level links and per-panel links — the
// server-owned shapes the provider must tolerate on read.
func serverDashboardLiteral() string {
	return `{
		"id": "0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0002",
		"createdAt": "2026-10-13T10:00:00Z",
		"updatedAt": "2026-10-13T10:00:00Z",
		"createdBy": "user@signoz.io",
		"updatedBy": "user@signoz.io",
		"orgId": "0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0003",
		"locked": false,
		"source": "user",
		"schemaVersion": "v6",
		"image": "",
		"name": "signoz-devops-status",
		"tags": [{"key": "tag", "value": "devops"}, {"key": "team", "value": "platform"}],
		"spec": {
			"display": {"name": "DevOps Status", "description": "Real-time infrastructure and application status dashboard"},
			"variables": [],
			"panels": {
				"abc123def456789a": {
					"kind": "Panel",
					"spec": {
						"display": {"name": "System Status", "description": ""},
						"plugin": {
							"kind": "signoz/WebComponentPanel",
							"spec": {"pluginId": "devops-status", "pluginVersion": "1.0.0", "element": "lax-devops-status", "attributes": {"env": "prod"}, "configKeys": ["env"]}
						},
						"queries": [],
						"links": [{"name": "runbook", "url": "https://runbook.example", "tooltip": "", "renderVariables": false, "targetBlank": true}]
					}
				}
			},
			"layouts": [{"kind": "Grid", "spec": {"items": [{"x": 0, "y": 0, "width": 6, "height": 4, "content": {"$ref": "#/spec/panels/abc123def456789a"}}]}}],
			"links": [{"name": "docs", "url": "https://docs.example", "tooltip": "", "renderVariables": false, "targetBlank": false}]
		}
	}`
}

func checkIngestionKeyPost() error {
	key := model.PostableIngestionKey{Name: "primary", Tags: []string{"team"}}

	raw, err := json.Marshal(key)
	if err != nil {
		return err
	}

	var serverKey gatewaytypes.PostableIngestionKey
	if err := decodeStrict(raw, &serverKey); err != nil {
		return fmt.Errorf("server strict decode (key): %w", err)
	}

	if serverKey.Name != "primary" || len(serverKey.Tags) != 1 {
		return fmt.Errorf("key: %+v", serverKey)
	}

	limit := model.PostableIngestionKeyLimit{
		KeyID:  "0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0004",
		Signal: "logs",
		Config: model.LimitConfig{Day: &model.LimitValue{Count: int64Ptr(100)}},
		Tags:   []string{"team"},
	}

	raw, err = json.Marshal(limit)
	if err != nil {
		return err
	}

	var serverLimit gatewaytypes.PostableIngestionKeyLimit
	if err := decodeStrict(raw, &serverLimit); err != nil {
		return fmt.Errorf("server strict decode (limit): %w", err)
	}

	if serverLimit.KeyID != limit.KeyID || serverLimit.Signal != "logs" ||
		serverLimit.Config.Day == nil || *serverLimit.Config.Day.Count != 100 {
		return fmt.Errorf("limit: %+v", serverLimit)
	}

	updatable := model.UpdatableIngestionKeyLimit{
		Config: model.LimitConfig{
			Day:    &model.LimitValue{Count: int64Ptr(200)},
			Second: &model.LimitValue{Size: int64Ptr(1024)},
		},
		Tags: []string{"team"},
	}

	raw, err = json.Marshal(updatable)
	if err != nil {
		return err
	}

	var serverUpdatable gatewaytypes.UpdatableIngestionKeyLimit
	if err := decodeStrict(raw, &serverUpdatable); err != nil {
		return fmt.Errorf("server strict decode (updatable): %w", err)
	}

	if serverUpdatable.Config.Day == nil || *serverUpdatable.Config.Day.Count != 200 || serverUpdatable.Config.Second == nil {
		return fmt.Errorf("updatable: %+v", serverUpdatable)
	}

	return nil
}

func checkIngestionKeyGet() error {
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	expiry := now.Add(30 * 24 * time.Hour)

	serverKey := gatewaytypes.IngestionKey{
		ID: "0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0004", Name: "primary", Value: "secret-value",
		ExpiresAt: expiry, Tags: []string{"team"}, CreatedAt: now, UpdatedAt: now,
		WorkspaceID: "0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0005",
		Limits: []gatewaytypes.Limit{{
			ID: "limit-1", Signal: "logs",
			Config:    gatewaytypes.LimitConfig{Day: &gatewaytypes.LimitValue{Count: int64Ptr(100)}},
			Tags:      []string{"team"},
			CreatedAt: now, UpdatedAt: now,
			KeyID:  "0197f6d6-a2b4-7cc2-a8c3-5f2b8e9d0004",
			Metric: gatewaytypes.LimitMetric{Day: &gatewaytypes.LimitMetricValue{Count: 12, Size: 345}},
		}},
	}

	raw, err := envelope("success", serverKey)
	if err != nil {
		return err
	}

	var key model.IngestionKey
	if err := unwrapGateway(raw, &key); err != nil {
		return fmt.Errorf("provider strict decode (key): %w", err)
	}

	if key.ID != serverKey.ID || key.Value != "secret-value" || key.WorkspaceID != serverKey.WorkspaceID {
		return fmt.Errorf("key scalars: %+v", key)
	}

	if !key.ExpiresAt.Equal(expiry) || !key.CreatedAt.Equal(now) {
		return fmt.Errorf("key times: %+v", key)
	}

	if len(key.Limits) != 1 || key.Limits[0].KeyID != serverKey.ID ||
		key.Limits[0].Metric.Day == nil || key.Limits[0].Metric.Day.Count != 12 {
		return fmt.Errorf("key limits: %+v", key.Limits)
	}

	list := gatewaytypes.GettableIngestionKeys{
		Keys:       []gatewaytypes.IngestionKey{serverKey},
		Pagination: gatewaytypes.Pagination{Page: 1, PerPage: 50, Pages: 1, Total: 1},
	}

	raw, err = envelope("success", list)
	if err != nil {
		return err
	}

	var listResp struct {
		Status string `json:"status"`
		Data   struct {
			Keys       []model.IngestionKey `json:"keys"`
			Pagination struct {
				Page    int `json:"page"`
				PerPage int `json:"per_page"`
				Pages   int `json:"pages"`
				Total   int `json:"total"`
			} `json:"_pagination"`
		} `json:"data"`
	}
	if err := unwrapGateway(raw, &listResp.Data); err != nil {
		return fmt.Errorf("provider strict decode (list): %w", err)
	}

	if listResp.Data.Pagination.Pages != 1 || len(listResp.Data.Keys) != 1 {
		return fmt.Errorf("list: %+v", listResp.Data)
	}

	createdRaw, err := envelope("success", gatewaytypes.GettableCreatedIngestionKey{ID: "id-1", Value: "secret-value"})
	if err != nil {
		return err
	}

	var created model.CreatedIngestionKey
	if err := unwrapGateway(createdRaw, &created); err != nil {
		return fmt.Errorf("provider strict decode (created key): %w", err)
	}

	if created.ID != "id-1" || created.Value != "secret-value" {
		return fmt.Errorf("created key: %+v", created)
	}

	createdLimitRaw, err := envelope("success", gatewaytypes.GettableCreatedIngestionKeyLimit{ID: "limit-9"})
	if err != nil {
		return err
	}

	var createdLimit model.CreatedIngestionKeyLimit
	if err := unwrapGateway(createdLimitRaw, &createdLimit); err != nil {
		return fmt.Errorf("provider strict decode (created limit): %w", err)
	}

	if createdLimit.ID != "limit-9" {
		return fmt.Errorf("created limit: %+v", createdLimit)
	}

	return nil
}

// unwrapGateway mirrors the client's doGateway: pull the payload out of the
// {"status","data"} envelope, then decode it into out with DisallowUnknownFields.
func unwrapGateway(raw []byte, out any) error {
	var env struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := decodeStrict(raw, &env); err != nil {
		return err
	}

	if env.Status != "success" {
		return fmt.Errorf("unexpected envelope status %q", env.Status)
	}

	return decodeStrict(env.Data, out)
}
