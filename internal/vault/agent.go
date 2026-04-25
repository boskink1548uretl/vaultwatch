package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AgentInfo represents the status of a Vault Agent process.
type AgentInfo struct {
	Initialized bool   `json:"initialized"`
	SealStatus  string `json:"seal_status"`
	Version     string `json:"version"`
	CacheEnabled bool  `json:"cache_enabled"`
}

// GetAgentInfo retrieves status information from a Vault Agent endpoint.
func (c *Client) GetAgentInfo(ctx context.Context) (*AgentInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/agent/v1/metrics", nil)
	if err != nil {
		return nil, fmt.Errorf("building agent info request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent info request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("agent endpoint not available (is this a Vault Agent?): %s", c.address)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from agent endpoint: %d", resp.StatusCode)
	}

	var info AgentInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding agent info response: %w", err)
	}
	return &info, nil
}

// IsAgentCacheEnabled returns true if the Vault Agent has caching enabled.
func (c *Client) IsAgentCacheEnabled(ctx context.Context) (bool, error) {
	info, err := c.GetAgentInfo(ctx)
	if err != nil {
		return false, err
	}
	return info.CacheEnabled, nil
}
