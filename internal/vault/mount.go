package vault

import (
	"context"
	"fmt"
	"path"
)

// MountInfo holds metadata about a Vault secrets engine mount.
type MountInfo struct {
	Path        string
	Type        string
	Description string
	Options     map[string]string
}

// ListMounts returns all secrets engine mounts visible to the authenticated token.
func (c *Client) ListMounts(ctx context.Context) ([]MountInfo, error) {
	secret, err := c.logical.ReadWithContext(ctx, "sys/mounts")
	if err != nil {
		return nil, fmt.Errorf("listing mounts: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no mount data returned from sys/mounts")
	}

	var mounts []MountInfo
	for rawPath, rawVal := range secret.Data {
		entry, ok := rawVal.(map[string]interface{})
		if !ok {
			continue
		}
		info := MountInfo{
			Path: path.Clean(rawPath),
		}
		if t, ok := entry["type"].(string); ok {
			info.Type = t
		}
		if d, ok := entry["description"].(string); ok {
			info.Description = d
		}
		if opts, ok := entry["options"].(map[string]interface{}); ok {
			info.Options = make(map[string]string, len(opts))
			for k, v := range opts {
				if s, ok := v.(string); ok {
					info.Options[k] = s
				}
			}
		}
		mounts = append(mounts, info)
	}
	return mounts, nil
}

// GetMount returns info for a specific mount path, or an error if not found.
func (c *Client) GetMount(ctx context.Context, mountPath string) (*MountInfo, error) {
	mounts, err := c.ListMounts(ctx)
	if err != nil {
		return nil, err
	}
	cleaned := path.Clean(mountPath)
	for _, m := range mounts {
		if m.Path == cleaned {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("mount %q not found", mountPath)
}
