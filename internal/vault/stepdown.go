package vault

import (
	"context"
	"fmt"
	"net/http"
)

// StepDownResult holds the outcome of a leader step-down request.
type StepDownResult struct {
	Success bool
	Message string
}

// StepDownLeader requests the active Vault node to step down as leader,
// triggering a new leader election among standby nodes.
func (c *Client) StepDownLeader(ctx context.Context) (*StepDownResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		c.address+"/v1/sys/step-down", nil)
	if err != nil {
		return nil, fmt.Errorf("stepdown: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stepdown: request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return &StepDownResult{Success: true, Message: "leader step-down accepted"}, nil
	case http.StatusForbidden:
		return &StepDownResult{Success: false, Message: "permission denied"}, nil
	case http.StatusBadRequest:
		return &StepDownResult{Success: false, Message: "node is not the active leader"}, nil
	default:
		return &StepDownResult{Success: false,
			Message: fmt.Sprintf("unexpected status: %d", resp.StatusCode)}, nil
	}
}

// IsLeader returns true when the connected Vault node is the active leader.
func (c *Client) IsLeader(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/leader", nil)
	if err != nil {
		return false, fmt.Errorf("isleader: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("isleader: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("isleader: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		IsSelf bool `json:"is_self"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return false, fmt.Errorf("isleader: decode: %w", err)
	}
	return body.IsSelf, nil
}
