# slack-fast-mcp

<!-- Badges -->
[![CI](https://github.com/kai-ko/slack-fast-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/kai-ko/slack-fast-mcp/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kai-ko/slack-fast-mcp)](https://github.com/kai-ko/slack-fast-mcp/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/kai-ko/slack-fast-mcp)](https://goreportcard.com/report/github.com/kai-ko/slack-fast-mcp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

The fastest Slack [MCP](https://modelcontextprotocol.io/) Server. Written in Go, starts in ~10ms.

Post messages, read history, and reply to threads — all from AI editors like [Cursor](https://cursor.com), [Windsurf](https://codeium.com/windsurf), [Claude Desktop](https://claude.ai/download), or your terminal.

🇯🇵 [日本語版 README はこちら](./README_ja.md)

<!-- TODO: Add demo GIF here -->
<!-- ![Demo](./docs/assets/demo.gif) -->

## Why slack-fast-mcp?

| | slack-fast-mcp | Node.js MCP | Python MCP |
|---|---|---|---|
| **Startup** | ~10ms | ~200-500ms | ~300-800ms |
| **Install** | Single binary | `npm install` | `pip install` |
| **Runtime** | None | Node.js | Python |
| **Binary size** | ~10MB | N/A | N/A |

MCP Servers start a new process for each request. **Startup speed directly impacts your experience.** slack-fast-mcp is a native Go binary — no runtime, no dependencies, just speed.

> Benchmark: startup time measured on Apple M1 (macOS). Actual values vary by hardware.

## Features

- **3 MCP Tools**: `slack_post_message`, `slack_get_history`, `slack_post_thread`
- **CLI Mode**: Use from the terminal as `slack-fast-mcp post`, `history`, `reply`
- **Setup Wizard**: Interactive `slack-fast-mcp setup` for easy configuration
- **Per-project Config**: `.slack-mcp.json` for project-specific Slack settings
- **Cross-platform**: macOS, Linux, Windows binaries available
- **Secure**: Environment variable references for tokens, hardcoded token warnings

## Quick Start

### 1. Install

#### Option A: Download binary (recommended)

Download the latest binary from [GitHub Releases](https://github.com/kai-ko/slack-fast-mcp/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/kai-ko/slack-fast-mcp/releases/latest/download/slack-fast-mcp_darwin_arm64 -o /usr/local/bin/slack-fast-mcp
chmod +x /usr/local/bin/slack-fast-mcp
```

<details>
<summary>Other platforms</summary>

```bash
# macOS (Intel)
curl -L https://github.com/kai-ko/slack-fast-mcp/releases/latest/download/slack-fast-mcp_darwin_amd64 -o /usr/local/bin/slack-fast-mcp
chmod +x /usr/local/bin/slack-fast-mcp

# Linux (x86_64)
curl -L https://github.com/kai-ko/slack-fast-mcp/releases/latest/download/slack-fast-mcp_linux_amd64 -o /usr/local/bin/slack-fast-mcp
chmod +x /usr/local/bin/slack-fast-mcp

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/kai-ko/slack-fast-mcp/releases/latest/download/slack-fast-mcp_windows_amd64.exe" -OutFile "$env:USERPROFILE\bin\slack-fast-mcp.exe"
```

> **Windows PATH:** If `$env:USERPROFILE\bin` is not in your PATH, add it:
> ```powershell
> [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
> ```
> Restart PowerShell after adding.

</details>

> **macOS Gatekeeper**: If you see a warning, run: `xattr -d com.apple.quarantine /usr/local/bin/slack-fast-mcp`

#### Option B: Go install

```bash
go install github.com/kai-ko/slack-fast-mcp/cmd/slack-fast-mcp@latest
```

#### Option C: Build from source

```bash
git clone https://github.com/kai-ko/slack-fast-mcp.git
cd slack-fast-mcp && make build
```

Verify the installation:

```bash
slack-fast-mcp version
```

### 2. Create a Slack App

> For a detailed walkthrough with screenshots, see the [Slack App Setup Guide](./docs/slack-app-setup.md).

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From scratch**
2. Add **Bot Token Scopes** under **OAuth & Permissions**:

   **Required:**
   - `chat:write` — Post messages
   - `channels:history` — Read public channel history
   - `channels:read` — Resolve channel names

   **Recommended (optional):**
   - `users:read` — Display usernames in history (without this, only user IDs are shown)
   - `groups:history` — Read private channel history
   - `groups:read` — Resolve private channel names

3. **Install** the app to your workspace
4. Copy the **Bot User OAuth Token** (`xoxb-...`)

### 3. Configure

Run the setup wizard (recommended):

```bash
slack-fast-mcp setup
```

Or configure manually:

```bash
# Set the token as an environment variable
export SLACK_BOT_TOKEN='xoxb-your-token-here'

# Create a project config (optional)
echo '{"token":"${SLACK_BOT_TOKEN}","default_channel":"general"}' > .slack-mcp.json
```

> **Note:** `export` sets the variable for the current terminal session only. To persist it, add the line to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) and restart your terminal.

### 4. Use with AI Editors

#### Cursor / Windsurf

Add to `.cursor/mcp.json` (or `.windsurf/mcp.json`):

```json
{
  "mcpServers": {
    "slack-fast-mcp": {
      "command": "/usr/local/bin/slack-fast-mcp",
      "args": [],
      "env": {
        "SLACK_BOT_TOKEN": "${SLACK_BOT_TOKEN}"
      }
    }
  }
}
```

#### Claude Desktop

Add to Claude Desktop's MCP config (Settings → Developer → MCP Servers):

```json
{
  "slack-fast-mcp": {
    "command": "/usr/local/bin/slack-fast-mcp",
    "args": [],
    "env": {
      "SLACK_BOT_TOKEN": "your-token-here"
    }
  }
}
```

> **Note:** slack-fast-mcp works with any MCP-compatible tool via stdio transport.

### 5. Invite the Bot to Your Channel

> **This step is required.** The bot cannot post to or read from a channel unless it has been invited.

In Slack, open the target channel and type:

```
/invite @your-bot-name
```

## MCP Tools

### `slack_post_message`

Post a message to a Slack channel.

```
slack_post_message(channel: "general", message: "Hello World!")
```

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID (defaults to config) |
| `message` | string | Yes | Message text (Slack mrkdwn supported) |

### `slack_get_history`

Get message history from a channel.

```
slack_get_history(channel: "general", limit: 10)
```

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID (defaults to config) |
| `limit` | integer | No | Number of messages (1-100, default: 10) |
| `oldest` | string | No | Start time (Unix timestamp) |
| `latest` | string | No | End time (Unix timestamp) |

### `slack_post_thread`

Reply to a message thread.

```
slack_post_thread(channel: "general", thread_ts: "1234567890.123456", message: "Reply!")
```

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID (defaults to config) |
| `thread_ts` | string | Yes | Thread timestamp to reply to |
| `message` | string | Yes | Reply text (Slack mrkdwn supported) |

## CLI Usage

```bash
# Start as MCP Server (default when no subcommand is given)
slack-fast-mcp serve

# Post a message
slack-fast-mcp post --channel general --message "Hello from CLI!"

# Get channel history
slack-fast-mcp history --channel general --limit 20

# Reply to a thread
slack-fast-mcp reply --channel general --thread-ts 1234567890.123456 --message "Reply here"

# Output in JSON format (pipe to jq for pretty output)
slack-fast-mcp history --channel general --json | jq '.messages[].text'

# Show version
slack-fast-mcp version

# Run setup wizard
slack-fast-mcp setup
```

<details>
<summary>Configuration details</summary>

## Configuration

### Priority (highest to lowest)

1. CLI flags (`--token`, `--channel`)
2. Environment variables (`SLACK_BOT_TOKEN`, `SLACK_DEFAULT_CHANNEL`)
3. Project config (`.slack-mcp.json`)
4. Global config (`~/.config/slack-fast-mcp/config.json`)

### `.slack-mcp.json`

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | Yes | Bot token. Use `${ENV_VAR}` to reference env vars |
| `default_channel` | string | No | Default channel name or ID |

### Environment Variables

| Variable | Description |
|---|---|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token |
| `SLACK_DEFAULT_CHANNEL` | Default channel |
| `SLACK_FAST_MCP_LOG_LEVEL` | Log level (debug/info/warn/error) |

</details>

## Security

### Token Management

- **Never hardcode tokens** in files committed to Git
- Use `${SLACK_BOT_TOKEN}` environment variable references in config files
- The tool **detects and warns** if it finds hardcoded tokens (starting with `xoxb-`, `xoxp-`, `xoxs-`)

### Recommended `.gitignore` entries

```gitignore
.slack-mcp.json
# If you hardcode tokens in Cursor config (not recommended):
# .cursor/mcp.json
```

### What this tool does NOT do

- Does **not** store any data locally (messages, tokens, or credentials)
- Does **not** have admin/management permissions — only reads and posts messages
- All communication with Slack is over **HTTPS**

### If a token is leaked

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Select your app → **OAuth & Permissions**
3. Click **Revoke Token** to invalidate the compromised token
4. Reinstall the app to generate a new token

## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `not_in_channel` | Bot not invited to channel | `/invite @your-bot-name` in the channel |
| `invalid_auth` | Token is invalid or expired | Regenerate at [api.slack.com/apps](https://api.slack.com/apps) |
| `channel_not_found` | Wrong channel name | Check spelling, don't include `#` prefix |
| `missing_scope` | OAuth scope not added | Add scope in Slack App settings, reinstall app |
| `token_not_configured` | No token set | Run `slack-fast-mcp setup` or set `SLACK_BOT_TOKEN` |

For more details, see the [Slack App Setup Guide](./docs/slack-app-setup.md).

## Roadmap

| Feature | Priority | Status |
|---|---|---|
| File upload support | Medium | Planned |
| Emoji reactions | Low | Planned |
| User search / mention | Low | Planned |
| Multi-workspace support | Low | Planned |
| HTTP transport (remote MCP) | Low | Planned |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

- [Report a bug](https://github.com/kai-ko/slack-fast-mcp/issues/new)
- [Request a feature](https://github.com/kai-ko/slack-fast-mcp/issues/new)
- Improve documentation

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## Building from Source

```bash
git clone https://github.com/kai-ko/slack-fast-mcp.git
cd slack-fast-mcp
make build
```

### Development

```bash
make test          # Run tests
make test-race     # Run tests with race detector
make quality       # Full quality gate (vet, build, test, coverage, smoke)
make smoke         # Smoke test the binary
make help          # Show all available targets
```

## Acknowledgments

Built with these excellent libraries:

- [mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP SDK
- [slack-go/slack](https://github.com/slack-go/slack) — Slack API client for Go
- [cobra](https://github.com/spf13/cobra) — CLI framework for Go

## License

[MIT License](./LICENSE)
