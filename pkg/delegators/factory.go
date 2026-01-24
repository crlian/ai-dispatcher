package delegators

import (
	"fmt"
	"strings"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// GetDelegator returns a delegator for the specified tool type
func GetDelegator(toolType trackers.ToolType) (Delegator, error) {
	switch toolType {
	case trackers.ClaudeCodeTool:
		return NewClaudeCodeDelegator(), nil
	case trackers.CodexTool:
		return NewCodexDelegator(), nil
	case trackers.OpenCodeTool:
		return NewOpenCodeDelegator(), nil
	default:
		return nil, fmt.Errorf("unknown tool type: %s", toolType)
	}
}

// GetDelegatorByName returns a delegator for the specified tool name (case-insensitive)
func GetDelegatorByName(name string) (Delegator, error) {
	toolType, err := trackers.ValidateToolType(name)
	if err != nil {
		return nil, err
	}
	return GetDelegator(toolType)
}

// GetAllDelegators returns delegators for all available tools
func GetAllDelegators() []Delegator {
	return []Delegator{
		NewClaudeCodeDelegator(),
		NewCodexDelegator(),
		NewOpenCodeDelegator(),
	}
}

// ValidateDelegatorAvailable checks if a delegator's tool is available
func ValidateDelegatorAvailable(delegator Delegator) error {
	// This would check if the tool command exists
	// For now, just return nil
	return nil
}

// FormatToolName formats a tool name for display
func FormatToolName(toolType trackers.ToolType) string {
	name := string(toolType)
	// Convert kebab-case to Title Case
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
