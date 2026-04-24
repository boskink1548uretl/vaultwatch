package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Group represents a Vault identity group.
type Group struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Policies []string          `json:"policies"`
	Metadata map[string]string `json:"metadata"`
}

// ListGroups returns all identity groups from Vault.
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/identity/group/id?list=true", nil)
	if err != nil {
		return nil, fmt.Errorf("build list groups request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Group{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list groups: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			KeyInfo map[string]Group `json:"key_info"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode list groups response: %w", err)
	}

	groups := make([]Group, 0, len(body.Data.KeyInfo))
	for id, g := range body.Data.KeyInfo {
		g.ID = id
		groups = append(groups, g)
	}
	return groups, nil
}

// GetGroup retrieves a single identity group by name.
func (c *Client) GetGroup(ctx context.Context, name string) (*Group, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/identity/group/name/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("build get group request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get group: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data Group `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode get group response: %w", err)
	}
	return &body.Data, nil
}
