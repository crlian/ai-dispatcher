package delegators

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
	"github.com/mattn/go-isatty"
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

	// Query asks the tool for input without executing (for council mode)
	// The prompt includes full conversation history and context
	Query(ctx context.Context, prompt string) (string, error)

	// GetToolName returns the name of the tool
	GetToolName() string

	// GetToolType returns the type of the tool
	GetToolType() trackers.ToolType

	// SetTimeout sets the execution timeout
	SetTimeout(timeout time.Duration)
}

type Parser interface {
	Parse() (string, error)
	GetAccumulated() string
}

// BaseDelegator provides common functionality for all delegators
type BaseDelegator struct {
	toolName   string
	toolType   trackers.ToolType
	command    string
	timeout    time.Duration
	parserType string
}

const (
	ParserTypeDefault = "default"
	ParserTypeClaude  = "claude"
	ParserTypeCodex   = "codex"
)

// NewBaseDelegator creates a new base delegator
func NewBaseDelegator(toolName string, toolType trackers.ToolType, command string) *BaseDelegator {
	return &BaseDelegator{
		toolName:   toolName,
		toolType:   toolType,
		command:    command,
		timeout:    DefaultTimeout,
		parserType: ParserTypeDefault,
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
// Uses streaming to parse real-time progress from Claude Code output
func (bd *BaseDelegator) ExecuteCommand(ctx context.Context, args []string) (*DelegationResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, bd.timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, bd.command, args...)

	// Get stdout pipe for streaming
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Capture stderr separately
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start timing
	start := time.Now()

	// Start command (non-blocking)
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Parse stream concurrently
	var output string
	var parseErr error
	var wg sync.WaitGroup
	wg.Add(1)

	// Track last activity time and spinner state
	var lastActivityTime atomic.Int64
	lastActivityTime.Store(time.Now().UnixNano())
	var spinnerShown atomic.Bool
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	var spinnerIdx atomic.Int32

	go func() {
		defer wg.Done()

		lineHandler := func(line string) {
			if line == "" {
				return
			}

			if spinnerShown.Load() {
				fmt.Fprint(os.Stderr, "\b\b")
				spinnerShown.Store(false)
			}
			fmt.Fprint(os.Stderr, line)
			fmt.Fprint(os.Stderr, "\n")
			os.Stderr.Sync()
			lastActivityTime.Store(time.Now().UnixNano())
		}

		var renderer *StreamingMarkdownRenderer
		if shouldUseColors() && bd.parserType != ParserTypeCodex {
			renderer = NewStreamingMarkdownRenderer(lineHandler)
			lineHandler = nil
		}

		var parser Parser
		switch bd.parserType {
		case ParserTypeCodex:
			parser = NewCodexStreamParser(stdoutPipe, func(line string) {
				if renderer != nil {
					renderer.Write([]byte(line + "\n"))
				} else if lineHandler != nil {
					lineHandler(line)
				}
			})
		default:
			parser = NewStreamParser(stdoutPipe, func(line string) {
				if renderer != nil {
					renderer.Write([]byte(line + "\n"))
				} else if lineHandler != nil {
					lineHandler(line)
				}
			})
		}

		output, parseErr = parser.Parse()

		if renderer != nil {
			renderer.Flush()
			renderer.Close()
		}
	}()

	// Show spinner when idle for more than 2 seconds
	spinnerTicker := time.NewTicker(300 * time.Millisecond)
	defer spinnerTicker.Stop()

	go func() {
		for range spinnerTicker.C {
			now := time.Now().UnixNano()
			lastActivity := lastActivityTime.Load()
			timeSinceActivity := time.Duration(now - lastActivity)

			// Show spinner if idle for more than 2 seconds
			if timeSinceActivity > 2*time.Second {
				if !spinnerShown.Load() {
					// First time showing spinner - add space and spinner
					spinnerShown.Store(true)
					fmt.Fprintf(os.Stderr, " %s", spinnerFrames[spinnerIdx.Load()%int32(len(spinnerFrames))])
					os.Stderr.Sync()
				} else {
					// Update spinner frame
					fmt.Fprintf(os.Stderr, "\b%s", spinnerFrames[spinnerIdx.Load()%int32(len(spinnerFrames))])
					os.Stderr.Sync()
				}
				spinnerIdx.Add(1)
			}
		}
	}()

	// Wait for streaming to complete
	wg.Wait()

	// Stop spinner ticker and clean up any remaining spinner
	spinnerTicker.Stop()
	if spinnerShown.Load() {
		fmt.Fprint(os.Stderr, "\b\b") // Remove spinner
	}

	// Wait for command to finish
	cmdErr := cmd.Wait()
	duration := time.Since(start)

	// Get exit code
	exitCode := 0
	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Combine output with stderr if present
	if stderr.Len() > 0 {
		if len(output) > 0 {
			output += "\n"
		}
		output += stderr.String()
	}

	// If parsing failed, return error
	if parseErr != nil {
		return nil, fmt.Errorf("stream parsing failed: %w", parseErr)
	}

	// Estimate tokens used (rough: ~4 chars per token)
	tokensUsed := len(output) / 4

	// Build result
	result := &DelegationResult{
		Success:    cmdErr == nil && exitCode == 0,
		Output:     output,
		TokensUsed: tokensUsed,
		Duration:   duration,
		ToolName:   bd.toolName,
		ExitCode:   exitCode,
	}

	if cmdErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("execution timeout after %v", bd.timeout)
		} else {
			result.Error = cmdErr.Error()
		}
	}

	return result, nil
}

