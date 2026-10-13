package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
)

// WebComponent panel kind used by the Lacrosse / noizu-forks SigNoz plugin host.
const PanelKindWebComponent = "signoz/WebComponentPanel"

// PluginDashboard is a high-level dashboard built from plugin pages/panels.
// It is expanded into a SigNoz V2 dashboard document (Perses-style panels).
type PluginDashboard struct {
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Pages       []PluginPage `json:"pages"`
}

// PluginPage groups plugin panels (a "page" in the UI / layout group).
type PluginPage struct {
	Name   string        `json:"name"`
	Panels []PluginPanel `json:"panels"`
}

// PluginPanel is one WebComponent plugin instance on a page.
type PluginPanel struct {
	ID         string                 `json:"id,omitempty"`
	Title      string                 `json:"title"`
	PluginID   string                 `json:"plugin_id"`
	Element    string                 `json:"element"`
	Attributes map[string]string      `json:"attributes,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	ConfigKeys []string               `json:"config_keys,omitempty"`
	X          int                    `json:"x"`
	Y          int                    `json:"y"`
	W          int                    `json:"w"`
	H          int                    `json:"h"`
}

// DashboardV2 is the wire format for POST/PUT /api/v2/dashboards.
type DashboardV2 struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Data        map[string]interface{} `json:"data"`
}

// DashboardV2Data is returned from GET /api/v2/dashboards/{id}.
type DashboardV2Data struct {
	ID          string                 `json:"id"`
	CreatedAt   string                 `json:"createdAt,omitempty"`
	CreatedBy   string                 `json:"createdBy,omitempty"`
	UpdatedAt   string                 `json:"updatedAt,omitempty"`
	UpdatedBy   string                 `json:"updatedBy,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	// Some SigNoz versions nest fields under data only.
	Raw map[string]interface{} `json:"-"`
}

func newPanelID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ToDashboardV2 expands plugin pages/panels into a V2 dashboard document.
func (p PluginDashboard) ToDashboardV2() DashboardV2 {
	panels := map[string]interface{}{}
	// layouts: map page name → layout items
	layouts := map[string]interface{}{}
	// ordered page list for UI
	var pageOrder []string

	for _, page := range p.Pages {
		pageName := page.Name
		if pageName == "" {
			pageName = "default"
		}
		pageOrder = append(pageOrder, pageName)
		var layoutItems []map[string]interface{}

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
			attrs := panel.Attributes
			if attrs == nil {
				attrs = map[string]string{}
			}
			props := panel.Properties
			if props == nil {
				props = map[string]interface{}{}
			}
			cfgKeys := panel.ConfigKeys
			if cfgKeys == nil {
				cfgKeys = []string{}
			}

			panels[id] = map[string]interface{}{
				"kind": "Panel",
				"spec": map[string]interface{}{
					"display": map[string]interface{}{
						"name": panel.Title,
					},
					"plugin": map[string]interface{}{
						"kind": PanelKindWebComponent,
						"spec": map[string]interface{}{
							"pluginId":   panel.PluginID,
							"element":    panel.Element,
							"attributes": attrs,
							"properties": props,
							"configKeys": cfgKeys,
						},
					},
					"queries": []interface{}{},
				},
			}

			layoutItems = append(layoutItems, map[string]interface{}{
				"i":      id,
				"x":      panel.X,
				"y":      panel.Y,
				"w":      w,
				"h":      h,
				"moved":  false,
				"static": false,
			})
		}
		layouts[pageName] = layoutItems
	}

	// Prefer a single "default" layout array if only one page (matches many UIs).
	var layoutWire interface{} = layouts
	if len(pageOrder) == 1 {
		layoutWire = layouts[pageOrder[0]]
	}

	data := map[string]interface{}{
		"panels":    panels,
		"layouts":   layoutWire,
		"variables": map[string]interface{}{},
		"pages":     pageOrder,
		// Lacrosse metadata for round-trip / debugging
		"lacrosse": map[string]interface{}{
			"kind":    "plugin_dashboard",
			"version": 1,
		},
	}

	return DashboardV2{
		Title:       p.Title,
		Description: p.Description,
		Tags:        p.Tags,
		Data:        data,
	}
}

