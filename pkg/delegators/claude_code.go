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
	// Stream JSON for real-time progress display (requires --verbose)
	args := []string{
		"-p",
		task,
		"--model",
		"haiku",
		"--output-format",
		"stream-json",
		"--include-partial-messages",
		"--verbose",
	}

	// Execute command
	result, err := ccd.ExecuteCommand(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("claude-code execution failed: %w", err)
	}

	return result, nil
}

// Query asks Claude for input in council mode (without executing)
func (ccd *ClaudeCodeDelegator) Query(ctx context.Context, prompt string) (string, error) {
	// Add strict prefix for language and behavior
	strictPrompt := "⚠️  CRITICAL: Respond in the SAME LANGUAGE as the user. " +
		"Maximum 2-3 short sentences. Do not explain who you are. Just answer directly.\n\n" + prompt

	// For council mode, we use the prompt directly which already includes
	// the system instructions and conversation history
	args := []string{
		"-p",
		strictPrompt,
		"--model",
		"haiku",
		// Use plain output for faster responses in chat mode
	}

	// Execute command WITHOUT streaming (clean output for council chat)
	result, err := ccd.ExecuteCommandSimple(ctx, args)
	if err != nil {
		return "", fmt.Errorf("claude-code query failed: %w", err)
	}

	return result.Output, nil
}
