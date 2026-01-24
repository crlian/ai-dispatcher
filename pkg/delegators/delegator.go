package delegators

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// DefaultTimeout is the default timeout for task execution
const DefaultTimeout = 5 * time.Minute

// DelegationResult represents the result of executing a task
type DelegationResult struct {
	Success    bool          `json:"success"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	TokensUsed int           `json:"tokens_used"`
	Duration   time.Duration `json:"duration"`
	ToolName   string        `json:"tool_name"`
	ExitCode   int           `json:"exit_code"`
}

// Delegator defines the interface for executing tasks with AI tools
type Delegator interface {
	// Execute runs the task and returns the result
	Execute(ctx context.Context, task string) (*DelegationResult, error)

	// GetToolName returns the name of the tool
	GetToolName() string

	// GetToolType returns the type of the tool
	GetToolType() trackers.ToolType

	// SetTimeout sets the execution timeout
	SetTimeout(timeout time.Duration)
}

// BaseDelegator provides common functionality for all delegators
type BaseDelegator struct {
	toolName string
	toolType trackers.ToolType
	command  string
	timeout  time.Duration
}

// NewBaseDelegator creates a new base delegator
func NewBaseDelegator(toolName string, toolType trackers.ToolType, command string) *BaseDelegator {
	return &BaseDelegator{
		toolName: toolName,
		toolType: toolType,
		command:  command,
		timeout:  DefaultTimeout,
	}
}

// GetToolName returns the tool name
func (bd *BaseDelegator) GetToolName() string {
	return bd.toolName
}

// GetToolType returns the tool type
func (bd *BaseDelegator) GetToolType() trackers.ToolType {
	return bd.toolType
}

// SetTimeout sets the execution timeout
func (bd *BaseDelegator) SetTimeout(timeout time.Duration) {
	bd.timeout = timeout
}

// ExecuteCommand executes a command with timeout and captures output
func (bd *BaseDelegator) ExecuteCommand(ctx context.Context, args []string) (*DelegationResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, bd.timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, bd.command, args...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start timing
	start := time.Now()

	// Execute command
	err := cmd.Run()
	duration := time.Since(start)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Combine output
	output := stdout.String()
	if stderr.Len() > 0 {
		if len(output) > 0 {
			output += "\n"
		}
		output += stderr.String()
	}

	// Estimate tokens used (rough: ~4 chars per token)
	tokensUsed := len(output) / 4

	// Build result
	result := &DelegationResult{
		Success:    err == nil && exitCode == 0,
		Output:     output,
		TokensUsed: tokensUsed,
		Duration:   duration,
		ToolName:   bd.toolName,
		ExitCode:   exitCode,
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("execution timeout after %v", bd.timeout)
		} else {
			result.Error = err.Error()
		}
	}

	return result, nil
}

// EstimateTokens estimates tokens from text (rough: ~4 characters per token)
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return len(text) / 4
}

// FormatDuration formats a duration as a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// TruncateOutput truncates output to a maximum length
func TruncateOutput(output string, maxLength int) string {
	if len(output) <= maxLength {
		return output
	}
	return output[:maxLength] + "... (truncated)"
}
