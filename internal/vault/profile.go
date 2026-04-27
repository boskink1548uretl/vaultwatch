package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// VaultProfile holds a summary of key Vault instance attributes.
type VaultProfile struct {
	Version     string `json:"version"`
	ClusterName string `json:"cluster_name"`
	ClusterID   string `json:"cluster_id"`
	HA          bool   `json:"ha_enabled"`
	Sealed      bool   `json:"sealed"`
	Initialized bool   `json:"initialized"`
}

// GetProfile fetches a high-level profile of the Vault instance.
func (c *Client) GetProfile(ctx context.Context) (*VaultProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/health?standbyok=true&perfstandbyok=true", nil)
	if err != nil {
		return nil, fmt.Errorf("profile: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("profile: request failed: %w", err)
	}
	defer resp.Body.Close()

	var raw struct {
		Version     string `json:"version"`
		ClusterName string `json:"cluster_name"`
		ClusterID   string `json:"cluster_id"`
		HA          bool   `json:"ha_enabled"`
		Sealed      bool   `json:"sealed"`
		Initialized bool   `json:"initialized"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("profile: decode response: %w", err)
	}

	return &VaultProfile{
		Version:     raw.Version,
		ClusterName: raw.ClusterName,
		ClusterID:   raw.ClusterID,
		HA:          raw.HA,
		Sealed:      raw.Sealed,
		Initialized: raw.Initialized,
	}, nil
}
