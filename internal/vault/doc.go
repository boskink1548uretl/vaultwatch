// Package vault provides a thin client wrapper around the HashiCorp Vault HTTP
// API used by vaultwatch.
//
// It exposes three primary capabilities:
//
//   - Health checking via [Client.IsHealthy]
//   - Secret metadata retrieval via [Client.GetSecretMetadata]
//   - Recursive secret path listing via [ListSecrets]
//   - Secret renewal via [Client.RenewSecret], which handles both KV v2
//     (re-write in place) and dynamic secrets (lease renewal).
//
// All methods accept a [context.Context] for timeout and cancellation support.
package vault
