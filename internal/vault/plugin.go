package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Plugin represents a registered Vault plugin.
type Plugin struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // auth, secret, database
	Version string `json:"version"`
	Builtin bool   `json:"builtin"`
	SHA256  string `json:"sha256"`
	Command string `json:"command"`
}

// pluginCatalogResponse is the raw API response for listing plugins.
type pluginCatalogResponse struct {
	Data struct {
		Detailed []Plugin `json:"detailed"`
	} `json:"data"`
}

// pluginSingleResponse is the raw API response for a single plugin lookup.
type pluginSingleResponse struct {
	Data Plugin `json:"data"`
}

// ListPlugins returns all registered plugins of the given type.
// Valid types are "auth", "secret", and "database".
// Pass an empty string to list all plugin types.
func (c *Client) ListPlugins(ctx context.Context, pluginType string) ([]Plugin, error) {
	path := "/v1/sys/plugins/catalog"
	if pluginType != "" {
		path = fmt.Sprintf("/v1/sys/plugins/catalog/%s", pluginType)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building list plugins request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("listing plugins: unexpected status %d", resp.StatusCode)
	}

	var result pluginCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding plugin list response: %w", err)
	}
	return result.Data.Detailed, nil
}

// GetPlugin retrieves details for a specific plugin by type and name.
func (c *Client) GetPlugin(ctx context.Context, pluginType, name string) (*Plugin, error) {
	if pluginType == "" || name == "" {
		return nil, fmt.Errorf("plugin type and name must not be empty")
	}

	path := fmt.Sprintf("/v1/sys/plugins/catalog/%s/%s", pluginType, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building get plugin request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting plugin %s/%s: %w", pluginType, name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting plugin %s/%s: unexpected status %d", pluginType, name, resp.StatusCode)
	}

	var result pluginSingleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding plugin response: %w", err)
	}
	return &result.Data, nil
}
