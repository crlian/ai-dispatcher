package delegators

import (
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

const maxBufferSize = 500

type StreamingMarkdownRenderer struct {
	mu          sync.Mutex
	renderer    *glamour.TermRenderer
	buffer      strings.Builder
	onLine      LineHandler
	lastLineLen int
	closed      bool
}

func NewStreamingMarkdownRenderer(onLine LineHandler) *StreamingMarkdownRenderer {
	if onLine == nil {
		onLine = func(line string) {}
	}

	smr := &StreamingMarkdownRenderer{
		onLine: onLine,
	}

	if shouldUseColors() {
		style := "dark"
		if envStyle := os.Getenv("GLAMOUR_STYLE"); envStyle != "" {
			style = envStyle
		}

		r, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle(style),
			glamour.WithWordWrap(100),
			glamour.WithEmoji(),
		)
		if err == nil {
			smr.renderer = r
		}
	}

	return smr
}

func (smr *StreamingMarkdownRenderer) Write(p []byte) (int, error) {
	smr.mu.Lock()
	defer smr.mu.Unlock()

	if smr.closed {
		return len(p), nil
	}

	smr.buffer.Write(p)
	smr.processBuffer()

	return len(p), nil
}

func (smr *StreamingMarkdownRenderer) processBuffer() {
	content := smr.buffer.String()

	for {
		lineEnd := strings.Index(content, "\n")
		if lineEnd == -1 {
			if len(content) >= maxBufferSize || !smr.rendererAvailable() {
				smr.flushBuffer(content)
				smr.buffer.Reset()
				content = ""
			}
			break
		}

		line := content[:lineEnd]
		remaining := content[lineEnd+1:]

		smr.renderLine(line)
		content = remaining
	}

	if content != "" {
		smr.buffer.Reset()
		smr.buffer.WriteString(content)
	}
}

func (smr *StreamingMarkdownRenderer) renderLine(line string) {
	if !smr.rendererAvailable() {
		smr.onLine(line)
		return
	}

	if line == "" {
		smr.onLine("")
		return
	}

	rendered, err := smr.renderer.Render(line + "\n")
	if err != nil {
		smr.onLine(line)
		return
	}

	rendered = strings.TrimSuffix(rendered, "\n")
	smr.onLine(rendered)
}

func (smr *StreamingMarkdownRenderer) flushBuffer(content string) {
	if content == "" {
		return
	}

	if !smr.rendererAvailable() {
		smr.onLine(content)
		return
	}

	rendered, err := smr.renderer.Render(content + "\n")
	if err != nil {
		smr.onLine(content)
		return
	}

	lines := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
	for _, line := range lines {
		smr.onLine(line)
	}
}

func (smr *StreamingMarkdownRenderer) rendererAvailable() bool {
	return smr.renderer != nil
}

func (smr *StreamingMarkdownRenderer) Close() error {
	smr.mu.Lock()
	defer smr.mu.Unlock()

	smr.closed = true

	if smr.buffer.Len() > 0 {
		smr.flushBuffer(smr.buffer.String())
		smr.buffer.Reset()
	}

	if smr.renderer != nil {
		err := smr.renderer.Close()
		smr.renderer = nil
		return err
	}

	return nil
}

func (smr *StreamingMarkdownRenderer) Flush() {
	smr.mu.Lock()
	defer smr.mu.Unlock()

	if smr.buffer.Len() > 0 {
		smr.flushBuffer(smr.buffer.String())
		smr.buffer.Reset()
	}
}
