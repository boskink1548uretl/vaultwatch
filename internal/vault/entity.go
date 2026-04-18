package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Entity represents a Vault identity entity.
type Entity struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Policies []string          `json:"policies"`
	Meta     map[string]string `json:"metadata"`
	Disabled bool              `json:"disabled"`
}

// ListEntities returns all identity entity names.
func (c *Client) ListEntities(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/identity/entity?list=true", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault: list entities returned %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body.Data.Keys, nil
}

// GetEntity returns a single entity by name.
func (c *Client) GetEntity(ctx context.Context, name string) (*Entity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/identity/entity/name/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault: get entity returned %d", resp.StatusCode)
	}

	var body struct {
		Data Entity `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &body.Data, nil
}
