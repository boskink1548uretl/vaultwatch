// Package monitor provides watchers and checkers for Vault operational health.
//
// ACL Watcher
//
// The ACLWatcher inspects the policies attached to the Vault token in use and
// reports two categories of findings:
//
//  1. Missing required policies — policies that the operator expects to be
//     present but are not currently attached to the token. These are emitted
//     as WARNING log lines so they surface in audit trails without blocking
//     normal operation.
//
//  2. Sensitive policy characteristics — the root policy and any policy rules
//     that use wildcard paths (e.g. "secret/*") are flagged as CRITICAL or
//     INFO respectively, prompting the operator to tighten access controls.
//
// Usage:
//
//	watcher := monitor.NewACLWatcher(vaultClient, []string{"ops", "default"}, os.Stderr)
//	if err := watcher.Check(ctx); err != nil {
//		log.Fatal(err)
//	}
package monitor
