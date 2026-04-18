package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// EntityWatcher checks for disabled or policy-less identity entities.
type EntityWatcher struct {
	client           *vault.Client
	requiredPolicies []string
	logger           *log.Logger
}

// NewEntityWatcher creates an EntityWatcher.
func NewEntityWatcher(client *vault.Client, requiredPolicies []string, logger *log.Logger) *EntityWatcher {
	if logger == nil {
		logger = log.Default()
	}
	return &EntityWatcher{client: client, requiredPolicies: requiredPolicies, logger: logger}
}

// Check lists all entities and logs warnings for disabled or bare entities.
func (ew *EntityWatcher) Check(ctx context.Context) error {
	names, err := ew.client.ListEntities(ctx)
	if err != nil {
		return fmt.Errorf("entity watcher: list: %w", err)
	}

	for _, name := range names {
		e, err := ew.client.GetEntity(ctx, name)
		if err != nil {
			ew.logger.Printf("[entity] error fetching %s: %v", name, err)
			continue
		}
		if e == nil {
			continue
		}
		if e.Disabled {
			ew.logger.Printf("[entity] WARNING: entity %q is disabled", e.Name)
		}
		if len(e.Policies) == 0 {
			ew.logger.Printf("[entity] WARNING: entity %q has no policies attached", e.Name)
		}
		for _, required := range ew.requiredPolicies {
			if !containsPolicy(e.Policies, required) {
				ew.logger.Printf("[entity] WARNING: entity %q missing required policy %q", e.Name, required)
			}
		}
	}
	return nil
}

func containsPolicy(policies []string, target string) bool {
	for _, p := range policies {
		if p == target {
			return true
		}
	}
	return false
}
