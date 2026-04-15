package vault

import (
	"context"
	"fmt"
	"path"
)

// SecretTags represents the custom metadata tags attached to a KV v2 secret.
type SecretTags map[string]string

// GetSecretTags retrieves the custom_metadata map from a KV v2 secret's
// metadata endpoint. Returns an empty map if no tags are present.
func (c *Client) GetSecretTags(ctx context.Context, secretPath string) (SecretTags, error) {
	mount := kvMountFromPath(secretPath)
	key := kvKeyFromPath(secretPath)

	metaPath := path.Join(mount, "metadata", key)

	secret, err := c.logical.ReadWithContext(ctx, metaPath)
	if err != nil {
		return nil, fmt.Errorf("reading tags for %q: %w", secretPath, err)
	}
	if secret == nil || secret.Data == nil {
		return SecretTags{}, nil
	}

	raw, ok := secret.Data["custom_metadata"]
	if !ok || raw == nil {
		return SecretTags{}, nil
	}

	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return SecretTags{}, nil
	}

	tags := make(SecretTags, len(rawMap))
	for k, v := range rawMap {
		if s, ok := v.(string); ok {
			tags[k] = s
		}
	}
	return tags, nil
}
