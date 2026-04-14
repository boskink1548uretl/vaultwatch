// Package alert implements notification backends for VaultWatch.
//
// A Notifier sends a Notification when a secret is approaching or has
// passed its expiration deadline. Notifications are constructed from
// monitor.CheckResult values via FromCheckResult.
//
// Built-in backends:
//
//	- StdoutNotifier  – writes human-readable lines to any io.Writer.
//
// Additional backends (e.g. Slack, PagerDuty, email) can be added by
// implementing the Notifier interface.
package alert
