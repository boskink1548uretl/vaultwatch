// Package report provides utilities for aggregating and formatting
// the results of a VaultWatch monitoring cycle into human-readable
// summaries.
//
// Usage:
//
//	results := checker.EvaluateAll(secrets)
//	summary := report.Build(results)
//	report.Write(os.Stdout, summary)
//
// The summary includes per-level counts (OK, Warning, Critical, Expired)
// and a tabular listing of each secret path with its remaining TTL.
//
// Output format:
//
// Secrets are grouped by status level and sorted by remaining TTL in
// ascending order, so the most urgent secrets appear first. Each row
// includes the secret path, status label, and a human-readable duration
// (e.g. "2d 4h", "expired").
//
// Filtering:
//
// Use [BuildFiltered] to restrict the summary to secrets at or above a
// given severity level, which is useful for alerting pipelines that
// should only emit output when actionable issues are present.
package report
