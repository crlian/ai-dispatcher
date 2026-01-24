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
		return nil, fmt.Errorf("codex is not currently supported (tool does not exist)")
	case OpenCodeTool:
		return nil, fmt.Errorf("opencode is not currently supported (tool does not exist)")
	default:
		return nil, fmt.Errorf("unknown tool type: %s", toolType)
	}
}

// GetAllTrackers returns trackers for all available tools
// Currently only Claude Code is supported
func GetAllTrackers() []UsageTracker {
	trackers := []UsageTracker{}

	// Add Claude Code tracker (only one currently supported)
	if tracker := NewClaudeCodeTracker(); tracker != nil {
		trackers = append(trackers, tracker)
	}

	// Codex and OpenCode are not currently supported
	// as the required tools don't exist

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
