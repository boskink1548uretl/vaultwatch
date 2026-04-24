package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Plugin represents a Vault plugin registration.
type Plugin struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Builtin bool   `json:"builtin"`
	SHA256  string `json:"sha256,omitempty"`
}

// pluginListResponse mirrors the Vault API response for listing plugins.
type pluginListResponse struct {
	Data struct {
		Detailed []Plugin `json:"detailed"`
	} `json:"data"`
}

// ListPlugins returns all registered plugins of the given type ("auth", "secret", "database").
// Pass an empty string to list all types.
func (c *Client) ListPlugins(ctx context.Context, pluginType string) ([]Plugin, error) {
	path := "/v1/sys/plugins/catalog"
	if pluginType != "" {
		path = fmt.Sprintf("/v1/sys/plugins/catalog/%s", pluginType)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list plugins: unexpected status %d", resp.StatusCode)
	}

	var result pluginListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode plugins: %w", err)
	}
	return result.Data.Detailed, nil
}

// GetPlugin returns a single plugin by type and name.
func (c *Client) GetPlugin(ctx context.Context, pluginType, name string) (*Plugin, error) {
	path := fmt.Sprintf("/v1/sys/plugins/catalog/%s/%s", pluginType, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get plugin: unexpected status %d", resp.StatusCode)
	}

	var wrapper struct {
		Data Plugin `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decode plugin: %w", err)
	}
	return &wrapper.Data, nil
}
