package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AuthMethod represents a Vault auth method configuration.
type AuthMethod struct {
	Type        string            `json:"type"`
	Path        string            `json:"path"`
	Description string            `json:"description"`
	Options     map[string]string `json:"options"`
	Local       bool              `json:"local"`
	CreatedAt   time.Time         `json:"created_time"`
}

// ListAuthMethods returns all enabled auth methods from Vault.
func (c *Client) ListAuthMethods(ctx context.Context) (map[string]*AuthMethod, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/auth", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list auth methods: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result map[string]*AuthMethod
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

// GetAuthMethod returns a single auth method by path.
func (c *Client) GetAuthMethod(ctx context.Context, path string) (*AuthMethod, error) {
	methods, err := c.ListAuthMethods(ctx)
	if err != nil {
		return nil, err
	}
	key := path + "/"
	m, ok := methods[key]
	if !ok {
		return nil, fmt.Errorf("auth method %q not found", path)
	}
	return m, nil
}
