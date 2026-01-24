# AI Router

Intelligent routing for AI coding assistants - optimize costs and availability automatically.

## Installation

```bash
npm install -g @crlian/ai-router
```

## Quick Start

```bash
# Check status of all AI tools
ai-router status

# Execute a task with intelligent routing
ai-router exec "fix bug in authentication"

# Show help
ai-router --help
```

## Features

- **Intelligent Routing**: Automatically selects the best AI tool
- **Cost Optimization**: Prioritizes free and cheaper alternatives
- **Usage Tracking**: Real-time monitoring of tool availability
- **Multi-Tool Support**: Works with Claude Code, Codex, and OpenCode

## Prerequisites

Install usage tracking tools:

```bash
npm install -g ccusage @ccusage/codex @ccusage/opencode
```

## Usage

### Check Tool Status

```bash
ai-router status
```

### Execute Tasks

```bash
# Basic execution
ai-router exec "add comment to main function"

# Verbose mode
ai-router exec "refactor service" --verbose

# Force specific tool
ai-router exec "task" --force opencode

# Dry run
ai-router exec "task" --dry-run
```

## Documentation

Full documentation: [github.com/crlian/ai-dispatcher](https://github.com/crlian/ai-dispatcher)

## Support

- Issues: [github.com/crlian/ai-dispatcher/issues](https://github.com/crlian/ai-dispatcher/issues)
- Discussions: [github.com/crlian/ai-dispatcher/discussions](https://github.com/crlian/ai-dispatcher/discussions)

## License

MIT © Cesar William
