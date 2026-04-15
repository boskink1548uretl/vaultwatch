// Package vault provides a thin client around the HashiCorp Vault HTTP API.
//
// It exposes the operations required by vaultwatch:
//
//   - NewClient – creates an authenticated Vault client.
//   - IsHealthy  – checks Vault's seal/init status.
//   - GetSecretMetadata – retrieves KV v2 metadata (including custom_metadata
//     fields such as "expires_at") for a single secret path.
//   - ListSecrets – recursively enumerates all secret paths under a given
//     KV v2 mount and path prefix, following folder entries automatically.
//
// All methods accept a context so callers can enforce timeouts and
// propagate cancellation through the scheduler loop.
package vault
