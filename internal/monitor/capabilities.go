package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// CapabilityClient defines the interface for fetching token capabilities.
type CapabilityClient interface {
	GetCapabilities(ctx context.Context, paths []string) (map[string][]string, error)
}

// CapabilityWatcher monitors whether the Vault token has expected capabilities on critical paths.
type CapabilityWatcher struct {
	client    CapabilityClient
	paths     []string
	required  []string
	logger    *log.Logger
}

// CapabilityIssue describes a missing or unexpected capability on a path.
type CapabilityIssue struct {
	Path    string
	Missing []string
}

// NewCapabilityWatcher creates a CapabilityWatcher for the given paths and required capabilities.
func NewCapabilityWatcher(client CapabilityClient, paths, required []string, w io.Writer) *CapabilityWatcher {
	if w == nil {
		w = os.Stderr
	}
	return &CapabilityWatcher{
		client:   client,
		paths:    paths,
		required: required,
		logger:   log.New(w, "[capability-watcher] ", log.LstdFlags),
	}
}

// Check evaluates the token's capabilities on each configured path and returns any issues found.
func (w *CapabilityWatcher) Check(ctx context.Context) ([]CapabilityIssue, error) {
	if len(w.paths) == 0 {
		return nil, nil
	}

	caps, err := w.client.GetCapabilities(ctx, w.paths)
	if err != nil {
		return nil, fmt.Errorf("get capabilities: %w", err)
	}

	var issues []CapabilityIssue
	for _, path := range w.paths {
		pathCaps := caps[path]
		var missing []string
		for _, req := range w.required {
			if !containsCapability(pathCaps, req) {
				missing = append(missing, req)
			}
		}
		if len(missing) > 0 {
			w.logger.Printf("path %q missing capabilities: %s", path, strings.Join(missing, ", "))
			issues = append(issues, CapabilityIssue{Path: path, Missing: missing})
		}
	}
	return issues, nil
}

func containsCapability(caps []string, target string) bool {
	for _, c := range caps {
		if c == target || c == "root" {
			return true
		}
	}
	return false
}
