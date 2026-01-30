package council

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crlian/ai-dispatcher/pkg/delegators"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

// maxHistoryContext is the number of recent messages to include in prompts
const maxHistoryContext = 3

// toolPatterns contains regex patterns with word boundaries for tool detection
var toolPatterns = map[string]*regexp.Regexp{
	"claude":   regexp.MustCompile(`(?i)\bclaude\b`),
	"codex":    regexp.MustCompile(`(?i)\bcodex\b`),
	"opencode": regexp.MustCompile(`(?i)\bopencode\b`),
}

// filePattern detects file paths in messages (e.g., orchestrator.go, src/main.js)
var filePattern = regexp.MustCompile(`[\w\-./]+\.(go|js|ts|tsx|jsx|py|rs|java|cpp|c|h|rb|php|swift|kt|scala|sh|yaml|yml|json|toml|md|sql)(?:\b|$)`)

// Response represents a tool's response
type Response struct {
	Tool    string
	Content string
	Error   error
}

// Orchestrator manages the council session and real tool connections
type Orchestrator struct {
	session        *Session
	delegators     map[string]delegators.Delegator
	availableTools map[string]bool // Track which tools are available
	timeout        time.Duration
	useMocks       bool // For testing without real tools
}

// NewOrchestrator creates a new council orchestrator
func NewOrchestrator() *Orchestrator {
	// Initialize all delegators with standardized keys
	allDelegators := delegators.GetAllDelegators()
	delegatorMap := make(map[string]delegators.Delegator)

	for _, d := range allDelegators {
		// Map to standard keys used throughout the codebase
		toolType := d.GetToolType()
		var key string
		switch toolType {
		case trackers.ClaudeCodeTool:
			key = "claude"
		case trackers.CodexTool:
			key = "codex"
		case trackers.OpenCodeTool:
			key = "opencode"
		default:
			key = strings.ToLower(d.GetToolName())
			key = strings.ReplaceAll(key, " ", "")
			key = strings.ReplaceAll(key, "-", "")
		}
		delegatorMap[key] = d
	}

	return &Orchestrator{
		session:        NewSession(),
		delegators:     delegatorMap,
		availableTools: make(map[string]bool), // Will be set by SetAvailableTools
		timeout:        2 * time.Minute,       // Shorter timeout for chat mode
		useMocks:       false,                 // Set to true for testing without real tools
	}
}

// NewMockOrchestrator creates an orchestrator with mock responses (for testing)
func NewMockOrchestrator() *Orchestrator {
	o := NewOrchestrator()
	o.useMocks = true
	return o
}

// GetSession returns the current session
func (o *Orchestrator) GetSession() *Session {
	return o.session
}

// SetUseMocks enables or disables mock mode
func (o *Orchestrator) SetUseMocks(useMocks bool) {
	o.useMocks = useMocks
}

// SetAvailableTools sets which tools are available for the council
func (o *Orchestrator) SetAvailableTools(available map[string]bool) {
	o.availableTools = available
}

// buildCouncilPrompt constructs a rich prompt with conversation history
func (o *Orchestrator) buildCouncilPrompt(currentMessage string) string {
	var prompt strings.Builder

	// System context - council mode with rules
	prompt.WriteString("[You are in a council with other AI tools (Claude, Codex, OpenCode) discussing a coding task. ")
	prompt.WriteString("Give your unique perspective. You may agree or disagree with others. ")
	prompt.WriteString("Rules: Match user's language. Max 2 sentences. No file changes. Be direct.]\n\n")

	// Include current file under discussion if set
	if currentFile := o.session.GetCurrentFile(); currentFile != "" {
		prompt.WriteString(fmt.Sprintf("[File under discussion: %s - read it if you need context]\n\n", currentFile))
	}

	// Add conversation history - include last N messages for better context
	history := o.session.GetHistory()

	// Calculate starting index for relevant history
	startIdx := len(history) - maxHistoryContext
	if startIdx < 0 {
		startIdx = 0
	}

	relevantHistory := history[startIdx:]
	// Filter out the last message if it's from the user (we add it as currentMessage)
	if len(relevantHistory) > 0 && relevantHistory[len(relevantHistory)-1].From == "user" {
		relevantHistory = relevantHistory[:len(relevantHistory)-1]
	}

	if len(relevantHistory) > 0 {
		prompt.WriteString("Context:\n")
		for _, msg := range relevantHistory {
			content := msg.Content
			// Truncate long messages
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.From, content))
		}
		prompt.WriteString("\n")
	}

	// Current message
	prompt.WriteString(fmt.Sprintf("User: %s\n", currentMessage))
	prompt.WriteString("Response:")

	return prompt.String()
}

