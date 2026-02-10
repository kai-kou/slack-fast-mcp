# Contributing to slack-fast-mcp

Thank you for your interest in contributing to slack-fast-mcp! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- **Go 1.25+** ([install](https://go.dev/doc/install))
- **Git**
- A **Slack workspace** for testing (optional, for integration tests)

### Setup

```bash
# Clone the repository
git clone https://github.com/kai-kou/slack-fast-mcp.git
cd slack-fast-mcp

# Install dependencies
go mod download

# Install Git hooks (recommended)
make setup-hooks

# Run tests
make test

# Build
make build
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Follow the existing code style and conventions
- Add tests for new functionality
- Update documentation if needed

### 3. Run Quality Checks

Before pushing, run the full quality gate:

```bash
make quality
```

This runs:
1. `go vet` — static analysis
2. Build verification
3. Tests with race detection
4. Coverage check (minimum 65%)
5. Smoke test (binary startup)
6. Test report generation

### 4. Commit and Push

```bash
git add .
git commit -m "feat: description of your change"
git push origin your-branch-name
```

### 5. Create a Pull Request

Open a pull request on GitHub. Include:
- A clear description of the change
- Why the change is needed
- How to test it

## Code Style

### Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective-go) guidelines
- Use `gofmt` / `goimports` for formatting
- Add comments for all exported types and functions
- Package comments should describe the package's purpose

### Project Structure

```
cmd/slack-fast-mcp/     # Entry point
internal/
  cli/                  # CLI commands (cobra)
  config/               # Configuration loading
  errors/               # Error types
  mcp/                  # MCP server + tool handlers
  slack/                # Slack API client
docs/                   # Documentation
scripts/                # Build & test scripts
```

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation only
- `test:` — Adding or updating tests
- `refactor:` — Code refactoring
- `ci:` — CI/CD changes
- `chore:` — Maintenance tasks

## Testing

### Unit Tests

```bash
make test              # Fast tests
make test-race         # With race detection
make test-cover        # With coverage report
```

### Integration Tests

Requires a Slack Bot Token and test channel:

```bash
SLACK_BOT_TOKEN=xoxb-xxx SLACK_TEST_CHANNEL=bot-test make test-integration
```

### Smoke Tests

```bash
make smoke
```

## Reporting Issues

When reporting a bug, please include:

1. **Version**: Output of `slack-fast-mcp version --json`
2. **OS/Platform**: e.g., macOS 15, Ubuntu 24.04
3. **Steps to reproduce**
4. **Expected behavior**
5. **Actual behavior**
6. **Logs**: Relevant error output (with `SLACK_FAST_MCP_LOG_LEVEL=debug` if applicable)

## Feature Requests

Feature requests are welcome! Please describe:

1. **Use case**: What problem are you trying to solve?
2. **Proposed solution**: How would you like it to work?
3. **Alternatives**: Any alternative solutions you've considered?

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](./LICENSE).
