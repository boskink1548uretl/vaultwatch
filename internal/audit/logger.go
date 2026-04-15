// Package audit provides structured audit logging for secret check events.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp  time.Time `json:"timestamp"`
	SecretPath string    `json:"secret_path"`
	Level      string    `json:"level"`
	DaysLeft   float64   `json:"days_left"`
	Message    string    `json:"message"`
}

// Logger writes audit entries as newline-delimited JSON.
type Logger struct {
	w io.Writer
}

// NewLogger returns a Logger that writes to the given path.
// Pass an empty path to write to stdout.
func NewLogger(path string) (*Logger, error) {
	if path == "" {
		return &Logger{w: os.Stdout}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	return &Logger{w: f}, nil
}

// Record writes an audit entry derived from a CheckResult.
func (l *Logger) Record(r monitor.CheckResult) error {
	entry := Entry{
		Timestamp:  time.Now().UTC(),
		SecretPath: r.SecretPath,
		Level:      r.Level.String(),
		DaysLeft:   r.TTL.Hours() / 24,
		Message:    buildMessage(r),
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(l.w, "%s\n", b)
	return err
}

func buildMessage(r monitor.CheckResult) string {
	if r.TTL <= 0 {
		return fmt.Sprintf("secret %s has expired", r.SecretPath)
	}
	days := r.TTL.Hours() / 24
	return fmt.Sprintf("secret %s expires in %.1f days", r.SecretPath, days)
}
