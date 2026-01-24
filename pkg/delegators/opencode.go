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
