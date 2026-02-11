# slack-fast-mcp

[![CI](https://github.com/kai-kou/slack-fast-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/kai-kou/slack-fast-mcp/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kai-kou/slack-fast-mcp)](https://github.com/kai-kou/slack-fast-mcp/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/kai-kou/slack-fast-mcp)](https://goreportcard.com/report/github.com/kai-kou/slack-fast-mcp)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kai-kou/slack-fast-mcp)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

**The fastest Slack [MCP](https://modelcontextprotocol.io/) server.** Written in Go, starts in ~10ms. No runtime, no dependencies — just a single binary.

Post messages, read history, and reply to threads from AI editors like [Cursor](https://cursor.com), [Windsurf](https://codeium.com/windsurf), [Claude Desktop](https://claude.ai/download), or your terminal.

[Japanese / 日本語版はこちら](./README_ja.md)

<p align="center">
  <img src="./docs/assets/hero-image.png" alt="slack-fast-mcp overview" width="800">
</p>

---

## Table of Contents

- [Why slack-fast-mcp?](#why-slack-fast-mcp)
- [Design Philosophy](#design-philosophy)
- [What Can You Do?](#what-can-you-do)
- [Quick Start](#quick-start)
- [MCP Tools](#mcp-tools)
- [CLI Usage](#cli-usage)
- [Configuration](#configuration)
- [Security](#security)
- [Troubleshooting](#troubleshooting)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)
- [License](#license)

---

## Why slack-fast-mcp?

MCP servers start a new process **for every request**. Startup speed directly impacts your experience.

| | slack-fast-mcp | Node.js MCP | Python MCP |
|---|---|---|---|
| **Startup** | ~10 ms | ~200–500 ms | ~300–800 ms |
| **Install** | Single binary | `npm install` | `pip install` |
| **Runtime** | None required | Node.js | Python |
| **Binary size** | ~10 MB | N/A | N/A |

> Benchmark: startup time measured on Apple M1 (macOS). Actual values vary by hardware.

### Features

- **5 MCP Tools** — `slack_post_message`, `slack_get_history`, `slack_post_thread`, `slack_add_reaction`, `slack_remove_reaction`
- **CLI Mode** — Use from the terminal: `post`, `history`, `reply`
- **Setup Wizard** — Interactive `slack-fast-mcp setup` for easy configuration
- **Per-project Config** — `.slack-mcp.json` for project-specific Slack settings
- **Cross-platform** — macOS, Linux, Windows binaries available
- **Secure** — Environment variable references for tokens, hardcoded-token warnings

---

## Design Philosophy

> **Do a few things exceptionally well, rather than everything adequately.**

There are [many Slack MCP servers](https://mcp.so/tag/slack) available. slack-fast-mcp takes a deliberately minimalist approach:

| | slack-fast-mcp | Feature-rich alternatives |
|---|---|---|
| **Design** | Minimalist — 5 focused tools | Full-featured — 8+ tools |
| **Auth** | Standard Bot Token (`xoxb`) | Bot / User / Browser tokens |
| **Dependencies** | ~15 (binary ~10 MB) | 100+ (binary ~15 MB) |
| **CLI mode** | Built-in terminal commands | MCP server only |
| **Setup** | Interactive wizard + per-project config | Environment variables |
| **Reactions** | Add/remove emoji reactions | Full reaction management |
| **DMs / Search** | Not supported | Supported |

### Choose slack-fast-mcp if you value:

- **Speed** — Sub-10ms cold starts on every request
- **Simplicity** — One binary, zero runtime dependencies, 3-minute setup
- **Developer experience** — CLI mode, setup wizard, project-scoped config
- **Clean auth** — Standard Slack Bot Token, no workarounds needed

### Consider alternatives if you need:

Message search, DM/Group DM support, browser-token authentication, or SSE/HTTP transports.

> **See also:** [korotovsky/slack-mcp-server](https://github.com/korotovsky/slack-mcp-server) (1k+ stars, Go, feature-rich) is a great option if you need advanced capabilities.

---

## What Can You Do?

Here are some real-world scenarios where slack-fast-mcp shines:

- **Daily standup from your editor** — Ask AI to post a progress update to `#daily-standup` without leaving your code
- **Pull request notifications** — Let AI post a summary to Slack when you finish a PR
- **Thread-based collaboration** — Read and reply to Slack threads directly from Cursor or Claude Desktop
- **CI/CD status reporting** — Pipe build results to a Slack channel via CLI
- **Emoji reactions** — Let AI react to messages with emoji (e.g. :thumbsup: for acknowledgment, :eyes: for "looking into it")
- **Team log / journal** — Auto-post session summaries to your personal `#times-*` channel

---

## Quick Start

### Prerequisites

- A [Slack workspace](https://slack.com/) you can add apps to
- One of: macOS, Linux, or Windows

### 1. Install

#### Option A: Download binary (recommended)

Download the latest binary from [GitHub Releases](https://github.com/kai-kou/slack-fast-mcp/releases):

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/kai-kou/slack-fast-mcp/releases/latest/download/slack-fast-mcp_Darwin_arm64.tar.gz
tar xzf slack-fast-mcp_Darwin_arm64.tar.gz
sudo mv slack-fast-mcp /usr/local/bin/
```

<details>
<summary>Other platforms</summary>

```bash
# macOS (Intel)
curl -LO https://github.com/kai-kou/slack-fast-mcp/releases/latest/download/slack-fast-mcp_Darwin_amd64.tar.gz
tar xzf slack-fast-mcp_Darwin_amd64.tar.gz
sudo mv slack-fast-mcp /usr/local/bin/

# Linux (x86_64)
curl -LO https://github.com/kai-kou/slack-fast-mcp/releases/latest/download/slack-fast-mcp_Linux_amd64.tar.gz
tar xzf slack-fast-mcp_Linux_amd64.tar.gz
sudo mv slack-fast-mcp /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/kai-kou/slack-fast-mcp/releases/latest/download/slack-fast-mcp_Windows_amd64.zip" -OutFile slack-fast-mcp.zip
Expand-Archive slack-fast-mcp.zip -DestinationPath "$env:USERPROFILE\bin"
```

> **Windows PATH:** If `$env:USERPROFILE\bin` is not in your PATH, add it:
> ```powershell
> [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
> ```
> Restart PowerShell after adding.

</details>

> **macOS Gatekeeper:** If you see a warning, run: `xattr -d com.apple.quarantine /usr/local/bin/slack-fast-mcp`

#### Option B: Go install

```bash
go install github.com/kai-kou/slack-fast-mcp/cmd/slack-fast-mcp@latest
```

#### Option C: Build from source

```bash
git clone https://github.com/kai-kou/slack-fast-mcp.git
cd slack-fast-mcp && make build
```

Verify:

```bash
slack-fast-mcp version
```

### 2. Create a Slack App

> For a detailed walkthrough with screenshots, see the [Slack App Setup Guide](./docs/slack-app-setup.md).

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From scratch**
2. Add **Bot Token Scopes** under **OAuth & Permissions**:

   | Scope | Purpose | Required? |
   |---|---|---|
   | `chat:write` | Post messages | **Yes** |
   | `channels:history` | Read public channel history | **Yes** |
   | `channels:read` | Resolve channel names | **Yes** |
   | `reactions:write` | Add/remove emoji reactions | Recommended |
   | `users:read` | Display usernames in history | Recommended |
   | `groups:history` | Read private channel history | Optional |
   | `groups:read` | Resolve private channel names | Optional |

3. **Install** the app to your workspace
4. Copy the **Bot User OAuth Token** (`xoxb-...`)

### 3. Configure

The easiest way — run the setup wizard:

```bash
slack-fast-mcp setup
```

Or configure manually:

```bash
export SLACK_BOT_TOKEN='xoxb-your-token-here'
```

<details>
<summary>Persisting the token across terminal sessions</summary>

The `export` command only sets the variable for the current session. To persist it (and for AI editors like Cursor to pick it up), add the export line to your shell profile:

| Shell | File to edit | How to check |
|---|---|---|
| **zsh** (macOS default) | `~/.zprofile` or `~/.zshrc` | `echo $SHELL` shows `/bin/zsh` |
| **bash** | `~/.bash_profile` or `~/.bashrc` | `echo $SHELL` shows `/bin/bash` |

```bash
# Example: Add to ~/.zprofile (macOS with zsh)
echo "export SLACK_BOT_TOKEN='xoxb-your-token-here'" >> ~/.zprofile
source ~/.zprofile
```

For detailed instructions per OS, see the [Slack App Setup Guide §5.1](./docs/slack-app-setup.md#51-方法a-環境変数で設定推奨).

</details>

### 4. Add to Your AI Editor

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

> **Note:** Claude Desktop may not support `${VAR}` environment variable expansion. If you set the token directly, ensure this config file is **not** committed to Git. It is typically stored in your user directory, so this is usually safe.

> slack-fast-mcp works with **any MCP-compatible tool** via stdio transport.

### 5. Invite the Bot

> **This step is required.** The bot cannot post to or read from a channel unless it has been invited.

In Slack, open the target channel and type:

```
/invite @your-bot-name
```

---

## MCP Tools

Five tools are available when connected as an MCP server:

### `slack_post_message`

Post a message to a Slack channel.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID. Defaults to config value |
| `message` | string | **Yes** | Message text ([Slack mrkdwn](https://api.slack.com/reference/surfaces/formatting) supported) |
| `display_name` | string | No | Sender name (appends `#name` hashtag to message) |

**Example:**

```
slack_post_message(channel: "general", message: "Hello from MCP!")
```

### `slack_get_history`

Get message history from a channel.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID. Defaults to config value |
| `limit` | integer | No | Number of messages, 1–100 (default: 10) |
| `oldest` | string | No | Start time (Unix timestamp) |
| `latest` | string | No | End time (Unix timestamp) |

**Example:**

```
slack_get_history(channel: "general", limit: 5)
```

### `slack_post_thread`

Reply to a message thread.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID. Defaults to config value |
| `thread_ts` | string | **Yes** | Timestamp of the parent message |
| `message` | string | **Yes** | Reply text ([Slack mrkdwn](https://api.slack.com/reference/surfaces/formatting) supported) |
| `display_name` | string | No | Sender name (appends `#name` hashtag to message) |

**Example:**

```
slack_post_thread(channel: "general", thread_ts: "1234567890.123456", message: "Got it!")
```

### `slack_add_reaction`

Add an emoji reaction to a message.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID. Defaults to config value |
| `timestamp` | string | **Yes** | Timestamp of the message to react to |
| `reaction` | string | **Yes** | Emoji name without colons (e.g. `thumbsup`, `heart`, `eyes`) |

**Example:**

```
slack_add_reaction(channel: "general", timestamp: "1234567890.123456", reaction: "thumbsup")
```

### `slack_remove_reaction`

Remove an emoji reaction from a message.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `channel` | string | No | Channel name or ID. Defaults to config value |
| `timestamp` | string | **Yes** | Timestamp of the message to remove reaction from |
| `reaction` | string | **Yes** | Emoji name without colons (e.g. `thumbsup`, `heart`, `eyes`) |

**Example:**

```
slack_remove_reaction(channel: "general", timestamp: "1234567890.123456", reaction: "thumbsup")
```

---

## CLI Usage

slack-fast-mcp also works as a standalone CLI tool:

```bash
# Post a message
slack-fast-mcp post --channel general --message "Hello from CLI!"

# Get channel history
slack-fast-mcp history --channel general --limit 20

# Reply to a thread
slack-fast-mcp reply --channel general --thread-ts 1234567890.123456 --message "Reply here"

# Add a reaction
slack-fast-mcp react --channel general --timestamp 1234567890.123456 --reaction thumbsup

# Remove a reaction
slack-fast-mcp unreact --channel general --timestamp 1234567890.123456 --reaction thumbsup

# JSON output (pipe to jq for pretty printing)
slack-fast-mcp history --channel general --json | jq '.messages[].text'

# Start as MCP server (default when no subcommand is given)
slack-fast-mcp serve

# Show version
slack-fast-mcp version

# Run setup wizard
slack-fast-mcp setup
```

---

## Configuration

Configuration is resolved in the following priority (highest first):

| Priority | Source | Example |
|---|---|---|
| 1 | CLI flags | `--token`, `--channel` |
| 2 | Environment variables | `SLACK_BOT_TOKEN`, `SLACK_DEFAULT_CHANNEL` |
| 3 | Project config file | `.slack-mcp.json` |
| 4 | Global config file | `~/.config/slack-fast-mcp/config.json` |

### Project Config: `.slack-mcp.json`

Create a `.slack-mcp.json` in your project root to set defaults:

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general",
  "display_name": "my-bot"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | **Yes** | Bot token. Use `${ENV_VAR}` to reference environment variables |
| `default_channel` | string | No | Default channel name or ID |
| `display_name` | string | No | Default sender name (appends `#name` hashtag to messages) |

### Environment Variables

| Variable | Description |
|---|---|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token |
| `SLACK_DEFAULT_CHANNEL` | Default channel name or ID |
| `SLACK_DISPLAY_NAME` | Default sender display name |
| `SLACK_FAST_MCP_LOG_LEVEL` | Log level: `debug`, `info`, `warn`, `error` |

---

## Security

### Token Management

- **Never hardcode tokens** in files committed to Git
- Use `${SLACK_BOT_TOKEN}` environment variable references in config files
- The tool **detects and warns** if it finds hardcoded tokens (starting with `xoxb-`, `xoxp-`, `xoxs-`)

### Recommended `.gitignore`

```gitignore
.slack-mcp.json
```

### What This Tool Does NOT Do

- Does **not** store any data locally (messages, tokens, or credentials)
- Does **not** have admin/management permissions — only reads and posts messages
- All communication with Slack is over **HTTPS**

### If a Token Is Leaked

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Select your app → **OAuth & Permissions**
3. Click **Revoke Token** to invalidate the compromised token
4. Reinstall the app to generate a new token

---

## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `not_in_channel` | Bot not invited to channel | `/invite @your-bot-name` in the channel |
| `invalid_auth` | Token is invalid or expired | Regenerate at [api.slack.com/apps](https://api.slack.com/apps) |
| `channel_not_found` | Wrong channel name | Check spelling; don't include `#` prefix |
| `missing_scope` | OAuth scope not added | Add scope in Slack App settings, then reinstall |
| `already_reacted` | Already reacted with this emoji | Use a different emoji or remove the existing reaction first |
| `no_reaction` | No reaction to remove | Check the emoji name — the bot can only remove its own reactions |
| `token_not_configured` | No token set | Run `slack-fast-mcp setup` or set `SLACK_BOT_TOKEN` |

For more details, see the [Slack App Setup Guide](./docs/slack-app-setup.md).

---

## Roadmap

| Feature | Priority | Status |
|---|---|---|
| File upload support | Medium | Planned |
| Emoji reactions | Low | **Done** |
| User search / mention | Low | Planned |
| Multi-workspace support | Low | Planned |
| HTTP transport (remote MCP) | Low | Planned |

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

- [Report a bug](https://github.com/kai-kou/slack-fast-mcp/issues/new)
- [Request a feature](https://github.com/kai-kou/slack-fast-mcp/issues/new)
- Improve documentation

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Development

```bash
git clone https://github.com/kai-kou/slack-fast-mcp.git
cd slack-fast-mcp

make build         # Build the binary
make test          # Run tests
make test-race     # Run tests with race detector
make quality       # Full quality gate (vet, build, test, coverage, smoke)
make smoke         # Smoke test the binary
make help          # Show all available targets
```

---

## Acknowledgments

Built with these excellent libraries:

- [mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP SDK
- [slack-go/slack](https://github.com/slack-go/slack) — Slack API client for Go
- [cobra](https://github.com/spf13/cobra) — CLI framework for Go

## License

[MIT License](./LICENSE)
