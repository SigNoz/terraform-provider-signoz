package model

// PluginManifest mirrors pkg/plugins/types.go Manifest — one catalog entry of
// GET /api/v1/plugins (and of the single-plugin GET with ?version=).
type PluginManifest struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	DisplayName string `json:"displayName"`
	Element     string `json:"element"`
	Source      string `json:"source"`
	BundlePath  string `json:"bundlePath,omitempty"`
	Description string `json:"description,omitempty"`
}
