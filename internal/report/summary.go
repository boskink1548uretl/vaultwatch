// Package report provides functionality for generating human-readable
// summaries of secret expiry check results.
package report

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// Summary holds aggregated results from a monitoring cycle.
type Summary struct {
	GeneratedAt time.Time
	Total       int
	OK          int
	Warning     int
	Critical    int
	Expired     int
	Results     []monitor.CheckResult
}

// Build aggregates a slice of CheckResults into a Summary.
func Build(results []monitor.CheckResult) Summary {
	s := Summary{
		GeneratedAt: time.Now().UTC(),
		Total:       len(results),
		Results:     results,
	}
	for _, r := range results {
		switch r.Level {
		case monitor.LevelOK:
			s.OK++
		case monitor.LevelWarning:
			s.Warning++
		case monitor.LevelCritical:
			s.Critical++
		case monitor.LevelExpired:
			s.Expired++
		}
	}
	return s
}

// Write renders the summary as a formatted report to the given writer.
func Write(w io.Writer, s Summary) error {
	line := strings.Repeat("-", 60)
	_, err := fmt.Fprintf(w,
		"%s\nVaultWatch Report — %s\n%s\nTotal: %d | OK: %d | Warning: %d | Critical: %d | Expired: %d\n%s\n",
		line,
		s.GeneratedAt.Format(time.RFC1123),
		line,
		s.Total, s.OK, s.Warning, s.Critical, s.Expired,
		line,
	)
	if err != nil {
		return err
	}
	for _, r := range s.Results {
		ttl := "expired"
		if r.TTL > 0 {
			ttl = r.TTL.Round(time.Second).String()
		}
		_, err = fmt.Fprintf(w, "[%-8s] %-40s TTL: %s\n", r.Level, r.SecretPath, ttl)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w, line)
	return err
}
