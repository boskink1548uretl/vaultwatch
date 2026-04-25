package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RemediationAction represents a suggested or applied fix for a vault issue.
type RemediationAction struct {
	Path        string `json:"path"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Applied     bool   `json:"applied"`
}

// RemediationResult holds the outcome of a remediation attempt.
type RemediationResult struct {
	Action  RemediationAction `json:"action"`
	Success bool              `json:"success"`
	Message string            `json:"message"`
}

// ApplyRemediation sends a remediation request to the Vault API.
// The action map is posted to the given path under sys/remediations.
func (c *Client) ApplyRemediation(ctx context.Context, action RemediationAction) (*RemediationResult, error) {
	if action.Path == "" {
		return nil, fmt.Errorf("remediation path must not be empty")
	}
	if action.Action == "" {
		return nil, fmt.Errorf("remediation action must not be empty")
	}

	body := map[string]string{
		"path":        action.Path,
		"action":      action.Action,
		"description": action.Description,
	}

	resp, err := c.post(ctx, "/v1/sys/remediations", body)
	if err != nil {
		return nil, fmt.Errorf("remediation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		action.Applied = true
		return &RemediationResult{
			Action:  action,
			Success: true,
			Message: "remediation applied successfully",
		}, nil
	}

	var errBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err == nil {
		if errs, ok := errBody["errors"].([]interface{}); ok && len(errs) > 0 {
			return nil, fmt.Errorf("vault error: %v", errs[0])
		}
	}
	return nil, fmt.Errorf("unexpected status %d from remediation endpoint", resp.StatusCode)
}
