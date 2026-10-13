package legacyclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/SigNoz/terraform-provider-signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const pluginsPath = "api/v1/plugins"

// pluginsEnvelope is the {status,data} wrapper of the plugin catalog API.
type pluginsEnvelope struct {
	Status string               `json:"status"`
	Data   model.PluginManifest `json:"data"`
}

type pluginsListEnvelope struct {
	Status string                 `json:"status"`
	Data   []model.PluginManifest `json:"data"`
}

// GetPlugin returns the catalog's default manifest for a plugin id (the server
// picks the last registered version when no version is requested).
func (c *Client) GetPlugin(ctx context.Context, id string) (*model.PluginManifest, error) {
	return c.getPlugin(ctx, id, "")
}

// GetPluginVersion returns the manifest for a specific plugin version.
func (c *Client) GetPluginVersion(ctx context.Context, id, version string) (*model.PluginManifest, error) {
	return c.getPlugin(ctx, id, version)
}

func (c *Client) getPlugin(ctx context.Context, id, version string) (*model.PluginManifest, error) {
	u, err := url.JoinPath(c.hostURL.String(), pluginsPath, id)
	if err != nil {
		return nil, err
	}
	if version != "" {
		parsed, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		q := parsed.Query()
		q.Set("version", version)
		parsed.RawQuery = q.Encode()
		u = parsed.String()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var env pluginsEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("get plugin %s: decode response: %w", id, err)
	}
	if env.Status != "success" {
		return nil, fmt.Errorf("get plugin %s: unexpected status %q in response", id, env.Status)
	}
	tflog.Debug(ctx, "GetPlugin", map[string]any{"id": id, "version": env.Data.Version})
	return &env.Data, nil
}

// ListPlugins returns the full plugin catalog.
func (c *Client) ListPlugins(ctx context.Context) ([]model.PluginManifest, error) {
	u, err := url.JoinPath(c.hostURL.String(), pluginsPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var env pluginsListEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("list plugins: decode response: %w", err)
	}
	if env.Status != "success" {
		return nil, fmt.Errorf("list plugins: unexpected status %q in response", env.Status)
	}
	return env.Data, nil
}
