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
package report
