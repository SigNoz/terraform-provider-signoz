package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// testPluginResource builds a pluginDashboardResource wired to an httptest
// backend serving the plugin catalog contract.
func testPluginResource(t *testing.T) *pluginDashboardResource {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeManifest := func(m model.PluginManifest) {
			b, _ := json.Marshal(struct {
				Status string               `json:"status"`
				Data   model.PluginManifest `json:"data"`
			}{Status: "success", Data: m})
			writeTestJSON(w, http.StatusOK, string(b))
		}

		switch {
		case r.URL.Path == "/api/v1/plugins":
			writeTestList(t, w)

		case r.URL.Path == "/api/v1/plugins/fleet-node-grid":
			writeManifest(model.PluginManifest{ID: "fleet-node-grid", Version: "0.2.0", Element: "lax-fleet-node-grid"})

		case r.URL.Path == "/api/v1/plugins/hello" && r.URL.Query().Get("version") == "":
			writeManifest(model.PluginManifest{ID: "hello", Version: "0.3.0", Element: "lax-hello-wc"})

		case r.URL.Path == "/api/v1/plugins/hello" && isKnownHelloVersion(r.URL.Query().Get("version")):
			writeManifest(model.PluginManifest{ID: "hello", Version: r.URL.Query().Get("version"), Element: "lax-hello-wc"})

		default:
			writeTestJSON(w, http.StatusNotFound, `{"status":"error","error":{"code":404,"msg":"not found"}}`)
		}
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	api, err := apiclients.New(srv.URL, "", "test", "", 5*time.Second, 1)
	if err != nil {
		t.Fatalf("build wrapped client: %v", err)
	}

	client, err := legacyclient.NewClient(api)
	if err != nil {
		t.Fatalf("build legacy client: %v", err)
	}

	return &pluginDashboardResource{client: client}
}

func isKnownHelloVersion(v string) bool {
	return v == "0.2.0" || v == "0.3.0"
}

func writeTestJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

func writeTestList(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	b, _ := json.Marshal(struct {
		Status string                 `json:"status"`
		Data   []model.PluginManifest `json:"data"`
	}{Status: "success", Data: []model.PluginManifest{
		{ID: "hello", Version: "0.2.0", Element: "lax-hello-wc"},
		{ID: "hello", Version: "0.3.0", Element: "lax-hello-wc"},
	}})

	writeTestJSON(w, http.StatusOK, string(b))
}

func TestResolvePluginVersions_AutoFillsLatestAndElement(t *testing.T) {
	r := testPluginResource(t)

	pd := model.PluginDashboard{
		Title: "Board",
		Pages: []model.PluginPage{{Name: "P", Panels: []model.PluginPanel{
			{Title: "Greeting", PluginID: "hello"},
		}}},
	}

	if err := r.resolvePluginVersions(context.Background(), &pd); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	panel := pd.Pages[0].Panels[0]
	if panel.PluginVersion != "0.3.0" {
		t.Fatalf("auto version: %q (server's last-registered pick)", panel.PluginVersion)
	}
	if panel.Element != "lax-hello-wc" {
		t.Fatalf("element: %q", panel.Element)
	}
}

func TestResolvePluginVersions_ExplicitVersionAndElementGuard(t *testing.T) {
	r := testPluginResource(t)

	pd := model.PluginDashboard{
		Title: "Board",
		Pages: []model.PluginPage{{Name: "P", Panels: []model.PluginPanel{
			{Title: "Grid", PluginID: "fleet-node-grid", PluginVersion: "0.2.0", Element: "lax-fleet-node-grid"},
		}}},
	}

	if err := r.resolvePluginVersions(context.Background(), &pd); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	panel := pd.Pages[0].Panels[0]
	if panel.PluginVersion != "0.2.0" || panel.Element != "lax-fleet-node-grid" {
		t.Fatalf("panel: %+v", panel)
	}

	// Mismatched element is an error naming both values.
	pd.Pages[0].Panels[0].Element = "lax-wrong"
	err := r.resolvePluginVersions(context.Background(), &pd)
	if err == nil || !strings.Contains(err.Error(), "lax-wrong") || !strings.Contains(err.Error(), "lax-fleet-node-grid") {
		t.Fatalf("element mismatch error: %v", err)
	}
}

