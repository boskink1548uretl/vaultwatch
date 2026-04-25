package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// RemediationClient is the interface required to apply remediations.
type RemediationClient interface {
	ApplyRemediation(ctx context.Context, action vault.RemediationAction) (*vault.RemediationResult, error)
}

// RemediationWatcher applies automatic remediations for secrets matching
// a configured action and path prefix.
type RemediationWatcher struct {
	client RemediationClient
	action string
	paths  []string
	logger *log.Logger
}

// NewRemediationWatcher creates a watcher that will attempt remediation
// on the given paths using the specified action (e.g. "rotate", "revoke").
func NewRemediationWatcher(client RemediationClient, action string, paths []string, out io.Writer) *RemediationWatcher {
	if out == nil {
		out = os.Stdout
	}
	return &RemediationWatcher{
		client: client,
		action: action,
		paths:  paths,
		logger: log.New(out, "[remediation] ", log.LstdFlags),
	}
}

// RunAll attempts remediation on all configured paths and returns a
// slice of results, continuing on individual failures.
func (w *RemediationWatcher) RunAll(ctx context.Context) []vault.RemediationResult {
	results := make([]vault.RemediationResult, 0, len(w.paths))
	for _, path := range w.paths {
		action := vault.RemediationAction{
			Path:        path,
			Action:      w.action,
			Description: fmt.Sprintf("auto-remediation: %s on %s", w.action, path),
		}
		result, err := w.client.ApplyRemediation(ctx, action)
		if err != nil {
			w.logger.Printf("failed to remediate %s: %v", path, err)
			results = append(results, vault.RemediationResult{
				Action:  action,
				Success: false,
				Message: err.Error(),
			})
			continue
		}
		w.logger.Printf("remediation applied: %s on %s", w.action, path)
		results = append(results, *result)
	}
	return results
}
