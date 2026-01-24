# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Installation

```bash
# Build for current platform
make build

# Build for all platforms (macOS Intel/ARM64, Linux amd64/arm64, Windows amd64)
make build-all

# Install binary to $GOPATH/bin
make install
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report (generates coverage.html)
make test-coverage
```

### Code Quality

```bash
# Lint code (requires golangci-lint)
make lint

# Format code
make fmt

# Run go vet
make vet

# Tidy go.mod
make tidy
```

### Dependencies

```bash
# Download and verify dependencies
make deps

# Download dependencies
go mod download
```

### Running the Binary

```bash
# Build and run locally
make run

# Run built binary directly
./bin/ai-router status
./bin/ai-router exec "task description"
```

## Architecture Overview

AI Router is a CLI tool that orchestrates multiple AI coding assistants (Claude Code, Codex, OpenCode) through an intelligent routing pipeline. The codebase follows a modular architecture with four main components:

### 1. **Complexity Analysis** (`pkg/analyzers/`)

The `ComplexityAnalyzer` classifies tasks into three levels: `Simple`, `Medium`, or `Complex`.

- **LLM Method**: Attempts to use the cheapest available AI tool to analyze the task (primary method)
- **Heuristic Method**: Uses keyword matching and word count analysis as a fallback
- Returns `ComplexityAnalysis` with: level classification, estimated tokens (rough: 4 chars/token), confidence score, and method used

Key insight: The analyzer tries LLM first but gracefully falls back to heuristic rules if no tool is available.

### 2. **Usage Tracking** (`pkg/trackers/`)

The `UsageTracker` interface fetches real-time availability and cost data for each tool.

**Core Concepts:**
- Tracks availability as a percentage (0-100) based on remaining cost in 5-hour window
- Cost limit varies by plan (Pro: $2.00, Max5: $4.00, Max20: $8.00)
- Availability threshold: tools with <5% capacity are considered unavailable
- Data is cached within a single request (not across commands)

**Trackers:**
- `ClaudeCodeTracker`: Claude Code tool - uses Anthropic OAuth API for utilization data (fully independent, no CLI dependencies)
- `CodexTracker`: Codex tool - wraps `@ccusage/codex` CLI command (embeds `BaseTracker`)
- `OpenCodeTracker`: OpenCode tool - wraps `@ccusage/opencode` CLI command (embeds `BaseTracker`, free tier, no cost limit)

**Architecture:**
- `ClaudeCodeTracker` is fully independent: directly implements all `UsageTracker` interface methods using the Anthropic API
- Other trackers embed `BaseTracker` which provides common functionality for executing CLI commands and parsing results

### 3. **Routing Engine** (`pkg/router/`)

The `DecisionEngine` combines complexity analysis and availability data to select the optimal tool.

**Decision Priority:**
1. Filter to available tools (>5% capacity)
2. Sort by: free tools first → cheaper tools → higher available capacity
3. Handle forced tool selection with warnings about low availability or exceeding limits

**Key Components:**
- `DecisionEngine`: Makes the routing decision
- `CostCalculator`: Estimates token costs and filters available tools
- `RoutingDecision`: Contains selected tool, reason, alternatives, and complexity analysis
- Formatting utilities: `FormatDecision()` and `FormatCost()`

### 4. **Task Execution** (`pkg/delegators/`)

The `Delegator` interface executes the task using the selected tool.

**Execution Flow:**
- Execute task with timeout protection (default: 5 minutes, configurable per command)
- Capture stdout and stderr
- Estimate tokens from output (~4 chars/token)
- Return `DelegationResult` with success status, output, duration, and exit code

**Delegators:**
- `ClaudeCodeDelegator`: Delegates to Claude Code CLI
- `CodexDelegator`: Delegates to Codex CLI
- `OpenCodeDelegator`: Delegates to OpenCode CLI

All delegators embed `BaseDelegator` which provides command execution with context timeout and output capture.

### 5. **CLI Commands** (`cmd/`)

Two main commands implement the user-facing interface:

**`status` Command:**
- Fetches status from all trackers
- Displays availability, remaining time, and current cost per tool
- Outputs formatted table or JSON

**`exec` Command:**
- Runs the complete 5-step pipeline (analyze → check availability → calculate costs → decide → execute)
- Flags: `--force`, `--verbose`, `--dry-run`, `--json`, `--timeout`
- Verbose mode shows each pipeline step with detailed status
- Dry-run mode shows the decision without executing the task

### Pipeline Flow (`cmd/exec.go`)

1. **Analyze Complexity**: Create `ComplexityAnalyzer`, call `AnalyzeComplexity(task)`
2. **Initialize Engine**: Create `DecisionEngine` with all trackers
3. **Check Availability**: Query tool statuses via engine
4. **Make Decision**: Call `engine.MakeDecision(complexity, forceTool)`
5. **Execute Task**: Get delegator, set timeout, execute, return result

The `PipelineResult` struct aggregates results from all stages for both text and JSON output.

## Key Implementation Patterns

### Error Handling

- Trackers return errors if the CLI tool is not installed (helpful message: "tool not installed: X (run: npm install -g X)")
- Analyzer gracefully falls back from LLM to heuristic if LLM analysis fails
- Engine returns detailed error messages if no tools are available
- Delegator captures timeout errors separately from execution errors

### Extensibility Points

- **Add new tracker**: Two patterns available:
  1. **API-based tracker** (like ClaudeCodeTracker): Implement `UsageTracker` interface directly with all 6 methods - no BaseTracker needed
  2. **CLI-based tracker** (like CodexTracker): Embed `BaseTracker`, override methods as needed, create tracker in `trackers/` file, add to `GetAllTrackers()` factory
- **Add new delegator**: Implement `Delegator` interface, create delegator in `delegators/`, add to `GetDelegator()` factory
- **Modify routing logic**: Update priority in `CostCalculator.SortEstimates()` in `pkg/router/calculator.go`
- **Change complexity keywords**: Update keyword lists in `heuristicAnalysis()` in `pkg/analyzers/complexity.go`

### Constants and Configuration

- Availability threshold: `AvailabilityThreshold = 5.0` (percent)
- Plan cost limits: `ProPlanCostLimit`, `Max5PlanCostLimit`, `Max20PlanCostLimit` in `pkg/trackers/tracker.go`
- Default execution timeout: `DefaultTimeout = 5 * time.Minute` in `pkg/delegators/delegator.go`
- Complexity tokens (heuristic): Simple=150, Medium=500, Complex=1500 in `pkg/analyzers/complexity.go`

## Testing

- Unit tests use mocks in `test/mocks/`
- Integration tests in `test/integration_test.go`
- Test files follow pattern `*_test.go` in same package
- Run specific test: `go test ./pkg/router -v -run TestName`

## Dependencies

- `github.com/spf13/cobra`: CLI framework (cobra library)
- `github.com/fatih/color`: Terminal color output
- `go.mod` version: Go 1.21+

## Deployment

The Makefile builds binaries for multiple platforms:
- macOS: `darwin-amd64`, `darwin-arm64`
- Linux: `linux-amd64`, `linux-arm64`
- Windows: `windows-amd64.exe`

Binaries include version info via build flags set in Makefile (`LDFLAGS`).
