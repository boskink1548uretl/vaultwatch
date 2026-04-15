package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// SlackNotifier sends alert notifications to a Slack webhook.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// slackPayload represents the JSON body sent to a Slack incoming webhook.
type slackPayload struct {
	Text        string            `json:"text"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

type slackAttachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
}

// NewSlackNotifier creates a SlackNotifier that posts to the given webhook URL.
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a Slack message for the given CheckResult if it warrants an alert.
func (s *SlackNotifier) Notify(result monitor.CheckResult) error {
	if result.Level == monitor.LevelNone {
		return nil
	}

	color := colorForLevel(result.Level)
	body := slackPayload{
		Text: fmt.Sprintf(":rotating_light: VaultWatch Alert: *%s*", result.SecretPath),
		Attachments: []slackAttachment{
			{
				Color:  color,
				Title:  fmt.Sprintf("Alert Level: %s", result.Level),
				Text:   fmt.Sprintf("Secret `%s` expires in %s.", result.SecretPath, result.TimeUntilExpiry.Round(time.Minute)),
				Footer: "vaultwatch",
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("slack: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func colorForLevel(level monitor.AlertLevel) string {
	switch level {
	case monitor.LevelWarning:
		return "warning"
	case monitor.LevelCritical:
		return "danger"
	case monitor.LevelExpired:
		return "#000000"
	default:
		return "good"
	}
}
