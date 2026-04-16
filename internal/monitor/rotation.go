package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// RotationPolicy defines what action to take when a secret reaches a threshold.
type RotationPolicy struct {
	// AutoRotate enables automatic rotation when a secret is at critical or expired level.
	AutoRotate bool
	// NewDataFn is called to generate replacement secret data for a given path.
	// If nil, an empty map is used.
	NewDataFn func(path string) map[string]interface{}
}

// RotationManager handles automatic secret rotation based on check results.
type RotationManager struct {
	client *vault.Client
	policy RotationPolicy
}

// NewRotationManager creates a RotationManager with the given client and policy.
func NewRotationManager(client *vault.Client, policy RotationPolicy) *RotationManager {
	return &RotationManager{client: client, policy: policy}
}

// MaybeRotate inspects a CheckResult and rotates the secret if policy requires it.
// Returns the RotateResult if rotation was performed, or nil if skipped.
func (rm *RotationManager) MaybeRotate(ctx context.Context, result CheckResult) (*vault.RotateResult, error) {
	if !rm.policy.AutoRotate {
		return nil, nil
	}

	if result.Level != LevelCritical && result.Level != LevelExpired {
		return nil, nil
	}

	newData := map[string]interface{}{}
	if rm.policy.NewDataFn != nil {
		newData = rm.policy.NewDataFn(result.Path)
	}

	log.Printf("[rotation] rotating secret %s (level=%s)", result.Path, result.Level)

	rr, err := rm.client.RotateSecret(ctx, result.Path, newData)
	if err != nil {
		return nil, fmt.Errorf("auto-rotate %s: %w", result.Path, err)
	}

	log.Printf("[rotation] rotated %s → version %d", rr.Path, rr.NewVersion)
	return rr, nil
}

// MaybeRotateAll runs MaybeRotate for each result in the provided slice, collecting
// all rotation results. If any rotation fails, the error is returned immediately.
func (rm *RotationManager) MaybeRotateAll(ctx context.Context, results []CheckResult) ([]*vault.RotateResult, error) {
	var rotated []*vault.RotateResult
	for _, result := range results {
		rr, err := rm.MaybeRotate(ctx, result)
		if err != nil {
			return rotated, err
		}
		if rr != nil {
			rotated = append(rotated, rr)
		}
	}
	return rotated, nil
}
