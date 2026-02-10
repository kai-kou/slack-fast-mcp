# 修正計画書: README.md / README_ja.md

**作成日**: 2026-02-10
**バージョン**: Ver.1
**対象ドキュメント**: README.md（英語版）、README_ja.md（日本語版）
**ステータス**: 全Phase修正完了（2026-02-10）

---

## 修正方針

本修正計画書は、7つの専門軸によるレビュー結果と統合レビューに基づき、**OSS 公開前に README を最適化するための具体的な修正内容**を定義する。

### 修正の3原則

1. **英語版・日本語版は必ず同時に修正する**（同期崩れ防止）
2. **既存の簡潔さ・読みやすさを維持する**（情報を増やしすぎない）
3. **初見ユーザーの「完走率」を最大化する**（セットアップ → 初回成功体験の導線）

---

## Phase 1: 必須修正（Must Fix）— 公開前に必ず対応

### 修正 1-1: README 冒頭の強化

**関連指摘**: S-01, S-02, P-03, RK-10, H-01

**修正内容（英語版）**:

```markdown
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
```

**修正内容（日本語版）**: 同等の内容を日本語で反映。バッジは同一のものを使用。

**修正理由**:
- バッジ追加により、プロジェクトの品質・活発さが一目でわかる
- MCP へのリンクにより、MCP を知らない読者も文脈を理解できる
- 対応エディタの明示により、Cursor 以外のユーザーも対象であることが伝わる
- デモ GIF のプレースホルダーを設置（GIF 制作は別タスク）

---

### 修正 1-2: Slack App スコープの完全化

**関連指摘**: L-02, L-03, RK-07

**修正内容（英語版）**: Quick Start の手順2を以下に差し替え:

```markdown
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
```

**修正内容（日本語版）**: 同等の内容を日本語で反映。

**修正理由**:
- 必須/推奨の区分を明確にすることで、ユーザーが必要なスコープを正しく設定できる
- `users:read` を省略した場合の影響を具体的に記載
- セットアップ詳細ガイドへのリンクを冒頭に追加

---

### 修正 1-3: Bot チャンネル招待の強調

**関連指摘**: E-01

**修正内容（英語版）**: Quick Start に独立したステップ5を追加:

```markdown
### 5. Invite the Bot to Your Channel

> ⚠️ **This step is required.** The bot cannot post to or read from a channel unless it has been invited.

In Slack, open the target channel and type:

```
/invite @your-bot-name
```
```

**修正内容（日本語版）**: 同等の内容を日本語で反映。

**修正理由**:
- `not_in_channel` は最も頻出するエラーであり、独立したステップとして視覚的に強調することで発生を予防

---

### 修正 1-4: セキュリティセクションの強化

**関連指摘**: RK-01, RK-02, RK-03, P-04, R-04

**修正内容（英語版）**: Security セクションを以下に差し替え:

```markdown
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
```

**修正内容（日本語版）**: 同等の内容を日本語で反映。

**修正理由**:
- セキュリティを重視するテックリード・チーム導入検討者への安心材料
- トークン漏洩時の対処法を明記することで、インシデント対応を迅速化
- `.gitignore` 例をコピペ可能にすることで、設定漏れを防止

---

### 修正 1-5: Use with Cursor を Use with AI Editors に拡大

**関連指摘**: S-04

**修正内容（英語版）**: 手順4のタイトルと内容を更新:

```markdown
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
```

**修正内容（日本語版）**: 同等の内容を日本語で反映。

**修正理由**:
- MCP は Cursor だけのものではなく、Windsurf、Claude Desktop 等にも対応
- 対応エディタを明示することで、潜在的なユーザー層を広げる

---

## Phase 2: 推奨修正（Should Fix）— 公開後早期に対応

### 修正 2-1: Contributing セクションの追加

**関連指摘**: S-03

**修正内容**: License セクションの直前に追加:

```markdown
## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

- 🐛 [Report a bug](https://github.com/kai-ko/slack-fast-mcp/issues/new)
- 💡 [Request a feature](https://github.com/kai-ko/slack-fast-mcp/issues/new)
- 📖 Improve documentation

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.
```

---

### 修正 2-2: `go install` インストール方法の追加

**関連指摘**: E-04

**修正内容**: Quick Start のダウンロードセクションに追加:

```markdown
### 1. Install

#### Option A: Download binary (recommended)
... (existing content) ...

#### Option B: Go install
```bash
go install github.com/kai-ko/slack-fast-mcp/cmd/slack-fast-mcp@latest
```

#### Option C: Build from source
```bash
git clone https://github.com/kai-ko/slack-fast-mcp.git
cd slack-fast-mcp && make build
```
```

---

### 修正 2-3: 環境変数の永続化方法の補足

**関連指摘**: E-03

**修正内容**: Configure セクションに注記を追加:

```markdown
> **Note:** `export` sets the variable for the current terminal session only. To persist it, add the line to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) and restart your terminal.
```

---

### 修正 2-4: トラブルシューティングセクションの追加

**関連指摘**: E-06

**修正内容**: Security セクションの後に追加:

```markdown
## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `not_in_channel` | Bot not invited to channel | `/invite @your-bot-name` in the channel |
| `invalid_auth` | Token is invalid or expired | Regenerate at [api.slack.com/apps](https://api.slack.com/apps) |
| `channel_not_found` | Wrong channel name | Check spelling, don't include `#` prefix |
| `missing_scope` | OAuth scope not added | Add scope in Slack App settings, reinstall app |