// ToJSON is a debug helper.
func (d DashboardV2) ToJSON() (string, error) {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToDashboardV1 expands plugin pages into a classic SigNoz v1/v4 dashboard.
// Each WebComponent panel is stored as a widget whose description embeds the
// lacrossePlugin spec + stub properties so stock SigNoz APIs can host the
// document until a plugin-host image renders WC panels natively.
func (p PluginDashboard) ToDashboardV1() *Dashboard {
	v2 := p.ToDashboardV2()
	panels, _ := v2.Data["panels"].(map[string]interface{})
	var widgets []map[string]interface{}
	var layout []map[string]interface{}
	// Flatten page layouts in page order
	pageOrder, _ := v2.Data["pages"].([]string)
	if pageOrder == nil {
		if pi, ok := v2.Data["pages"].([]interface{}); ok {
			for _, x := range pi {
				if s, ok := x.(string); ok {
					pageOrder = append(pageOrder, s)
				}
			}
		}
	}
	// Collect layout items from layouts map or array
	var layoutItems []map[string]interface{}
	switch L := v2.Data["layouts"].(type) {
	case []map[string]interface{}:
		layoutItems = L
	case []interface{}:
		for _, it := range L {
			if m, ok := it.(map[string]interface{}); ok {
				layoutItems = append(layoutItems, m)
			}
		}
	case map[string]interface{}:
		for _, pname := range pageOrder {
			if arr, ok := L[pname].([]map[string]interface{}); ok {
				layoutItems = append(layoutItems, arr...)
			} else if arr, ok := L[pname].([]interface{}); ok {
				for _, it := range arr {
					if m, ok := it.(map[string]interface{}); ok {
						layoutItems = append(layoutItems, m)
					}
				}
			}
		}
		if len(layoutItems) == 0 {
			for _, v := range L {
				if arr, ok := v.([]interface{}); ok {
					for _, it := range arr {
						if m, ok := it.(map[string]interface{}); ok {
							layoutItems = append(layoutItems, m)
						}
					}
				}
			}
		}
	}

	// Index panels in stable order of layout items
	seen := map[string]bool{}
	for _, li := range layoutItems {
		id, _ := li["i"].(string)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		panel, ok := panels[id].(map[string]interface{})
		if !ok {
			continue
		}
		spec, _ := panel["spec"].(map[string]interface{})
		display, _ := spec["display"].(map[string]interface{})
		title, _ := display["name"].(string)
		plugin := spec["plugin"]
		meta, _ := json.Marshal(map[string]interface{}{
			"lacrossePlugin": plugin,
			"stubMode":       true,
			"kind":           "plugin_panel",
		})
		widgets = append(widgets, map[string]interface{}{
			"id":              id,
			"title":           title,
			"description":     string(meta),
			"panelTypes":      "graph",
			"nullZeroValues":  "zero",
			"opacity":         "1",
			"isStacked":       false,
			"yAxisUnit":       "none",
			"fillSpans":       false,
			"softMax":         0,
			"softMin":         0,
			"stackedBarChart": false,
			"query": map[string]interface{}{
				"id":             "A",
				"queryType":      "builder",
				"promql":         []interface{}{},
				"clickhouse_sql": []interface{}{},
				"builder": map[string]interface{}{
					"queryData":     []interface{}{},
					"queryFormulas": []interface{}{},
				},
			},
		})
		// normalize layout keys to h,i,w,x,y
		layout = append(layout, map[string]interface{}{
			"i":      id,
			"x":      li["x"],
			"y":      li["y"],
			"w":      li["w"],
			"h":      li["h"],
			"moved":  false,
			"static": false,
		})
	}

	return &Dashboard{
		CollapsableRowsMigrated: true,
		Description:             p.Description + " [plugin-stub-widgets]",
		Layout:                  layout,
		Name:                    p.Title,
		Title:                   p.Title,
		Tags:                    p.Tags,
		UploadedGrafana:         false,
		Variables:               map[string]interface{}{},
		Version:                 "v4",
		Widgets:                 widgets,
		PanelMap:                map[string]interface{}{},
	}
}
