package delegators

import (
	"context"
	"fmt"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// ClaudeCodeDelegator executes tasks using Claude Code
type ClaudeCodeDelegator struct {
	*BaseDelegator
}

// NewClaudeCodeDelegator creates a new Claude Code delegator
func NewClaudeCodeDelegator() *ClaudeCodeDelegator {
	return &ClaudeCodeDelegator{
		BaseDelegator: NewBaseDelegator(
			"Claude Code",
			trackers.ClaudeCodeTool,
			"claude",
		),
	}
}

// Execute runs a task using Claude Code
func (ccd *ClaudeCodeDelegator) Execute(ctx context.Context, task string) (*DelegationResult, error) {
	// Build command arguments
	// Using print mode (-p) for non-interactive execution with Haiku model
	args := []string{
		"-p",
		task,
		"--model",
		"haiku",
	}

	// Execute command
	result, err := ccd.ExecuteCommand(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("claude-code execution failed: %w", err)
	}

	return result, nil
}
