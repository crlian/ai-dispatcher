package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/crlian/ai-dispatcher/pkg/router"
	"github.com/crlian/ai-dispatcher/pkg/trackers"
)

var (
	statusJSON bool
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all AI tools",
	Long: `Display the current status of all AI coding assistants including:
  • Available capacity percentage
  • Remaining time until limit reset
  • Current cost in 5-hour window
  • Availability status`,
	Run: runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output in JSON format")
}

func runStatus(cmd *cobra.Command, args []string) {
	// Get all trackers
	allTrackers := trackers.GetAllTrackers()

	// Create decision engine
	engine := router.NewDecisionEngine(allTrackers)

	// Get tool status
	statuses, err := engine.GetToolStatus()
	if err != nil {
		exitWithError(fmt.Errorf("failed to get tool status: %w", err))
	}

	// Output based on format
	if statusJSON {
		outputJSON(statuses)
	} else {
		outputTable(statuses)
	}
}

// outputJSON outputs status in JSON format
func outputJSON(statuses []*router.ToolStatus) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(statuses); err != nil {
		exitWithError(fmt.Errorf("failed to encode JSON: %w", err))
	}
}

// outputTable outputs status in a formatted table
func outputTable(statuses []*router.ToolStatus) {
	// Print header
	fmt.Println()
	fmt.Println("📊 AI Tools Status Report")
	fmt.Println()

	// Create tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print table header
	fmt.Fprintf(w, "%-12s\t%10s\t%14s\t%10s\t%s\n", "Tool", "Available", "Remaining Time", "Cost (5h)", "Status")
	fmt.Fprintf(w, "%-12s\t%10s\t%14s\t%10s\t%s\n", "────", "─────────", "──────────────", "──────────", "──────")

	// Color functions
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()

	// Print each tool status
	for _, status := range statuses {
		// Format available percentage
		availStr := fmt.Sprintf("%.1f%%", status.Available)

		// Format remaining time
		var timeStr string
		if status.RemainingTime > 0 {
			hours := status.RemainingTime / 60
			minutes := status.RemainingTime % 60
			if hours > 0 {
				timeStr = fmt.Sprintf("%dh %dm", hours, minutes)
			} else {
				timeStr = fmt.Sprintf("%dm", minutes)
			}
		} else {
			timeStr = "N/A"
		}

		// Format cost
		costStr := router.FormatCost(status.CurrentCost)

		// Format status with color
		var statusStr string
		switch status.Status {
		case "available":
			statusStr = green("✓ Available")
		case "low":
			statusStr = yellow("⚡ Low")
		case "limited":
			statusStr = red("✗ Limited")
		case "error":
			statusStr = gray("⚠ Error")
			availStr = "N/A"
			timeStr = "N/A"
			costStr = "N/A"
		default:
			statusStr = status.Status
		}

		// Print row
		fmt.Fprintf(w, "%-12s\t%10s\t%14s\t%10s\t%s\n",
			status.ToolName,
			availStr,
			timeStr,
			costStr,
			statusStr,
		)

		// Print error if any
		if status.Error != "" {
			fmt.Fprintf(w, "\t%s\t\t\t\n", gray("↳ "+status.Error))
		}
	}

	w.Flush()
	fmt.Println()

	// Print legend
	fmt.Println("Legend:")
	fmt.Printf("  %s - Tool has >20%% capacity available\n", green("✓ Available"))
	fmt.Printf("  %s - Tool has 5-20%% capacity available\n", yellow("⚡ Low"))
	fmt.Printf("  %s - Tool has <5%% capacity available\n", red("✗ Limited"))
	fmt.Println()
}
