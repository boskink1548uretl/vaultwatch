package vault

import (
	"context"
	"fmt"
	"time"
)

// TokenInfo holds metadata about the current Vault token.
type TokenInfo struct {
	Accessor    string
	DisplayName string
	Policies    []string
	TTL         time.Duration
	Renewable   bool
	ExpireTime  time.Time
}

// GetTokenInfo retrieves metadata about the currently authenticated token.
func (c *Client) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	path := "auth/token/lookup-self"
	secret, err := c.logical.ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("token lookup failed: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("token lookup returned empty response")
	}

	ttlRaw, _ := secret.Data["ttl"].(float64)
	ttl := time.Duration(ttlRaw) * time.Second

	renewable, _ := secret.Data["renewable"].(bool)
	displayName, _ := secret.Data["display_name"].(string)
	accessor, _ := secret.Data["accessor"].(string)

	var policies []string
	if raw, ok := secret.Data["policies"].([]interface{}); ok {
		for _, p := range raw {
			if s, ok := p.(string); ok {
				policies = append(policies, s)
			}
		}
	}

	expireTime := time.Now().Add(ttl)

	return &TokenInfo{
		Accessor:    accessor,
		DisplayName: displayName,
		Policies:    policies,
		TTL:         ttl,
		Renewable:   renewable,
		ExpireTime:  expireTime,
	}, nil
}

// IsTokenExpiringSoon returns true if the token TTL is within the given threshold.
func (c *Client) IsTokenExpiringSoon(ctx context.Context, threshold time.Duration) (bool, error) {
	info, err := c.GetTokenInfo(ctx)
	if err != nil {
		return false, err
	}
	return info.TTL <= threshold, nil
}
