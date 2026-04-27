package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TelemetrySnapshot holds a subset of Vault telemetry metrics.
type TelemetrySnapshot struct {
	Counters []TelemetryMetric `json:"Counters"`
	Gauges   []TelemetryMetric `json:"Gauges"`
	Summaries []TelemetryMetric `json:"Summaries"`
}

// TelemetryMetric represents a single named metric with its current value.
type TelemetryMetric struct {
	Name   string             `json:"Name"`
	Labels map[string]string  `json:"Labels"`
	Count  int                `json:"Count"`
	Sum    float64            `json:"Sum"`
	Min    float64            `json:"Min"`
	Max    float64            `json:"Max"`
	Mean   float64            `json:"Mean"`
}

// GetTelemetry fetches the current telemetry snapshot from Vault.
// Vault must have unauthenticated_metrics_access enabled or the token must
// have access to sys/metrics.
func (c *Client) GetTelemetry(ctx context.Context) (*TelemetrySnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/metrics?format=json", nil)
	if err != nil {
		return nil, fmt.Errorf("building telemetry request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing telemetry request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("telemetry endpoint not found (is telemetry enabled?)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from telemetry endpoint: %d", resp.StatusCode)
	}

	var snap TelemetrySnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return nil, fmt.Errorf("decoding telemetry response: %w", err)
	}
	return &snap, nil
}
