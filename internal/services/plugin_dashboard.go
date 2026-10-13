package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apiclients"
	"github.com/SigNoz/terraform-provider-signoz/internal/legacyclient"
	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                   = &pluginDashboardResource{}
	_ resource.ResourceWithConfigure      = &pluginDashboardResource{}
	_ resource.ResourceWithImportState    = &pluginDashboardResource{}
	_ resource.ResourceWithValidateConfig = &pluginDashboardResource{}
)

// NewPluginDashboardResource manages SigNoz v2 dashboards
// (schemaVersion v6, POST/PUT /api/v2/dashboards) composed of WebComponent
// plugin panels (signoz/WebComponentPanel).
func NewPluginDashboardResource() resource.Resource {
	return &pluginDashboardResource{}
}

type pluginDashboardResource struct {
	client *legacyclient.Client
}

type pluginDashboardModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	Title       types.String      `tfsdk:"title"`
	Description types.String      `tfsdk:"description"`
	Tags        types.List        `tfsdk:"tags"`
	Pages       []pluginPageModel `tfsdk:"page"`
	// Canonical POST body JSON (computed) for debugging / drift inspection.
	Document types.String `tfsdk:"document"`
}

type pluginPageModel struct {
	Name   types.String       `tfsdk:"name"`
	Panels []pluginPanelModel `tfsdk:"plugin_panel"`
}

type pluginPanelModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	PluginID      types.String `tfsdk:"plugin_id"`
	PluginVersion types.String `tfsdk:"plugin_version"`
	Element       types.String `tfsdk:"element"`
	Attributes    types.Map    `tfsdk:"attributes"`
	Properties    types.String `tfsdk:"properties"` // JSON object string
	ConfigKeys    types.List   `tfsdk:"config_keys"`
	X             types.Int64  `tfsdk:"x"`
	Y             types.Int64  `tfsdk:"y"`
	W             types.Int64  `tfsdk:"w"`
	H             types.Int64  `tfsdk:"h"`
}

func (r *pluginDashboardResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = SigNozPluginDashboard
}

func (r *pluginDashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, ok := req.ProviderData.(*apiclients.WrappedClient)
	if !ok {
		addErr(&resp.Diagnostics, fmt.Errorf("expected *apiclients.WrappedClient, got %T", req.ProviderData), operationConfigure, SigNozPluginDashboard)
		return
	}

	legacyClient, err := legacyclient.NewClient(api)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build legacy SigNoz client", err.Error())
		return
	}

	r.client = legacyClient
}

