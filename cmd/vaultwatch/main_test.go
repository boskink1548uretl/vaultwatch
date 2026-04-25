package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runAsCrasher re-executes the current test binary with BE_CRASHER=1 and the
// given test name, then asserts that the subprocess exits with a non-zero
// status code. It returns true when the caller should act as the subprocess.
func runAsCrasher(t *testing.T) bool {
	t.Helper()
	if os.Getenv("BE_CRASHER") == "1" {
		return true
	}
	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return false // expected non-zero exit — test passes
	}
	t.Fatalf("expected non-zero exit, got: %v", err)
	return false
}

// TestMain_MissingConfig verifies the binary exits non-zero when config is absent.
func TestMain_MissingConfig(t *testing.T) {
	if runAsCrasher(t) {
		os.Args = []string{"vaultwatch", "--config", "/nonexistent/path.yaml"}
		main()
	}
}

// TestMain_InvalidConfig verifies the binary exits non-zero on a malformed config.
func TestMain_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(cfgPath, []byte(":::invalid yaml:::"), 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	if runAsCrasher(t) {
		os.Args = []string{"vaultwatch", "--config", cfgPath}
		main()
	}
}
