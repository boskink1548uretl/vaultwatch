package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// LoginResult holds the token and metadata returned after a successful auth login.
type LoginResult struct {
	ClientToken   string            `json:"client_token"`
	Accessor      string            `json:"accessor"`
	Policies      []string          `json:"policies"`
	LeaseDuration int               `json:"lease_duration"`
	Renewable     bool              `json:"renewable"`
	Metadata      map[string]string `json:"metadata"`
}

// loginResponse wraps the Vault API auth response envelope.
type loginResponse struct {
	Auth *LoginResult `json:"auth"`
}

// LoginWithAppRole authenticates using the AppRole method and returns a LoginResult.
func (c *Client) LoginWithAppRole(ctx context.Context, roleID, secretID string) (*LoginResult, error) {
	if roleID == "" {
		return nil, fmt.Errorf("roleID must not be empty")
	}
	if secretID == "" {
		return nil, fmt.Errorf("secretID must not be empty")
	}

	payload := map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	body, err := c.post(ctx, "/v1/auth/approle/login", payload)
	if err != nil {
		return nil, fmt.Errorf("approle login request failed: %w", err)
	}

	var resp loginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}
	if resp.Auth == nil {
		return nil, fmt.Errorf("login response contained no auth block")
	}
	return resp.Auth, nil
}

// LoginWithToken validates that the provided token is accepted by Vault and
// returns a LoginResult populated from the token's own-lookup endpoint.
func (c *Client) LoginWithToken(ctx context.Context, token string) (*LoginResult, error) {
	if token == "" {
		return nil, fmt.Errorf("token must not be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token lookup request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("token is invalid or expired (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from token lookup", resp.StatusCode)
	}

	var data struct {
		Data struct {
			Accessor      string   `json:"accessor"`
			Policies      []string `json:"policies"`
			LeaseDuration int      `json:"ttl"`
			Renewable     bool     `json:"renewable"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode token lookup response: %w", err)
	}
	return &LoginResult{
		ClientToken:   token,
		Accessor:      data.Data.Accessor,
		Policies:      data.Data.Policies,
		LeaseDuration: data.Data.LeaseDuration,
		Renewable:     data.Data.Renewable,
	}, nil
}
