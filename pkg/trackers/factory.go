package trackers

import "fmt"

// GetTracker returns a tracker for the specified tool type
func GetTracker(toolType ToolType) (UsageTracker, error) {
	switch toolType {
	case ClaudeCodeTool:
		tracker := NewClaudeCodeTracker()
		if tracker == nil {
			return nil, fmt.Errorf("claude-code tracker not available")
		}
		return tracker, nil
	case CodexTool:
		tracker := NewCodexTracker()
		if tracker == nil {
			return nil, fmt.Errorf("codex tracker not available")
		}
		return tracker, nil
	case OpenCodeTool:
		return nil, fmt.Errorf("opencode is not currently supported (tool does not exist)")
	default:
		return nil, fmt.Errorf("unknown tool type: %s", toolType)
	}
}

// GetAllTrackers returns trackers for all available tools
// Currently Claude Code and Codex are supported
func GetAllTrackers() []UsageTracker {
	trackers := []UsageTracker{}

	// Add Claude Code tracker
	if tracker := NewClaudeCodeTracker(); tracker != nil {
		trackers = append(trackers, tracker)
	}

	// Add Codex tracker
	if tracker := NewCodexTracker(); tracker != nil {
		trackers = append(trackers, tracker)
	}

	// OpenCode is not currently supported

	return trackers
}

// GetTrackerByName returns a tracker for the specified tool name (case-insensitive)
func GetTrackerByName(name string) (UsageTracker, error) {
	toolType, err := ValidateToolType(name)
	if err != nil {
		return nil, err
	}
	return GetTracker(toolType)
}