// Broadcast sends a prompt to all available tools and returns a channel of responses.
// Responses are sent as they arrive. The channel is closed when all tools have responded.
func (o *Orchestrator) Broadcast(message string) <-chan Response {
	responseChan := make(chan Response)

	go func() {
		defer close(responseChan)

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
		defer cancel()

		// Build rich prompt with history
		prompt := o.buildCouncilPrompt(message)

		// Query each tool (skip if not available)
		toolNames := []string{"claude", "codex", "opencode"}
		for _, toolName := range toolNames {
			if o.useMocks {
				// Mock mode stays sequential (it's fast anyway)
				response := o.getMockResponse(toolName)
				responseChan <- Response{
					Tool:    toolName,
					Content: response,
				}
				o.session.AddMessage(toolName, response)
				continue
			}

			// Check if tool is available - SKIP if not available
			if available, ok := o.availableTools[toolName]; !ok || !available {
				continue
			}

			// Use real delegator with goroutines for parallel execution
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				if delegator, ok := o.delegators[name]; ok {
					content, err := delegator.Query(ctx, prompt)
					if err != nil {
						responseChan <- Response{
							Tool:    name,
							Content: fmt.Sprintf("[Error: %v]", err),
							Error:   err,
						}
					} else {
						responseChan <- Response{
							Tool:    name,
							Content: content,
						}
						o.session.AddMessage(name, content)
					}
				}
			}(toolName)
		}

		wg.Wait()
	}()

	return responseChan
}

// Query sends a prompt to a specific tool
func (o *Orchestrator) Query(tool, message string) Response {
	o.session.SetLastTool(tool)

	// Check if tool is available (return error if not, unless in mock mode)
	if !o.useMocks {
		if available, ok := o.availableTools[tool]; !ok || !available {
			return Response{
				Tool:    tool,
				Content: "",
				Error:   fmt.Errorf("tool '%s' is not available", tool),
			}
		}
	}

	if o.useMocks {
		// Use mock response for testing
		response := o.getMockResponse(tool)
		o.session.AddMessage(tool, response)
		return Response{
			Tool:    tool,
			Content: response,
		}
	}

	// Use real delegator
	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()

	// Build rich prompt with history
	prompt := o.buildCouncilPrompt(message)

	if delegator, ok := o.delegators[tool]; ok {
		content, err := delegator.Query(ctx, prompt)
		if err != nil {
			return Response{
				Tool:    tool,
				Content: "",
				Error:   err,
			}
		}
		o.session.AddMessage(tool, content)
		return Response{
			Tool:    tool,
			Content: content,
		}
	}

	return Response{
		Tool:    tool,
		Content: "",
		Error:   fmt.Errorf("tool '%s' not found", tool),
	}
}

// Execute runs a task with a tool (real execution)
func (o *Orchestrator) Execute(tool string) string {
	// If no tool specified, use last mentioned
	if tool == "" {
		tool = o.session.GetLastTool()
	}

	if tool == "" {
		return "No hay herramienta seleccionada. Menciona una herramienta primero."
	}

	o.session.SetLastTool(tool)

	if o.useMocks {
		// Mock execution
		return fmt.Sprintf("[Mock] Ejecutando con %s...\n%s", tool, o.generateMockFiles())
	}

	// Find the original task from history
	originalTask := o.findOriginalTask()
	if originalTask == "" {
		return "No se encontró la tarea original para ejecutar."
	}

	// Use real delegator
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Longer timeout for execution
	defer cancel()

	if delegator, ok := o.delegators[tool]; ok {
		delegator.SetTimeout(5 * time.Minute)
		result, err := delegator.Execute(ctx, originalTask)
		if err != nil {
			return fmt.Sprintf("Error ejecutando con %s: %v", tool, err)
		}

		if result.Success {
			return fmt.Sprintf("✓ %s completó la tarea exitosamente\n\nOutput:\n%s",
				tool, result.Output)
		} else {
			return fmt.Sprintf("✗ %s falló en la ejecución\nError: %s\n\nOutput:\n%s",
				tool, result.Error, result.Output)
		}
	}

	return fmt.Sprintf("Herramienta '%s' no encontrada", tool)
}

