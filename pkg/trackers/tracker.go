package trackers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ToolType represents the type of AI coding assistant
type ToolType string

const (
	ClaudeCodeTool ToolType = "claude-code"
	CodexTool      ToolType = "codex"
	OpenCodeTool   ToolType = "opencode"
)

// AvailabilityThreshold is the minimum percentage required to consider a tool available
const AvailabilityThreshold = 5.0

// Cost limits per plan (estimated based on real usage patterns)
// These values are based on observed 5-hour window limits
const (
	ProPlanCostLimit   = 2.00  // Pro plan approximate limit
	Max5PlanCostLimit  = 4.00  // Max5 plan approximate limit
	Max20PlanCostLimit = 8.00  // Max20 plan observed limit
	DefaultCostLimit   = 8.00  // Use Max20 as default (most permissive)
)

// UsageTracker defines the interface for tracking AI tool usage and availability
type UsageTracker interface {
	// GetAvailablePercentage returns the percentage of available capacity (0-100)
	GetAvailablePercentage() (float64, error)

	// GetRemainingTime returns the remaining time in minutes before limit resets
	GetRemainingTime() (int, error)

	// GetTotalCost5hWindow returns the total cost spent in the 5-hour window
	GetTotalCost5hWindow() (float64, error)

	// IsAvailable returns true if the tool has more than 5% capacity available
	IsAvailable() (bool, error)

	// GetToolName returns the name of the tool
	GetToolName() string

	// GetToolType returns the type of the tool
	GetToolType() ToolType
}

// UsageData represents the parsed JSON output from ccusage blocks --active
type UsageData struct {
	Blocks []BlockData `json:"blocks"`
}

// BlockData represents a single 5-hour block from ccusage
type BlockData struct {
	IsActive   bool               `json:"isActive"`
	CostUSD    float64            `json:"costUSD"`
	Projection ProjectionData     `json:"projection"`
	LimitStatus *TokenLimitStatus `json:"tokenLimitStatus,omitempty"`
}

// ProjectionData represents projected usage
type ProjectionData struct {
	RemainingMinutes int     `json:"remainingMinutes"`
	TotalCost        float64 `json:"totalCost"`
}

// TokenLimitStatus represents token limit information
type TokenLimitStatus struct {
	Limit      int     `json:"limit"`
	PercentUsed float64 `json:"percentUsed"`
	Status     string  `json:"status"`
}

// BaseTracker provides common functionality for all trackers
type BaseTracker struct {
	toolName  string
	toolType  ToolType
	command   string
	args      []string
	cache     *UsageData
	costLimit float64 // Cost limit for the 5-hour window
}

// NewBaseTracker creates a new base tracker with the given configuration
func NewBaseTracker(toolName string, toolType ToolType, command string, args []string, costLimit float64) *BaseTracker {
	return &BaseTracker{
		toolName:  toolName,
		toolType:  toolType,
		command:   command,
		args:      args,
		costLimit: costLimit,
	}
}

// GetToolName returns the tool name
func (b *BaseTracker) GetToolName() string {
	return b.toolName
}

// GetToolType returns the tool type
func (b *BaseTracker) GetToolType() ToolType {
	return b.toolType
}

// FetchData executes the command and parses the JSON output
func (b *BaseTracker) FetchData() (*UsageData, error) {
	// Return cached data if available
	if b.cache != nil {
		return b.cache, nil
	}

	// Check if command exists
	if _, err := exec.LookPath(b.command); err != nil {
		return nil, fmt.Errorf("tool not installed: %s (run: npm install -g %s)", b.command, b.command)
	}

	// Execute command
	cmd := exec.Command(b.command, b.args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s: %w\nOutput: %s", b.command, err, string(output))
	}

	// Parse JSON output
	var data UsageData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w\nOutput: %s", b.command, err, string(output))
	}

	// Cache the data
	b.cache = &data
	return &data, nil
}

// GetAvailablePercentage returns the available percentage based on cost
func (b *BaseTracker) GetAvailablePercentage() (float64, error) {
	data, err := b.FetchData()
	if err != nil {
		return 0, err
	}

	// Find active block
	var activeBlock *BlockData
	for i := range data.Blocks {
		if data.Blocks[i].IsActive {
			activeBlock = &data.Blocks[i]
			break
		}
	}

	if activeBlock == nil {
		return 0, fmt.Errorf("no active block found")
	}

	// Calculate available percentage based on cost
	// % Available = ((CostLimit - CurrentCost) / CostLimit) * 100
	currentCost := activeBlock.CostUSD
	available := ((b.costLimit - currentCost) / b.costLimit) * 100

	// Ensure it's between 0 and 100
	if available < 0 {
		available = 0
	}
	if available > 100 {
		available = 100
	}

	return available, nil
}

// GetRemainingTime returns the remaining time in minutes
func (b *BaseTracker) GetRemainingTime() (int, error) {
	data, err := b.FetchData()
	if err != nil {
		return 0, err
	}

	// Find active block
	for _, block := range data.Blocks {
		if block.IsActive {
			return block.Projection.RemainingMinutes, nil
		}
	}

	return 0, fmt.Errorf("no active block found")
}

// GetTotalCost5hWindow returns the total cost in the 5-hour window
func (b *BaseTracker) GetTotalCost5hWindow() (float64, error) {
	data, err := b.FetchData()
	if err != nil {
		return 0, err
	}

	// Find active block
	for _, block := range data.Blocks {
		if block.IsActive {
			return block.CostUSD, nil
		}
	}

	return 0, fmt.Errorf("no active block found")
}

// IsAvailable returns true if available percentage is above threshold
func (b *BaseTracker) IsAvailable() (bool, error) {
	available, err := b.GetAvailablePercentage()
	if err != nil {
		return false, err
	}
	return available >= AvailabilityThreshold, nil
}

// ClearCache clears the cached data
func (b *BaseTracker) ClearCache() {
	b.cache = nil
}

// ValidateToolType checks if the given string is a valid tool type
func ValidateToolType(toolType string) (ToolType, error) {
	normalized := strings.ToLower(strings.TrimSpace(toolType))

	switch normalized {
	case string(ClaudeCodeTool):
		return ClaudeCodeTool, nil
	case string(CodexTool):
		return CodexTool, nil
	case string(OpenCodeTool):
		return OpenCodeTool, nil
	default:
		return "", errors.New("invalid tool type: must be one of [claude-code, codex, opencode]")
	}
}

// GetAllToolTypes returns all available tool types
func GetAllToolTypes() []ToolType {
	return []ToolType{ClaudeCodeTool, CodexTool, OpenCodeTool}
}
