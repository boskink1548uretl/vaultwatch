package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// BackendInfo holds information about a secrets backend (mount).
type BackendInfo struct {
	Path        string            `json:"path"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Options     map[string]string `json:"options"`
	Local       bool              `json:"local"`
	SealWrap    bool              `json:"seal_wrap"`
}

// ListBackends returns all enabled secrets backends from Vault.
func (c *Client) ListBackends(ctx context.Context) ([]BackendInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("build list backends request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list backends request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list backends: unexpected status %d", resp.StatusCode)
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode backends response: %w", err)
	}

	var backends []BackendInfo
	for path, data := range raw {
		var info BackendInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}
		info.Path = path
		backends = append(backends, info)
	}
	return backends, nil
}

// GetBackend returns details for a single secrets backend by mount path.
func (c *Client) GetBackend(ctx context.Context, mountPath string) (*BackendInfo, error) {
	backends, err := c.ListBackends(ctx)
	if err != nil {
		return nil, err
	}
	for _, b := range backends {
		if b.Path == mountPath || b.Path == mountPath+"/" {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("backend %q not found", mountPath)
}
