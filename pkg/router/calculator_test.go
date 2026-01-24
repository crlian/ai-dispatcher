package router

import (
	"testing"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

func TestGetPricing(t *testing.T) {
	calculator := &CostCalculator{}

	tests := []struct {
		name     string
		toolType trackers.ToolType
		expected float64
	}{
		{
			name:     "claude-code pricing",
			toolType: trackers.ClaudeCodeTool,
			expected: ClaudeCodePricePer1k,
		},
		{
			name:     "codex pricing",
			toolType: trackers.CodexTool,
			expected: CodexPricePer1k,
		},
		{
			name:     "opencode pricing",
			toolType: trackers.OpenCodeTool,
			expected: OpenCodePricePer1k,
		},
		{
			name:     "unknown tool pricing",
			toolType: trackers.ToolType("unknown"),
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.getPricing(tt.toolType)
			if result != tt.expected {
				t.Errorf("getPricing() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortEstimates(t *testing.T) {
	calculator := &CostCalculator{}

	estimates := []*CostEstimate{
		{
			Tool:             trackers.ClaudeCodeTool,
			ToolName:         "Claude Code",
			EstimatedCost:    0.015,
			AvailablePercent: 50.0,
			IsAvailable:      true,
			WillExceedLimit:  false,
		},
		{
			Tool:             trackers.OpenCodeTool,
			ToolName:         "OpenCode",
			EstimatedCost:    0.0,
			AvailablePercent: 80.0,
			IsAvailable:      true,
			WillExceedLimit:  false,
		},
		{
			Tool:             trackers.CodexTool,
			ToolName:         "Codex",
			EstimatedCost:    0.010,
			AvailablePercent: 30.0,
			IsAvailable:      true,
			WillExceedLimit:  false,
		},
	}

	sorted := calculator.SortEstimates(estimates)

	// OpenCode (free) should be first
	if sorted[0].Tool != trackers.OpenCodeTool {
		t.Errorf("First sorted estimate should be OpenCode, got %v", sorted[0].Tool)
	}

	// Codex (cheaper) should be second
	if sorted[1].Tool != trackers.CodexTool {
		t.Errorf("Second sorted estimate should be Codex, got %v", sorted[1].Tool)
	}

	// Claude Code (expensive) should be third
	if sorted[2].Tool != trackers.ClaudeCodeTool {
		t.Errorf("Third sorted estimate should be Claude Code, got %v", sorted[2].Tool)
	}
}

func TestFilterAvailable(t *testing.T) {
	calculator := &CostCalculator{}

	estimates := []*CostEstimate{
		{
			Tool:            trackers.ClaudeCodeTool,
			IsAvailable:     true,
			WillExceedLimit: false,
		},
		{
			Tool:            trackers.CodexTool,
			IsAvailable:     false,
			WillExceedLimit: false,
		},
		{
			Tool:            trackers.OpenCodeTool,
			IsAvailable:     true,
			WillExceedLimit: true,
		},
	}

	filtered := calculator.FilterAvailable(estimates)

	if len(filtered) != 1 {
		t.Errorf("FilterAvailable() returned %d estimates, want 1", len(filtered))
	}

	if filtered[0].Tool != trackers.ClaudeCodeTool {
		t.Errorf("FilterAvailable() returned wrong tool: got %v, want %v",
			filtered[0].Tool, trackers.ClaudeCodeTool)
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		name     string
		cost     float64
		expected string
	}{
		{
			name:     "zero cost",
			cost:     0.0,
			expected: "$0.000",
		},
		{
			name:     "small cost",
			cost:     0.015,
			expected: "$0.015",
		},
		{
			name:     "large cost",
			cost:     1.234,
			expected: "$1.234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCost(tt.cost)
			if result != tt.expected {
				t.Errorf("FormatCost() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetBestEstimate(t *testing.T) {
	calculator := &CostCalculator{}

	estimates := []*CostEstimate{
		{
			Tool:             trackers.ClaudeCodeTool,
			EstimatedCost:    0.015,
			IsAvailable:      true,
			WillExceedLimit:  false,
			AvailablePercent: 50.0,
		},
		{
			Tool:             trackers.OpenCodeTool,
			EstimatedCost:    0.0,
			IsAvailable:      true,
			WillExceedLimit:  false,
			AvailablePercent: 80.0,
		},
	}

	best := calculator.GetBestEstimate(estimates)

	if best.Tool != trackers.OpenCodeTool {
		t.Errorf("GetBestEstimate() = %v, want %v", best.Tool, trackers.OpenCodeTool)
	}

	// Test with empty list
	emptyBest := calculator.GetBestEstimate([]*CostEstimate{})
	if emptyBest != nil {
		t.Errorf("GetBestEstimate() with empty list should return nil, got %v", emptyBest)
	}
}

func TestGetToolPricing(t *testing.T) {
	pricing := GetToolPricing()

	if len(pricing) != 3 {
		t.Errorf("GetToolPricing() returned %d prices, want 3", len(pricing))
	}

	expectedPricing := map[trackers.ToolType]float64{
		trackers.ClaudeCodeTool: ClaudeCodePricePer1k,
		trackers.CodexTool:      CodexPricePer1k,
		trackers.OpenCodeTool:   OpenCodePricePer1k,
	}

	for toolType, expectedPrice := range expectedPricing {
		if price, ok := pricing[toolType]; !ok {
			t.Errorf("GetToolPricing() missing price for %v", toolType)
		} else if price != expectedPrice {
			t.Errorf("GetToolPricing()[%v] = %v, want %v", toolType, price, expectedPrice)
		}
	}
}
