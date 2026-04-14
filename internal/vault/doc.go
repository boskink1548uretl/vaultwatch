// Package vault provides a thin wrapper around the HashiCorp Vault API client
// tailored for vaultwatch's secret expiration monitoring use case.
//
// It exposes:
//   - Client: a configured Vault API client with health-check support.
//   - SecretMetadata: a struct capturing path, TTL, expiration time, and
//     renewability of a Vault secret lease.
//
// Usage:
//
//	client, err := vault.NewClient("https://vault.example.com", os.Getenv("VAULT_TOKEN"))
//	if err != nil { ... }
//
//	if err := client.IsHealthy(); err != nil { ... }
//
//	meta, err := client.GetSecretMetadata("secret/myapp/db")
//	if err != nil { ... }
//	fmt.Printf("Expires at: %s (TTL: %s)\n", meta.ExpiresAt, meta.TTL)
package vault
