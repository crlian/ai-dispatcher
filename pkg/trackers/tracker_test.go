package trackers

import (
	"testing"
)

func TestValidateToolType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ToolType
		wantErr  bool
	}{
		{
			name:     "valid claude-code",
			input:    "claude-code",
			expected: ClaudeCodeTool,
			wantErr:  false,
		},
		{
			name:     "valid codex",
			input:    "codex",
			expected: CodexTool,
			wantErr:  false,
		},
		{
			name:     "valid opencode",
			input:    "opencode",
			expected: OpenCodeTool,
			wantErr:  false,
		},
		{
			name:     "valid with uppercase",
			input:    "CLAUDE-CODE",
			expected: ClaudeCodeTool,
			wantErr:  false,
		},
		{
			name:     "valid with spaces",
			input:    "  codex  ",
			expected: CodexTool,
			wantErr:  false,
		},
		{
			name:     "invalid tool type",
			input:    "invalid-tool",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateToolType(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateToolType() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateToolType() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("ValidateToolType() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestGetAllToolTypes(t *testing.T) {
	toolTypes := GetAllToolTypes()

	if len(toolTypes) != 3 {
		t.Errorf("GetAllToolTypes() returned %d tools, want 3", len(toolTypes))
	}

	expectedTools := map[ToolType]bool{
		ClaudeCodeTool: true,
		CodexTool:      true,
		OpenCodeTool:   true,
	}

	for _, toolType := range toolTypes {
		if !expectedTools[toolType] {
			t.Errorf("GetAllToolTypes() returned unexpected tool: %v", toolType)
		}
	}
}

func TestGetTracker(t *testing.T) {
	tests := []struct {
		name     string
		toolType ToolType
		wantErr  bool
	}{
		{
			name:     "claude-code tracker",
			toolType: ClaudeCodeTool,
			wantErr:  false,
		},
		{
			name:     "codex tracker",
			toolType: CodexTool,
			wantErr:  true, // Codex is not supported
		},
		{
			name:     "opencode tracker",
			toolType: OpenCodeTool,
			wantErr:  true, // OpenCode is not supported
		},
		{
			name:     "invalid tracker",
			toolType: ToolType("invalid"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker, err := GetTracker(tt.toolType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTracker() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetTracker() unexpected error: %v", err)
				}
				if tracker == nil {
					t.Errorf("GetTracker() returned nil tracker")
				}
				if tracker.GetToolType() != tt.toolType {
					t.Errorf("GetTracker() returned wrong tool type: got %v, want %v",
						tracker.GetToolType(), tt.toolType)
				}
			}
		})
	}
}

func TestGetAllTrackers(t *testing.T) {
	trackers := GetAllTrackers()

	// Currently only Claude Code is supported
	if len(trackers) != 1 {
		t.Errorf("GetAllTrackers() returned %d trackers, want 1", len(trackers))
	}

	// Verify it's Claude Code
	if len(trackers) > 0 {
		if trackers[0].GetToolType() != ClaudeCodeTool {
			t.Errorf("GetAllTrackers() returned wrong tool type: got %v, want %v",
				trackers[0].GetToolType(), ClaudeCodeTool)
		}
	}
}
