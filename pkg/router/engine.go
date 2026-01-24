package router

import (
	"fmt"
	"strings"

	"github.com/crlian/ai-dispatcher/pkg/analyzers"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// RoutingDecision represents the decision made by the routing engine
type RoutingDecision struct {
	SelectedTool trackers.ToolType             `json:"selected_tool"`
	SelectedName string                        `json:"selected_name"`
	Reason       string                        `json:"reason"`
	Alternatives []*CostEstimate               `json:"alternatives"`
	SelectedCost *CostEstimate                 `json:"selected_cost"`
	Complexity   *analyzers.ComplexityAnalysis `json:"complexity"`
	WasForced    bool                          `json:"was_forced"`
}

// DecisionEngine makes routing decisions for task execution
type DecisionEngine struct {
	calculator *CostCalculator
	trackers   []trackers.UsageTracker
}

// NewDecisionEngine creates a new decision engine
func NewDecisionEngine(trackers []trackers.UsageTracker) *DecisionEngine {
	return &DecisionEngine{
		calculator: NewCostCalculator(trackers),
		trackers:   trackers,
	}
}

// MakeDecision determines the best tool to use for a task
// If forceTool is specified, it will attempt to use that tool
func (de *DecisionEngine) MakeDecision(
	analysis *analyzers.ComplexityAnalysis,
	forceTool string,
) (*RoutingDecision, error) {
	// Calculate costs for all tools
	estimates, err := de.calculator.CalculateCosts(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate costs: %w", err)
	}

	// Handle forced tool selection
	if forceTool != "" {
		return de.handleForcedTool(forceTool, estimates, analysis)
	}

	// Filter to available tools
	available := de.calculator.FilterAvailable(estimates)
	if len(available) == 0 {
		return nil, fmt.Errorf("no tools available - all tools have exceeded their limits or are unavailable")
	}

	// Sort by priority and select best
	sorted := de.calculator.SortEstimates(available)
	selected := sorted[0]

	// Build reason
	reason := de.buildReason(selected, analysis, sorted)

	return &RoutingDecision{
		SelectedTool: selected.Tool,
		SelectedName: selected.ToolName,
		Reason:       reason,
		Alternatives: sorted[1:], // All other options
		SelectedCost: selected,
		Complexity:   analysis,
		WasForced:    false,
	}, nil
}

// handleForcedTool handles the case where a specific tool is forced
func (de *DecisionEngine) handleForcedTool(
	forceTool string,
	estimates []*CostEstimate,
	analysis *analyzers.ComplexityAnalysis,
) (*RoutingDecision, error) {
	// Validate and normalize tool name
	toolType, err := trackers.ValidateToolType(forceTool)
	if err != nil {
		return nil, fmt.Errorf("invalid forced tool: %w", err)
	}

	// Find the forced tool in estimates
	var selected *CostEstimate
	var alternatives []*CostEstimate

	for _, est := range estimates {
		if est.Tool == toolType {
			selected = est
		} else {
			alternatives = append(alternatives, est)
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("forced tool %s not found in available tools", forceTool)
	}

	// Build reason for forced selection
	reason := fmt.Sprintf("Using %s (forced by --force flag)", selected.ToolName)
	if !selected.IsAvailable {
		reason += fmt.Sprintf(" - WARNING: Tool has low availability (%.1f%%)", selected.AvailablePercent)
	}
	if selected.WillExceedLimit {
		reason += " - WARNING: This may exceed usage limits"
	}

	return &RoutingDecision{
		SelectedTool: selected.Tool,
		SelectedName: selected.ToolName,
		Reason:       reason,
		Alternatives: alternatives,
		SelectedCost: selected,
		Complexity:   analysis,
		WasForced:    true,
	}, nil
}

// buildReason constructs a human-readable explanation for the routing decision
func (de *DecisionEngine) buildReason(
	selected *CostEstimate,
	analysis *analyzers.ComplexityAnalysis,
	allEstimates []*CostEstimate,
) string {
	var parts []string

	// Start with the selection
	if selected.EstimatedCost == 0 {
		parts = append(parts, fmt.Sprintf(
			"Selected %s (free tier) with %.1f%% capacity available",
			selected.ToolName,
			selected.AvailablePercent,
		))
	} else {
		parts = append(parts, fmt.Sprintf(
			"Selected %s (est. cost: %s) with %.1f%% capacity available",
			selected.ToolName,
			FormatCost(selected.EstimatedCost),
			selected.AvailablePercent,
		))
	}

	// Add complexity context
	parts = append(parts, fmt.Sprintf(
		"Task complexity: %s (~%d tokens, confidence: %.1f%%, method: %s)",
		analysis.Level,
		analysis.Tokens,
		analysis.Confidence*100,
		analysis.Method,
	))

	// Add reasoning if available
	if analysis.Reasoning != "" {
		parts = append(parts, fmt.Sprintf("Reason: %s", analysis.Reasoning))
	}

	// Mention alternatives if available
	if len(allEstimates) > 1 {
		altNames := make([]string, 0)
		for i := 1; i < len(allEstimates) && i < 3; i++ {
			alt := allEstimates[i]
			if alt.EstimatedCost == 0 {
				altNames = append(altNames, fmt.Sprintf("%s (free)", alt.ToolName))
			} else {
				altNames = append(altNames, fmt.Sprintf("%s (%s)", alt.ToolName, FormatCost(alt.EstimatedCost)))
			}
		}
		if len(altNames) > 0 {
			parts = append(parts, fmt.Sprintf("Alternatives: %s", strings.Join(altNames, ", ")))
		}
	}

	return strings.Join(parts, ". ")
}

// GetAvailableTools returns a list of currently available tools
func (de *DecisionEngine) GetAvailableTools() ([]trackers.ToolType, error) {
	available := make([]trackers.ToolType, 0)

	for _, tracker := range de.trackers {
		isAvailable, err := tracker.IsAvailable()
		if err != nil {
			continue
		}
		if isAvailable {
			available = append(available, tracker.GetToolType())
		}
	}

	return available, nil
}

// GetToolStatus returns status information for all tools
func (de *DecisionEngine) GetToolStatus() ([]*ToolStatus, error) {
	statuses := make([]*ToolStatus, 0, len(de.trackers))

	for _, tracker := range de.trackers {
		status, err := de.getToolStatus(tracker)
		if err != nil {
			// Add error status
			statuses = append(statuses, &ToolStatus{
				Tool:      tracker.GetToolType(),
				ToolName:  tracker.GetToolName(),
				Available: 0,
				Status:    "error",
				Error:     err.Error(),
			})
			continue
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// ToolStatus represents the current status of a tool
type ToolStatus struct {
	Tool          trackers.ToolType `json:"tool"`
	ToolName      string            `json:"tool_name"`
	Available     float64           `json:"available_percent"`
	RemainingTime int               `json:"remaining_time_minutes"`
	CurrentCost   float64           `json:"current_cost_5h"`
	IsAvailable   bool              `json:"is_available"`
	Status        string            `json:"status"` // "available", "low", "limited", "error"
	Error         string            `json:"error,omitempty"`
}

// getToolStatus retrieves status for a single tool
func (de *DecisionEngine) getToolStatus(tracker trackers.UsageTracker) (*ToolStatus, error) {
	available, err := tracker.GetAvailablePercentage()
	if err != nil {
		return nil, err
	}

	remainingTime, err := tracker.GetRemainingTime()
	if err != nil {
		return nil, err
	}

	currentCost, err := tracker.GetTotalCost5hWindow()
	if err != nil {
		return nil, err
	}

	isAvailable, err := tracker.IsAvailable()
	if err != nil {
		return nil, err
	}

	// Determine status
	var status string
	if !isAvailable {
		status = "limited"
	} else if available < 20.0 {
		status = "low"
	} else {
		status = "available"
	}

	return &ToolStatus{
		Tool:          tracker.GetToolType(),
		ToolName:      tracker.GetToolName(),
		Available:     available,
		RemainingTime: remainingTime,
		CurrentCost:   currentCost,
		IsAvailable:   isAvailable,
		Status:        status,
	}, nil
}

// FormatDecision formats a routing decision as a human-readable string
func FormatDecision(decision *RoutingDecision) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("📍 Routing Decision\n"))
	builder.WriteString(fmt.Sprintf("   Tool: %s\n", decision.SelectedName))
	if decision.SelectedCost != nil {
		builder.WriteString(fmt.Sprintf("   Cost: %s\n", FormatCost(decision.SelectedCost.EstimatedCost)))
		builder.WriteString(fmt.Sprintf("   Tokens: ~%d\n", decision.SelectedCost.EstimatedTokens))
		builder.WriteString(fmt.Sprintf("   Available: %.1f%%\n", decision.SelectedCost.AvailablePercent))
	}
	builder.WriteString(fmt.Sprintf("   Reason: %s\n", decision.Reason))

	if decision.WasForced {
		builder.WriteString("   ⚠️  Tool selection was forced\n")
	}

	return builder.String()
}
