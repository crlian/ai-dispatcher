package council

import (
	"time"
)

// Message represents a single message in the council session
type Message struct {
	From    string // "user", "claude", "codex", "opencode"
	Content string
	Time    time.Time
}

// Session holds the state of a council session
type Session struct {
	History   []Message
	LastTool  string // Last mentioned tool
	StartedAt time.Time
}

// NewSession creates a new council session
func NewSession() *Session {
	return &Session{
		History:   make([]Message, 0),
		StartedAt: time.Now(),
	}
}

// AddMessage adds a message to the session history
func (s *Session) AddMessage(from, content string) {
	s.History = append(s.History, Message{
		From:    from,
		Content: content,
		Time:    time.Now(),
	})
}

// SetLastTool sets the last mentioned tool
func (s *Session) SetLastTool(tool string) {
	s.LastTool = tool
}

// GetLastTool returns the last mentioned tool
func (s *Session) GetLastTool() string {
	return s.LastTool
}

// GetHistory returns all messages in the session
func (s *Session) GetHistory() []Message {
	return s.History
}
