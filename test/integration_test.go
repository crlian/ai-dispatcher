package test

import (
	"testing"

	"github.com/crlian/ai-dispatcher/pkg/analyzers"
	"github.com/crlian/ai-dispatcher/pkg/router"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
	"github.com/crlian/ai-dispatcher/test/mocks"
)

// TestFullRoutingPipeline tests the complete routing pipeline with mocks
func TestFullRoutingPipeline(t *testing.T) {
	// Test case 1: Simple task with Claude Code available
	t.Run("simple task uses claude code", func(t *testing.T) {
		// Create mock trackers (only Claude Code is supported)
		mockTrackers := []trackers.UsageTracker{
			createMockTracker("Claude Code", trackers.ClaudeCodeTool, 50.0, 3.30),
		}

		// Analyze complexity
		analyzer := analyzers.NewComplexityAnalyzer(mockTrackers)
		analysis, err := analyzer.AnalyzeComplexity("fix typo in README")
		if err != nil {
			t.Fatalf("AnalyzeComplexity() error = %v", err)
		}

		// Note: LLM analysis (mocked) returns "medium", but that's okay for this test
		// We're primarily testing that the router selects the free tool
		if analysis.Level == "" {
			t.Error("Analysis level should not be empty")
		}

		// Make routing decision
		engine := router.NewDecisionEngine(mockTrackers)
		decision, err := engine.MakeDecision(analysis, "")
		if err != nil {
			t.Fatalf("MakeDecision() error = %v", err)
		}

		// Should select Claude Code (only available tool)
		if decision.SelectedTool != trackers.ClaudeCodeTool {
			t.Errorf("Expected Claude Code, got %v", decision.SelectedTool)
		}

		// Should have some cost estimation
		if decision.SelectedCost == nil {
			t.Error("Expected cost estimation, got nil")
		}
	})

	// Test case 2: Forced tool selection
	t.Run("forced tool selection", func(t *testing.T) {
		// Create mock trackers with Claude Code
		mockTrackers := []trackers.UsageTracker{
			createMockTracker("Claude Code", trackers.ClaudeCodeTool, 50.0, 3.30),
		}

		analyzer := analyzers.NewComplexityAnalyzer(mockTrackers)
		analysis, err := analyzer.AnalyzeComplexity("refactor entire authentication system")
		if err != nil {
			t.Fatalf("AnalyzeComplexity() error = %v", err)
		}

		// Force Claude Code
		engine := router.NewDecisionEngine(mockTrackers)
		decision, err := engine.MakeDecision(analysis, "claude-code")
		if err != nil {
			t.Fatalf("MakeDecision() error = %v", err)
		}

		if decision.SelectedTool != trackers.ClaudeCodeTool {
			t.Errorf("Expected Claude Code (forced), got %v", decision.SelectedTool)
		}

		if !decision.WasForced {
			t.Error("Decision should be marked as forced")
		}
	})

	// Test case 3: No tools available (low availability)
	t.Run("no tools available", func(t *testing.T) {
		// Create trackers with low availability (<5%)
		unavailableTrackers := []trackers.UsageTracker{
			createMockTracker("Claude Code", trackers.ClaudeCodeTool, 2.0, 6.47), // 2% available
		}

		analyzer := analyzers.NewComplexityAnalyzer(unavailableTrackers)
		analysis, err := analyzer.AnalyzeComplexity("simple task")
		if err != nil {
			t.Fatalf("AnalyzeComplexity() error = %v", err)
		}

		engine := router.NewDecisionEngine(unavailableTrackers)
		_, err = engine.MakeDecision(analysis, "")
		if err == nil {
			t.Error("Expected error when no tools available, got nil")
		}
	})
}

// TestCostCalculation tests the cost calculation logic
func TestCostCalculation(t *testing.T) {
	mockTrackers := []trackers.UsageTracker{
		createMockTracker("Claude Code", trackers.ClaudeCodeTool, 50.0, 3.30),
	}

	// Create complexity analysis
	analysis := &analyzers.ComplexityAnalysis{
		Level:      analyzers.Medium,
		Tokens:     500,
		Confidence: 0.9,
		Method:     "heuristic",
	}

	// Calculate costs
	calculator := router.NewCostCalculator(mockTrackers)
	estimates, err := calculator.CalculateCosts(analysis)
	if err != nil {
		t.Fatalf("CalculateCosts() error = %v", err)
	}

	if len(estimates) != 1 {
		t.Errorf("Expected 1 estimate, got %d", len(estimates))
	}

	// Verify cost calculation for Claude Code
	if len(estimates) > 0 {
		estimate := estimates[0]
		if estimate.Tool != trackers.ClaudeCodeTool {
			t.Errorf("Expected Claude Code estimate, got %v", estimate.Tool)
		}

		expectedCost := 500 * router.ClaudeCodePricePer1k / 1000.0
		if estimate.EstimatedCost != expectedCost {
			t.Errorf("Claude Code cost = %v, want %v", estimate.EstimatedCost, expectedCost)
		}
	}
}

// TestToolStatus tests the tool status reporting
func TestToolStatus(t *testing.T) {
	mockTrackers := []trackers.UsageTracker{
		createMockTracker("Claude Code", trackers.ClaudeCodeTool, 50.0, 3.30),
	}

	engine := router.NewDecisionEngine(mockTrackers)
	statuses, err := engine.GetToolStatus()
	if err != nil {
		t.Fatalf("GetToolStatus() error = %v", err)
	}

	if len(statuses) != 1 {
		t.Errorf("Expected 1 status, got %d", len(statuses))
	}

	// Check Claude Code status
	if len(statuses) > 0 {
		status := statuses[0]
		if status.Tool != trackers.ClaudeCodeTool {
			t.Errorf("Expected Claude Code status, got %v", status.Tool)
		}
		if status.Status != "available" {
			t.Errorf("Claude Code status = %v, want available", status.Status)
		}
	}
}

// Helper function to create a mock tracker
// available: percentage available (0-100)
// The cost is calculated automatically based on the available percentage
func createMockTracker(name string, toolType trackers.ToolType, available float64, _ float64) *mocks.MockTracker {
	mock := mocks.NewMockTracker(name, toolType)
	mock.SetAvailable(available) // This automatically sets the cost
	return mock
}
