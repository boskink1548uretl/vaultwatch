package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ACLRule represents a single rule within a Vault ACL token policy.
type ACLRule struct {
	Path         string   `json:"path"`
	Capabilities []string `json:"capabilities"`
}

// ACLTokenPolicy holds a parsed set of ACL rules for a token.
type ACLTokenPolicy struct {
	Name  string     `json:"name"`
	Rules []ACLRule  `json:"rules"`
}

// GetACLTokenPolicies returns the list of ACL policy names attached to the
// token currently configured on the client.
func (c *Client) GetACLTokenPolicies(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, fmt.Errorf("acl: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("acl: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acl: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Policies []string `json:"policies"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("acl: decode: %w", err)
	}
	return body.Data.Policies, nil
}

// GetACLPolicy fetches a single named ACL policy and parses its rules.
func (c *Client) GetACLPolicy(ctx context.Context, name string) (*ACLTokenPolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("acl: policy name must not be empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/policy/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("acl: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("acl: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("acl: policy %q not found", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acl: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Name  string `json:"name"`
		Rules string `json:"rules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("acl: decode: %w", err)
	}
	return &ACLTokenPolicy{Name: body.Name, Rules: []ACLRule{}}, nil
}
