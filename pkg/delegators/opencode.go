package delegators

import (
	"context"
	"fmt"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// OpenCodeDelegator executes tasks using OpenCode
type OpenCodeDelegator struct {
	*BaseDelegator
}

// NewOpenCodeDelegator creates a new OpenCode delegator
func NewOpenCodeDelegator() *OpenCodeDelegator {
	return &OpenCodeDelegator{
		BaseDelegator: NewBaseDelegator(
			"OpenCode",
			trackers.OpenCodeTool,
			"opencode",
		),
	}
}

// Execute runs a task using OpenCode
func (ocd *OpenCodeDelegator) Execute(ctx context.Context, task string) (*DelegationResult, error) {
	// Build command arguments
	args := []string{
		"run",
		task,
	}

	// Execute command
	result, err := ocd.ExecuteCommand(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("opencode execution failed: %w", err)
	}

	return result, nil
}

// Query asks OpenCode for input in council mode (without executing)
func (ocd *OpenCodeDelegator) Query(ctx context.Context, prompt string) (string, error) {
	// OpenCode might have a chat or ask mode
	// For now, we'll use run with a modified prompt that asks for plan only
	args := []string{
		"run",
		prompt + "\n\nIMPORTANT: Only describe your approach. Do NOT modify any files.",
	}

	// Execute command WITHOUT streaming (clean output for council chat)
	result, err := ocd.ExecuteCommandSimple(ctx, args)
	if err != nil {
		return "", fmt.Errorf("opencode query failed: %w", err)
	}

	return result.Output, nil
}
