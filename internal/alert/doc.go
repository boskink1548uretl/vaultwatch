// Package alert provides notifier implementations for sending Vault secret
// expiration alerts to various destinations.
//
// Supported notifiers:
//   - StdoutNotifier: writes human-readable alerts to standard output.
//   - SlackNotifier: posts rich formatted messages to a Slack webhook URL.
//   - PagerDutyNotifier: triggers incidents via the PagerDuty Events API v2.
//
// All notifiers accept a monitor.CheckResult and only send an alert when the
// result level is Warning, Critical, or Expired. LevelOK results are silently
// ignored, making it safe to call Notify on every evaluation tick.
package alert