func TestResolvePluginVersions_MissingPluginListsCatalog(t *testing.T) {
	r := testPluginResource(t)

	pd := model.PluginDashboard{
		Title: "Board",
		Pages: []model.PluginPage{{Name: "P", Panels: []model.PluginPanel{
			{Title: "Ghost", PluginID: "no-such-plugin"},
		}}},
	}

	err := r.resolvePluginVersions(context.Background(), &pd)
	if err == nil {
		t.Fatal("expected error for missing plugin")
	}

	// The error must name the plugin and the available catalog entries with
	// their versions, so a typo is fixable from the plan alone.
	if !strings.Contains(err.Error(), "no-such-plugin") ||
		!strings.Contains(err.Error(), "hello (0.2.0, 0.3.0)") {
		t.Fatalf("error should list catalog: %v", err)
	}
}

func TestResolvePluginVersions_MissingVersionListsVersions(t *testing.T) {
	r := testPluginResource(t)

	pd := model.PluginDashboard{
		Title: "Board",
		Pages: []model.PluginPage{{Name: "P", Panels: []model.PluginPanel{
			{Title: "Greeting", PluginID: "hello", PluginVersion: "9.9.9"},
		}}},
	}

	err := r.resolvePluginVersions(context.Background(), &pd)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "9.9.9") || !strings.Contains(err.Error(), "hello (0.2.0, 0.3.0)") {
		t.Fatalf("error should list versions: %v", err)
	}
}

func gettableForStateTests() *model.GettableDashboardV2 {
	g := &model.GettableDashboardV2{
		ID:   "dash-1",
		Name: "board",
		Tags: []model.GettableTag{{Key: "tag", Value: "lacrosse"}},
	}
	g.Spec.Display = model.DisplayV2{Name: "Board"}
	g.Spec.Panels = map[string]model.PanelV2{
		"panel-a": {
			Kind: model.PanelEnvelopeKind,
			Spec: model.PanelSpecV2{
				Display: model.DisplayV2{Name: "Greeting"},
				Plugin: model.PanelPluginV2{
					Kind: model.PanelKindWebComponent,
					Spec: model.WebComponentPanelSpec{
						PluginID:      "hello",
						PluginVersion: "0.3.0",
						Element:       "lax-hello-wc",
					},
				},
				Queries: []map[string]any{},
			},
		},
	}
	g.Spec.Layouts = []model.LayoutV2{{
		Kind: "Grid",
		Spec: model.GridLayoutSpecV2{
			Display: &model.GridLayoutDisplayV2{Title: "Overview"},
			Items: []model.GridItemV2{{
				X: 0, Y: 0, Width: 6, Height: 4,
				Content: model.JSONPointerV2{Ref: "#/spec/panels/panel-a"},
			}},
		},
	}}

	return g
}

func TestStateFromGettable_NormalizesEmptyCollections(t *testing.T) {
	state := stateFromGettable(gettableForStateTests())

	if state.ID.ValueString() != "dash-1" || state.Name.ValueString() != "board" {
		t.Fatalf("identity: %+v", state)
	}
	if state.Title.ValueString() != "Board" {
		t.Fatalf("title: %s", state.Title)
	}
	if !state.Description.IsNull() {
		t.Fatalf("empty description must be null, got %s", state.Description)
	}
	if state.Tags.IsNull() || len(state.Tags.Elements()) != 1 {
		t.Fatalf("tags: %+v", state.Tags)
	}

	if len(state.Pages) != 1 || state.Pages[0].Name.ValueString() != "Overview" {
		t.Fatalf("pages: %+v", state.Pages)
	}

	panel := state.Pages[0].Panels[0]
	if panel.ID.ValueString() != "panel-a" {
		t.Fatalf("panel id must be the server map key: %s", panel.ID)
	}
	if panel.PluginVersion.ValueString() != "0.3.0" || panel.Element.ValueString() != "lax-hello-wc" {
		t.Fatalf("panel: %+v", panel)
	}
	if !panel.Attributes.IsNull() || !panel.Properties.IsNull() || !panel.ConfigKeys.IsNull() {
		t.Fatalf("empty collections must be null: attrs=%v props=%v keys=%v",
			panel.Attributes, panel.Properties, panel.ConfigKeys)
	}
	if panel.X.ValueInt64() != 0 || panel.W.ValueInt64() != 6 || panel.H.ValueInt64() != 4 {
		t.Fatalf("geometry: %+v", panel)
	}
	if state.Document.IsNull() || !strings.Contains(state.Document.ValueString(), `"generateName": true`) {
		t.Fatalf("document: %s", state.Document)
	}
}

