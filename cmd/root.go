package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information (set by main package)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "ai-dispatcher",
	Short: "AI Dispatcher - Intelligent routing for AI coding assistants",
	Long: `AI Dispatcher orchestrates multiple AI coding assistants (Claude Code, Codex, OpenCode)
and routes tasks intelligently to optimize costs and availability.

Features:
  • Automatic complexity analysis
  • Smart cost optimization
  • Availability tracking
  • Usage limit monitoring
  • Multi-tool support`,
	Version: Version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"AI Dispatcher version %s\nBuild time: %s\nGit commit: %s\n",
		Version, BuildTime, GitCommit,
	))

	// Add subcommands
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(execCmd)
}

// exitWithError prints error and exits
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
