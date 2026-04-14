// Package alert provides notification backends for VaultWatch alerts.
package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/monitor"
)

// Severity mirrors the monitor package level for alert routing.
type Severity string

const (
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
	SeverityExpired  Severity = "EXPIRED"
)

// Notification holds the data sent to a notifier.
type Notification struct {
	SecretPath string
	Severity   Severity
	ExpiresAt  time.Time
	TimeLeft   time.Duration
	Message    string
}

// Notifier is the interface implemented by all alert backends.
type Notifier interface {
	Send(n Notification) error
	Name() string
}

// FromCheckResult converts a monitor.CheckResult into a Notification.
// Returns nil if the result requires no alert.
func FromCheckResult(r monitor.CheckResult) *Notification {
	var sev Severity
	switch r.Level {
	case monitor.LevelWarning:
		sev = SeverityWarning
	case monitor.LevelCritical:
		sev = SeverityCritical
	case monitor.LevelExpired:
		sev = SeverityExpired
	default:
		return nil
	}

	msg := fmt.Sprintf("[%s] Secret '%s' expires in %s (at %s)",
		sev, r.SecretPath, r.TimeLeft.Round(time.Second), r.ExpiresAt.UTC().Format(time.RFC3339))

	return &Notification{
		SecretPath: r.SecretPath,
		Severity:   sev,
		ExpiresAt:  r.ExpiresAt,
		TimeLeft:   r.TimeLeft,
		Message:    msg,
	}
}

// StdoutNotifier writes alert notifications to an io.Writer (default: os.Stdout).
type StdoutNotifier struct {
	w io.Writer
}

// NewStdoutNotifier creates a StdoutNotifier writing to the given writer.
// If w is nil, os.Stdout is used.
func NewStdoutNotifier(w io.Writer) *StdoutNotifier {
	if w == nil {
		w = os.Stdout
	}
	return &StdoutNotifier{w: w}
}

// Name returns the backend identifier.
func (s *StdoutNotifier) Name() string { return "stdout" }

// Send writes the notification message followed by a newline.
func (s *StdoutNotifier) Send(n Notification) error {
	_, err := fmt.Fprintln(s.w, n.Message)
	return err
}
