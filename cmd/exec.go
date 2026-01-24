package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/crlian/ai-dispatcher/pkg/analyzers"
	"github.com/crlian/ai-dispatcher/pkg/delegators"
	"github.com/crlian/ai-dispatcher/pkg/router"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

var (
	execForce   string
	execVerbose bool
	execDryRun  bool
	execJSON    bool
	execTimeout time.Duration
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [task]",
	Short: "Execute a coding task with intelligent routing",
	Long: `Execute a coding task using the best available AI tool.

The router will:
  1. Analyze task complexity
  2. Check tool availability
  3. Calculate costs
  4. Select the optimal tool
  5. Execute the task

Examples:
  ai-router exec "fix bug in auth.go"
  ai-router exec "refactor user service" --verbose
  ai-router exec "add comments" --force opencode
  ai-router exec "implement feature" --dry-run`,
	Args: cobra.ExactArgs(1),
	Run:  runExec,
}

func init() {
	execCmd.Flags().StringVar(&execForce, "force", "", "Force use of specific tool (claude-code, codex, opencode)")
	execCmd.Flags().BoolVarP(&execVerbose, "verbose", "v", false, "Show detailed execution information")
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false, "Show routing decision without executing")
	execCmd.Flags().BoolVar(&execJSON, "json", false, "Output result in JSON format")
	execCmd.Flags().DurationVar(&execTimeout, "timeout", 5*time.Minute, "Execution timeout")
}

func runExec(cmd *cobra.Command, args []string) {
	task := args[0]

	// Validate task
	if strings.TrimSpace(task) == "" {
		exitWithError(fmt.Errorf("task cannot be empty"))
	}

	// Execute the pipeline
	result := executePipeline(task)

	// Output based on format
	if execJSON {
		outputExecJSON(result)
	} else {
		outputExecText(result)
	}

	// Exit with error if execution failed
	if result.ExecutionResult != nil && !result.ExecutionResult.Success {
		os.Exit(1)
	}
}

// PipelineResult contains the complete result of the execution pipeline
type PipelineResult struct {
	Task            string                        `json:"task"`
	Complexity      *analyzers.ComplexityAnalysis `json:"complexity"`
	Decision        *router.RoutingDecision       `json:"decision"`
	ExecutionResult *delegators.DelegationResult  `json:"execution_result,omitempty"`
	DryRun          bool                          `json:"dry_run"`
	Error           string                        `json:"error,omitempty"`
	TotalDuration   time.Duration                 `json:"total_duration"`
}

// executePipeline runs the complete routing and execution pipeline
func executePipeline(task string) *PipelineResult {
	start := time.Now()
	result := &PipelineResult{
		Task:   task,
		DryRun: execDryRun,
	}

	// Step 1: Analyze complexity
	if execVerbose {
		fmt.Println()
		fmt.Println("🔍 Step 1/5: Analyzing task complexity...")
	}

	allTrackers := trackers.GetAllTrackers()
	analyzer := analyzers.NewComplexityAnalyzer(allTrackers)

	complexity, err := analyzer.AnalyzeComplexity(task)
	if err != nil {
		result.Error = fmt.Sprintf("complexity analysis failed: %v", err)
		result.TotalDuration = time.Since(start)
		return result
	}
	result.Complexity = complexity

	if execVerbose {
		fmt.Printf("   Level: %s\n", complexity.Level)
		fmt.Printf("   Tokens: ~%d\n", complexity.Tokens)
		fmt.Printf("   Method: %s (confidence: %.0f%%)\n", complexity.Method, complexity.Confidence*100)
		fmt.Printf("   Reasoning: %s\n", complexity.Reasoning)
	}

	// Step 2: Create decision engine
	if execVerbose {
		fmt.Println()
		fmt.Println("⚙️  Step 2/5: Initializing decision engine...")
	}

	engine := router.NewDecisionEngine(allTrackers)

	// Step 3: Check availability
	if execVerbose {
		fmt.Println()
		fmt.Println("📊 Step 3/5: Checking tool availability...")

		statuses, err := engine.GetToolStatus()
		if err == nil {
			for _, status := range statuses {
				statusIcon := "✓"
				if !status.IsAvailable {
					statusIcon = "✗"
				}
				fmt.Printf("   %s %s: %.1f%% available\n",
					statusIcon, status.ToolName, status.Available)
			}
		}
	}

	// Step 4: Make routing decision
	if execVerbose {
		fmt.Println()
		fmt.Println("🎯 Step 4/5: Making routing decision...")
	}

	decision, err := engine.MakeDecision(complexity, execForce)
	if err != nil {
		result.Error = fmt.Sprintf("routing decision failed: %v", err)
		result.TotalDuration = time.Since(start)
		return result
	}
	result.Decision = decision

	if execVerbose || execDryRun {
		fmt.Println()
		printDecision(decision)
	}

	// Step 5: Execute (if not dry-run)
	if !execDryRun {
		if execVerbose {
			fmt.Println()
			fmt.Println("🚀 Step 5/5: Executing task...")
		}

		delegator, err := delegators.GetDelegator(decision.SelectedTool)
		if err != nil {
			result.Error = fmt.Sprintf("failed to get delegator: %v", err)
			result.TotalDuration = time.Since(start)
			return result
		}

		// Set timeout
		delegator.SetTimeout(execTimeout)

		// Execute task
		ctx := context.Background()
		execResult, err := delegator.Execute(ctx, task)
		if err != nil {
			result.Error = fmt.Sprintf("execution failed: %v", err)
			result.TotalDuration = time.Since(start)
			return result
		}
		result.ExecutionResult = execResult
	}

	result.TotalDuration = time.Since(start)
	return result
}

