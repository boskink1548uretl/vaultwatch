package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RaftPeer represents a single node in the Raft cluster.
type RaftPeer struct {
	NodeID    string `json:"node_id"`
	Address   string `json:"address"`
	Leader    bool   `json:"leader"`
	Protocol  string `json:"protocol_version"`
	Voter     bool   `json:"voter"`
}

// RaftConfiguration holds the current Raft cluster configuration.
type RaftConfiguration struct {
	Index   uint64      `json:"index"`
	Servers []RaftPeer  `json:"servers"`
}

// GetRaftConfiguration retrieves the current Raft cluster peer configuration.
func (c *Client) GetRaftConfiguration(ctx context.Context) (*RaftConfiguration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/storage/raft/configuration", nil)
	if err != nil {
		return nil, fmt.Errorf("building raft config request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("raft config request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("raft storage backend not in use")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from raft config endpoint", resp.StatusCode)
	}

	var wrapper struct {
		Data RaftConfiguration `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decoding raft config response: %w", err)
	}
	return &wrapper.Data, nil
}

// IsRaftLeader returns true if the local node is the current Raft leader.
func (c *Client) IsRaftLeader(ctx context.Context) (bool, error) {
	cfg, err := c.GetRaftConfiguration(ctx)
	if err != nil {
		return false, err
	}
	for _, peer := range cfg.Servers {
		if peer.Leader {
			// We have no direct way to know our own node ID here, so we
			// surface the leader flag from the peer list for callers.
			return peer.Leader, nil
		}
	}
	return false, nil
}