For more details, see the [Slack App Setup Guide](./docs/slack-app-setup.md#8-トラブルシューティング).
```

---

### 修正 2-5: バイナリサイズの実測値更新

**関連指摘**: L-01

**修正手順**:
1. `make build` でバイナリをビルド
2. `ls -lh` でサイズを確認
3. 比較表の `~10MB` を実測値に更新（例: `~12MB`）
4. 要件定義書の `~10-15MB` との整合性を確認

---

### 修正 2-6: Windows インストール手順の補完

**関連指摘**: L-05

**修正内容**: Other platforms の `<details>` 内に追加:

```markdown
> **Windows PATH:** If `$env:USERPROFILE\bin` is not in your PATH, add it:
> ```powershell
> [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
> ```
> Restart PowerShell after adding.
```

---

### 修正 2-7: `serve` コマンドの記載

**関連指摘**: L-04

**修正内容**: CLI Usage セクションに追加:

```markdown
# Start as MCP Server (default when no subcommand is given)
slack-fast-mcp serve
```

---

## Phase 3: 任意修正（Nice to Have）— 余裕がある時に対応

| # | 修正内容 | 関連指摘 | 備考 |
|---|---|---|---|
| 3-1 | デモ GIF の作成・追加 | P-01 | 別タスクとして GIF 制作を計画 |
| 3-2 | ロードマップセクションの追加 | S-05 | 設計書 §8 から抜粋 |
| 3-3 | CLI パイプ連携例の追加 | P-02 | `jq` との連携例 |
| 3-4 | MCP Tools にワンライナー使用例 | R-02 | 各ツールの冒頭に1行例 |
| 3-5 | Configuration の折りたたみ整理 | R-03 | `<details>` 活用 |
| 3-6 | ベンチマーク条件の注記 | RK-05 | 計測環境の記載 |
| 3-7 | Acknowledgments セクション | H-05 | 依存ライブラリへの謝辞 |
| 3-8 | バイナリチェックサムの検証手順 | RK-08 | GoReleaser の checksums.txt |
| 3-9 | MCP の1行説明とリンク | RK-10 | Phase 1 で冒頭に統合済み |
| 3-10 | 開発動機の一文 | H-04 | 任意 |

---

## 修正後の README 構成（想定）

```
# slack-fast-mcp
  [バッジ: CI | Release | Go Report Card | License]
  1行説明（MCP リンク付き）
  対応エディタの明示
  言語切替リンク
  [デモ GIF（Phase 3）]

## Why slack-fast-mcp?
  比較表（現行のまま）
  起動モデルの簡潔な説明

## Features
  現行のまま

## Quick Start
  ### 1. Install
    Option A: Download binary
    Option B: go install  ← NEW
    Option C: Build from source
  ### 2. Create a Slack App
    セットアップガイドへのリンク  ← NEW
    必須/推奨スコープの区分  ← IMPROVED
  ### 3. Configure
    setup wizard（推奨）
    手動設定（補足）
    環境変数永続化の注記  ← NEW
  ### 4. Use with AI Editors  ← RENAMED
    Cursor / Windsurf
    Claude Desktop  ← NEW
  ### 5. Invite the Bot  ← NEW (独立ステップ化)

## MCP Tools
  現行のまま

## CLI Usage
  serve コマンド追加  ← NEW
  現行のまま

## Configuration
  現行のまま

## Security  ← EXPANDED
  Token Management
  .gitignore entries
  What this tool does NOT do  ← NEW
  If a token is leaked  ← NEW

## Troubleshooting  ← NEW
  よくあるエラーと対処法

## Contributing  ← NEW

## Building from Source
  現行のまま

## License
  現行のまま
```

---

## 同期チェックリスト

修正実施時に以下を確認すること:

- [ ] 英語版と日本語版で同一の修正が反映されているか
- [ ] 新規追加セクションが両版に存在するか
- [ ] バッジの URL が正しいか（リポジトリ名・オーナー名）
- [ ] 内部リンク（`./docs/...`）が正しく機能するか
- [ ] `<details>` タグが正しく閉じられているか
- [ ] コードブロックの言語指定が正しいか
- [ ] 表のフォーマットが崩れていないか（GitHub でのプレビューで確認）

---

## 見積もり

| Phase | 修正数 | 推定工数 | 対応時期 |
|---|---|---|---|
| Phase 1（必須） | 5件 | 2-3時間 | OSS 公開前 |
| Phase 2（推奨） | 7件 | 1-2時間 | 公開後1週間以内 |
| Phase 3（任意） | 10件 | 2-3時間 | 余裕がある時 |

---

## 参照レビュー結果

| レビュー結果 | ファイル |
|---|---|
| ① 戦略 | `reviews/README/review_strategy.md` |
| ② 論理・MECE | `reviews/README/review_logic-mece.md` |
| ③ 実行設計 | `reviews/README/review_execution.md` |
| ④ ポジション | `reviews/README/review_perspective.md` |
| ⑤ 可読性 | `reviews/README/review_readability.md` |
| ⑥ ヒューマンライズ | `reviews/README/review_humanize.md` |
| ⑦ リスク | `reviews/README/review_risk.md` |
| 統合レビュー | `reviews/README/統合レビュー結果.md` |
