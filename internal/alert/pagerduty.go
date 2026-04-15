package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	client         *http.Client
	eventsURL      string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	DedupKey    string    `json:"dedup_key"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

// NewPagerDutyNotifier creates a PagerDutyNotifier with the given integration key.
func NewPagerDutyNotifier(integrationKey string) *PagerDutyNotifier {
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 10 * time.Second},
		eventsURL:      pagerDutyEventsURL,
	}
}

// Notify sends a PagerDuty event if the check result warrants an alert.
func (p *PagerDutyNotifier) Notify(result monitor.CheckResult) error {
	if result.Level == monitor.LevelOK {
		return nil
	}

	severity := severityForLevel(result.Level)
	summary := fmt.Sprintf("[%s] Vault secret %s expires in %s",
		result.Level, result.SecretPath, result.TimeRemaining.Round(time.Minute))
	if result.Level == monitor.LevelExpired {
		summary = fmt.Sprintf("[EXPIRED] Vault secret %s has expired", result.SecretPath)
	}

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		DedupKey:    fmt.Sprintf("vaultwatch-%s", result.SecretPath),
		Payload: pdDetails{
			Summary:   summary,
			Source:    "vaultwatch",
			Severity:  severity,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := p.client.Post(p.eventsURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func severityForLevel(level monitor.AlertLevel) string {
	switch level {
	case monitor.LevelCritical, monitor.LevelExpired:
		return "critical"
	case monitor.LevelWarning:
		return "warning"
	default:
		return "info"
	}
}