func (r *pluginDashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyList, _ := types.ListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a SigNoz **v2** dashboard (Perses spec, `schemaVersion v6`) populated with " +
			"**WebComponent plugin panels** (`signoz/WebComponentPanel`). Panel versions and elements are " +
			"resolved against the server plugin catalog at apply time. Use this to manage plugin-host " +
			"dashboards as code — see the shipped devops-status example " +
			"(`plugin_id = \"devops-status\"`, `element = \"lax-devops-status\"`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Dashboard UUID from SigNoz.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server-generated DNS name (immutable once created; regeneration requires replacement).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Dashboard title shown in SigNoz UI.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Dashboard description.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(emptyList),
				MarkdownDescription: "Tags. A plain string maps to tag key `tag`; use `key:value` for any other key.",
			},
			"document": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Canonical v2 dashboard JSON document (the POST body) for inspection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"page": schema.ListNestedBlock{
				MarkdownDescription: "Dashboard page; maps to one Grid layout. At least one `page` block is required — the framework lists blocks as optional, but a plan with no page fails validation.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Layout title (e.g. Overview, DevTime).",
						},
					},
					Blocks: map[string]schema.Block{
						"plugin_panel": schema.ListNestedBlock{
							MarkdownDescription: "WebComponent plugin panel on this page.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Panel id (the server panels-map key). Auto-generated when omitted and kept in state for stability.",
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"title": schema.StringAttribute{
										Required: true,
									},
									"plugin_id": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Registered plugin id in the server catalog (GET /api/v1/plugins).",
									},
									"plugin_version": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Plugin version to pin. When omitted, the server catalog's default version is used and recorded here.",
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"element": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Custom element tag (must contain a hyphen). Filled from the plugin catalog manifest when omitted; a mismatch with the manifest is an error.",
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"attributes": schema.MapAttribute{
										ElementType:         types.StringType,
										Optional:            true,
										MarkdownDescription: "HTML attributes on the custom element.",
									},
									"properties": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "JSON object of JS properties for the element. Stored normalized (sorted keys, compact); formatting differences in config show as cosmetic diffs.",
									},
									"config_keys": schema.ListAttribute{
										ElementType:         types.StringType,
										Optional:            true,
										MarkdownDescription: "Config keys this panel instance needs resolved.",
									},
									"x": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(0),
									},
									"y": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(0),
									},
									"w": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(6),
									},
									"h": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(4),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *pluginDashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pluginDashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pd, diags := planToPluginDashboard(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.resolvePluginVersions(ctx, &pd); err != nil {
		resp.Diagnostics.AddError("Error resolving plugin catalog", err.Error())
		return
	}

	created, err := r.client.CreateDashboardV2(ctx, pd.ToPostable())
	if err != nil {
		resp.Diagnostics.AddError("Error creating plugin dashboard", err.Error())
		return
	}

	state := stateFromGettable(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pluginDashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pluginDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetDashboardV2(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err, operationRead, SigNozPluginDashboard)
		return
	}

	refreshed := stateFromGettable(got)
	// Keep the dashboard id even when the server echoes it empty (defensive).
	refreshed.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

func (r *pluginDashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pluginDashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state pluginDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pd, diags := planToPluginDashboard(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.resolvePluginVersions(ctx, &pd); err != nil {
		resp.Diagnostics.AddError("Error resolving plugin catalog", err.Error())
		return
	}

	server, err := r.client.GetDashboardV2(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading plugin dashboard before update", err.Error())
		return
	}

	// name is immutable server-side: always send the one from state.
	updated, err := r.client.UpdateDashboardV2(ctx, state.ID.ValueString(), updatableForUpdate(pd, server, state.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error updating plugin dashboard", err.Error())
		return
	}

	refreshed := stateFromGettable(updated)
	refreshed.ID = state.ID
	refreshed.Name = state.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

// updatableForUpdate builds the PUT body from the plan, carrying the
// server-owned fields — the dashboard image and each panel's links — from the
// live dashboard so the update does not blank them.
func updatableForUpdate(pd model.PluginDashboard, server *model.GettableDashboardV2, name string) model.UpdatableDashboardV2 {
	pd.CarryServerPreserved(model.FromGettable(*server))
	return pd.ToUpdatable(name)
}

func (r *pluginDashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pluginDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDashboardV2(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addErr(&resp.Diagnostics, err, operationDelete, SigNozPluginDashboard)
	}
}

func (r *pluginDashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig rejects grid and tag violations before apply. The server
// rejects them at write time; failing at plan time keeps applies clean.
func (r *pluginDashboardResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config pluginDashboardModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validatePluginDashboardConfig(ctx, config)...)
}

const (
	pluginDashboardGridColumns   = 12
	pluginDashboardMaxPanelsPage = 100
	pluginDashboardMaxPages      = 500
	pluginDashboardMaxTags       = 10
)

// reservedTagKeys mirrors the server's reservedDSLKeys: a tag key colliding
// with a list-query column name would make the dashboard DSL ambiguous, so
// the server rejects them case-insensitively at write time.
var reservedTagKeys = map[string]struct{}{
	"name":        {},
	"description": {},
	"created_at":  {},
	"updated_at":  {},
	"created_by":  {},
	"locked":      {},
	"source":      {},
}

func validatePluginDashboardConfig(ctx context.Context, config pluginDashboardModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(config.Pages) > pluginDashboardMaxPages {
		diags.AddAttributeError(
			path.Root("page"),
			"Too many pages",
			fmt.Sprintf("Dashboard has %d pages, but at most %d are allowed.", len(config.Pages), pluginDashboardMaxPages),
		)
	}

	if !config.Tags.IsNull() && !config.Tags.IsUnknown() {
		var tags []string
		diags.Append(config.Tags.ElementsAs(ctx, &tags, false)...)

		if len(tags) > pluginDashboardMaxTags {
			diags.AddAttributeError(
				path.Root("tags"),
				"Too many tags",
				fmt.Sprintf("Dashboard has %d tags, but at most %d are allowed.", len(tags), pluginDashboardMaxTags),
			)
		}

		for i, tag := range tags {
			c := strings.Index(tag, ":")
			if c < 0 {
				// Plain strings map to the server key "tag", never reserved.
				continue
			}

			key := strings.ToLower(strings.TrimSpace(tag[:c]))
			if _, reserved := reservedTagKeys[key]; reserved {
				diags.AddAttributeError(
					path.Root("tags").AtListIndex(i),
					"Reserved tag key",
					fmt.Sprintf("Tag key %q is reserved by the server (reserved: name, description, created_at, updated_at, created_by, locked, source).", key),
				)
			}
		}
	}

	for pi, page := range config.Pages {
		if len(page.Panels) > pluginDashboardMaxPanelsPage {
			diags.AddAttributeError(
				pagePath(pi),
				"Too many panels",
				fmt.Sprintf("Page %q has %d panels, but at most %d are allowed.", page.Name.ValueString(), len(page.Panels), pluginDashboardMaxPanelsPage),
			)
		}

		rects := make([]gridRect, len(page.Panels))
		known := make([]bool, len(page.Panels))
		for fi, panel := range page.Panels {
			rects[fi], known[fi] = panel.gridRect()
			if !known[fi] {
				continue
			}

			g := rects[fi]
			pageName := page.Name.ValueString()
			title := panel.Title.ValueString()

			if g.w < 1 {
				diags.AddAttributeError(panelPath(pi, fi, "w"), "Invalid panel width",
					fmt.Sprintf("Panel %q on page %q has width %d; width must be at least 1.", title, pageName, g.w))
			}
			if g.h < 1 {
				diags.AddAttributeError(panelPath(pi, fi, "h"), "Invalid panel height",
					fmt.Sprintf("Panel %q on page %q has height %d; height must be at least 1.", title, pageName, g.h))
			}
			if g.x < 0 {
				diags.AddAttributeError(panelPath(pi, fi, "x"), "Invalid panel x",
					fmt.Sprintf("Panel %q on page %q has x %d; x must be at least 0.", title, pageName, g.x))
			}
			if g.y < 0 {
				diags.AddAttributeError(panelPath(pi, fi, "y"), "Invalid panel y",
					fmt.Sprintf("Panel %q on page %q has y %d; y must be at least 0.", title, pageName, g.y))
			}
			if g.x+g.w > pluginDashboardGridColumns {
				diags.AddAttributeError(panelPath(pi, fi, "w"), "Panel exceeds grid width",
					fmt.Sprintf("Panel %q on page %q spans x=%d w=%d, past the %d-column grid (x + w <= %d).",
						title, pageName, g.x, g.w, pluginDashboardGridColumns, pluginDashboardGridColumns))
			}
		}

		for i := 0; i < len(rects); i++ {
			if !known[i] {
				continue
			}
			for j := i + 1; j < len(rects); j++ {
				if !known[j] {
					continue
				}
				a, b := rects[i], rects[j]

				if a.x < b.x+b.w && b.x < a.x+a.w && a.y < b.y+b.h && b.y < a.y+a.h {
					diags.AddAttributeError(panelPath(pi, j, ""), "Overlapping panels",
						fmt.Sprintf("Panel %q [x=%d y=%d w=%d h=%d] overlaps panel %q [x=%d y=%d w=%d h=%d] on page %q.",
							config.Pages[pi].Panels[j].Title.ValueString(), b.x, b.y, b.w, b.h,
							config.Pages[pi].Panels[i].Title.ValueString(), a.x, a.y, a.w, a.h,
							config.Pages[pi].Name.ValueString()))
				}
			}
		}
	}

	return diags
}

type gridRect struct {
	x, y, w, h int64
}

// gridRect resolves a panel's rectangle the way the plan will: omitted
// attributes fall back to the schema defaults, which only apply at plan time.
// known is false while any attribute is still unknown — validation waits for
// the plan instead of guessing.
func (panel pluginPanelModel) gridRect() (gridRect, bool) {
	g := gridRect{w: 6, h: 4}

	for _, c := range []struct {
		v   types.Int64
		def int64
		out *int64
	}{
		{panel.X, 0, &g.x},
		{panel.Y, 0, &g.y},
		{panel.W, 6, &g.w},
		{panel.H, 4, &g.h},
	} {
		switch {
		case c.v.IsNull():
			*c.out = c.def
		case c.v.IsUnknown():
			return gridRect{}, false
		default:
			*c.out = c.v.ValueInt64()
		}
	}

	return g, true
}

func pagePath(pageIdx int) path.Path {
	return path.Root("page").AtListIndex(pageIdx)
}

func panelPath(pageIdx, panelIdx int, attr string) path.Path {
	p := pagePath(pageIdx).AtName("plugin_panel").AtListIndex(panelIdx)
	if attr != "" {
		p = p.AtName(attr)
	}
	return p
}

// resolvePluginVersions pins every panel against the server plugin catalog:
// an explicit panel version must exist, otherwise the catalog's default
// manifest fills the version (and the element when the panel omits it).
func (r *pluginDashboardResource) resolvePluginVersions(ctx context.Context, pd *model.PluginDashboard) error {
	for pi := range pd.Pages {
		for fi := range pd.Pages[pi].Panels {
			panel := &pd.Pages[pi].Panels[fi]

			var (
				manifest *model.PluginManifest
				err      error
			)

			if panel.PluginVersion != "" {
				manifest, err = r.client.GetPluginVersion(ctx, panel.PluginID, panel.PluginVersion)
			} else {
				manifest, err = r.client.GetPlugin(ctx, panel.PluginID)
			}

			if err != nil {
				if isNotFound(err) {
					return fmt.Errorf("panel %q (plugin %q): %w", panel.Title, panel.PluginID, r.pluginMissingErr(ctx, panel))
				}
				return fmt.Errorf("panel %q (plugin %q): %w", panel.Title, panel.PluginID, err)
			}

			if manifest.Element != "" {
				switch {
				case panel.Element == "":
					panel.Element = manifest.Element
				case panel.Element != manifest.Element:
					return fmt.Errorf("panel %q (plugin %q): element %q does not match catalog element %q",
						panel.Title, panel.PluginID, panel.Element, manifest.Element)
				}
			}

			if manifest.Version != "" {
				panel.PluginVersion = manifest.Version
			}
		}
	}

	tflog.Debug(ctx, "Resolved plugin versions", map[string]any{"title": pd.Title})

	return nil
}

// pluginMissingErr lists the plugin ids (with versions) the server offers, to
// make a bad plugin_id or typo'd version diagnosable from the plan alone.
func (r *pluginDashboardResource) pluginMissingErr(ctx context.Context, panel *model.PluginPanel) error {
	plugins, listErr := r.client.ListPlugins(ctx)
	if listErr != nil {
		return fmt.Errorf("plugin %q version %q not found in catalog", panel.PluginID, panel.PluginVersion)
	}

	byID := map[string][]string{}
	for _, p := range plugins {
		byID[p.ID] = append(byID[p.ID], p.Version)
	}

	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		versions := byID[id]
		sort.Strings(versions)
		parts = append(parts, fmt.Sprintf("%s (%s)", id, strings.Join(versions, ", ")))
	}

	return fmt.Errorf("plugin %q version %q not found in catalog; available: %s",
		panel.PluginID, panel.PluginVersion, strings.Join(parts, ", "))
}

func isNotFound(err error) bool {
	var apiErr *apiclients.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

func planToPluginDashboard(ctx context.Context, plan pluginDashboardModel) (model.PluginDashboard, diag.Diagnostics) {
	var diags diag.Diagnostics

	pd := model.PluginDashboard{
		Title:       plan.Title.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		diags.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		pd.Tags = tags
	}

	if len(plan.Pages) == 0 {
		diags.AddError("Invalid plugin dashboard", "At least one page { ... } block is required.")
		return model.PluginDashboard{}, diags
	}

	for _, page := range plan.Pages {
		pp := model.PluginPage{Name: page.Name.ValueString()}

		if len(page.Panels) == 0 {
			diags.AddError("Invalid page", fmt.Sprintf("page %q needs at least one plugin_panel", pp.Name))
			continue
		}

		for _, panel := range page.Panels {
			var attrs map[string]string
			if !panel.Attributes.IsNull() && !panel.Attributes.IsUnknown() {
				diags.Append(panel.Attributes.ElementsAs(ctx, &attrs, false)...)
			}

			var props map[string]interface{}
			if !panel.Properties.IsNull() && !panel.Properties.IsUnknown() && panel.Properties.ValueString() != "" {
				if err := json.Unmarshal([]byte(panel.Properties.ValueString()), &props); err != nil {
					diags.AddError("Invalid properties JSON", err.Error())
				}
			}

			var cfgKeys []string
			if !panel.ConfigKeys.IsNull() && !panel.ConfigKeys.IsUnknown() {
				diags.Append(panel.ConfigKeys.ElementsAs(ctx, &cfgKeys, false)...)
			}

			pp.Panels = append(pp.Panels, model.PluginPanel{
				ID:            panel.ID.ValueString(),
				Title:         panel.Title.ValueString(),
				PluginID:      panel.PluginID.ValueString(),
				PluginVersion: panel.PluginVersion.ValueString(),
				Element:       panel.Element.ValueString(),
				Attributes:    attrs,
				Properties:    props,
				ConfigKeys:    cfgKeys,
				X:             int(panel.X.ValueInt64()),
				Y:             int(panel.Y.ValueInt64()),
				W:             int(panel.W.ValueInt64()),
				H:             int(panel.H.ValueInt64()),
			})
		}

		pd.Pages = append(pd.Pages, pp)
	}

	if diags.HasError() {
		return model.PluginDashboard{}, diags
	}

	return pd, diags
}

// stateFromGettable rebuilds TF state from a server dashboard: title and
// description from spec.display, pages from layouts, panels keyed by their
// server map key. Empty collections normalize to null so unset config matches.
func stateFromGettable(g *model.GettableDashboardV2) pluginDashboardModel {
	pd := model.FromGettable(*g)

	state := pluginDashboardModel{
		ID:          types.StringValue(g.ID),
		Name:        types.StringValue(g.Name),
		Title:       types.StringValue(pd.Title),
		Description: types.StringNull(),
		Tags:        types.ListValueMust(types.StringType, []attr.Value{}),
	}

	if pd.Description != "" {
		state.Description = types.StringValue(pd.Description)
	}

	tagVals := make([]attr.Value, 0, len(pd.Tags))
	for _, t := range pd.Tags {
		tagVals = append(tagVals, types.StringValue(t))
	}
	if len(tagVals) > 0 {
		state.Tags = types.ListValueMust(types.StringType, tagVals)
	}

	doc, err := pd.DocumentJSON()
	if err == nil {
		state.Document = types.StringValue(doc)
	}

	for _, page := range pd.Pages {
		pm := pluginPageModel{Name: types.StringValue(page.Name)}

		for _, panel := range page.Panels {
			panelModel := pluginPanelModel{
				ID:            types.StringValue(panel.ID),
				Title:         types.StringValue(panel.Title),
				PluginID:      types.StringValue(panel.PluginID),
				PluginVersion: types.StringNull(),
				Element:       types.StringValue(panel.Element),
				Attributes:    types.MapNull(types.StringType),
				Properties:    types.StringNull(),
				ConfigKeys:    types.ListNull(types.StringType),
				X:             types.Int64Value(int64(panel.X)),
				Y:             types.Int64Value(int64(panel.Y)),
				W:             types.Int64Value(int64(panel.W)),
				H:             types.Int64Value(int64(panel.H)),
			}

			if panel.PluginVersion != "" {
				panelModel.PluginVersion = types.StringValue(panel.PluginVersion)
			}

			if len(panel.Attributes) > 0 {
				attrVals := map[string]attr.Value{}
				for k, v := range panel.Attributes {
					attrVals[k] = types.StringValue(v)
				}
				panelModel.Attributes = types.MapValueMust(types.StringType, attrVals)
			}

			if props := model.PropertiesJSON(panel.Properties); props != "" {
				panelModel.Properties = types.StringValue(props)
			}

			if len(panel.ConfigKeys) > 0 {
				keyVals := make([]attr.Value, 0, len(panel.ConfigKeys))
				for _, k := range panel.ConfigKeys {
					keyVals = append(keyVals, types.StringValue(k))
				}
				panelModel.ConfigKeys = types.ListValueMust(types.StringType, keyVals)
			}

			pm.Panels = append(pm.Panels, panelModel)
		}

		state.Pages = append(state.Pages, pm)
	}

	return state
}
