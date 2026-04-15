package monitor

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// PolicyFetcher abstracts fetching policy data from Vault.
type PolicyFetcher interface {
	ListPolicies(ctx context.Context) ([]string, error)
	GetPolicy(ctx context.Context, name string) (PolicyRules, error)
}

// PolicyRules holds the name and HCL rules of a Vault ACL policy.
type PolicyRules interface {
	GetName() string
	GetRules() string
}

// PolicyChecker inspects Vault policies for risky capability grants.
type PolicyChecker struct {
	fetcher  PolicyFetcher
	risky    []string
}

// NewPolicyChecker creates a PolicyChecker with default risky capabilities.
func NewPolicyChecker(fetcher PolicyFetcher) *PolicyChecker {
	return &PolicyChecker{
		fetcher: fetcher,
		risky:   []string{"sudo", "delete"},
	}
}

// PolicyFinding represents a potentially risky policy grant.
type PolicyFinding struct {
	PolicyName string
	Capability string
	Detail     string
}

// Audit lists all policies and flags those containing risky capabilities.
func (pc *PolicyChecker) Audit(ctx context.Context) ([]PolicyFinding, error) {
	names, err := pc.fetcher.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing policies: %w", err)
	}

	var findings []PolicyFinding
	for _, name := range names {
		if name == "root" {
			log.Printf("[policy] skipping built-in root policy")
			continue
		}
		pol, err := pc.fetcher.GetPolicy(ctx, name)
		if err != nil {
			log.Printf("[policy] could not fetch %q: %v", name, err)
			continue
		}
		for _, cap := range pc.risky {
			if strings.Contains(pol.GetRules(), cap) {
				findings = append(findings, PolicyFinding{
					PolicyName: pol.GetName(),
					Capability: cap,
					Detail:     fmt.Sprintf("policy %q grants %q capability", pol.GetName(), cap),
				})
			}
		}
	}
	return findings, nil
}
