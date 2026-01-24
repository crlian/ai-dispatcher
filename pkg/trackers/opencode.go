package trackers

// OpenCodeTracker is currently not supported
// The @ccusage/opencode package does not exist as a real tool
// This is kept for potential future implementation
type OpenCodeTracker struct {
	*BaseTracker
}

// NewOpenCodeTracker creates a new tracker for OpenCode (NOT IMPLEMENTED)
// Returns nil as OpenCode tracking is not currently available
func NewOpenCodeTracker() *OpenCodeTracker {
	// OpenCode tracking is not implemented
	// The original plan assumed tools that don't exist
	return nil
}
