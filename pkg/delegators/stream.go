package delegators

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type LineHandler func(line string)

type StreamParser struct {
	reader     io.Reader
	onLine     LineHandler
	lineBuffer strings.Builder
}

func NewStreamParser(reader io.Reader, onLine LineHandler) *StreamParser {
	if onLine == nil {
		onLine = func(line string) {}
	}
	return &StreamParser{
		reader: reader,
		onLine: onLine,
	}
}

func (sp *StreamParser) Parse() (string, error) {
	scanner := bufio.NewScanner(sp.reader)
	var fullOutput strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal(line, &event); err != nil {
			sp.processText(string(line), &fullOutput)
			continue
		}

		eventType, ok := event["type"].(string)
		if !ok {
			sp.processText(string(line), &fullOutput)
			continue
		}

		if eventType == "stream_event" {
			if streamEvent, ok := event["event"].(map[string]interface{}); ok {
				if eventSubType, ok := streamEvent["type"].(string); ok && eventSubType == "content_block_delta" {
					if delta, ok := streamEvent["delta"].(map[string]interface{}); ok {
						if deltaType, ok := delta["type"].(string); ok && deltaType == "text_delta" {
							if text, ok := delta["text"].(string); ok && text != "" {
								sp.processText(text, &fullOutput)
							}
						}
					}
				}
			}
		}

		if eventType == "assistant" {
			if message, ok := event["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].([]interface{}); ok {
					for _, item := range content {
						if contentItem, ok := item.(map[string]interface{}); ok {
							if text, ok := contentItem["text"].(string); ok && text != "" {
								if sp.lineBuffer.Len() == 0 {
									sp.processText(text, &fullOutput)
								}
							}
						}
					}
				}
			}
		}

		if eventType == "result" {
			if result, ok := event["result"].(string); ok && result != "" {
				if sp.lineBuffer.Len() == 0 {
					sp.processText(result, &fullOutput)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("stream parsing error: %w", err)
	}

	if sp.lineBuffer.Len() > 0 {
		line := sp.lineBuffer.String()
		sp.onLine(line)
		fullOutput.WriteString(line)
		sp.lineBuffer.Reset()
	}

	return fullOutput.String(), nil
}

func (sp *StreamParser) processText(text string, output *strings.Builder) {
	sp.lineBuffer.WriteString(text)

	for {
		content := sp.lineBuffer.String()
		newlineIdx := strings.Index(content, "\n")
		if newlineIdx == -1 {
			break
		}

		line := content[:newlineIdx]
		remaining := content[newlineIdx+1:]

		sp.onLine(line)
		output.WriteString(line)
		output.WriteString("\n")

		sp.lineBuffer.Reset()
		sp.lineBuffer.WriteString(remaining)
	}
}

func (sp *StreamParser) GetAccumulated() string {
	if sp.lineBuffer.Len() > 0 {
		return sp.lineBuffer.String()
	}
	return ""
}
