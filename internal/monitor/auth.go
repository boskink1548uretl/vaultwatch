package monitor

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// AuthWatcher monitors Vault auth methods for misconfigurations.
type AuthWatcher struct {
	client *vault.Client
	logger *log.Logger
	// RiskyTypes flags auth method types considered high-risk if enabled.
	RiskyTypes []string
}

// NewAuthWatcher creates an AuthWatcher with sensible defaults.
func NewAuthWatcher(client *vault.Client, logger *log.Logger) *AuthWatcher {
	return &AuthWatcher{
		client:     client,
		logger:     logger,
		RiskyTypes: []string{"userpass", "ldap"},
	}
}

// AuthFinding describes a potential issue with an auth method.
type AuthFinding struct {
	Path    string
	Type    string
	Message string
}

// Check inspects all enabled auth methods and returns findings.
func (w *AuthWatcher) Check(ctx context.Context) ([]AuthFinding, error) {
	methods, err := w.client.ListAuthMethods(ctx)
	if err != nil {
		return nil, fmt.Errorf("list auth methods: %w", err)
	}

	var findings []AuthFinding
	for path, m := range methods {
		if w.isRisky(m.Type) {
			f := AuthFinding{
				Path:    strings.TrimSuffix(path, "/"),
				Type:    m.Type,
				Message: fmt.Sprintf("auth method %q (%s) is enabled and may pose a risk", path, m.Type),
			}
			findings = append(findings, f)
			w.logger.Printf("[WARN] %s", f.Message)
		}
	}
	return findings, nil
}

func (w *AuthWatcher) isRisky(authType string) bool {
	for _, r := range w.RiskyTypes {
		if r == authType {
			return true
		}
	}
	return false
}
