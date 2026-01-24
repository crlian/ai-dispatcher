package trackers

// CodexTracker is currently not supported
// The @ccusage/codex package does not exist as a real tool
// This is kept for potential future implementation
type CodexTracker struct {
	*BaseTracker
}

// NewCodexTracker creates a new tracker for Codex (NOT IMPLEMENTED)
// Returns nil as Codex tracking is not currently available
func NewCodexTracker() *CodexTracker {
	// Codex tracking is not implemented
	// The original plan assumed tools that don't exist
	return nil
}
