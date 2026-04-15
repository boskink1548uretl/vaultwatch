package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/report"
)

func makeResult(path string, level monitor.AlertLevel, ttl time.Duration) monitor.CheckResult {
	return monitor.CheckResult{
		SecretPath: path,
		Level:      level,
		TTL:        ttl,
	}
}

func TestBuild_CountsCorrectly(t *testing.T) {
	results := []monitor.CheckResult{
		makeResult("secret/a", monitor.LevelOK, 72*time.Hour),
		makeResult("secret/b", monitor.LevelWarning, 48*time.Hour),
		makeResult("secret/c", monitor.LevelCritical, 6*time.Hour),
		makeResult("secret/d", monitor.LevelExpired, 0),
		makeResult("secret/e", monitor.LevelOK, 100*time.Hour),
	}

	s := report.Build(results)

	if s.Total != 5 {
		t.Errorf("expected Total=5, got %d", s.Total)
	}
	if s.OK != 2 {
		t.Errorf("expected OK=2, got %d", s.OK)
	}
	if s.Warning != 1 {
		t.Errorf("expected Warning=1, got %d", s.Warning)
	}
	if s.Critical != 1 {
		t.Errorf("expected Critical=1, got %d", s.Critical)
	}
	if s.Expired != 1 {
		t.Errorf("expected Expired=1, got %d", s.Expired)
	}
}

func TestBuild_EmptyResults(t *testing.T) {
	s := report.Build(nil)
	if s.Total != 0 || s.OK != 0 {
		t.Errorf("expected all-zero summary for empty input")
	}
}

func TestWrite_ContainsSecretPaths(t *testing.T) {
	results := []monitor.CheckResult{
		makeResult("secret/db/password", monitor.LevelCritical, 4*time.Hour),
		makeResult("secret/api/key", monitor.LevelOK, 200*time.Hour),
	}
	s := report.Build(results)

	var buf bytes.Buffer
	if err := report.Write(&buf, s); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "secret/db/password") {
		t.Error("expected output to contain 'secret/db/password'")
	}
	if !strings.Contains(output, "secret/api/key") {
		t.Error("expected output to contain 'secret/api/key'")
	}
	if !strings.Contains(output, "Critical") {
		t.Error("expected output to contain 'Critical'")
	}
}

func TestWrite_ExpiredShowsExpiredLabel(t *testing.T) {
	results := []monitor.CheckResult{
		makeResult("secret/old", monitor.LevelExpired, 0),
	}
	s := report.Build(results)

	var buf bytes.Buffer
	if err := report.Write(&buf, s); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "expired") {
		t.Error("expected 'expired' label for expired secret")
	}
}