// printDecision prints the routing decision with colors
func printDecision(decision *router.RoutingDecision) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println("📍 Routing Decision")
	fmt.Printf("   %s: %s\n", cyan("Selected tool"), decision.SelectedName)

	if decision.SelectedCost != nil {
		costStr := router.FormatCost(decision.SelectedCost.EstimatedCost)
		if decision.SelectedCost.EstimatedCost == 0 {
			costStr = green(costStr)
		}
		fmt.Printf("   %s: %s\n", cyan("Estimated cost"), costStr)
		fmt.Printf("   %s: ~%d\n", cyan("Estimated tokens"), decision.SelectedCost.EstimatedTokens)
		fmt.Printf("   %s: %.1f%%\n", cyan("Available capacity"), decision.SelectedCost.AvailablePercent)
	}

	if decision.WasForced {
		fmt.Printf("   %s\n", yellow("⚠️  Tool selection was forced"))
	}

	if execVerbose && len(decision.Alternatives) > 0 {
		fmt.Printf("\n   %s:\n", cyan("Alternatives"))
		for i, alt := range decision.Alternatives {
			if i >= 2 {
				break // Show max 2 alternatives
			}
			status := "available"
			if !alt.IsAvailable {
				status = "unavailable"
			}
			fmt.Printf("      • %s (%s, %s)\n",
				alt.ToolName,
				router.FormatCost(alt.EstimatedCost),
				status,
			)
		}
	}
}

// outputExecText outputs execution result in text format
func outputExecText(result *PipelineResult) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println()

	if result.Error != "" {
		fmt.Printf("%s %s: %s\n", red("✗"), red("Error"), result.Error)
		return
	}

	if result.DryRun {
		fmt.Printf("%s %s\n", cyan("ℹ"), "Dry run completed - no task was executed")
		fmt.Printf("   Total time: %s\n", delegators.FormatDuration(result.TotalDuration))
		return
	}

	if result.ExecutionResult == nil {
		fmt.Println("No execution result")
		return
	}

	exec := result.ExecutionResult

	if exec.Success {
		fmt.Printf("%s Task completed successfully\n", green("✓"))
		fmt.Printf("   Tool: %s\n", exec.ToolName)
		fmt.Printf("   Duration: %s\n", delegators.FormatDuration(exec.Duration))
		fmt.Printf("   Tokens used: ~%d\n", exec.TokensUsed)

		if execVerbose && exec.Output != "" {
			fmt.Println()
			fmt.Println("Output:")
			fmt.Println(delegators.TruncateOutput(exec.Output, 1000))
		}
	} else {
		fmt.Printf("%s Task failed\n", red("✗"))
		fmt.Printf("   Tool: %s\n", exec.ToolName)
		fmt.Printf("   Duration: %s\n", delegators.FormatDuration(exec.Duration))
		fmt.Printf("   Exit code: %d\n", exec.ExitCode)

		if exec.Error != "" {
			fmt.Printf("   Error: %s\n", exec.Error)
		}

		if exec.Output != "" {
			fmt.Println()
			fmt.Println("Output:")
			fmt.Println(delegators.TruncateOutput(exec.Output, 1000))
		}
	}

	fmt.Println()
}

// outputExecJSON outputs execution result in JSON format
func outputExecJSON(result *PipelineResult) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		exitWithError(fmt.Errorf("failed to encode JSON: %w", err))
	}
}