// findOriginalTask extracts the last user message (the most recent/refined task)
func (o *Orchestrator) findOriginalTask() string {
	history := o.session.GetHistory()
	// Search backwards to find the LAST user message (most recent task)
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].From == "user" {
			return history[i].Content
		}
	}
	return ""
}

// getMockResponse returns a random mock response for a tool
func (o *Orchestrator) getMockResponse(tool string) string {
	mockResponses := map[string][]string{
		"claude": {
			"Propongo usar JWT middleware. Es stateless y escalable.",
			"Sugiero implementar OAuth2 para mejor seguridad a largo plazo.",
			"Podemos usar sessions con Redis para persistencia.",
			"API keys serían más simples para este caso de uso.",
			"Considera usar middleware de auth existente y adaptarlo.",
			"Implementaré auth con roles y permisos desde el inicio.",
		},
		"codex": {
			"JWT es bueno pero considera refresh tokens para seguridad.",
			"OAuth2 permite invalidar tokens, mejor para producción.",
			"Sessions tradicionales son más simples de debuggear.",
			"API Gateway con auth centralizado podría ser mejor arquitectura.",
			"Considera usar un servicio de auth externo como Auth0.",
			"Implementaré auth básico primero, luego iteramos.",
		},
		"opencode": {
			"Versión simple con JWT y sin dependencias externas.",
			"Implementación rápida con middleware básico.",
			"Auth minimalista, solo lo necesario.",
			"Podemos usar variables de entorno para secrets.",
			"No over-engineering, auth simple y funcional.",
		},
	}

	responses, ok := mockResponses[tool]
	if !ok || len(responses) == 0 {
		return "No tengo una opinión sobre eso."
	}
	return responses[rand.Intn(len(responses))]
}

// generateMockFiles generates random file modification messages
func (o *Orchestrator) generateMockFiles() string {
	files := []string{
		"auth/middleware.go",
		"auth/jwt.go",
		"routes/auth.go",
		"config/security.go",
		"middleware/auth.go",
		"handlers/login.go",
		"models/user.go",
		"tests/auth_test.go",
	}

	var result strings.Builder
	result.WriteString("Archivos modificados:\n")
	for _, file := range files[:3] { // Just show 3 files for mock
		result.WriteString(fmt.Sprintf("  ✓ %s (modificado)\n", file))
	}

	return result.String()
}

// DetectTool checks if a message mentions a specific tool using word boundaries
func DetectTool(message string) string {
	firstPos := -1
	firstTool := ""

	for tool, pattern := range toolPatterns {
		loc := pattern.FindStringIndex(message)
		if loc != nil && (firstPos == -1 || loc[0] < firstPos) {
			firstPos = loc[0]
			firstTool = tool
		}
	}

	return firstTool
}

// DetectFile extracts the first file path mentioned in a message
func DetectFile(message string) string {
	match := filePattern.FindString(message)
	return match
}

// Plan generates an execution plan using the specified tool
func (o *Orchestrator) Plan(tool string) (*Plan, error) {
	// Resolve tool (use LastTool if not specified)
	if tool == "" {
		tool = o.session.GetLastTool()
	}
	if tool == "" {
		return nil, fmt.Errorf("no hay herramienta seleccionada")
	}

	// Find the original task from history
	task := o.findOriginalTask()
	if task == "" {
		return nil, fmt.Errorf("no se encontro la tarea original")
	}

	// Mock mode
	if o.useMocks {
		return GetMockPlan(tool, task, o.session), nil
	}

	// Verify tool availability
	if available, ok := o.availableTools[tool]; !ok || !available {
		return nil, fmt.Errorf("tool '%s' no esta disponible", tool)
	}

	// Build prompt and query
	prompt := buildPlanPrompt(o.session, task)

	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()

	delegator, ok := o.delegators[tool]
	if !ok {
		return nil, fmt.Errorf("tool '%s' no encontrada", tool)
	}

	response, err := delegator.Query(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response
	return ParsePlanResponse(tool, task, response)
}
