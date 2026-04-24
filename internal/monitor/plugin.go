package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// PluginWatcher checks for non-builtin plugins that may pose a security risk
// or that lack a pinned SHA256 checksum.
type PluginWatcher struct {
	client *vault.Client
	logger *log.Logger
	types  []string
}

// NewPluginWatcher creates a PluginWatcher for the given plugin types.
// If types is empty, it defaults to ["auth", "secret", "database"].
func NewPluginWatcher(client *vault.Client, logger *log.Logger, types []string) *PluginWatcher {
	if len(types) == 0 {
		types = []string{"auth", "secret", "database"}
	}
	return &PluginWatcher{
		client: client,
		logger: logger,
		types:  types,
	}
}

// PluginFinding describes a plugin that requires attention.
type PluginFinding struct {
	Name    string
	Type    string
	Version string
	Reason  string
}

// Check inspects all registered plugins and returns findings for any
// non-builtin plugin that is missing a SHA256 checksum.
func (pw *PluginWatcher) Check(ctx context.Context) ([]PluginFinding, error) {
	var findings []PluginFinding

	for _, pt := range pw.types {
		plugins, err := pw.client.ListPlugins(ctx, pt)
		if err != nil {
			return nil, fmt.Errorf("list plugins (%s): %w", pt, err)
		}

		for _, p := range plugins {
			if p.Builtin {
				continue
			}
			if p.SHA256 == "" {
				findings = append(findings, PluginFinding{
					Name:    p.Name,
					Type:    p.Type,
					Version: p.Version,
					Reason:  "non-builtin plugin missing SHA256 checksum",
				})
				pw.logger.Printf("[WARN] plugin %s/%s has no SHA256 pin", pt, p.Name)
			}
		}
	}

	return findings, nil
}
