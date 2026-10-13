package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Contract constants mirroring SigNoz v2 dashboards
// (pkg/types/dashboardtypes/perses_dashboard.go): schemaVersion "v6", the
// panel envelope kind "Panel", the WebComponent plugin envelope kind, and the
// key plain TF tag strings map to.
const (
	DashboardSchemaVersionV2 = "v6"
	PanelEnvelopeKind        = "Panel"
	PanelKindWebComponent    = "signoz/WebComponentPanel"
	DefaultTagKey            = "tag"
)

// PluginDashboard is a high-level dashboard built from plugin pages/panels.
// It is expanded into a SigNoz v2 dashboard (Perses spec) via ToPostable and
// rebuilt from a server response via FromGettable.
type PluginDashboard struct {
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Pages       []PluginPage `json:"pages"`
}

// PluginPage groups plugin panels (one Grid layout on the server).
type PluginPage struct {
	Name   string        `json:"name"`
	Panels []PluginPanel `json:"panels"`
}

// PluginPanel is one WebComponent plugin instance on a page.
type PluginPanel struct {
	ID            string                 `json:"id,omitempty"`
	Title         string                 `json:"title"`
	PluginID      string                 `json:"plugin_id"`
	PluginVersion string                 `json:"plugin_version,omitempty"`
	Element       string                 `json:"element"`
	Attributes    map[string]string      `json:"attributes,omitempty"`
	Properties    map[string]interface{} `json:"properties,omitempty"`
	ConfigKeys    []string               `json:"config_keys,omitempty"`
	X             int                    `json:"x"`
	Y             int                    `json:"y"`
	W             int                    `json:"w"`
	H             int                    `json:"h"`
}

// PostableTag mirrors tagtypes.PostableTag.
type PostableTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GettableTag mirrors tagtypes.GettableTag.
type GettableTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostableDashboardV2 is the POST /api/v2/dashboards body. The server decodes
// every level with DisallowUnknownFields, so only contract fields exist here.
type PostableDashboardV2 struct {
	SchemaVersion string          `json:"schemaVersion"`
	Name          string          `json:"name"`
	GenerateName  bool            `json:"generateName"`
	Tags          []PostableTag   `json:"tags"`
	Spec          DashboardSpecV2 `json:"spec"`
}

// UpdatableDashboardV2 is the PUT /api/v2/dashboards/{id} body. Name is the
// server's DNS name and is immutable — always send the name from state.
type UpdatableDashboardV2 struct {
	SchemaVersion string          `json:"schemaVersion"`
	Name          string          `json:"name"`
	Tags          []PostableTag   `json:"tags"`
	Spec          DashboardSpecV2 `json:"spec"`
}

// GettableDashboardV2 is the dashboard object inside the {status,data}
// envelope of POST/GET/PUT responses.
type GettableDashboardV2 struct {
	ID            string          `json:"id"`
	CreatedAt     string          `json:"createdAt,omitempty"`
	UpdatedAt     string          `json:"updatedAt,omitempty"`
	CreatedBy     string          `json:"createdBy,omitempty"`
	UpdatedBy     string          `json:"updatedBy,omitempty"`
	OrgID         string          `json:"orgId,omitempty"`
	Locked        bool            `json:"locked"`
	Source        string          `json:"source,omitempty"`
	SchemaVersion string          `json:"schemaVersion,omitempty"`
	Name          string          `json:"name"`
	Tags          []GettableTag   `json:"tags"`
	Spec          DashboardSpecV2 `json:"spec"`
}

// DashboardSpecV2 mirrors dashboardtypes.DashboardSpec. Variables, duration,
// refreshInterval and links are not managed by the provider; variables are
// decoded opaquely so a read of a dashboard edited elsewhere still round-trips.
type DashboardSpecV2 struct {
	Display         DisplayV2          `json:"display"`
	Variables       []map[string]any   `json:"variables"`
	Panels          map[string]PanelV2 `json:"panels"`
	Layouts         []LayoutV2         `json:"layouts"`
	Duration        string             `json:"duration,omitempty"`
	RefreshInterval string             `json:"refreshInterval,omitempty"`
}

