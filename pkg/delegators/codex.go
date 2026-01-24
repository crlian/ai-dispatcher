package delegators

import (
	"context"
	"fmt"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// CodexDelegator executes tasks using Codex
type CodexDelegator struct {
	*BaseDelegator
}

// NewCodexDelegator creates a new Codex delegator
func NewCodexDelegator() *CodexDelegator {
	return &CodexDelegator{
		BaseDelegator: NewBaseDelegator(
			"Codex",
			trackers.CodexTool,
			"codex",
		),
	}
}

// Execute runs a task using Codex
func (cd *CodexDelegator) Execute(ctx context.Context, task string) (*DelegationResult, error) {
	// Build command arguments
	args := []string{
		"exec",
		task,
	}

	// Execute command
	result, err := cd.ExecuteCommand(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("codex execution failed: %w", err)
	}

	return result, nil
}
