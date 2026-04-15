package vault

import (
	"context"
	"fmt"
	"path"
	"strings"
)

// ListSecrets returns all secret paths under the given mount and prefix.
// It recursively traverses folders returned by the KV v2 list endpoint.
func (c *Client) ListSecrets(ctx context.Context, mount, prefix string) ([]string, error) {
	listPath := path.Join(mount, "metadata", prefix)
	secret, err := c.logical.ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("listing secrets at %q: %w", listPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keys, ok := secret.Data["keys"]
	if !ok {
		return nil, nil
	}

	rawKeys, ok := keys.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys type at %q", listPath)
	}

	var results []string
	for _, k := range rawKeys {
		key, ok := k.(string)
		if !ok {
			continue
		}
		full := path.Join(prefix, key)
		if strings.HasSuffix(key, "/") {
			// It's a folder — recurse
			sub, err := c.ListSecrets(ctx, mount, full)
			if err != nil {
				return nil, err
			}
			results = append(results, sub...)
		} else {
			results = append(results, full)
		}
	}
	return results, nil
}
