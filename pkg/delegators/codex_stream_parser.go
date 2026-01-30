package delegators

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type CodexLineHandler func(line string)

type CodexStreamParser struct {
	reader     io.Reader
	onLine     CodexLineHandler
	lineBuffer strings.Builder
}

func NewCodexStreamParser(reader io.Reader, onLine CodexLineHandler) *CodexStreamParser {
	if onLine == nil {
		onLine = func(line string) {}
	}
	return &CodexStreamParser{
		reader: reader,
		onLine: onLine,
	}
}

func (sp *CodexStreamParser) Parse() (string, error) {
	scanner := bufio.NewScanner(sp.reader)
	var fullOutput strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}

		switch eventType {
		case "delta":
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if content, ok := delta["content"].(string); ok && content != "" {
					sp.processText(content, &fullOutput)
				}
			}

		case "message":
			if contentRaw, ok := event["content"].(string); ok && contentRaw != "" {
				if sp.lineBuffer.Len() == 0 {
					sp.processText(contentRaw, &fullOutput)
				}
			} else if contentArray, ok := event["content"].([]interface{}); ok {
				for _, item := range contentArray {
					if itemStr, ok := item.(string); ok && itemStr != "" {
						if sp.lineBuffer.Len() == 0 {
							sp.processText(itemStr, &fullOutput)
						}
					}
				}
			}

		case "item.completed":
			if item, ok := event["item"].(map[string]interface{}); ok {
				sp.extractItemContent(item, &fullOutput)
			}

		case "item.started":
			if item, ok := event["item"].(map[string]interface{}); ok {
				if cmd, ok := item["command"].(string); ok && cmd != "" {
					sp.processText("$ "+cmd, &fullOutput)
				}
			}

		case "result":
			if result, ok := event["result"].(string); ok && result != "" {
				if sp.lineBuffer.Len() == 0 {
					sp.processText(result, &fullOutput)
				}
			}

		case "error":
			if errorMsg, ok := event["error"].(string); ok && errorMsg != "" {
				sp.processText("[Error] "+errorMsg, &fullOutput)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("codex stream parsing error: %w", err)
	}

	if sp.lineBuffer.Len() > 0 {
		line := sp.lineBuffer.String()
		sp.onLine(line)
		fullOutput.WriteString(line)
		sp.lineBuffer.Reset()
	}

	return fullOutput.String(), nil
}

func (sp *CodexStreamParser) processText(text string, output *strings.Builder) {
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

func (sp *CodexStreamParser) extractItemContent(item map[string]interface{}, output *strings.Builder) {
	text, hasText := item["text"].(string)
	if hasText && text != "" {
		// Solo mostrar reasoning sustancial (>200 chars), filtrar planificación intermedia
		if len(text) > 200 {
			sp.processText(text+"\n", output)
		}
	}

	if itemType, ok := item["type"].(string); ok {
		if itemType == "command_execution" {
			if aggregatedOutput, ok := item["aggregated_output"].(string); ok && aggregatedOutput != "" {
				sp.processText(aggregatedOutput, output)
			}
		}
	}
}

func (sp *CodexStreamParser) GetAccumulated() string {
	if sp.lineBuffer.Len() > 0 {
		return sp.lineBuffer.String()
	}
	return ""
}
