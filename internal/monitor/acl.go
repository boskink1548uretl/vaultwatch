package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// ACLClient is the subset of vault.Client used by ACLWatcher.
type ACLClient interface {
	GetACLTokenPolicies(ctx context.Context) ([]string, error)
	GetACLPolicy(ctx context.Context, name string) (*vault.ACLTokenPolicy, error)
}

// ACLWatcher checks that the configured Vault token holds a required set of
// policies and warns when sensitive wildcard policies are attached.
type ACLWatcher struct {
	client           ACLClient
	requiredPolicies []string
	logger           *log.Logger
}

// NewACLWatcher creates an ACLWatcher with the given required policy names.
func NewACLWatcher(client ACLClient, required []string, w io.Writer) *ACLWatcher {
	return &ACLWatcher{
		client:           client,
		requiredPolicies: required,
		logger:           log.New(w, "[acl] ", 0),
	}
}

// Check fetches the token's current policies and reports any issues.
func (a *ACLWatcher) Check(ctx context.Context) error {
	attached, err := a.client.GetACLTokenPolicies(ctx)
	if err != nil {
		return fmt.Errorf("acl watcher: fetch policies: %w", err)
	}

	attachedSet := make(map[string]struct{}, len(attached))
	for _, p := range attached {
		attachedSet[p] = struct{}{}
	}

	for _, req := range a.requiredPolicies {
		if _, ok := attachedSet[req]; !ok {
			a.logger.Printf("WARNING: required policy %q is not attached to the current token", req)
		}
	}

	for _, name := range attached {
		if strings.EqualFold(name, "root") {
			a.logger.Printf("CRITICAL: token has root policy attached — consider using a scoped token")
			continue
		}
		pol, err := a.client.GetACLPolicy(ctx, name)
		if err != nil {
			a.logger.Printf("WARNING: could not fetch policy %q: %v", name, err)
			continue
		}
		for _, rule := range pol.Rules {
			if strings.HasSuffix(rule.Path, "*") {
				a.logger.Printf("INFO: policy %q contains wildcard path %q", name, rule.Path)
			}
		}
	}
	return nil
}
