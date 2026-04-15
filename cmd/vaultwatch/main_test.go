package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain_MissingConfig verifies the binary exits non-zero when config is absent.
func TestMain_MissingConfig(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		os.Args = []string{"vaultwatch", "--config", "/nonexistent/path.yaml"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MissingConfig")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return // expected non-zero exit
	}
	t.Fatalf("expected non-zero exit, got: %v", err)
}

// TestMain_InvalidConfig verifies the binary exits non-zero on a malformed config.
func TestMain_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(cfgPath, []byte(":::invalid yaml:::"), 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	if os.Getenv("BE_CRASHER") == "1" {
		os.Args = []string{"vaultwatch", "--config", cfgPath}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_InvalidConfig")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("expected non-zero exit, got: %v", err)
}
