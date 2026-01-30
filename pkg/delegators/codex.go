package delegators

import (
	"context"
	"fmt"
	"strings"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// CodexDelegator executes tasks using Codex
type CodexDelegator struct {
	*BaseDelegator
}

// NewCodexDelegator creates a new Codex delegator
func NewCodexDelegator() *CodexDelegator {
	bd := NewBaseDelegator(
		"Codex",
		trackers.CodexTool,
		"codex",
	)
	// Use default parser (passes output directly without NDJSON parsing)
	bd.parserType = ParserTypeDefault
	return &CodexDelegator{
		BaseDelegator: bd,
	}
}

// Execute runs a task using Codex
func (cd *CodexDelegator) Execute(ctx context.Context, task string) (*DelegationResult, error) {
	// Build command arguments
	// Using default output format (clean, structured display)
	// Full-auto to avoid approval prompts
	args := []string{
		"exec",
		"--full-auto",
		"--model", "gpt-5.2-codex",
		"-c", "model_reasoning_effort=low", // Use low reasoning to save tokens
		"--sandbox", "read-only",
		"--skip-git-repo-check",
		"--",
		task,
	}

	// Execute command
	result, err := cd.ExecuteCommand(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("codex execution failed: %w", err)
	}

	return result, nil
}

// Query asks Codex for input in council mode (without executing)
func (cd *CodexDelegator) Query(ctx context.Context, prompt string) (string, error) {
	// Add strict prefix to force concise responses and match user's language
	strictPrompt := "EXTREMELY IMPORTANT: Respond in the SAME LANGUAGE as the user. " +
		"Maximum 2 short sentences. NO markdown, NO lists, NO headers. " +
		"Just plain direct text.\n\n" + prompt

	// For council mode, we use exec command WITHOUT --full-auto
	// so Codex can analyze the project context
	// Sandbox allows read access to workspace so it can see the code
	args := []string{
		"exec",
		"--model", "gpt-5.2-codex",
		"-c", "model_reasoning_effort=low", // Use low reasoning to save tokens
		"--sandbox", "workspace-write", // Allow access to workspace for context
		"--",
		strictPrompt,
	}

	// Execute command WITHOUT streaming (clean output for council chat)
	result, err := cd.ExecuteCommandSimple(ctx, args)
	if err != nil {
		return "", fmt.Errorf("codex query failed: %w", err)
	}

	// Parse Codex output to extract only the useful response
	// Codex outputs metadata, then "thinking", then "codex" section with the actual response
	output := cd.parseCodexOutput(result.Output)

	return output, nil
}

// parseCodexOutput extracts the actual response from Codex verbose output
func (cd *CodexDelegator) parseCodexOutput(output string) string {
	// Find the "codex" section which contains the actual response
	lines := strings.Split(output, "\n")
	var result []string
	inResponse := false

	for i, line := range lines {
		// Skip metadata headers at the start
		if !inResponse {
			if strings.HasPrefix(line, "OpenAI Codex") ||
				strings.HasPrefix(line, "workdir:") ||
				strings.HasPrefix(line, "model:") ||
				strings.HasPrefix(line, "provider:") ||
				strings.HasPrefix(line, "approval:") ||
				strings.HasPrefix(line, "sandbox:") ||
				strings.HasPrefix(line, "reasoning") ||
				strings.HasPrefix(line, "session id:") ||
				strings.HasPrefix(line, "mcp startup:") ||
				line == "--------" ||
				strings.HasPrefix(line, "user") {
				continue
			}

			// Look for "thinking" section - skip it
			if strings.HasPrefix(line, "thinking") {
				continue
			}

			// Look for the actual response after "codex" marker
			if line == "codex" && i+1 < len(lines) {
				inResponse = true
				continue
			}
		}

		// Once we're in the response, include everything except trailing metadata
		if inResponse {
			// Stop if we hit tokens used line
			if strings.HasPrefix(line, "tokens used") ||
				isOnlyNumbersAndCommas(strings.TrimSpace(line)) {
				break
			}
			result = append(result, line)
		}
	}

	// Clean up
	for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
		result = result[:len(result)-1]
	}

	if len(result) == 0 {
		// Try to return everything after "codex" line
		if idx := strings.Index(output, "\ncodex\n"); idx != -1 {
			return strings.TrimSpace(output[idx+8:])
		}
		return output
	}

	return strings.Join(result, "\n")
}

// isOnlyNumbersAndCommas checks if a string contains only digits, commas, or dots
func isOnlyNumbersAndCommas(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || r == ',' || r == '.') {
			return false
		}
	}
	return true
}
