package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/crlian/ai-dispatcher/pkg/council"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// loadingSpinner shows an animated loading indicator
type loadingSpinner struct {
	stopChan chan struct{}
	wg       sync.WaitGroup
	message  string
}

func newLoadingSpinner(message string) *loadingSpinner {
	return &loadingSpinner{
		stopChan: make(chan struct{}),
		message:  message,
	}
}

func (s *loadingSpinner) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-s.stopChan:
				// Clear the spinner line
				fmt.Print("\r\033[K")
				return
			default:
				if s.message == "" {
					fmt.Printf("\r%s", frames[i%len(frames)])
				} else {
					fmt.Printf("\r%s %s", frames[i%len(frames)], s.message)
				}
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *loadingSpinner) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

var councilReal bool

// councilCmd represents the council command
var councilCmd = &cobra.Command{
	Use:   "council",
	Short: "Interactive council mode with multiple AI tools",
	Long: `Enter an interactive council session where multiple AI tools can discuss
and propose solutions to your coding tasks.

In council mode:
  • All tools respond to the first question
  • Mention a tool by name to direct questions to it
  • Type "plan [tool]" to see an execution plan
  • Type "ejecuta [tool]" to run with the last mentioned tool
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
	councilCmd.Flags().BoolVar(&councilReal, "real", false, "Connect to real AI tools (requires installation)")
}

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
	fmt.Println("   Commands: plan [tool] | ejecuta [tool] | exit | quit")
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

		// Check for plan command
		if strings.HasPrefix(lowerInput, "plan") {
			tool := ""
			parts := strings.Fields(input)
			if len(parts) > 1 {
				tool = strings.ToLower(parts[1])
			}

			fmt.Println()
			spinner := newLoadingSpinner("")
			spinner.Start()

			plan, err := orch.Plan(tool)
			spinner.Stop()

			if err != nil {
				fmt.Println(councilColor("[council]") + " " + err.Error())
			} else {
				displayPlan(plan)
			}
			fmt.Println()
			continue
		}

		// Check for ejecuta command
		if strings.HasPrefix(lowerInput, "ejecuta") {
			tool := ""
			parts := strings.Fields(input)
			if len(parts) > 1 {
				tool = strings.ToLower(parts[1])
			}

			// If no tool specified, use the last mentioned
			if tool == "" {
				tool = orch.GetSession().GetLastTool()
			}

			// Verify tool availability before executing (only in real mode)
			if tool != "" && councilReal && !availableTools[tool] {
				fmt.Println()
				fmt.Println(councilColor("[council]") + " " + fmt.Sprintf("⚠️  %s no está disponible actualmente.", tool))
				fmt.Println()
				continue
			}

			fmt.Println()
			// Start loading spinner for execution
			spinner := newLoadingSpinner("")
			spinner.Start()

			// Execute with tool
			result := orch.Execute(tool)
			spinner.Stop()

			// Handle case where no tool is selected
			if tool == "" || toolColors[tool] == nil {
				fmt.Println(councilColor("[council]") + " " + result)
			} else {
				fmt.Println(toolColors[tool]("["+tool+"]") + " " + result)
			}
			fmt.Println()
			continue
		}

		// Add user message to history
		orch.GetSession().AddMessage("user", input)

		// Detect if user mentions a file and track it
		if file := council.DetectFile(input); file != "" {
			orch.GetSession().SetCurrentFile(file)
		}

		// Detect if a specific tool is mentioned
		mentionedTool := council.DetectTool(input)

		if mentionedTool != "" {
			// Direct question to specific tool
			orch.GetSession().SetLastTool(mentionedTool)
			fmt.Println()

			// Start loading spinner
			spinner := newLoadingSpinner("")
			spinner.Start()

			response := orch.Query(mentionedTool, input)
			spinner.Stop()

			// Only display if tool responded successfully
			if response.Error == nil && response.Content != "" {
				fmt.Println(toolColors[response.Tool]("["+response.Tool+"]") + "  " + response.Content)
				fmt.Println()
			}
		} else {
			// First message or broadcast to all
			responseChan := orch.Broadcast(input)
			fmt.Println()

			// Start loading spinner (no message, just the spinner)
			spinner := newLoadingSpinner("")
			spinner.Start()

			for resp := range responseChan {
				// Stop current spinner
				spinner.Stop()

				// Only display if tool responded successfully
				if resp.Error == nil && resp.Content != "" {
					fmt.Println(toolColors[resp.Tool]("["+resp.Tool+"]") + "  " + resp.Content)
				}

				// Start new spinner for remaining responses
				spinner = newLoadingSpinner("")
				spinner.Start()
			}

			// Stop final spinner when channel closes
			spinner.Stop()
			fmt.Println()
		}
	}
}

// displayPlan renders the plan with Lipgloss styling
func displayPlan(plan *council.Plan) {
	// Colors
	purple := lipgloss.Color("99")
	green := lipgloss.Color("42")
	yellow := lipgloss.Color("214")

	// Header box
	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(0, 2).
		Bold(true)

	header := fmt.Sprintf("Plan de %s", plan.Tool)
	fmt.Println(headerStyle.Render(header))
	fmt.Println()

	// Summary
	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		PaddingLeft(2)
	fmt.Println(summaryStyle.Render(plan.Summary))
	fmt.Println()

	// Files table
	if len(plan.Files) > 0 {
		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(purple)).
			Headers("", "Archivo", "Cambio").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().Bold(true).Foreground(purple)
				}
				return lipgloss.NewStyle().Padding(0, 1)
			})

		for _, f := range plan.Files {
			icon := getActionIcon(f.Action)
			t.Row(icon, f.Path, f.Summary)
		}
		fmt.Println(t)
	}

	// Dependencies
	if len(plan.Dependencies) > 0 {
		fmt.Println()
		depStyle := lipgloss.NewStyle().Foreground(green)
		fmt.Println(depStyle.Render("Dependencias:"))
		for _, d := range plan.Dependencies {
			fmt.Printf("  %s %s\n", depStyle.Render("*"), d)
		}
	}

	// Risks
	if len(plan.Risks) > 0 {
		fmt.Println()
		riskStyle := lipgloss.NewStyle().Foreground(yellow)
		fmt.Println(riskStyle.Render("Consideraciones:"))
		for _, r := range plan.Risks {
			fmt.Printf("  %s %s\n", riskStyle.Render("!"), r)
		}
	}
}

// getActionIcon returns a styled icon for the file action type
func getActionIcon(action string) string {
	switch action {
	case "create":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("+")
	case "modify":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("~")
	case "delete":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("-")
	default:
		return "*"
	}
}
