# Contributing to AI Router

Thank you for your interest in contributing to AI Router! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

Before creating a bug report:
1. Check the [existing issues](https://github.com/crlian/ai-dispatcher/issues) to avoid duplicates
2. Collect information about the bug:
   - AI Router version (`ai-router --version`)
   - Operating system and version
   - Go version (if building from source)
   - Steps to reproduce
   - Expected vs actual behavior

Create a bug report with:
- Clear, descriptive title
- Detailed description
- Steps to reproduce
- Code samples (if applicable)
- Error messages or logs

### Suggesting Features

Feature requests are welcome! Please:
1. Check existing issues for similar suggestions
2. Describe the problem your feature would solve
3. Explain your proposed solution
4. Consider alternative solutions
5. Provide examples of how the feature would be used

### Pull Requests

#### Setup Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ai-dispatcher.git
   cd ai-dispatcher
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/crlian/ai-dispatcher.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Create a branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. Make your changes
2. Add tests for new functionality
3. Run tests:
   ```bash
   make test
   ```

4. Format code:
   ```bash
   make fmt
   ```

5. Run linter:
   ```bash
   make lint
   ```

6. Commit with clear message:
   ```bash
   git commit -m "feat: add support for new AI tool"
   ```

#### Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Examples:
```
feat: add GitHub Copilot support
fix: handle timeout errors correctly
docs: update installation instructions
test: add integration tests for router
```

#### Pull Request Process

1. Update documentation for any changed functionality
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md (if applicable)
5. Submit PR with:
   - Clear title and description
   - Link to related issues
   - Screenshots (for UI changes)
   - Breaking changes notes (if any)

6. Wait for review and address feedback
7. Once approved, a maintainer will merge your PR

## Development Guidelines

### Code Style

- Follow Go best practices and idioms
- Use `gofmt` for formatting
- Keep functions focused and small
- Write clear, self-documenting code
- Add comments for complex logic
- Use meaningful variable names

### Testing

- Write tests for new functionality
- Maintain >80% code coverage
- Include unit, integration, and E2E tests
- Use table-driven tests where appropriate
- Mock external dependencies

Example test structure:
```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if tt.wantErr && err == nil {
                t.Error("expected error, got nil")
            }
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Add inline comments for complex code
- Update command help text
- Include examples in documentation
- Keep API documentation current

### Project Structure

- `cmd/` - CLI commands
- `pkg/` - Reusable packages
- `test/` - Test files and fixtures
- `scripts/` - Build and utility scripts
- `npm/` - npm wrapper package

### Adding New AI Tools

To add support for a new AI tool:

1. Create tracker in `pkg/trackers/`:
   ```go
   type NewToolTracker struct {
       *BaseTracker
   }

   func NewNewToolTracker() *NewToolTracker {
       return &NewToolTracker{
           BaseTracker: NewBaseTracker(
               "New Tool",
               NewToolType,
               "command",
               []string{"args"},
           ),
       }
   }
   ```

2. Add tool type constant:
   ```go
   const NewToolType ToolType = "new-tool"
   ```

3. Create delegator in `pkg/delegators/`
4. Update factory functions
5. Add pricing in `pkg/router/calculator.go`
6. Add tests
7. Update documentation

## Release Process

Releases are handled by maintainers:

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create and push version tag:
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

4. GitHub Actions will:
   - Build binaries for all platforms
   - Create GitHub release
   - Publish to npm

## Getting Help

- Ask questions in [GitHub Discussions](https://github.com/crlian/ai-dispatcher/discussions)
- Check existing documentation
- Look at similar issues or PRs
- Reach out to maintainers

## Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes
- Git commit history

Thank you for contributing to AI Router!
