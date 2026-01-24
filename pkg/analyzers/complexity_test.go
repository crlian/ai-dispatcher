package analyzers

import (
	"strings"
	"testing"
)

func TestHeuristicAnalysis(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}

	tests := []struct {
		name          string
		task          string
		expectedLevel ComplexityLevel
		expectedMin   int // Minimum expected tokens
		expectedMax   int // Maximum expected tokens
	}{
		{
			name:          "simple task - fix typo",
			task:          "fix typo in README",
			expectedLevel: Simple,
			expectedMin:   100,
			expectedMax:   200,
		},
		{
			name:          "simple task - add comment",
			task:          "add comment to function",
			expectedLevel: Simple,
			expectedMin:   100,
			expectedMax:   200,
		},
		{
			name:          "medium task - standard fix",
			task:          "fix authentication bug in login handler",
			expectedLevel: Medium,
			expectedMin:   400,
			expectedMax:   600,
		},
		{
			name:          "complex task - refactor",
			task:          "refactor the entire authentication system to use JWT tokens",
			expectedLevel: Complex,
			expectedMin:   1000,
			expectedMax:   2000,
		},
		{
			name:          "complex task - architecture",
			task:          "implement new microservice architecture for user management",
			expectedLevel: Complex,
			expectedMin:   1000,
			expectedMax:   2000,
		},
		{
			name:          "very short task",
			task:          "fix bug",
			expectedLevel: Simple,
			expectedMin:   100,
			expectedMax:   200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.heuristicAnalysis(tt.task)

			if result.Level != tt.expectedLevel {
				t.Errorf("heuristicAnalysis() level = %v, want %v", result.Level, tt.expectedLevel)
			}

			if result.Tokens < tt.expectedMin || result.Tokens > tt.expectedMax {
				t.Errorf("heuristicAnalysis() tokens = %v, want between %v and %v",
					result.Tokens, tt.expectedMin, tt.expectedMax)
			}

			if result.Method != "heuristic" {
				t.Errorf("heuristicAnalysis() method = %v, want heuristic", result.Method)
			}

			if result.Confidence != 0.6 {
				t.Errorf("heuristicAnalysis() confidence = %v, want 0.6", result.Confidence)
			}

			if result.Reasoning == "" {
				t.Error("heuristicAnalysis() reasoning should not be empty")
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty text",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "hello",
			expected: 1, // 5 chars / 4 = 1.25 -> 1
		},
		{
			name:     "medium text",
			text:     "This is a test string with some words",
			expected: 9, // 38 chars / 4 = 9.5 -> 9
		},
		{
			name:     "long text",
			text:     strings.Repeat("test ", 100),
			expected: 125, // 500 chars / 4 = 125
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			if result != tt.expected {
				t.Errorf("EstimateTokens() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetComplexityDescription(t *testing.T) {
	tests := []struct {
		name     string
		level    ComplexityLevel
		contains string
	}{
		{
			name:     "simple description",
			level:    Simple,
			contains: "Simple",
		},
		{
			name:     "medium description",
			level:    Medium,
			contains: "Medium",
		},
		{
			name:     "complex description",
			level:    Complex,
			contains: "Complex",
		},
		{
			name:     "unknown description",
			level:    ComplexityLevel("unknown"),
			contains: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetComplexityDescription(tt.level)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("GetComplexityDescription() = %v, should contain %v", result, tt.contains)
			}
		})
	}
}

func TestComplexityAnalysisMethods(t *testing.T) {
	tests := []struct {
		name              string
		analysis          *ComplexityAnalysis
		expectIsComplex   bool
		expectIsMediumOrHigher bool
	}{
		{
			name:              "simple analysis",
			analysis:          &ComplexityAnalysis{Level: Simple},
			expectIsComplex:   false,
			expectIsMediumOrHigher: false,
		},
		{
			name:              "medium analysis",
			analysis:          &ComplexityAnalysis{Level: Medium},
			expectIsComplex:   false,
			expectIsMediumOrHigher: true,
		},
		{
			name:              "complex analysis",
			analysis:          &ComplexityAnalysis{Level: Complex},
			expectIsComplex:   true,
			expectIsMediumOrHigher: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.analysis.IsComplex(); result != tt.expectIsComplex {
				t.Errorf("IsComplex() = %v, want %v", result, tt.expectIsComplex)
			}

			if result := tt.analysis.IsMediumOrHigher(); result != tt.expectIsMediumOrHigher {
				t.Errorf("IsMediumOrHigher() = %v, want %v", result, tt.expectIsMediumOrHigher)
			}
		})
	}
}
