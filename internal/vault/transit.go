package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TransitKeyInfo holds metadata about a Vault Transit encryption key.
type TransitKeyInfo struct {
	Name            string
	Type            string
	DeletionAllowed bool
	Exportable      bool
	MinDecryptVersion int
	LatestVersion   int
}

// ListTransitKeys returns the names of all transit keys under the given mount.
func (c *Client) ListTransitKeys(ctx context.Context, mount string) ([]string, error) {
	path := fmt.Sprintf("/v1/%s/keys", mount)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path+"?list=true", nil)
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
		return nil, fmt.Errorf("transit list keys: unexpected status %d", resp.StatusCode)
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

// GetTransitKey returns metadata for a specific transit key.
func (c *Client) GetTransitKey(ctx context.Context, mount, keyName string) (*TransitKeyInfo, error) {
	path := fmt.Sprintf("/v1/%s/keys/%s", mount, keyName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+path, nil)
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
		return nil, fmt.Errorf("transit get key: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Name              string `json:"name"`
			Type              string `json:"type"`
			DeletionAllowed   bool   `json:"deletion_allowed"`
			Exportable        bool   `json:"exportable"`
			MinDecryptVersion int    `json:"min_decryption_version"`
			LatestVersion     int    `json:"latest_version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &TransitKeyInfo{
		Name:              body.Data.Name,
		Type:              body.Data.Type,
		DeletionAllowed:   body.Data.DeletionAllowed,
		Exportable:        body.Data.Exportable,
		MinDecryptVersion: body.Data.MinDecryptVersion,
		LatestVersion:     body.Data.LatestVersion,
	}, nil
}
