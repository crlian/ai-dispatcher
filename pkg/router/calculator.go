package router

import (
	"fmt"
	"sort"

	"github.com/crlian/ai-dispatcher/pkg/analyzers"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// Pricing per 1k tokens for each tool (in USD)
// Note: Actual pricing varies by model (Sonnet, Haiku, Opus)
// These are average estimates for calculation purposes
const (
	ClaudeCodePricePer1k = 0.030 // Average ~$0.03 per 1k tokens
	CodexPricePer1k      = 0.000 // Not supported
	OpenCodePricePer1k   = 0.000 // Not supported
)

// CostEstimate represents the estimated cost for using a specific tool
type CostEstimate struct {
	Tool             trackers.ToolType `json:"tool"`
	ToolName         string            `json:"tool_name"`
	EstimatedCost    float64           `json:"estimated_cost"`
	EstimatedTokens  int               `json:"estimated_tokens"`
	AvailablePercent float64           `json:"available_percent"`
	CurrentCost5h    float64           `json:"current_cost_5h"`
	WillExceedLimit  bool              `json:"will_exceed_limit"`
	IsAvailable      bool              `json:"is_available"`
	Confidence       float64           `json:"confidence"`
}

// CostCalculator calculates costs for different AI tools
type CostCalculator struct {
	trackers []trackers.UsageTracker
}

// NewCostCalculator creates a new cost calculator
func NewCostCalculator(trackers []trackers.UsageTracker) *CostCalculator {
	return &CostCalculator{
		trackers: trackers,
	}
}

// CalculateCosts calculates cost estimates for all available tools
func (cc *CostCalculator) CalculateCosts(analysis *analyzers.ComplexityAnalysis) ([]*CostEstimate, error) {
	estimates := make([]*CostEstimate, 0, len(cc.trackers))

	for _, tracker := range cc.trackers {
		estimate, err := cc.calculateForTool(tracker, analysis)
		if err != nil {
			// Log error but continue with other tools
			continue
		}
		estimates = append(estimates, estimate)
	}

	if len(estimates) == 0 {
		return nil, fmt.Errorf("no tools available for cost calculation")
	}

	return estimates, nil
}

// calculateForTool calculates cost estimate for a specific tool
func (cc *CostCalculator) calculateForTool(tracker trackers.UsageTracker, analysis *analyzers.ComplexityAnalysis) (*CostEstimate, error) {
	// Get tool pricing
	pricePerToken := cc.getPricing(tracker.GetToolType()) / 1000.0

	// Calculate estimated cost
	estimatedCost := float64(analysis.Tokens) * pricePerToken

	// Get current usage
	available, err := tracker.GetAvailablePercentage()
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

	// Check if adding this task would exceed limits
	// Assuming a limit based on available percentage and current cost
	willExceedLimit := !isAvailable || available < 10.0

	return &CostEstimate{
		Tool:             tracker.GetToolType(),
		ToolName:         tracker.GetToolName(),
		EstimatedCost:    estimatedCost,
		EstimatedTokens:  analysis.Tokens,
		AvailablePercent: available,
		CurrentCost5h:    currentCost,
		WillExceedLimit:  willExceedLimit,
		IsAvailable:      isAvailable,
		Confidence:       analysis.Confidence,
	}, nil
}

// getPricing returns the price per 1k tokens for a tool type
func (cc *CostCalculator) getPricing(toolType trackers.ToolType) float64 {
	switch toolType {
	case trackers.ClaudeCodeTool:
		return ClaudeCodePricePer1k
	case trackers.CodexTool:
		return CodexPricePer1k
	case trackers.OpenCodeTool:
		return OpenCodePricePer1k
	default:
		return 0.0
	}
}

// SortEstimates sorts cost estimates by priority
// Priority: available > free > cheaper > expensive
func (cc *CostCalculator) SortEstimates(estimates []*CostEstimate) []*CostEstimate {
	sorted := make([]*CostEstimate, len(estimates))
	copy(sorted, estimates)

	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]

		// 1. Prioritize available tools
		if a.IsAvailable != b.IsAvailable {
			return a.IsAvailable
		}

		// 2. Prioritize tools that won't exceed limit
		if a.WillExceedLimit != b.WillExceedLimit {
			return !a.WillExceedLimit
		}

		// 3. Prioritize free tools
		if a.EstimatedCost == 0 && b.EstimatedCost != 0 {
			return true
		}
		if a.EstimatedCost != 0 && b.EstimatedCost == 0 {
			return false
		}

		// 4. Sort by cost (cheaper first)
		if a.EstimatedCost != b.EstimatedCost {
			return a.EstimatedCost < b.EstimatedCost
		}

		// 5. Sort by available percentage (more available first)
		return a.AvailablePercent > b.AvailablePercent
	})

	return sorted
}

// GetBestEstimate returns the best cost estimate based on priority
func (cc *CostCalculator) GetBestEstimate(estimates []*CostEstimate) *CostEstimate {
	if len(estimates) == 0 {
		return nil
	}

	sorted := cc.SortEstimates(estimates)
	return sorted[0]
}

// FilterAvailable filters estimates to only include available tools
func (cc *CostCalculator) FilterAvailable(estimates []*CostEstimate) []*CostEstimate {
	filtered := make([]*CostEstimate, 0)
	for _, estimate := range estimates {
		if estimate.IsAvailable && !estimate.WillExceedLimit {
			filtered = append(filtered, estimate)
		}
	}
	return filtered
}

// FormatCost formats a cost value as a string
func FormatCost(cost float64) string {
	if cost == 0 {
		return "$0.000"
	}
	return fmt.Sprintf("$%.3f", cost)
}

// GetToolPricing returns pricing information for all tools
func GetToolPricing() map[trackers.ToolType]float64 {
	return map[trackers.ToolType]float64{
		trackers.ClaudeCodeTool: ClaudeCodePricePer1k,
		trackers.CodexTool:      CodexPricePer1k,
		trackers.OpenCodeTool:   OpenCodePricePer1k,
	}
}
