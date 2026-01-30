# AI Dispatcher

Intelligent tool router for AI coding assistants - automatically selects the optimal AI tool based on real-time availability and task complexity.

## The Problem

You have multiple AI coding assistants installed:
- **Claude Code** - Powerful, specialized for complex tasks
- **Cursor** - Excellent for rapid development
- **OpenCode** - Free alternative with solid capabilities

But each time you need assistance, you face a decision:
- Which tool should I use?
- Is Claude Code available or at capacity?
- Would Cursor be faster for this specific task?
- Should I use the free OpenCode to minimize costs?
- What will the estimated cost be?

Manual selection leads to wasted time, suboptimal tool choices, and unnecessary spending.

## The Solution

AI Dispatcher automatically answers these questions:

```bash
$ ai-dispatcher exec "Refactor authentication to use JWT"

Analyzing task complexity...
   Level: COMPLEX (1500 tokens estimated)

Checking tool availability...
   Claude Code: 80% available (optimal for complexity)
   Cursor: 45% available
   OpenCode: 100% available

Routing decision: Claude Code selected
   Reason: Best for complex tasks with sufficient availability
   Estimated cost: $0.25
   Estimated time: 8-12 minutes

Executing task...
Completed in 8 min 32 sec

Result: Authentication refactored to JWT
   Files modified: 2
   Tests: PASSED
```

AI Dispatcher is a command-line tool that automatically selects the best AI coding assistant based on real-time availability and task complexity analysis.

## Key Features

- **Intelligent Routing**: Automatically selects the optimal tool based on real-time availability and task complexity
- **Cost Optimization**: Prioritizes free and cheaper alternatives when applicable, estimated to save 30-50% on subscription costs
- **Complexity Analysis**: Uses LLM and heuristic methods to classify task complexity (Simple/Medium/Complex)
- **Real-Time Availability**: Monitors tool capacity and remaining quota within 5-hour usage windows
- **Multi-Tool Support**: Works with Claude Code, Cursor, OpenCode, and designed for extensibility
- **Flexible Execution Modes**: Dry-run mode, forced tool selection, verbose output, custom timeouts
- **MCP Support**: Native Model Context Protocol integration for tool communication
- **Advanced Features**: Watch Mode for continuous monitoring, Approval Gate for production workflows

## How It Differs From Alternatives

| Feature | AI Dispatcher | Agency | CLI Agent Orchestrator |
|---------|---------------|--------|------------------------|
| **Primary Use** | Select best tool for single task | Parallelize multiple tasks | Coordinate supervisor-worker agents |
| **Input** | Single task description | Multiple different tasks | Complex hierarchical workflows |
| **Output** | One result from selected tool | Multiple merged results | Orchestrated workflow results |
| **User Complexity** | Simple and focused | Advanced (TUI dashboard) | Enterprise-grade |
| **Learning Curve** | Minimal (one command) | Moderate (TUI + configuration) | Steep (architecture knowledge) |
| **Problem Solved** | Which tool should I use? | How can I run 5 tasks in parallel? | How can I coordinate multiple agents? |

**Real-world examples:**
- **Agency**: I want to write tests AND documentation AND optimize code all at the same time
- **AI Dispatcher**: I want to refactor code and automatically have the system choose the best available tool
- **CLI Agent Orchestrator**: I want a supervisor agent delegating to specialist agents with fault tolerance

## Installation

### From npm (Recommended)

```bash
npm install -g ai-dispatcher
```

### From Source

```bash
git clone https://github.com/crlian/ai-dispatcher.git
cd ai-dispatcher
make install
```

### Download Binary

