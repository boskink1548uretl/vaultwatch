package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HAStatus represents the high-availability status of a Vault cluster.
type HAStatus struct {
	ClusterID   string    `json:"cluster_id"`
	ClusterName string    `json:"cluster_name"`
	LeaderAddr  string    `json:"leader_address"`
	IsLeader    bool      `json:"is_self"`
	Nodes       []HANode  `json:"nodes"`
}

// HANode represents a single node in the HA cluster.
type HANode struct {
	Hostname    string `json:"hostname"`
	APIAddr     string `json:"api_address"`
	ClusterAddr string `json:"cluster_address"`
	ActiveNode  bool   `json:"active_node"`
}

// GetHAStatus returns the high-availability status from /v1/sys/ha-status.
func (c *Client) GetHAStatus(ctx context.Context) (*HAStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/ha-status", nil)
	if err != nil {
		return nil, fmt.Errorf("ha status: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ha status: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ha status: unexpected status %d", resp.StatusCode)
	}

	var wrapper struct {
		Data HAStatus `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("ha status: decode: %w", err)
	}
	return &wrapper.Data, nil
}

// IsHALeader returns true if the current node is the active HA leader.
func (c *Client) IsHALeader(ctx context.Context) (bool, error) {
	status, err := c.GetHAStatus(ctx)
	if err != nil {
		return false, err
	}
	if status == nil {
		return false, nil
	}
	return status.IsLeader, nil
}
