// Package audit provides structured, append-only audit logging for
// VaultWatch secret check events.
//
// Each check result is serialised as a newline-delimited JSON record
// containing the secret path, alert level, remaining TTL in days, a
// human-readable message, and a UTC timestamp.
//
// Usage:
//
//	logger, err := audit.NewLogger("/var/log/vaultwatch/audit.log")
//	if err != nil { ... }
//	logger.Record(checkResult)
//
// Pass an empty path to write audit entries to stdout instead.
package audit
