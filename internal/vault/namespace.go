package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Namespace represents a Vault namespace with its path and metadata.
type Namespace struct {
	Path        string            `json:"path"`
	ID          string            `json:"id"`
	CustomMeta  map[string]string `json:"custom_metadata"`
}

// NamespaceLister can list and retrieve Vault namespaces.
type NamespaceLister interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	GetNamespace(ctx context.Context, path string) (*Namespace, error)
}

// ListNamespaces returns all child namespace paths under the current namespace.
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/namespaces?list=true", nil)
	if err != nil {
		return nil, fmt.Errorf("building list namespaces request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status listing namespaces: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding namespace list: %w", err)
	}
	return result.Data.Keys, nil
}

// GetNamespace retrieves metadata for a specific namespace by path.
func (c *Client) GetNamespace(ctx context.Context, path string) (*Namespace, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/namespaces/"+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building get namespace request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting namespace %q: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("namespace %q not found", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status getting namespace %q: %d", path, resp.StatusCode)
	}

	var result struct {
		Data Namespace `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding namespace %q: %w", path, err)
	}
	return &result.Data, nil
}
