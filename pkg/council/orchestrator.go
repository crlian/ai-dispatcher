package council

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/crlian/ai-dispatcher/pkg/delegators"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

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
func (o *Orchestrator) buildCouncilPrompt(currentMessage string, isFirstMessage bool) string {
	var prompt strings.Builder

	// System context - keep it minimal and consistent
	if isFirstMessage {
		prompt.WriteString("INSTRUCTIONS:\n")
		prompt.WriteString("1. Language: Respond in the SAME LANGUAGE as the user.\n")
		prompt.WriteString("2. Length: Maximum 2-3 short sentences.\n")
		prompt.WriteString("3. Do not modify files. Only consult.\n")
		prompt.WriteString("4. Participate in the debate as an expert. Do not explain who you are.\n\n")
	}

	// Add conversation history if available - ONLY LAST 2 MESSAGES to save tokens
	history := o.session.GetHistory()
	if len(history) > 0 {
		prompt.WriteString("Contexto:\n")
		// Get last 2 messages only
		startIdx := len(history) - 2
		if startIdx < 0 {
			startIdx = 0
		}
		for i := startIdx; i < len(history); i++ {
			msg := history[i]
			switch msg.From {
			case "user":
				prompt.WriteString(fmt.Sprintf("Usuario: %s\n", msg.Content))
			case "claude", "codex", "opencode":
				// Include specific tool name so other tools know who said what
				prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.From, msg.Content))
			}
		}
		prompt.WriteString("\n")
	}

	// Current message
	prompt.WriteString(fmt.Sprintf("Usuario: %s\n", currentMessage))
	prompt.WriteString("Respuesta:")

	return prompt.String()
}

// restingMessages returns appropriate "unavailable" message for each tool
func (o *Orchestrator) restingMessage(toolName string) string {
	messages := map[string]string{
		"claude":   "☕ Taking a break...",
		"codex":    "🔋 Recharging...",
		"opencode": "📴 Offline...",
	}
	if msg, ok := messages[toolName]; ok {
		return msg
	}
	return "Unavailable"
}

// Broadcast sends a prompt to all available tools and returns their responses
func (o *Orchestrator) Broadcast(message string) []Response {
	responses := make([]Response, 0)
	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()

	// Build rich prompt with history
	prompt := o.buildCouncilPrompt(message, true)

	// Query each tool (skip if not available)
	toolNames := []string{"claude", "codex", "opencode"}
	for _, toolName := range toolNames {
		if o.useMocks {
			// Use mock responses for testing
			response := o.getMockResponse(toolName)
			responses = append(responses, Response{
				Tool:    toolName,
				Content: response,
			})
			o.session.AddMessage(toolName, response)
		} else {
			// Check if tool is available - SKIP if not available
			if available, ok := o.availableTools[toolName]; !ok || !available {
				// Tool is not available, don't add to responses at all
				continue
			}

			// Use real delegator
			if delegator, ok := o.delegators[toolName]; ok {
				content, err := delegator.Query(ctx, prompt)
				if err != nil {
					// Only add errors if it's an available tool that failed
					responses = append(responses, Response{
						Tool:    toolName,
						Content: fmt.Sprintf("[Error: %v]", err),
						Error:   err,
					})
				} else {
					responses = append(responses, Response{
						Tool:    toolName,
						Content: content,
					})
					o.session.AddMessage(toolName, content)
				}
			}
		}
	}

	return responses
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

	// Build rich prompt with history (not first message)
	prompt := o.buildCouncilPrompt(message, false)

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

// findOriginalTask extracts the first user message (the original task)
func (o *Orchestrator) findOriginalTask() string {
	history := o.session.GetHistory()
	for _, msg := range history {
		if msg.From == "user" {
			return msg.Content
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
	return responses[0] // In real mode we'd randomize, but for simplicity return first
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

// DetectTool checks if a message mentions a specific tool
func DetectTool(message string) string {
	message = strings.ToLower(message)

	// Find positions of each tool mention
	claudePos := strings.Index(message, "claude")
	codexPos := strings.Index(message, "codex")
	opencodePos := strings.Index(message, "opencode")

	// Return the one that appears first (lowest index, but not -1)
	firstPos := -1
	firstTool := ""

	if claudePos != -1 {
		firstPos = claudePos
		firstTool = "claude"
	}

	if codexPos != -1 && (firstPos == -1 || codexPos < firstPos) {
		firstPos = codexPos
		firstTool = "codex"
	}

	if opencodePos != -1 && (firstPos == -1 || opencodePos < firstPos) {
		firstPos = opencodePos
		firstTool = "opencode"
	}

	return firstTool
}
