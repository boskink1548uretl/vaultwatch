// Package vault provides a client for interacting with HashiCorp Vault's
// HTTP API.
//
// # Profile
//
// The profile functions retrieve a high-level summary of the Vault instance
// including its version, cluster identity, HA status, seal state, and
// initialization status. This is primarily used for health dashboards and
// startup diagnostics.
//
// Example usage:
//
//	client, _ := vault.NewClient(addr, token)
//	profile, err := client.GetProfile(ctx)
//	if err != nil { ... }
//	fmt.Println(profile.Version, profile.ClusterName)
package vault
