package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RateQuota represents a Vault rate-limit quota rule.
type RateQuota struct {
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	Rate          float64 `json:"rate"`
	Interval      int     `json:"interval"`
	BlockInterval int     `json:"block_interval"`
}

// ListQuotas returns all rate-limit quota rule names from Vault.
func (c *Client) ListQuotas(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/quotas/rate-limit?list=true", nil)
	if err != nil {
		return nil, fmt.Errorf("quota list request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quota list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("quota list: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("quota list decode: %w", err)
	}
	return body.Data.Keys, nil
}

// GetQuota retrieves a single rate-limit quota rule by name.
func (c *Client) GetQuota(ctx context.Context, name string) (*RateQuota, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/quotas/rate-limit/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("get quota request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get quota: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get quota: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data RateQuota `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("get quota decode: %w", err)
	}
	return &body.Data, nil
}
