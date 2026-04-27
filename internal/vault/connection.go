package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ConnectionInfo holds details about the current Vault connection state.
type ConnectionInfo struct {
	Address     string        `json:"address"`
	Reachable   bool          `json:"reachable"`
	Latency     time.Duration `json:"latency_ms"`
	TLSEnabled  bool          `json:"tls_enabled"`
	StatusCode  int           `json:"status_code"`
}

// CheckConnection probes the Vault address and returns connection diagnostics.
func (c *Client) CheckConnection(ctx context.Context) (*ConnectionInfo, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/health", nil)
	if err != nil {
		return nil, fmt.Errorf("building connection check request: %w", err)
	}

	resp, err := c.http.Do(req)
	latency := time.Since(start)
	if err != nil {
		return &ConnectionInfo{
			Address:    c.address,
			Reachable:  false,
			Latency:    latency,
			TLSEnabled: isTLS(c.address),
		}, nil
	}
	defer resp.Body.Close()

	return &ConnectionInfo{
		Address:    c.address,
		Reachable:  true,
		Latency:    latency,
		TLSEnabled: isTLS(c.address),
		StatusCode: resp.StatusCode,
	}, nil
}

// isTLS returns true if the address uses the https scheme.
func isTLS(address string) bool {
	return len(address) >= 8 && address[:8] == "https://"
}
