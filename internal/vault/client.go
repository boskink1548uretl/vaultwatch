package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	api     *vaultapi.Client
	Address string
}

// SecretMetadata holds expiration info for a Vault secret.
type SecretMetadata struct {
	Path       string
	ExpiresAt  time.Time
	TTL        time.Duration
	Renewable  bool
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	if token != "" {
		api.SetToken(token)
	}

	return &Client{
		api:     api,
		Address: address,
	}, nil
}

// GetSecretMetadata retrieves lease/expiration metadata for a KV secret path.
func (c *Client) GetSecretMetadata(path string) (*SecretMetadata, error) {
	secret, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	var ttl time.Duration
	var expiresAt time.Time
	var renewable bool

	if secret.LeaseDuration > 0 {
		ttl = time.Duration(secret.LeaseDuration) * time.Second
		expiresAt = time.Now().Add(ttl)
	}
	if secret.Renewable {
		renewable = true
	}

	return &SecretMetadata{
		Path:      path,
		ExpiresAt: expiresAt,
		TTL:       ttl,
		Renewable: renewable,
	}, nil
}

// IsHealthy checks whether the Vault server is reachable and unsealed.
func (c *Client) IsHealthy() error {
	health, err := c.api.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}
	return nil
}
