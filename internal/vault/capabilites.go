package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CapabilitiesResult holds the capabilities of a token on given paths.
type CapabilitiesResult struct {
	Capabilities map[string][]string `json:"capabilities"`
}

// GetCapabilities returns the capabilities of the current token for the given paths.
func (c *Client) GetCapabilities(ctx context.Context, paths []string) (map[string][]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths must not be empty")
	}

	body := map[string]interface{}{
		"paths": paths,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal capabilities request: %w", err)
	}

	resp, err := c.rawPost(ctx, "/v1/sys/capabilities-self", data)
	if err != nil {
		return nil, fmt.Errorf("capabilities-self request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied on capabilities-self")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from capabilities-self", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode capabilities response: %w", err)
	}

	caps := make(map[string][]string)
	for _, path := range paths {
		if raw, ok := result[path]; ok {
			if items, ok := raw.([]interface{}); ok {
				for _, item := range items {
					if s, ok := item.(string); ok {
						caps[path] = append(caps[path], s)
					}
				}
			}
		}
	}
	return caps, nil
}

// HasCapability checks whether the current token has a specific capability on a path.
func (c *Client) HasCapability(ctx context.Context, path, capability string) (bool, error) {
	caps, err := c.GetCapabilities(ctx, []string{path})
	if err != nil {
		return false, err
	}
	for _, cap := range caps[path] {
		if cap == capability || cap == "root" {
			return true, nil
		}
	}
	return false, nil
}