func TestStateFromGettable_PropertiesAndAttributesPopulated(t *testing.T) {
	g := gettableForStateTests()

	panel := g.Spec.Panels["panel-a"]
	panel.Spec.Plugin.Spec.Attributes = map[string]string{"data-source": "stub"}
	panel.Spec.Plugin.Spec.Properties = map[string]any{
		"mode": "stub",
		"stub": map[string]any{"x": float64(1)},
	}
	panel.Spec.Plugin.Spec.ConfigKeys = []string{"endpoints"}
	g.Spec.Panels["panel-a"] = panel

	state := stateFromGettable(g)
	panelModel := state.Pages[0].Panels[0]

	var attrs map[string]string
	if err := panelModel.Attributes.ElementsAs(context.Background(), &attrs, false); err != nil {
		t.Fatal(err)
	}
	if attrs["data-source"] != "stub" {
		t.Fatalf("attributes: %+v", attrs)
	}
	if panelModel.Properties.ValueString() != `{"mode":"stub","stub":{"x":1}}` {
		t.Fatalf("properties normalized: %s", panelModel.Properties)
	}
	if len(panelModel.ConfigKeys.Elements()) != 1 {
		t.Fatalf("config keys: %+v", panelModel.ConfigKeys)
	}
}

func stringList(vals ...string) types.List {
	listVals := make([]attr.Value, 0, len(vals))
	for _, v := range vals {
		listVals = append(listVals, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, listVals)
}

func configPanel(title string, x, y, w, h types.Int64) pluginPanelModel {
	return pluginPanelModel{
		Title: types.StringValue(title),
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
	}
}

func TestValidatePluginDashboardConfig_DefaultedGeometryIsValid(t *testing.T) {
	ctx := context.Background()

	config := pluginDashboardModel{
		Tags: types.ListNull(types.StringType),
		Pages: []pluginPageModel{{
			Name: types.StringValue("Overview"),
			// x/y/w/h omitted in config: null here, filled by plan defaults.
			Panels: []pluginPanelModel{
				configPanel("A", types.Int64Null(), types.Int64Null(), types.Int64Null(), types.Int64Null()),
				configPanel("B", types.Int64Value(6), types.Int64Null(), types.Int64Null(), types.Int64Null()),
			},
		}},
	}

	diags := validatePluginDashboardConfig(ctx, config)
	if diags.HasError() {
		t.Fatalf("omitted geometry must fall back to builder defaults, got: %v", diags)
	}
}

func TestValidatePluginDashboardConfig_GridViolations(t *testing.T) {
	ctx := context.Background()

	config := pluginDashboardModel{
		Tags: types.ListNull(types.StringType),
		Pages: []pluginPageModel{{
			Name: types.StringValue("Overview"),
			Panels: []pluginPanelModel{
				configPanel("Zero", types.Int64Value(0), types.Int64Value(0), types.Int64Value(0), types.Int64Value(4)),
				configPanel("Wide", types.Int64Value(8), types.Int64Value(0), types.Int64Value(6), types.Int64Value(4)),
				configPanel("Negative", types.Int64Value(-1), types.Int64Value(0), types.Int64Value(2), types.Int64Value(2)),
			},
		}},
	}

	diags := validatePluginDashboardConfig(ctx, config)
	if !diags.HasError() {
		t.Fatal("expected grid violations")
	}

	paths := map[string]bool{}
	for _, d := range diags.Errors() {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok {
			paths[withPath.Path().String()] = true
		}
	}

	for _, want := range []string{
		`page[0].plugin_panel[0].w`, // w=0
		`page[0].plugin_panel[1].w`, // x+w = 14 > 12
		`page[0].plugin_panel[2].x`, // x=-1
	} {
		if !paths[want] {
			t.Fatalf("expected error at %s, got: %v", want, diags)
		}
	}
}

func TestValidatePluginDashboardConfig_Overlap(t *testing.T) {
	ctx := context.Background()

	config := pluginDashboardModel{
		Tags: types.ListNull(types.StringType),
		Pages: []pluginPageModel{{
			Name: types.StringValue("Overview"),
			Panels: []pluginPanelModel{
				configPanel("A", types.Int64Value(0), types.Int64Value(0), types.Int64Value(6), types.Int64Value(4)),
				configPanel("B", types.Int64Value(4), types.Int64Value(2), types.Int64Value(6), types.Int64Value(4)),
			},
		}},
	}

	diags := validatePluginDashboardConfig(ctx, config)
	if !diags.HasError() {
		t.Fatal("expected overlap error")
	}

	var msg string
	for _, d := range diags.Errors() {
		if d.Summary() == "Overlapping panels" {
			msg = d.Detail()
		}
	}
	if msg == "" {
		t.Fatalf("expected an Overlapping panels diagnostic, got: %v", diags)
	}
	if !strings.Contains(msg, `"A"`) || !strings.Contains(msg, `"B"`) {
		t.Fatalf("overlap error must name both panels: %s", msg)
	}
}

func TestValidatePluginDashboardConfig_Tags(t *testing.T) {
	ctx := context.Background()

	// Plain strings map to key "tag" — "name" as a value is fine.
	clean := pluginDashboardModel{
		Tags:  stringList("name", "team:platform"),
		Pages: []pluginPageModel{{Name: types.StringValue("P"), Panels: []pluginPanelModel{configPanel("A", types.Int64Null(), types.Int64Null(), types.Int64Null(), types.Int64Null())}}},
	}
	if diags := validatePluginDashboardConfig(ctx, clean); diags.HasError() {
		t.Fatalf("plain strings never collide with reserved keys: %v", diags)
	}

	// "LOCKED:x" must trip the case-insensitive reserved check.
	reserved := pluginDashboardModel{
		Tags:  stringList("team:platform", "LOCKED:x"),
		Pages: []pluginPageModel{{Name: types.StringValue("P"), Panels: []pluginPanelModel{configPanel("A", types.Int64Null(), types.Int64Null(), types.Int64Null(), types.Int64Null())}}},
	}
	diags := validatePluginDashboardConfig(ctx, reserved)
	if !diags.HasError() {
		t.Fatal("expected reserved tag key error")
	}
	found := false
	for _, d := range diags.Errors() {
		withPath, ok := d.(diag.DiagnosticWithPath)
		if ok && withPath.Path().String() == "tags[1]" && strings.Contains(d.Detail(), "locked") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected reserved-key error at tags[1]: %v", diags)
	}

	// More than 10 tags is rejected.
	many := make([]string, 11)
	for i := range many {
		many[i] = fmt.Sprintf("k%d:v", i)
	}
	tooMany := pluginDashboardModel{
		Tags:  stringList(many...),
		Pages: []pluginPageModel{{Name: types.StringValue("P"), Panels: []pluginPanelModel{configPanel("A", types.Int64Null(), types.Int64Null(), types.Int64Null(), types.Int64Null())}}},
	}
	diags = validatePluginDashboardConfig(ctx, tooMany)
	if !diags.HasError() {
		t.Fatalf("expected too-many-tags error, got: %v", diags)
	}
	found = false
	for _, d := range diags.Errors() {
		if strings.Contains(d.Summary(), "Too many tags") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Too many tags summary, got: %v", diags)
	}
}

func TestUpdatableForUpdate_CarriesServerOwnedFields(t *testing.T) {
	server := gettableForStateTests()
	server.Image = "icon-dashboards/fleet.svg"
	panel := server.Spec.Panels["panel-a"]
	panel.Spec.Links = []map[string]any{{"url": "https://status.example"}}
	server.Spec.Panels["panel-a"] = panel

	pd := model.PluginDashboard{
		Title: "Renamed board",
		Pages: []model.PluginPage{{Name: "Overview", Panels: []model.PluginPanel{
			{ID: "panel-a", Title: "Greeting", PluginID: "hello", Element: "lax-hello-wc", W: 6, H: 4},
		}}},
	}

	up := updatableForUpdate(pd, server, "board")

	if up.Image != "icon-dashboards/fleet.svg" {
		t.Fatalf("image not carried: %q", up.Image)
	}
	if up.Name != "board" {
		t.Fatalf("name must come from state: %q", up.Name)
	}
	if up.Spec.Display.Name != "Renamed board" {
		t.Fatalf("plan title must win: %+v", up.Spec.Display)
	}
	if links := up.Spec.Panels["panel-a"].Spec.Links; len(links) != 1 || links[0]["url"] != "https://status.example" {
		t.Fatalf("links not carried: %#v", links)
	}
}
