package main

import (
	"os"

	"github.com/crlian/ai-dispatcher/cmd"
)

// Version information (set by build flags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Set version info in cmd package
	cmd.Version = Version
	cmd.BuildTime = BuildTime
	cmd.GitCommit = GitCommit

	// Execute root command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
