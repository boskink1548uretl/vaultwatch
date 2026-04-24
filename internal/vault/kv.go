package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KVVersion represents the KV secrets engine version.
type KVVersion int

const (
	KVv1 KVVersion = 1
	KVv2 KVVersion = 2
)

// KVMetadata holds metadata about a KV v2 secret.
type KVMetadata struct {
	Path            string
	CurrentVersion  int
	OldestVersion   int
	CreatedTime     time.Time
	UpdatedTime     time.Time
	MaxVersions     int
	DeleteVersionAfter string
}

// GetKVVersion detects whether the given mount path uses KV v1 or v2.
func (c *Client) GetKVVersion(ctx context.Context, mount string) (KVVersion, error) {
	url := fmt.Sprintf("%s/v1/sys/mounts/%s", c.address, mount)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("kv version request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("kv version request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("mount %q not found", mount)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status %d for mount %q", resp.StatusCode, mount)
	}

	var body struct {
		Options struct {
			Version string `json:"version"`
		} `json:"options"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("decode mount info: %w", err)
	}

	if body.Options.Version == "2" {
		return KVv2, nil
	}
	return KVv1, nil
}

// GetKVMetadata returns metadata for a KV v2 secret path.
func (c *Client) GetKVMetadata(ctx context.Context, mount, path string) (*KVMetadata, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.address, mount, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kv metadata request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kv metadata request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret %q not found in mount %q", path, mount)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			CurrentVersion  int    `json:"current_version"`
			OldestVersion   int    `json:"oldest_version"`
			CreatedTime     string `json:"created_time"`
			UpdatedTime     string `json:"updated_time"`
			MaxVersions     int    `json:"max_versions"`
			DeleteVersionAfter string `json:"delete_version_after"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode kv metadata: %w", err)
	}

	created, _ := time.Parse(time.RFC3339Nano, result.Data.CreatedTime)
	updated, _ := time.Parse(time.RFC3339Nano, result.Data.UpdatedTime)

	return &KVMetadata{
		Path:               fmt.Sprintf("%s/%s", mount, path),
		CurrentVersion:     result.Data.CurrentVersion,
		OldestVersion:      result.Data.OldestVersion,
		CreatedTime:        created,
		UpdatedTime:        updated,
		MaxVersions:        result.Data.MaxVersions,
		DeleteVersionAfter: result.Data.DeleteVersionAfter,
	}, nil
}
