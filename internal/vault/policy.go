package vault

import (
	"context"
	"fmt"
	"net/http"
)

// PolicyInfo holds metadata about a Vault policy.
type PolicyInfo struct {
	Name  string
	Rules string
}

// GetPolicy retrieves a named ACL policy from Vault.
func (c *Client) GetPolicy(ctx context.Context, name string) (*PolicyInfo, error) {
	path := fmt.Sprintf("/v1/sys/policies/acl/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building policy request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("policy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("policy %q not found", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching policy %q", resp.StatusCode, name)
	}

	var body struct {
		Data struct {
			Name   string `json:"name"`
			Policy string `json:"policy"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("decoding policy response: %w", err)
	}

	return &PolicyInfo{
		Name:  body.Data.Name,
		Rules: body.Data.Policy,
	}, nil
}

// ListPolicies returns all ACL policy names from Vault.
func (c *Client) ListPolicies(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/policies/acl", nil)
	if err != nil {
		return nil, fmt.Errorf("building list-policies request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	q := req.URL.Query()
	q.Set("list", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list-policies request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d listing policies", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("decoding list-policies response: %w", err)
	}
	return body.Data.Keys, nil
}
