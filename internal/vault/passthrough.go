package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PassthroughEntry represents a single passthrough secret entry.
type PassthroughEntry struct {
	Path    string            `json:"path"`
	Data    map[string]string `json:"data"`
	Options map[string]string `json:"options,omitempty"`
}

// GetPassthrough reads a passthrough (generic) secret at the given path.
func (c *Client) GetPassthrough(ctx context.Context, path string) (*PassthroughEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v1/%s", c.address, path), nil)
	if err != nil {
		return nil, fmt.Errorf("passthrough: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("passthrough: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("passthrough: path %q not found", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("passthrough: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("passthrough: decode: %w", err)
	}

	return &PassthroughEntry{
		Path: path,
		Data: body.Data,
	}, nil
}

// ListPassthrough lists keys under a passthrough path.
func (c *Client) ListPassthrough(ctx context.Context, path string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "LIST",
		fmt.Sprintf("%s/v1/%s", c.address, path), nil)
	if err != nil {
		return nil, fmt.Errorf("passthrough: build list request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("passthrough: list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("passthrough: list unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("passthrough: decode list: %w", err)
	}
	return body.Data.Keys, nil
}
