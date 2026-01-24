package analyzers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// ComplexityLevel represents the complexity classification of a task
type ComplexityLevel string

const (
	Simple  ComplexityLevel = "simple"
	Medium  ComplexityLevel = "medium"
	Complex ComplexityLevel = "complex"
)

// ComplexityAnalysis contains the result of analyzing a task's complexity
type ComplexityAnalysis struct {
	Level      ComplexityLevel `json:"level"`      // Classification: simple, medium, or complex
	Tokens     int             `json:"tokens"`     // Estimated tokens needed
	Reasoning  string          `json:"reasoning"`  // Explanation of the classification
	Confidence float64         `json:"confidence"` // Confidence score (0.0-1.0)
	Method     string          `json:"method"`     // "llm" or "heuristic"
}

// ComplexityAnalyzer analyzes task complexity using LLM with heuristic fallback
type ComplexityAnalyzer struct {
	trackers []trackers.UsageTracker
	timeout  time.Duration
}

// NewComplexityAnalyzer creates a new complexity analyzer
func NewComplexityAnalyzer(trackers []trackers.UsageTracker) *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		trackers: trackers,
		timeout:  10 * time.Second,
	}
}

// AnalyzeComplexity analyzes the complexity of a task
// It first attempts to use an LLM, then falls back to heuristic analysis
func (ca *ComplexityAnalyzer) AnalyzeComplexity(task string) (*ComplexityAnalysis, error) {
	// Try LLM analysis first
	analysis, err := ca.llmAnalysis(task)
	if err == nil {
		return analysis, nil
	}

	// Fallback to heuristic analysis
	return ca.heuristicAnalysis(task), nil
}

// llmAnalysis uses the cheapest available LLM to analyze task complexity
func (ca *ComplexityAnalyzer) llmAnalysis(task string) (*ComplexityAnalysis, error) {
	// Find the tool with the most available capacity
	var bestTracker trackers.UsageTracker
	var maxAvailable float64 = 0

	for _, tracker := range ca.trackers {
		available, err := tracker.GetAvailablePercentage()
		if err != nil {
			continue
		}
		if available > maxAvailable {
			maxAvailable = available
			bestTracker = tracker
		}
	}

	if bestTracker == nil || maxAvailable < 5.0 {
		return nil, fmt.Errorf("no LLM available for complexity analysis")
	}

	// Construct prompt for LLM
	// In a real implementation, this would use the prompt to call the actual LLM tool
	// For now, we use a simple mock response
	_ = task // Acknowledge task parameter

	// Execute the tool with a mock response (demonstration only)
	toolName := bestTracker.GetToolName()
	cmd := exec.Command("echo", fmt.Sprintf(`{"level": "medium", "tokens": 500, "reasoning": "Using %s for analysis"}`, toolName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("LLM execution failed: %w", err)
	}

	// Parse LLM response
	var result struct {
		Level     string `json:"level"`
		Tokens    int    `json:"tokens"`
		Reasoning string `json:"reasoning"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &ComplexityAnalysis{
		Level:      ComplexityLevel(result.Level),
		Tokens:     result.Tokens,
		Reasoning:  result.Reasoning,
		Confidence: 0.9, // High confidence for LLM analysis
		Method:     "llm",
	}, nil
}

// heuristicAnalysis performs rule-based complexity analysis
func (ca *ComplexityAnalyzer) heuristicAnalysis(task string) *ComplexityAnalysis {
	taskLower := strings.ToLower(task)
	wordCount := len(strings.Fields(task))

	// Complex keywords
	complexKeywords := []string{
		"refactor", "architecture", "migrate", "redesign",
		"implement", "create new", "build", "design",
		"multiple", "entire", "all", "system",
	}

	// Simple keywords
	simpleKeywords := []string{
		"fix typo", "add comment", "rename", "delete",
		"update text", "change color", "format",
	}

	// Check for complex keywords
	complexCount := 0
	for _, keyword := range complexKeywords {
		if strings.Contains(taskLower, keyword) {
			complexCount++
		}
	}

	// Check for simple keywords
	simpleCount := 0
	for _, keyword := range simpleKeywords {
		if strings.Contains(taskLower, keyword) {
			simpleCount++
		}
	}

	// Determine complexity level
	var level ComplexityLevel
	var tokens int
	var reasoning string

	if complexCount > 0 || wordCount > 20 {
		level = Complex
		tokens = 1500
		reasoning = fmt.Sprintf("Task appears complex (word count: %d, complex keywords: %d)", wordCount, complexCount)
	} else if simpleCount > 0 || wordCount < 5 {
		level = Simple
		tokens = 150
		reasoning = fmt.Sprintf("Task appears simple (word count: %d, simple keywords: %d)", wordCount, simpleCount)
	} else {
		level = Medium
		tokens = 500
		reasoning = fmt.Sprintf("Task appears medium complexity (word count: %d)", wordCount)
	}

	return &ComplexityAnalysis{
		Level:      level,
		Tokens:     tokens,
		Reasoning:  reasoning,
		Confidence: 0.6, // Lower confidence for heuristic
		Method:     "heuristic",
	}
}

// EstimateTokens estimates the number of tokens for a given task description
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return len(text) / 4
}

// GetComplexityDescription returns a human-readable description of the complexity level
func GetComplexityDescription(level ComplexityLevel) string {
	switch level {
	case Simple:
		return "Simple task (quick fix, small change)"
	case Medium:
		return "Medium task (moderate feature or bug fix)"
	case Complex:
		return "Complex task (architecture, refactoring, multiple components)"
	default:
		return "Unknown complexity"
	}
}

// IsComplex returns true if the complexity level is complex
func (ca *ComplexityAnalysis) IsComplex() bool {
	return ca.Level == Complex
}

// IsMediumOrHigher returns true if the complexity is medium or higher
func (ca *ComplexityAnalysis) IsMediumOrHigher() bool {
	return ca.Level == Medium || ca.Level == Complex
}
