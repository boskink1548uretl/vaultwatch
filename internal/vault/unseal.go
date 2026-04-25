package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// UnsealStatus represents the result of submitting an unseal key share.
type UnsealStatus struct {
	Sealed    bool `json:"sealed"`
	T         int  `json:"t"` // threshold
	N         int  `json:"n"` // total shares
	Progress  int  `json:"progress"`
	Version   string `json:"version"`
}

// SubmitUnsealKey sends a single unseal key share to Vault.
// Returns the current unseal progress status.
func (c *Client) SubmitUnsealKey(ctx context.Context, key string) (*UnsealStatus, error) {
	if key == "" {
		return nil, fmt.Errorf("unseal key must not be empty")
	}

	body := map[string]string{"key": key}
	resp, err := c.put(ctx, "/v1/sys/unseal", body)
	if err != nil {
		return nil, fmt.Errorf("submitting unseal key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from unseal endpoint", resp.StatusCode)
	}

	var status UnsealStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding unseal response: %w", err)
	}
	return &status, nil
}

// ResetUnsealProgress cancels any in-progress unseal attempt.
func (c *Client) ResetUnsealProgress(ctx context.Context) (*UnsealStatus, error) {
	body := map[string]bool{"reset": true}
	resp, err := c.put(ctx, "/v1/sys/unseal", body)
	if err != nil {
		return nil, fmt.Errorf("resetting unseal progress: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from unseal reset", resp.StatusCode)
	}

	var status UnsealStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding unseal reset response: %w", err)
	}
	return &status, nil
}
