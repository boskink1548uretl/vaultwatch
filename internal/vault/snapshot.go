package vault

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// SnapshotInfo holds metadata about a Vault raft snapshot response.
type SnapshotInfo struct {
	StatusCode int
	Size       int64
}

// TakeSnapshot requests a Raft snapshot from Vault and writes it to w.
// It returns metadata about the snapshot on success.
func (c *Client) TakeSnapshot(ctx context.Context, w io.Writer) (*SnapshotInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/storage/raft/snapshot", nil)
	if err != nil {
		return nil, fmt.Errorf("snapshot: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("snapshot: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snapshot: unexpected status %d", resp.StatusCode)
	}

	n, err := io.Copy(w, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("snapshot: reading body: %w", err)
	}

	return &SnapshotInfo{StatusCode: resp.StatusCode, Size: n}, nil
}
