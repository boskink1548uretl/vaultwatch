// Package scheduler implements a periodic polling loop for VaultWatch.
//
// The Scheduler fetches secret metadata from Vault at a configurable interval,
// evaluates each secret against warning and critical thresholds via the monitor
// package, and dispatches any resulting alerts through the configured notifier.
//
// Typical usage:
//
//	s := scheduler.New(client, checker, notifier, 5*time.Minute, paths)
//	if err := s.Run(ctx); err != nil && err != context.Canceled {
//		log.Fatalf("scheduler error: %v", err)
//	}
package scheduler