// ExecuteCommandSimple executes a command without streaming visual feedback
// Used for council mode where we just want the final output without spinners
func (bd *BaseDelegator) ExecuteCommandSimple(ctx context.Context, args []string) (*DelegationResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, bd.timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, bd.command, args...)

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start timing
	start := time.Now()

	// Execute command (blocking)
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

	// Estimate tokens used
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

// shouldUseColors checks if terminal supports colors and formatting
func shouldUseColors() bool {
	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Respect FORCE_COLOR environment variable
	if forceColor := os.Getenv("FORCE_COLOR"); forceColor != "" {
		return forceColor != "0"
	}

	// Check if TERM is dumb
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Check if stderr is a terminal
	return isatty.IsTerminal(os.Stderr.Fd())
}

// RenderMarkdown renders markdown content for better terminal display
// Uses glamour for beautiful formatting when terminal supports it,
// falls back to plain text rendering for piped/redirected output
func RenderMarkdown(content string) string {
	// Check if we should use colors and ANSI formatting
	if !shouldUseColors() {
		return renderPlainMarkdown(content)
	}

	// Try glamour rendering with proper terminal support
	style := "dark" // default
	if envStyle := os.Getenv("GLAMOUR_STYLE"); envStyle != "" {
		style = envStyle
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(100),
		glamour.WithEmoji(),
	)
	if err != nil {
		// Fallback to plain text on renderer creation error
		return renderPlainMarkdown(content)
	}

	out, err := r.Render(content)
	if err != nil {
		// Fallback to plain text on rendering error
		return renderPlainMarkdown(content)
	}

	return out
}

// renderPlainMarkdown is the fallback plain text markdown renderer
// Used when terminal doesn't support colors or when glamour fails
func renderPlainMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// Handle headings - remove markdown syntax and add visual separation
		if strings.HasPrefix(line, "##") {
			heading := strings.TrimPrefix(line, "## ")
			heading = strings.TrimPrefix(heading, "# ")
			result = append(result, "")
			result = append(result, strings.ToUpper(heading))
			result = append(result, strings.Repeat("─", len(heading)))
			result = append(result, "")
		} else if strings.HasPrefix(line, "#") {
			heading := strings.TrimPrefix(line, "# ")
			result = append(result, "")
			result = append(result, strings.ToUpper(heading))
			result = append(result, strings.Repeat("═", len(heading)))
			result = append(result, "")
		} else {
			// Remove markdown bold/italic markers for cleaner display
			cleaned := line
			cleaned = strings.ReplaceAll(cleaned, "**", "")
			cleaned = strings.ReplaceAll(cleaned, "__", "")
			cleaned = strings.ReplaceAll(cleaned, "*", "")
			cleaned = strings.ReplaceAll(cleaned, "_", "")
			result = append(result, cleaned)
		}
	}

	return strings.Join(result, "\n")
}
