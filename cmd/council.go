package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/crlian/ai-dispatcher/pkg/council"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	councilMock bool
)

// councilCmd represents the council command
var councilCmd = &cobra.Command{
	Use:   "council",
	Short: "Interactive council mode with multiple AI tools",
	Long: `Enter an interactive council session where multiple AI tools can discuss
and propose solutions to your coding tasks.

In council mode:
  • All tools respond to the first question
  • Mention a tool by name to direct questions to it
  • Type "ejecuta" to run with the last mentioned tool
  • Type "exit" or "quit" to leave

By default, runs in MOCK mode with fake responses for testing.
Use --real flag to connect to actual AI tools (requires them to be installed).

Examples:
  ai-dispatcher council              # Mock mode (default)
  ai-dispatcher council --real       # Connect to real tools
  [council] > Implement authentication
  [claude]  I propose using JWT middleware
  [codex]   Consider OAuth2 for better security
  [council] > codex, why OAuth2?
  [codex]   It allows token revocation on the server
  [council] > ejecuta`,
	Run: runCouncil,
}

func init() {
	councilCmd.Flags().BoolVar(&councilMock, "mock", true, "Use mock responses (default: true)")
	councilCmd.Flags().BoolVar(&councilReal, "real", false, "Connect to real AI tools (requires installation)")
}

var councilReal bool

// checkToolAvailability checks which tools are installed and available
func checkToolAvailability() map[string]bool {
	available := make(map[string]bool)
	allTrackers := trackers.GetAllTrackers()

	for _, tracker := range allTrackers {
		// Map tool names to standard keys
		toolType := tracker.GetToolType()
		var toolKey string
		switch toolType {
		case trackers.ClaudeCodeTool:
			toolKey = "claude"
		case trackers.CodexTool:
			toolKey = "codex"
		case trackers.OpenCodeTool:
			toolKey = "opencode"
		default:
			toolKey = strings.ToLower(tracker.GetToolName())
			toolKey = strings.ReplaceAll(toolKey, " ", "")
			toolKey = strings.ReplaceAll(toolKey, "-", "")
		}

		// Check if tool is available
		isAvail, _ := tracker.IsAvailable()
		available[toolKey] = isAvail
	}

	return available
}

// ToolStatus represents the status message for a tool
type ToolStatus struct {
	Name      string
	Available bool
	Message   string
}

// getToolStatusMessages returns status messages for all tools
func getToolStatusMessages(available map[string]bool) []ToolStatus {
	statuses := []ToolStatus{
		{
			Name:      "claude",
			Available: available["claude"],
			Message:   "Claude Code is taking a break ☕",
		},
		{
			Name:      "codex",
			Available: available["codex"],
			Message:   "OpenAI Codex is recharging 🔋",
		},
		{
			Name:      "opencode",
			Available: available["opencode"],
			Message:   "OpenCode is offline 📴",
		},
	}
	return statuses
}

func runCouncil(cmd *cobra.Command, args []string) {
	// Initialize colors
	councilColor := color.New(color.FgCyan).SprintFunc()
	claudeColor := color.New(color.FgGreen).SprintFunc()
	codexColor := color.New(color.FgBlue).SprintFunc()
	opencodeColor := color.New(color.FgYellow).SprintFunc()

	// Color map for tools
	toolColors := map[string]func(a ...interface{}) string{
		"claude":   claudeColor,
		"codex":    codexColor,
		"opencode": opencodeColor,
	}

	// Initialize orchestrator
	useReal := councilReal
	var orch *council.Orchestrator
	var availableTools map[string]bool

	if useReal {
		orch = council.NewOrchestrator()
		orch.SetUseMocks(false)
		// Check which tools are available
		availableTools = checkToolAvailability()
		orch.SetAvailableTools(availableTools)
	} else {
		orch = council.NewMockOrchestrator()
		// In mock mode, all tools are "available"
		availableTools = map[string]bool{
			"claude":   true,
			"codex":    true,
			"opencode": true,
		}
		orch.SetAvailableTools(availableTools)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("🏛️  Council Mode - Interactive AI Panel")
	if useReal {
		fmt.Println("   Mode: REAL - Connecting to actual AI tools")
		// Show availability status
		toolStatuses := getToolStatusMessages(availableTools)
		hasUnavailable := false
		for _, status := range toolStatuses {
			if !status.Available {
				if !hasUnavailable {
					fmt.Println()
					fmt.Println("   Some council members are currently unavailable:")
					hasUnavailable = true
				}
				fmt.Printf("   • %s\n", status.Message)
			}
		}
		if hasUnavailable {
			fmt.Println()
		}
	} else {
		fmt.Println("   Mode: MOCK - Using simulated responses for testing")
		fmt.Println("   Tip: Use --real flag to connect to actual tools")
	}
	fmt.Println()
	fmt.Println("   Type your questions or tasks. Available tools will respond initially.")
	fmt.Println("   Mention a tool by name (claude, codex, opencode) to direct questions.")
	fmt.Println("   Commands: ejecuta [tool] | exit | quit")
	fmt.Println()

	for {
		// Show prompt
		fmt.Print(councilColor("[council] ") + "> ")

		// Read input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			break
		}

		// Trim and clean input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check for exit
		lowerInput := strings.ToLower(input)
		if lowerInput == "exit" || lowerInput == "quit" {
			fmt.Println()
			fmt.Println("👋 Leaving council mode.")
			fmt.Println()
			break
		}

		// Check for ejecuta command
		if strings.HasPrefix(lowerInput, "ejecuta") {
			tool := ""
			parts := strings.Fields(input)
			if len(parts) > 1 {
				tool = strings.ToLower(parts[1])
			}

			// Execute with tool (or last mentioned if not specified)
			result := orch.Execute(tool)
			fmt.Println()
			if tool == "" && orch.GetSession().GetLastTool() != "" {
				tool = orch.GetSession().GetLastTool()
			}
			fmt.Println(toolColors[tool]("["+tool+"]") + " " + result)
			fmt.Println()
			continue
		}

		// Add user message to history
		orch.GetSession().AddMessage("user", input)

		// Detect if a specific tool is mentioned
		mentionedTool := council.DetectTool(input)

		if mentionedTool != "" {
			// Direct question to specific tool
			orch.GetSession().SetLastTool(mentionedTool)
			response := orch.Query(mentionedTool, input)
			// Only display if tool responded successfully
			if response.Error == nil && response.Content != "" {
				fmt.Println()
				fmt.Println(toolColors[response.Tool]("["+response.Tool+"]") + "  " + response.Content)
				fmt.Println()
			}
		} else {
			// First message or broadcast to all
			responses := orch.Broadcast(input)
			fmt.Println()
			for _, resp := range responses {
				// Only display if tool responded successfully
				if resp.Error == nil && resp.Content != "" {
					fmt.Println(toolColors[resp.Tool]("["+resp.Tool+"]") + "  " + resp.Content)
				}
			}
			fmt.Println()
		}
	}
}
