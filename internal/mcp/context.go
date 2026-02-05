// Package mcp provides MCP server functionality for interactive CV optimization.
package mcp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// InteractiveContext contains all data needed by the MCP server subprocess.
type InteractiveContext struct {
	// ApplicationDir is the path to the application folder (e.g., "applications/acme-corp")
	ApplicationDir string `json:"application_dir"`
	// BaseCV is the contents of the user's base CV markdown
	BaseCV string `json:"base_cv"`
	// JobDescription is the contents of the job description
	JobDescription string `json:"job_description"`
	// ATSMode indicates whether to optimize for ATS systems
	ATSMode bool `json:"ats_mode"`
	// Model is the Claude model to use (may be empty for default)
	Model string `json:"model,omitempty"`
}

// Encode serializes the context to a base64-encoded JSON string.
func (c *InteractiveContext) Encode() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DecodeContext deserializes a base64-encoded JSON string into an InteractiveContext.
func DecodeContext(encoded string) (*InteractiveContext, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	var ctx InteractiveContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &ctx, nil
}