Pre-built binaries are available on the [releases page](https://github.com/crlian/ai-dispatcher/releases).

## Why Use AI Dispatcher?

**Cost Reduction**
- Automatically uses free tools when sufficient
- Routes to cheaper alternatives when appropriate
- Estimated 30-50% cost reduction on AI tool subscriptions

**Time Efficiency**
- Eliminates decision paralysis about tool selection
- Automatically uses the fastest available option for your task
- Single command instead of managing multiple tools

**Better Results**
- Matches task complexity to tool capability
- Uses specialized tools for their intended strengths
- Prevents suboptimal tool selection

**Visibility and Control**
- Real-time visibility into tool availability and capacity
- Detailed explanations of routing decisions
- Cost tracking and usage transparency

**Safety and Flexibility**
- Dry-run mode to preview decisions without execution
- Approval Gate mode for controlled production workflows
- Manual override capability when needed

## Prerequisites

AI Dispatcher requires at least one AI coding assistant to be installed. Optional tracking tools enable availability monitoring:

### Claude Code

```bash
npm install -g claude
npm install -g ccusage
```

### Cursor

```bash
# Install from https://www.cursor.com/
# No additional tools required
```

### OpenCode

```bash
npm install -g @ccusage/opencode
```

Note: You do not need to install all tools. AI Dispatcher automatically detects available tools on your system.

## Quick Start

### 1. Check Tool Status

View current availability of all installed tools:

```bash
$ ai-dispatcher status
```

Output:
```
AI Tools Status Report

Tool           Available  Remaining Time   Cost (5h)
------------------------------------------------------
Claude Code       78.0%          3h 59m      $1.234
Cursor            45.0%          2h 30m      $0.567
OpenCode         100.0%          5h 00m      $0.000

Status Legend:
  Available - Tool has >20% remaining capacity
  Low - Tool has 5-20% remaining capacity
  Limited - Tool has <5% remaining capacity (skipped unless forced)
```

### 2. Execute a Task

Submit a task and let AI Dispatcher automatically select the best tool:

```bash
$ ai-dispatcher exec "Refactor the authentication service to use JWT instead of sessions"
```

Output:
```
Step 1/5: Analyzing task complexity...
   Level: complex
   Tokens: ~1500
   Method: llm (confidence: 90%)

Step 2/5: Initializing decision engine...

Step 3/5: Checking tool availability...
   Claude Code: 78.0% available
   Cursor: 45.0% available
   OpenCode: 100.0% available

Step 4/5: Making routing decision...

Routing Decision
   Selected tool: Claude Code
   Reason: Best for complex tasks with sufficient availability
   Estimated cost: $0.25
   Available capacity: 78.0%

Step 5/5: Executing task...
   Delegating to Claude Code...
   Connecting to MCP server...
   Processing: Analyzing auth service...
   Refactoring JWT authentication...
   Running tests...

Task completed successfully
   Duration: 8 min 32 sec
   Tool used: Claude Code
   Files modified: 3
   Tests: PASSED
   Cost: $0.24
```

### 3. Execution Modes

**Dry-run mode** - Preview the routing decision without execution:
```bash
$ ai-dispatcher exec "add validation" --dry-run
```

**Verbose mode** - Detailed output of the complete pipeline:
```bash
$ ai-dispatcher exec "optimize queries" --verbose
```

**Force specific tool** - Override automatic selection:
```bash
$ ai-dispatcher exec "simple fix" --force opencode
```

**JSON output** - Machine-readable format for scripting:
```bash
$ ai-dispatcher exec "task description" --json
```

**Custom timeout** - Extend or reduce execution timeout:
```bash
$ ai-dispatcher exec "long-running task" --timeout 30m
```

## Commands

### status

Display the current status of all AI tools:

```bash
ai-dispatcher status
ai-dispatcher status --json
```

### exec

Execute a task with automatic tool routing:

```bash
ai-dispatcher exec "your task description"
ai-dispatcher exec "task" --verbose
ai-dispatcher exec "task" --dry-run
ai-dispatcher exec "task" --force opencode
ai-dispatcher exec "task" --timeout 10m
ai-dispatcher exec "task" --json
```

### council

Interactive council mode - Multiple AI tools discuss and debate before execution:

```bash
ai-dispatcher council              # Start interactive session (mock mode by default)
ai-dispatcher council --real       # Connect to real AI tools
```

In council mode:
- All available tools respond to the initial question
- Mention a tool by name (`claude`, `codex`, `opencode`) to direct questions to it
- Tools can debate and reference each other's responses
- Type `ejecuta` to execute with the last mentioned tool
- Type `exit` or `quit` to leave

Example session:
```bash
[council] > Should we use JWT or session-based auth?
[claude]  JWT is stateless and scalable...
[codex]   Sessions are simpler to debug...
[council] > claude, why JWT over sessions?
[claude]  JWT allows horizontal scaling without shared storage...
[council] > ejecuta claude
[claude]  Executing... Files modified: auth/jwt.go
```

### Flags

- `--force <tool>`: Override automatic selection (claude-code, cursor, opencode)
- `--verbose, -v`: Show detailed pipeline information
- `--dry-run`: Display routing decision without executing
- `--json`: Output results in JSON format
- `--timeout <duration>`: Set execution timeout (default: 5m)

## How It Works

AI Dispatcher follows a five-step intelligent pipeline:

### Step 1: Complexity Analysis

The system analyzes task complexity using two methods:
- **LLM Analysis**: Uses the cheapest available tool to evaluate complexity (primary method)
- **Heuristic Fallback**: Rule-based analysis when LLM is unavailable

Classification:
- **Simple**: Quick fixes, comments, renaming (approximately 50-200 tokens)
- **Medium**: Small features, bug fixes (approximately 200-1000 tokens)
- **Complex**: Architecture changes, refactoring, multi-file edits (approximately 1000+ tokens)

### Step 2: Availability Check

Each tool's availability is determined from:
- **Claude Code**: Anthropic OAuth API (actual utilization percentage)
- **Cursor**: CLI integration for remaining capacity
- **OpenCode**: CLI integration for quota status

Capacity levels:
- **Available**: Greater than 20% remaining capacity
- **Low**: 5-20% remaining capacity
- **Limited**: Less than 5% remaining capacity (skipped unless forced)

### Step 3: Cost Calculation

Costs are estimated based on:
- Token estimation from complexity analysis (approximately 4 characters = 1 token)
- Tool-specific pricing per 1 million tokens:
  - Claude Code: approximately $3 per 1M tokens
  - Cursor: approximately $2 per 1M tokens
  - OpenCode: $0 (free tier)

### Step 4: Routing Decision

The decision engine evaluates tools in this order:
1. Filter to available tools (>5% capacity)
2. Rank by complexity match
3. Evaluate cost versus availability trade-offs
4. Select highest-value option (quality, speed, cost)

Decision examples:
- Simple task with free tool available → OpenCode (minimize cost)
- Complex task with specialized tool required → Claude Code (maximize quality)
- Medium task with multiple suitable options → Cursor (balance cost and capability)

### Step 5: Task Execution and Monitoring

The selected tool executes the task with:
- Timeout protection (default 5 minutes, configurable)
- Output capture (stdout and stderr)
- Token usage estimation
- Execution duration tracking
- Error handling with graceful degradation

## Configuration

AI Dispatcher uses sensible defaults suitable for most use cases. Advanced configuration options are planned for future releases:

- Custom tool-to-command mappings
- Custom complexity thresholds
- Custom pricing per tool
- Custom availability thresholds
- Configuration files (`.ai-dispatcher.yml`)

## Development

### Prerequisites

- Go 1.21 or higher
- Make
- Node.js 14+ (for npm integration)

### Building from Source

```bash
git clone https://github.com/crlian/ai-dispatcher.git
cd ai-dispatcher

make deps          # Download dependencies
make build         # Build for current platform
make build-all     # Build for all platforms (macOS, Linux, Windows)
make test          # Run tests
make test-coverage # Run tests with coverage report
make lint          # Lint code
make fmt           # Format code
```

### Project Structure

```
ai-dispatcher/
├── cmd/                  # CLI commands
│   ├── root.go
│   ├── status.go
│   └── exec.go
├── pkg/
│   ├── analyzers/       # Complexity analysis
│   ├── trackers/        # Usage tracking and availability
│   ├── router/          # Routing decision engine
│   └── delegators/      # Task execution
├── test/
│   ├── mocks/           # Test mocks and fixtures
│   └── integration_test.go
├── npm/                 # npm package wrapper
└── scripts/             # Build and release scripts
```

## Roadmap

### Phase 1: Core (Current Implementation)
- [x] Multi-tool routing (Claude Code, Cursor, OpenCode)
- [x] Real-time availability checking
- [x] Complexity analysis (LLM + Heuristic)
- [x] Cost calculation and optimization
- [x] Dry-run mode

### Phase 2: Integration (Next)
- [ ] Tmux Mode: Interactive terminal sessions
- [ ] MCP Server Integration: Native Model Context Protocol support
- [ ] Configuration file support
- [ ] Custom pricing configuration

### Phase 3: Intelligence (Planned)
- [ ] Watch Mode: Automatic monitoring and fixing of failed tests/errors
- [ ] Approval Gate: Manual approval workflow for production safety
- [ ] Learning mode with historical decision tracking
- [ ] ML-based routing optimization

### Phase 4: Enterprise (Future)
- [ ] Web dashboard for cost visualization and analytics
- [ ] Historical cost tracking and reports
- [ ] Multi-project support
- [ ] Team collaboration features
- [ ] CI/CD integration (GitHub Actions, GitLab CI)

### Phase 5: Advanced (Long-term Vision)
- [ ] Custom agent integration
- [ ] Event-driven execution
- [ ] Distributed task queue
- [ ] Multi-provider orchestration

## Contributing

Contributions are welcome. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and the development process.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Technical Foundation

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Fatih Color](https://github.com/fatih/color) - Terminal output formatting
- [ccusage](https://github.com/crlian/ccusage) - Usage tracking integration

## Related Projects

- [Agency](https://github.com/tobias-walle/agency) - Multi-agent parallelization
- [CLI Agent Orchestrator](https://github.com/awslabs/cli-agent-orchestrator) - Supervisor-worker orchestration
