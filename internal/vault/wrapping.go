package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WrappedSecret holds metadata about a response-wrapped secret token.
type WrappedSecret struct {
	Token          string `json:"token"`
	Accessor       string `json:"accessor"`
	TTL            int    `json:"ttl"`
	CreationTime   string `json:"creation_time"`
	CreationPath   string `json:"creation_path"`
	WrappedAccessor string `json:"wrapped_accessor"`
}

// WrapSecret wraps the data at the given path using Vault's response-wrapping
// with the specified TTL (e.g. "300s", "5m").
func (c *Client) WrapSecret(ctx context.Context, path string, ttl string) (*WrappedSecret, error) {
	if path == "" {
		return nil, fmt.Errorf("wrapping: path must not be empty")
	}
	if ttl == "" {
		ttl = "300s"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/"+path, nil)
	if err != nil {
		return nil, fmt.Errorf("wrapping: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("X-Vault-Wrap-TTL", ttl)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wrapping: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("wrapping: path %q not found", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrapping: unexpected status %d", resp.StatusCode)
	}

	var envelope struct {
		WrapInfo *WrappedSecret `json:"wrap_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("wrapping: decode response: %w", err)
	}
	if envelope.WrapInfo == nil {
		return nil, fmt.Errorf("wrapping: response contained no wrap_info (wrapping may be disabled)")
	}
	return envelope.WrapInfo, nil
}

// LookupWrappingToken inspects a wrapping token without unwrapping it,
// returning its metadata.
func (c *Client) LookupWrappingToken(ctx context.Context, wrappingToken string) (*WrappedSecret, error) {
	if wrappingToken == "" {
		return nil, fmt.Errorf("lookup wrapping token: token must not be empty")
	}

	body, err := jsonBody(map[string]string{"token": wrappingToken})
	if err != nil {
		return nil, fmt.Errorf("lookup wrapping token: encode body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.address+"/v1/sys/wrapping/lookup", body)
	if err != nil {
		return nil, fmt.Errorf("lookup wrapping token: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lookup wrapping token: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lookup wrapping token: token not found or already unwrapped")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookup wrapping token: unexpected status %d", resp.StatusCode)
	}

	var envelope struct {
		Data *WrappedSecret `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("lookup wrapping token: decode response: %w", err)
	}
	return envelope.Data, nil
}
