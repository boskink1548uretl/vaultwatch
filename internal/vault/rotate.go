package vault

import (
	"context"
	"fmt"
	"time"
)

// RotateResult holds the outcome of a secret rotation attempt.
type RotateResult struct {
	Path      string
	RotatedAt time.Time
	NewVersion int
	Err       error
}

// RotateSecret deletes and re-writes a KV v2 secret at the given path,
// effectively rotating it by incrementing its version. The caller supplies
// newData as the replacement payload.
func (c *Client) RotateSecret(ctx context.Context, path string, newData map[string]interface{}) (*RotateResult, error) {
	kvPath := toKVDataPath(path)

	secret, err := c.vault.KVv2(kvMountFromPath(path)).Put(ctx, kvKeyFromPath(path), newData)
	if err != nil {
		return nil, fmt.Errorf("rotate secret %s: %w", path, err)
	}

	var version int
	if secret != nil && secret.VersionMetadata != nil {
		version = secret.VersionMetadata.Version
	}

	_ = kvPath // used for audit trail in future

	return &RotateResult{
		Path:       path,
		RotatedAt:  time.Now().UTC(),
		NewVersion: version,
	}, nil
}

// kvMountFromPath extracts the mount prefix from a secrets path.
// e.g. "secret/myapp/db" → "secret"
func kvMountFromPath(path string) string {
	for i, ch := range path {
		if ch == '/' {
			return path[:i]
		}
	}
	return path
}

// kvKeyFromPath extracts the key portion from a secrets path.
// e.g. "secret/myapp/db" → "myapp/db"
func kvKeyFromPath(path string) string {
	for i, ch := range path {
		if ch == '/' {
			return path[i+1:]
		}
	}
	return path
}
