// Package monitor provides secret expiry checking and automated rotation
// capabilities for the vaultwatch CLI tool.
//
// # Checker
//
// The Checker evaluates each secret's remaining TTL against configured
// warning and critical thresholds, producing a CheckResult with an
// AlertLevel (none, warning, critical, or expired).
//
// # RotationManager
//
// The RotationManager consumes CheckResults and, when AutoRotate is
// enabled, calls the Vault client to rotate secrets that have reached
// the critical or expired threshold. Callers supply a NewDataFn to
// generate replacement secret payloads on a per-path basis.
//
// # Alert Levels
//
// Alert levels are ordered by severity:
//
//	- AlertLevelNone     – TTL is within acceptable bounds
//	- AlertLevelWarning  – TTL has fallen below the warning threshold
//	- AlertLevelCritical – TTL has fallen below the critical threshold
//	- AlertLevelExpired  – the secret's TTL has reached zero
package monitor
