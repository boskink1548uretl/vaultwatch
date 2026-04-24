package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// GroupWatcherClient is the interface for group-related Vault operations.
type GroupWatcherClient interface {
	ListGroups(ctx context.Context) ([]vault.Group, error)
	GetGroup(ctx context.Context, name string) (*vault.Group, error)
}

// GroupWatcher monitors Vault identity groups for policy compliance.
type GroupWatcher struct {
	client          GroupWatcherClient
	riskyPolicies   []string
	logger          *log.Logger
}

// GroupRisk describes a group that holds a risky policy assignment.
type GroupRisk struct {
	GroupName   string
	GroupID     string
	RiskyPolicy string
}

// NewGroupWatcher creates a GroupWatcher with the given client and risky policy list.
func NewGroupWatcher(client GroupWatcherClient, riskyPolicies []string, logger *log.Logger) *GroupWatcher {
	if logger == nil {
		logger = log.Default()
	}
	if len(riskyPolicies) == 0 {
		riskyPolicies = []string{"root", "admin", "superuser"}
	}
	return &GroupWatcher{
		client:        client,
		riskyPolicies: riskyPolicies,
		logger:        logger,
	}
}

// Audit lists all groups and returns any that are assigned a risky policy.
func (w *GroupWatcher) Audit(ctx context.Context) ([]GroupRisk, error) {
	groups, err := w.client.ListGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("group audit: list groups: %w", err)
	}

	var risks []GroupRisk
	for _, g := range groups {
		for _, policy := range g.Policies {
			if w.isRisky(policy) {
				w.logger.Printf("[WARN] group %q (id=%s) holds risky policy %q",
					g.Name, g.ID, policy)
				risks = append(risks, GroupRisk{
					GroupName:   g.Name,
					GroupID:     g.ID,
					RiskyPolicy: policy,
				})
				break
			}
		}
	}
	return risks, nil
}

func (w *GroupWatcher) isRisky(policy string) bool {
	for _, r := range w.riskyPolicies {
		if r == policy {
			return true
		}
	}
	return false
}
