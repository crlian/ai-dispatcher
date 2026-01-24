package mocks

import (
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// MockTracker is a mock implementation of UsageTracker for testing
type MockTracker struct {
	toolName      string
	toolType      trackers.ToolType
	available     float64
	remainingTime int
	totalCost     float64
	isAvailable   bool
	shouldError   bool
	errorMessage  string
	costLimit     float64 // Cost limit for percentage calculation
}

// NewMockTracker creates a new mock tracker with default values
func NewMockTracker(toolName string, toolType trackers.ToolType) *MockTracker {
	return &MockTracker{
		toolName:      toolName,
		toolType:      toolType,
		available:     50.0,
		remainingTime: 120,
		totalCost:     3.30, // Half of default cost limit
		isAvailable:   true,
		shouldError:   false,
		costLimit:     trackers.DefaultCostLimit,
	}
}

// GetAvailablePercentage returns the mocked available percentage
func (m *MockTracker) GetAvailablePercentage() (float64, error) {
	if m.shouldError {
		return 0, m.makeError()
	}
	return m.available, nil
}

// GetRemainingTime returns the mocked remaining time
func (m *MockTracker) GetRemainingTime() (int, error) {
	if m.shouldError {
		return 0, m.makeError()
	}
	return m.remainingTime, nil
}

// GetTotalCost5hWindow returns the mocked total cost
func (m *MockTracker) GetTotalCost5hWindow() (float64, error) {
	if m.shouldError {
		return 0, m.makeError()
	}
	return m.totalCost, nil
}

// IsAvailable returns the mocked availability status
func (m *MockTracker) IsAvailable() (bool, error) {
	if m.shouldError {
		return false, m.makeError()
	}
	return m.isAvailable, nil
}

// GetToolName returns the tool name
func (m *MockTracker) GetToolName() string {
	return m.toolName
}

// GetToolType returns the tool type
func (m *MockTracker) GetToolType() trackers.ToolType {
	return m.toolType
}

// SetAvailable sets the available percentage and adjusts cost accordingly
func (m *MockTracker) SetAvailable(available float64) {
	m.available = available
	m.isAvailable = available >= trackers.AvailabilityThreshold

	// Calculate total cost based on available percentage
	// available% = ((costLimit - totalCost) / costLimit) * 100
	// Solving for totalCost: totalCost = costLimit * (1 - available/100)
	m.totalCost = m.costLimit * (1 - available/100.0)
}

// SetRemainingTime sets the remaining time
func (m *MockTracker) SetRemainingTime(remainingTime int) {
	m.remainingTime = remainingTime
}

// SetTotalCost sets the total cost
func (m *MockTracker) SetTotalCost(totalCost float64) {
	m.totalCost = totalCost
}

// SetError configures the mock to return an error
func (m *MockTracker) SetError(shouldError bool, message string) {
	m.shouldError = shouldError
	m.errorMessage = message
}

// makeError creates an error with the configured message
func (m *MockTracker) makeError() error {
	if m.errorMessage != "" {
		return &MockError{message: m.errorMessage}
	}
	return &MockError{message: "mock error"}
}

// MockError is a simple error type for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