// DisplayV2 mirrors dashboardtypes.Display (name required, description always
// serialized server-side).
type DisplayV2 struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PanelV2 mirrors dashboardtypes.Panel.
type PanelV2 struct {
	Kind string      `json:"kind"`
	Spec PanelSpecV2 `json:"spec"`
}

// PanelSpecV2 mirrors dashboardtypes.PanelSpec. WebComponent panels must carry
// an empty (non-null) queries list; queries are decoded opaquely so a read of
// a panel edited elsewhere still round-trips.
type PanelSpecV2 struct {
	Display DisplayV2        `json:"display"`
	Plugin  PanelPluginV2    `json:"plugin"`
	Queries []map[string]any `json:"queries"`
}

// PanelPluginV2 mirrors dashboardtypes.PanelPlugin.
type PanelPluginV2 struct {
	Kind string                `json:"kind"`
	Spec WebComponentPanelSpec `json:"spec"`
}

// WebComponentPanelSpec mirrors dashboardtypes.WebComponentPanelSpec.
type WebComponentPanelSpec struct {
	PluginID      string            `json:"pluginId,omitempty"`
	PluginVersion string            `json:"pluginVersion,omitempty"`
	Element       string            `json:"element,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	Properties    map[string]any    `json:"properties,omitempty"`
	ConfigKeys    []string          `json:"configKeys,omitempty"`
}

// LayoutV2 mirrors dashboardtypes.Layout.
type LayoutV2 struct {
	Kind string           `json:"kind"`
	Spec GridLayoutSpecV2 `json:"spec"`
}

// GridLayoutSpecV2 mirrors perses dashboard.GridLayoutSpec.
type GridLayoutSpecV2 struct {
	Display *GridLayoutDisplayV2 `json:"display,omitempty"`
	Items   []GridItemV2         `json:"items"`
}

// GridLayoutDisplayV2 mirrors perses dashboard.GridLayoutDisplay.
type GridLayoutDisplayV2 struct {
	Title string `json:"title"`
}

// GridItemV2 mirrors perses dashboard.GridItem. Content references the panel
// map by JSON pointer.
type GridItemV2 struct {
	X       int           `json:"x"`
	Y       int           `json:"y"`
	Width   int           `json:"width"`
	Height  int           `json:"height"`
	Content JSONPointerV2 `json:"content"`
}

// JSONPointerV2 mirrors perses common.JSONRef (only $ref is on the wire).
type JSONPointerV2 struct {
	Ref string `json:"$ref"`
}

func newPanelID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// TagsToPostable maps TF tag strings onto the server's {key,value} pairs.
// "k:v" becomes {key: k, value: v} (split at the first colon); a plain string
// becomes {key: "tag", value: s}.
func TagsToPostable(tags []string) []PostableTag {
	out := make([]PostableTag, 0, len(tags))
	for _, s := range tags {
		key, value := DefaultTagKey, s
		if i := strings.Index(s, ":"); i >= 0 {
			key, value = s[:i], s[i+1:]
		}
		out = append(out, PostableTag{Key: key, Value: value})
	}
	return out
}

// TagsFromGettable reverses TagsToPostable exactly: {tag, v} renders as the
// plain string v, everything else as "key:value".
func TagsFromGettable(tags []GettableTag) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.Key == DefaultTagKey {
			out = append(out, t.Value)
			continue
		}
		out = append(out, t.Key+":"+t.Value)
	}
	return out
}

// ToPostable expands plugin pages/panels into a POSTable v2 dashboard. A panel
// id is generated when unset and used as the panels map key, so a dashboard
// round-tripped through FromGettable keeps stable panel ids.
func (p PluginDashboard) ToPostable() PostableDashboardV2 {
	postable := PostableDashboardV2{
		SchemaVersion: DashboardSchemaVersionV2,
		GenerateName:  true,
		Tags:          TagsToPostable(p.Tags),
		Spec: DashboardSpecV2{
			Display:   DisplayV2{Name: p.Title, Description: p.Description},
			Variables: []map[string]any{},
			Panels:    map[string]PanelV2{},
			Layouts:   []LayoutV2{},
		},
	}

	for _, page := range p.Pages {
		layout := LayoutV2{
			Kind: "Grid",
			Spec: GridLayoutSpecV2{
				Display: &GridLayoutDisplayV2{Title: page.Name},
				Items:   []GridItemV2{},
			},
		}

		for _, panel := range page.Panels {
			id := panel.ID
			if id == "" {
				id = newPanelID()
			}
			w, h := panel.W, panel.H
			if w <= 0 {
				w = 6
			}
			if h <= 0 {
				h = 4
			}

			postable.Spec.Panels[id] = PanelV2{
				Kind: PanelEnvelopeKind,
				Spec: PanelSpecV2{
					Display: DisplayV2{Name: panel.Title},
					Plugin: PanelPluginV2{
						Kind: PanelKindWebComponent,
						Spec: WebComponentPanelSpec{
							PluginID:      panel.PluginID,
							PluginVersion: panel.PluginVersion,
							Element:       panel.Element,
							Attributes:    panel.Attributes,
							Properties:    panel.Properties,
							ConfigKeys:    panel.ConfigKeys,
						},
					},
					Queries: []map[string]any{},
				},
			}
			layout.Spec.Items = append(layout.Spec.Items, GridItemV2{
				X: panel.X, Y: panel.Y, Width: w, Height: h,
				Content: JSONPointerV2{Ref: "#/spec/panels/" + id},
			})
		}

		postable.Spec.Layouts = append(postable.Spec.Layouts, layout)
	}

	return postable
}

// ToUpdatable renders the PUT body; name must be the name held in state (it is
// immutable server-side).
func (p PluginDashboard) ToUpdatable(name string) UpdatableDashboardV2 {
	postable := p.ToPostable()
	return UpdatableDashboardV2{
		SchemaVersion: DashboardSchemaVersionV2,
		Name:          name,
		Tags:          postable.Tags,
		Spec:          postable.Spec,
	}
}

// FromGettable rebuilds the TF-shaped dashboard from a server response: title
// and description from spec.display, tags reversed from {key,value}, pages
// from layouts in order, and each panel from its layout item plus the panels
// map entry (the map key is the panel id).
func FromGettable(g GettableDashboardV2) PluginDashboard {
	pd := PluginDashboard{
		Title:       g.Spec.Display.Name,
		Description: g.Spec.Display.Description,
		Tags:        TagsFromGettable(g.Tags),
	}

	for _, layout := range g.Spec.Layouts {
		page := PluginPage{}
		if layout.Spec.Display != nil {
			page.Name = layout.Spec.Display.Title
		}

		for _, item := range layout.Spec.Items {
			id := panelIDFromRef(item.Content.Ref)
			panel, ok := g.Spec.Panels[id]
			if !ok {
				continue
			}
			if panel.Spec.Plugin.Kind != PanelKindWebComponent {
				continue
			}
			spec := panel.Spec.Plugin.Spec
			page.Panels = append(page.Panels, PluginPanel{
				ID:            id,
				Title:         panel.Spec.Display.Name,
				PluginID:      spec.PluginID,
				PluginVersion: spec.PluginVersion,
				Element:       spec.Element,
				Attributes:    spec.Attributes,
				Properties:    spec.Properties,
				ConfigKeys:    spec.ConfigKeys,
				X:             item.X,
				Y:             item.Y,
				W:             item.Width,
				H:             item.Height,
			})
		}

		pd.Pages = append(pd.Pages, page)
	}

	return pd
}

func panelIDFromRef(ref string) string {
	const prefix = "#/spec/panels/"
	return strings.TrimPrefix(ref, prefix)
}

// DocumentJSON renders the canonical POST body this dashboard would send —
// deterministic (no server timestamps), for the TF `document` attribute.
func (p PluginDashboard) DocumentJSON() (string, error) {
	b, err := json.MarshalIndent(p.ToPostable(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal dashboard document: %w", err)
	}
	return string(b), nil
}

// PropertiesJSON normalises a panel's properties map to a stable JSON string
// (Go marshals map keys sorted) so plans stay stable across reads.
func PropertiesJSON(props map[string]interface{}) string {
	if len(props) == 0 {
		return ""
	}
	b, err := json.Marshal(props)
	if err != nil {
		return ""
	}
	return string(b)
}
