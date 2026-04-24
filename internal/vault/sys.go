package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SysInfo holds high-level Vault system information.
type SysInfo struct {
	Version       string `json:"version"`
	ClusterName   string `json:"cluster_name"`
	ClusterID     string `json:"cluster_id"`
	HA            bool   `json:"ha_enabled"`
	Initialized   bool   `json:"initialized"`
	Sealed        bool   `json:"sealed"`
	Standby       bool   `json:"standby"`
	ReplicationDR struct {
		Mode string `json:"mode"`
	} `json:"replication_dr_mode"`
	ReplicationPerf struct {
		Mode string `json:"mode"`
	} `json:"replication_performance_mode"`
}

// GetSysInfo retrieves general system information from the /sys/health endpoint
// and the /sys/leader endpoint to build a composite SysInfo value.
func (c *Client) GetSysInfo(ctx context.Context) (*SysInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/health?standbyok=true&sealedok=true&uninitok=true", nil)
	if err != nil {
		return nil, fmt.Errorf("sys info request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sys info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("sys info: endpoint not found")
	}

	var info SysInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("sys info decode: %w", err)
	}
	return &info, nil
}
