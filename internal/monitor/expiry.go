// Package monitor provides functionality for checking secret expiration
// and determining which secrets require alerting based on configured thresholds.
package monitor

import (
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// AlertLevel represents the urgency of a secret expiration alert.
type AlertLevel string

const (
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
	AlertLevelExpired  AlertLevel = "expired"
)

// ExpiryAlert holds information about a secret that requires attention.
type ExpiryAlert struct {
	SecretPath  string
	ExpiresAt   time.Time
	TimeUntil   time.Duration
	Level       AlertLevel
	Message     string
}

// Checker evaluates secret metadata against configured thresholds.
type Checker struct {
	WarningThreshold  time.Duration
	CriticalThreshold time.Duration
}

// NewChecker creates a Checker with the provided alert thresholds.
func NewChecker(warning, critical time.Duration) *Checker {
	return &Checker{
		WarningThreshold:  warning,
		CriticalThreshold: critical,
	}
}

// Evaluate inspects secret metadata and returns an ExpiryAlert if the
// secret is within a warning or critical threshold, or already expired.
// Returns nil if the secret does not require alerting.
func (c *Checker) Evaluate(path string, meta *vault.SecretMetadata) *ExpiryAlert {
	if meta == nil || meta.ExpiresAt.IsZero() {
		return nil
	}

	now := time.Now()
	timeUntil := meta.ExpiresAt.Sub(now)

	var level AlertLevel
	var message string

	switch {
	case timeUntil <= 0:
		level = AlertLevelExpired
		message = fmt.Sprintf("secret %q has expired", path)
	case timeUntil <= c.CriticalThreshold:
		level = AlertLevelCritical
		message = fmt.Sprintf("secret %q expires in %s (critical threshold: %s)", path, timeUntil.Round(time.Second), c.CriticalThreshold)
	case timeUntil <= c.WarningThreshold:
		level = AlertLevelWarning
		message = fmt.Sprintf("secret %q expires in %s (warning threshold: %s)", path, timeUntil.Round(time.Second), c.WarningThreshold)
	default:
		return nil
	}

	return &ExpiryAlert{
		SecretPath: path,
		ExpiresAt:  meta.ExpiresAt,
		TimeUntil:  timeUntil,
		Level:      level,
		Message:    message,
	}
}
