// Package monitor provides secret expiry checking and automated rotation
// capabilities for the vaultwatch CLI tool.
//
// The Checker evaluates each secret's remaining TTL against configured
// warning and critical thresholds, producing a CheckResult with an
// AlertLevel (none, warning, critical, or expired).
//
// The RotationManager consumes CheckResults and, when AutoRotate is
// enabled, calls the Vault client to rotate secrets that have reached
// the critical or expired threshold. Callers supply a NewDataFn to
// generate replacement secret payloads on a per-path basis.
package monitor
