package council

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// FileAction represents a planned change to a file
type FileAction struct {
	Path    string `json:"path"`    // File path
	Action  string `json:"action"`  // "create", "modify", "delete"
	Summary string `json:"summary"` // Brief description of change
}

// Plan represents the structured execution plan
type Plan struct {
	Tool         string       `json:"-"`            // Tool that generated the plan
	Task         string       `json:"-"`            // Original task
	Summary      string       `json:"summary"`      // Approach description
	Files        []FileAction `json:"files"`        // List of changes
	Dependencies []string     `json:"dependencies"` // Packages to install
	Risks        []string     `json:"risks"`        // Considerations/risks
	Confidence   float64      `json:"confidence"`   // Plan confidence (0-1)
}

// buildPlanPrompt constructs the prompt for generating a plan
func buildPlanPrompt(session *Session, task string) string {
	var prompt strings.Builder

	prompt.WriteString(`Analyze this coding task and provide an execution plan.
Output ONLY valid JSON, no markdown or extra text.

Format:
{
  "summary": "2-3 sentence approach description",
  "files": [
    {"path": "relative/path.ext", "action": "create|modify|delete", "summary": "what changes"}
  ],
  "dependencies": ["packages to install"],
  "risks": ["considerations"],
  "confidence": 0.8
}

`)

	if file := session.GetCurrentFile(); file != "" {
		prompt.WriteString(fmt.Sprintf("File under discussion: %s\n", file))
	}

	// Last 5 messages of context
	history := session.GetHistory()
	start := len(history) - 5
	if start < 0 {
		start = 0
	}
	if len(history[start:]) > 0 {
		prompt.WriteString("\nContext:\n")
		for _, msg := range history[start:] {
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.From, content))
		}
	}

	prompt.WriteString(fmt.Sprintf("\nTask: %s", task))
	return prompt.String()
}

// extractJSON extracts JSON from text that may contain markdown or other content
func extractJSON(text string) string {
	// Try to find JSON block in markdown
	jsonBlockPattern := regexp.MustCompile("(?s)```(?:json)?\\s*({.*?})\\s*```")
	if matches := jsonBlockPattern.FindStringSubmatch(text); len(matches) > 1 {
		return matches[1]
	}

	// Try to find raw JSON object
	jsonPattern := regexp.MustCompile(`(?s)\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	if match := jsonPattern.FindString(text); match != "" {
		return match
	}

	return text
}

// ParsePlanResponse parses the tool response into a Plan struct
func ParsePlanResponse(tool, task, response string) (*Plan, error) {
	jsonStr := extractJSON(response)

	var plan Plan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		// If JSON parsing fails, create a basic plan from the response
		return &Plan{
			Tool:    tool,
			Task:    task,
			Summary: response,
			Files:   []FileAction{},
		}, nil
	}

	plan.Tool = tool
	plan.Task = task

	return &plan, nil
}

// GetMockPlan returns a mock plan for testing
func GetMockPlan(tool, task string, session *Session) *Plan {
	files := []FileAction{}

	// Extract files mentioned in history
	for _, msg := range session.GetHistory() {
		if file := DetectFile(msg.Content); file != "" {
			// Avoid duplicates
			found := false
			for _, f := range files {
				if f.Path == file {
					found = true
					break
				}
			}
			if !found {
				files = append(files, FileAction{
					Path:    file,
					Action:  "modify",
					Summary: "Cambios discutidos en council",
				})
			}
		}
	}

	// Mock files if none found
	if len(files) == 0 {
		files = []FileAction{
			{Path: "pkg/auth/middleware.go", Action: "create", Summary: "Nuevo middleware de auth"},
			{Path: "pkg/routes/auth.go", Action: "modify", Summary: "Agregar rutas de auth"},
			{Path: "config/security.go", Action: "create", Summary: "Configuracion de seguridad"},
		}
	}

	return &Plan{
		Tool:         tool,
		Task:         task,
		Summary:      fmt.Sprintf("[Mock] %s propone implementar con patrones estandar de Go", tool),
		Files:        files,
		Dependencies: []string{"github.com/golang-jwt/jwt/v5"},
		Risks:        []string{"Asegurar compatibilidad con sesiones existentes"},
		Confidence:   0.8,
	}
}
