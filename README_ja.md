# slack-fast-mcp

<!-- Badges -->
[![CI](https://github.com/kai-ko/slack-fast-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/kai-ko/slack-fast-mcp/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kai-ko/slack-fast-mcp)](https://github.com/kai-ko/slack-fast-mcp/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/kai-ko/slack-fast-mcp)](https://goreportcard.com/report/github.com/kai-ko/slack-fast-mcp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

最速の Slack [MCP](https://modelcontextprotocol.io/) Server。Go で書かれ、起動時間わずか ~10ms。

AI エディタ（[Cursor](https://cursor.com)、[Windsurf](https://codeium.com/windsurf)、[Claude Desktop](https://claude.ai/download)）やターミナルから、メッセージ投稿・履歴取得・スレッド返信が可能です。

🇬🇧 [English README](./README.md)

<!-- TODO: デモ GIF を追加 -->
<!-- ![Demo](./docs/assets/demo.gif) -->

## なぜ slack-fast-mcp？

| | slack-fast-mcp | Node.js MCP | Python MCP |
|---|---|---|---|
| **起動速度** | ~10ms | ~200-500ms | ~300-800ms |
| **インストール** | バイナリ配置のみ | `npm install` | `pip install` |
| **ランタイム** | 不要 | Node.js 必要 | Python 必要 |
| **バイナリサイズ** | ~10MB | N/A | N/A |

MCP Server はリクエストごとにプロセスが起動します。**起動速度がそのまま体感速度に直結します。** slack-fast-mcp は Go のネイティブバイナリ — ランタイム不要、依存なし、とにかく速い。

> ベンチマーク: Apple M1（macOS）での起動時間計測。実測値はハードウェアにより異なります。

## 機能

- **MCP ツール 3種**: `slack_post_message`, `slack_get_history`, `slack_post_thread`
- **CLI モード**: ターミナルから `slack-fast-mcp post`, `history`, `reply` で直接操作
- **セットアップウィザード**: `slack-fast-mcp setup` で対話形式の初期設定
- **プロジェクト別設定**: `.slack-mcp.json` でプロジェクトごとの Slack 設定を管理
- **クロスプラットフォーム**: macOS, Linux, Windows 対応バイナリ
- **セキュリティ**: トークンの環境変数参照、直書き検出・警告

## クイックスタート

### 1. インストール

#### 方法 A: バイナリダウンロード（推奨）

[GitHub Releases](https://github.com/kai-ko/slack-fast-mcp/releases) から最新バイナリをダウンロード:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/kai-ko/slack-fast-mcp/releases/latest/download/slack-fast-mcp_darwin_arm64 -o /usr/local/bin/slack-fast-mcp
chmod +x /usr/local/bin/slack-fast-mcp
```

<details>
<summary>その他のプラットフォーム</summary>

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

> **Windows PATH 設定:** `$env:USERPROFILE\bin` が PATH に含まれていない場合:
> ```powershell
> [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
> ```
> 設定後、PowerShell を再起動してください。

</details>

> **macOS Gatekeeper 警告**: 警告が表示される場合は `xattr -d com.apple.quarantine /usr/local/bin/slack-fast-mcp` を実行してください

#### 方法 B: Go install

```bash
go install github.com/kai-ko/slack-fast-mcp/cmd/slack-fast-mcp@latest
```

#### 方法 C: ソースからビルド

```bash
git clone https://github.com/kai-ko/slack-fast-mcp.git
cd slack-fast-mcp && make build
```

インストールの確認:

```bash
slack-fast-mcp version
```

### 2. Slack App を作成

> スクリーンショット付きの詳細手順は [Slack App セットアップガイド](./docs/slack-app-setup.md) を参照してください。

1. [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From scratch**
2. **OAuth & Permissions** で **Bot Token Scopes** を追加:

   **必須:**
   - `chat:write` — メッセージ投稿
   - `channels:history` — パブリックチャンネルの履歴取得
   - `channels:read` — チャンネル名の解決

   **推奨（任意）:**
   - `users:read` — 履歴でユーザー名を表示（未設定の場合、ユーザーIDのみ表示）
   - `groups:history` — プライベートチャンネルの履歴取得
   - `groups:read` — プライベートチャンネル名の解決

3. ワークスペースにアプリを**インストール**
4. **Bot User OAuth Token**（`xoxb-...`）をコピー

### 3. 設定

セットアップウィザードを実行（推奨）:

```bash
slack-fast-mcp setup
```

または手動で設定:

```bash
# トークンを環境変数に設定
export SLACK_BOT_TOKEN='xoxb-your-token-here'

# プロジェクト設定を作成（任意）
echo '{"token":"${SLACK_BOT_TOKEN}","default_channel":"general"}' > .slack-mcp.json
```

> **注意:** `export` は現在のターミナルセッションのみ有効です。永続化するにはシェルプロファイル（`~/.zshrc`、`~/.bashrc` 等）に追記し、ターミナルを再起動してください。

### 4. AI エディタで使用

#### Cursor / Windsurf

`.cursor/mcp.json`（または `.windsurf/mcp.json`）に追加:

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

Claude Desktop の MCP 設定（Settings → Developer → MCP Servers）に追加:

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

> **補足:** slack-fast-mcp は stdio transport を使用するすべての MCP 対応ツールで動作します。

### 5. Bot をチャンネルに招待

> **この手順は必須です。** Bot は招待されていないチャンネルには投稿・閲覧できません。

Slack で対象チャンネルを開き、以下を入力:

```
/invite @your-bot-name
```

## MCP ツール

### `slack_post_message`

Slack チャンネルにメッセージを投稿します。

```
slack_post_message(channel: "general", message: "Hello World!")
```

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel` | string | No | チャンネル名 or ID（設定ファイルのデフォルト値を使用） |
| `message` | string | Yes | メッセージ本文（Slack mrkdwn 対応） |

### `slack_get_history`

チャンネルのメッセージ履歴を取得します。

```
slack_get_history(channel: "general", limit: 10)
```

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel` | string | No | チャンネル名 or ID |
| `limit` | integer | No | 取得件数（1-100、デフォルト: 10） |
| `oldest` | string | No | 取得開始時刻（Unix timestamp） |
| `latest` | string | No | 取得終了時刻（Unix timestamp） |

### `slack_post_thread`

スレッドに返信を投稿します。

```
slack_post_thread(channel: "general", thread_ts: "1234567890.123456", message: "Reply!")
```

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel` | string | No | チャンネル名 or ID |
| `thread_ts` | string | Yes | 返信先メッセージのタイムスタンプ |
| `message` | string | Yes | 返信メッセージ本文（Slack mrkdwn 対応） |

## CLI の使い方

```bash
# MCP Server として起動（サブコマンド省略時のデフォルト）
slack-fast-mcp serve

# メッセージ投稿
slack-fast-mcp post --channel general --message "Hello from CLI!"

# チャンネル履歴取得
slack-fast-mcp history --channel general --limit 20

# スレッド返信
slack-fast-mcp reply --channel general --thread-ts 1234567890.123456 --message "返信します"

# JSON 形式で出力（jq と連携して整形）
slack-fast-mcp history --channel general --json | jq '.messages[].text'

# バージョン表示
slack-fast-mcp version

# セットアップウィザード
slack-fast-mcp setup
```

<details>
<summary>設定の詳細</summary>

## 設定

### 優先順位（高い順）

1. CLI フラグ（`--token`, `--channel`）
2. 環境変数（`SLACK_BOT_TOKEN`, `SLACK_DEFAULT_CHANNEL`）
3. プロジェクト設定（`.slack-mcp.json`）
4. グローバル設定（`~/.config/slack-fast-mcp/config.json`）

### `.slack-mcp.json`

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general"
}
```

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `token` | string | Yes | Bot トークン。`${ENV_VAR}` で環境変数を参照可能 |
| `default_channel` | string | No | デフォルトチャンネル名 or ID |

### 環境変数

| 変数名 | 説明 |
|---|---|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token |
| `SLACK_DEFAULT_CHANNEL` | デフォルトチャンネル |
| `SLACK_FAST_MCP_LOG_LEVEL` | ログレベル（debug/info/warn/error） |

</details>

## セキュリティ

### トークン管理

- 設定ファイルにトークンを**直書きしない**（Git にコミットされる可能性あり）
- `${SLACK_BOT_TOKEN}` 形式で環境変数を参照する
- トークン直書き（`xoxb-`、`xoxp-`、`xoxs-` で始まる文字列）を**検出し警告**します

### 推奨 `.gitignore` 設定

```gitignore
.slack-mcp.json
# Cursor 設定にトークンを直書きする場合（非推奨）:
# .cursor/mcp.json
```

### このツールが行わないこと

- ローカルにデータを**保存しません**（メッセージ、トークン、認証情報）
- 管理者権限を**持ちません** — メッセージの投稿・閲覧のみ
- Slack との通信はすべて **HTTPS** 経由

### トークンが漏洩した場合

1. [api.slack.com/apps](https://api.slack.com/apps) にアクセス
2. 対象のアプリを選択 → **OAuth & Permissions**
3. **Revoke Token** をクリックして漏洩したトークンを無効化
4. アプリを再インストールして新しいトークンを生成

## トラブルシューティング

| エラー | 原因 | 対処法 |
|---|---|---|
| `not_in_channel` | Bot がチャンネルに未招待 | チャンネルで `/invite @your-bot-name` を実行 |
| `invalid_auth` | トークンが無効または期限切れ | [api.slack.com/apps](https://api.slack.com/apps) で再生成 |
| `channel_not_found` | チャンネル名が間違っている | スペルを確認、`#` プレフィックスは不要 |
| `missing_scope` | OAuth スコープが未追加 | Slack App 設定でスコープを追加し、アプリを再インストール |
| `token_not_configured` | トークンが未設定 | `slack-fast-mcp setup` を実行、または `SLACK_BOT_TOKEN` を設定 |

詳しくは [Slack App セットアップガイド](./docs/slack-app-setup.md) を参照してください。

## ロードマップ

| 機能 | 優先度 | ステータス |
|---|---|---|
| ファイルアップロード | 中 | 計画中 |
| 絵文字リアクション | 低 | 計画中 |
| ユーザー検索・メンション | 低 | 計画中 |
| マルチワークスペース対応 | 低 | 計画中 |
| HTTP transport（リモート MCP） | 低 | 計画中 |

## コントリビュート

コントリビュート大歓迎です！お気軽に Pull Request を送ってください。

- [バグ報告](https://github.com/kai-ko/slack-fast-mcp/issues/new)
- [機能リクエスト](https://github.com/kai-ko/slack-fast-mcp/issues/new)
- ドキュメント改善

ガイドラインは [CONTRIBUTING.md](./CONTRIBUTING.md) を参照してください。

## ソースからビルド

```bash
git clone https://github.com/kai-ko/slack-fast-mcp.git
cd slack-fast-mcp
make build
```

### 開発

```bash
make test          # テスト実行
make test-race     # Race detector 付きテスト
make quality       # 品質ゲート（vet, build, test, coverage, smoke）
make smoke         # スモークテスト
make help          # ヘルプ表示
```

## 謝辞

以下の優れたライブラリを活用しています:

- [mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP SDK
- [slack-go/slack](https://github.com/slack-go/slack) — Go Slack API クライアント
- [cobra](https://github.com/spf13/cobra) — Go CLI フレームワーク

## ライセンス

[MIT License](./LICENSE)
