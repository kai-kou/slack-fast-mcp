# slack-fast-mcp

[![CI](https://github.com/kai-kou/slack-fast-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/kai-kou/slack-fast-mcp/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kai-kou/slack-fast-mcp)](https://github.com/kai-kou/slack-fast-mcp/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/kai-kou/slack-fast-mcp)](https://goreportcard.com/report/github.com/kai-kou/slack-fast-mcp)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kai-kou/slack-fast-mcp)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

**最速の Slack [MCP](https://modelcontextprotocol.io/) サーバー。** Go 製、起動時間わずか ~10ms。ランタイム不要、依存なし — バイナリひとつで完結。

AI エディタ（[Cursor](https://cursor.com)、[Windsurf](https://codeium.com/windsurf)、[Claude Desktop](https://claude.ai/download)）やターミナルから、メッセージ投稿・履歴取得・スレッド返信が可能です。

[English / English README](./README.md)

<!-- TODO: 実際のデモ GIF に差し替え（asciinema + svg-term-cli で録画） -->
<!-- ![Demo](./docs/assets/demo.gif) -->

---

## 目次

- [なぜ slack-fast-mcp？](#なぜ-slack-fast-mcp)
- [何ができる？](#何ができる)
- [クイックスタート](#クイックスタート)
- [MCP ツール](#mcp-ツール)
- [CLI の使い方](#cli-の使い方)
- [設定](#設定)
- [セキュリティ](#セキュリティ)
- [トラブルシューティング](#トラブルシューティング)
- [ロードマップ](#ロードマップ)
- [コントリビュート](#コントリビュート)
- [謝辞](#謝辞)
- [ライセンス](#ライセンス)

---

## なぜ slack-fast-mcp？

MCP サーバーはリクエストごとに**新しいプロセスを起動**します。起動速度がそのまま体感速度に直結します。

| | slack-fast-mcp | Node.js MCP | Python MCP |
|---|---|---|---|
| **起動速度** | ~10 ms | ~200–500 ms | ~300–800 ms |
| **インストール** | バイナリ配置のみ | `npm install` | `pip install` |
| **ランタイム** | 不要 | Node.js 必要 | Python 必要 |
| **バイナリサイズ** | ~10 MB | N/A | N/A |

> ベンチマーク: Apple M1（macOS）での起動時間計測。実測値はハードウェアにより異なります。

### 特徴

- **MCP ツール 3 種** — `slack_post_message`, `slack_get_history`, `slack_post_thread`
- **CLI モード** — ターミナルから `post`, `history`, `reply` で直接操作
- **セットアップウィザード** — `slack-fast-mcp setup` で対話形式の初期設定
- **プロジェクト別設定** — `.slack-mcp.json` でプロジェクトごとの Slack 設定を管理
- **クロスプラットフォーム** — macOS, Linux, Windows 対応バイナリ
- **セキュリティ** — トークンの環境変数参照、直書き検出・警告

---

## 何ができる？

slack-fast-mcp が活躍する実際のユースケース：

- **エディタからデイリースタンドアップ** — コードから離れずに AI に `#daily-standup` への進捗投稿を指示
- **PR 通知** — PR 完了時に AI が Slack へサマリーを自動投稿
- **スレッドベースのコラボレーション** — Cursor や Claude Desktop から直接 Slack スレッドを読み・返信
- **CI/CD ステータス報告** — CLI でビルド結果を Slack チャンネルにパイプ
- **チームログ / 分報** — セッション要約を個人の `#times-*` チャンネルへ自動投稿

---

## クイックスタート

### 前提条件

- アプリを追加できる [Slack ワークスペース](https://slack.com/)
- macOS、Linux、Windows のいずれか

### 1. インストール

#### 方法 A: バイナリダウンロード（推奨）

[GitHub Releases](https://github.com/kai-kou/slack-fast-mcp/releases) から最新バイナリをダウンロード：

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/kai-kou/slack-fast-mcp/releases/latest/download/slack-fast-mcp_Darwin_arm64.tar.gz
tar xzf slack-fast-mcp_Darwin_arm64.tar.gz
sudo mv slack-fast-mcp /usr/local/bin/
```

<details>
<summary>その他のプラットフォーム</summary>

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

> **Windows PATH 設定:** `$env:USERPROFILE\bin` が PATH に含まれていない場合：
> ```powershell
> [Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\bin", "User")
> ```
> 設定後、PowerShell を再起動してください。

</details>

> **macOS Gatekeeper 警告:** 警告が表示される場合は `xattr -d com.apple.quarantine /usr/local/bin/slack-fast-mcp` を実行してください

#### 方法 B: Go install

```bash
go install github.com/kai-kou/slack-fast-mcp/cmd/slack-fast-mcp@latest
```

#### 方法 C: ソースからビルド

```bash
git clone https://github.com/kai-kou/slack-fast-mcp.git
cd slack-fast-mcp && make build
```

インストールの確認：

```bash
slack-fast-mcp version
```

### 2. Slack App を作成

> スクリーンショット付きの詳細手順は [Slack App セットアップガイド](./docs/slack-app-setup.md) を参照してください。

1. [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From scratch**
2. **OAuth & Permissions** で **Bot Token Scopes** を追加：

   | スコープ | 用途 | 必須？ |
   |---|---|---|
   | `chat:write` | メッセージ投稿 | **必須** |
   | `channels:history` | パブリックチャンネルの履歴取得 | **必須** |
   | `channels:read` | チャンネル名の解決 | **必須** |
   | `users:read` | 履歴でユーザー名を表示 | 推奨 |
   | `groups:history` | プライベートチャンネルの履歴取得 | 任意 |
   | `groups:read` | プライベートチャンネル名の解決 | 任意 |

3. ワークスペースにアプリを**インストール**
4. **Bot User OAuth Token**（`xoxb-...`）をコピー

### 3. 設定

もっとも簡単な方法 — セットアップウィザードを実行：

```bash
slack-fast-mcp setup
```

または手動で設定：

```bash
export SLACK_BOT_TOKEN='xoxb-your-token-here'
```

<details>
<summary>ターミナル再起動後もトークンを維持する方法</summary>

`export` コマンドは現在のセッションのみ有効です。セッション終了後も維持し、Cursor などの AI エディタから参照できるようにするには、シェルプロファイルに追記してください：

| シェル | 編集するファイル | 確認方法 |
|---|---|---|
| **zsh**（macOS デフォルト） | `~/.zprofile` または `~/.zshrc` | `echo $SHELL` が `/bin/zsh` |
| **bash** | `~/.bash_profile` または `~/.bashrc` | `echo $SHELL` が `/bin/bash` |

```bash
# 例: ~/.zprofile に追記（macOS + zsh の場合）
echo "export SLACK_BOT_TOKEN='xoxb-your-token-here'" >> ~/.zprofile
source ~/.zprofile
```

OS ごとの詳しい設定手順は [Slack App セットアップガイド §5.1](./docs/slack-app-setup.md#51-方法a-環境変数で設定推奨) を参照してください。

</details>

### 4. AI エディタに追加

#### Cursor / Windsurf

`.cursor/mcp.json`（または `.windsurf/mcp.json`）に追加：

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

Claude Desktop の MCP 設定（Settings → Developer → MCP Servers）に追加：

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

> **注意:** Claude Desktop は `${VAR}` 形式の環境変数展開をサポートしていない場合があります。トークンを直接設定する場合は、この設定ファイルが **Git にコミットされないこと**を確認してください。通常、このファイルはユーザーディレクトリ内に保存されるため問題ありません。

> slack-fast-mcp は stdio transport を使用する**すべての MCP 対応ツール**で動作します。

### 5. Bot をチャンネルに招待

> **この手順は必須です。** Bot は招待されていないチャンネルには投稿・閲覧できません。

Slack で対象チャンネルを開き、以下を入力：

```
/invite @your-bot-name
```

---

## MCP ツール

MCP サーバーとして接続すると、3 つのツールが利用可能です：

### `slack_post_message`

Slack チャンネルにメッセージを投稿します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel` | string | No | チャンネル名 or ID。設定ファイルのデフォルト値を使用 |
| `message` | string | **Yes** | メッセージ本文（[Slack mrkdwn](https://api.slack.com/reference/surfaces/formatting) 対応） |
| `display_name` | string | No | 送信者名（メッセージ末尾に `#名前` ハッシュタグを付与） |

**使用例：**

```
slack_post_message(channel: "general", message: "MCPからこんにちは！")
```

### `slack_get_history`

チャンネルのメッセージ履歴を取得します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel` | string | No | チャンネル名 or ID。設定ファイルのデフォルト値を使用 |
| `limit` | integer | No | 取得件数、1–100（デフォルト: 10） |
| `oldest` | string | No | 取得開始時刻（Unix timestamp） |
| `latest` | string | No | 取得終了時刻（Unix timestamp） |

**使用例：**

```
slack_get_history(channel: "general", limit: 5)
```

### `slack_post_thread`

スレッドに返信を投稿します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel` | string | No | チャンネル名 or ID。設定ファイルのデフォルト値を使用 |
| `thread_ts` | string | **Yes** | 返信先メッセージのタイムスタンプ |
| `message` | string | **Yes** | 返信メッセージ本文（[Slack mrkdwn](https://api.slack.com/reference/surfaces/formatting) 対応） |
| `display_name` | string | No | 送信者名（メッセージ末尾に `#名前` ハッシュタグを付与） |

**使用例：**

```
slack_post_thread(channel: "general", thread_ts: "1234567890.123456", message: "了解！")
```

---

## CLI の使い方

slack-fast-mcp はスタンドアロンの CLI ツールとしても動作します：

```bash
# メッセージ投稿
slack-fast-mcp post --channel general --message "CLIからこんにちは！"

# チャンネル履歴取得
slack-fast-mcp history --channel general --limit 20

# スレッド返信
slack-fast-mcp reply --channel general --thread-ts 1234567890.123456 --message "返信します"

# JSON 形式で出力（jq と連携して整形）
slack-fast-mcp history --channel general --json | jq '.messages[].text'

# MCP サーバーとして起動（サブコマンド省略時のデフォルト）
slack-fast-mcp serve

# バージョン表示
slack-fast-mcp version

# セットアップウィザード
slack-fast-mcp setup
```

---

## 設定

設定は以下の優先順位で解決されます（高い順）：

| 優先度 | ソース | 例 |
|---|---|---|
| 1 | CLI フラグ | `--token`, `--channel` |
| 2 | 環境変数 | `SLACK_BOT_TOKEN`, `SLACK_DEFAULT_CHANNEL` |
| 3 | プロジェクト設定ファイル | `.slack-mcp.json` |
| 4 | グローバル設定ファイル | `~/.config/slack-fast-mcp/config.json` |

### プロジェクト設定: `.slack-mcp.json`

プロジェクトルートに `.slack-mcp.json` を作成してデフォルト値を設定：

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general",
  "display_name": "my-bot"
}
```

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `token` | string | **Yes** | Bot トークン。`${ENV_VAR}` で環境変数を参照可能 |
| `default_channel` | string | No | デフォルトチャンネル名 or ID |
| `display_name` | string | No | デフォルトの送信者名（メッセージ末尾に `#名前` ハッシュタグを付与） |

### 環境変数

| 変数名 | 説明 |
|---|---|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token |
| `SLACK_DEFAULT_CHANNEL` | デフォルトチャンネル名 or ID |
| `SLACK_DISPLAY_NAME` | デフォルトの送信者表示名 |
| `SLACK_FAST_MCP_LOG_LEVEL` | ログレベル: `debug`, `info`, `warn`, `error` |

---

## セキュリティ

### トークン管理

- 設定ファイルにトークンを**直書きしない**（Git にコミットされる可能性あり）
- `${SLACK_BOT_TOKEN}` 形式で環境変数を参照する
- トークン直書き（`xoxb-`、`xoxp-`、`xoxs-` で始まる文字列）を**検出し警告**します

### 推奨 `.gitignore`

```gitignore
.slack-mcp.json
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

---

## トラブルシューティング

| エラー | 原因 | 対処法 |
|---|---|---|
| `not_in_channel` | Bot がチャンネルに未招待 | チャンネルで `/invite @your-bot-name` を実行 |
| `invalid_auth` | トークンが無効または期限切れ | [api.slack.com/apps](https://api.slack.com/apps) で再生成 |
| `channel_not_found` | チャンネル名が間違っている | スペルを確認、`#` プレフィックスは不要 |
| `missing_scope` | OAuth スコープが未追加 | Slack App 設定でスコープを追加し、アプリを再インストール |
| `token_not_configured` | トークンが未設定 | `slack-fast-mcp setup` を実行、または `SLACK_BOT_TOKEN` を設定 |

詳しくは [Slack App セットアップガイド](./docs/slack-app-setup.md) を参照してください。

---

## ロードマップ

| 機能 | 優先度 | ステータス |
|---|---|---|
| ファイルアップロード | 中 | 計画中 |
| 絵文字リアクション | 低 | 計画中 |
| ユーザー検索・メンション | 低 | 計画中 |
| マルチワークスペース対応 | 低 | 計画中 |
| HTTP transport（リモート MCP） | 低 | 計画中 |

---

## コントリビュート

コントリビュート大歓迎です！お気軽に Pull Request を送ってください。

- [バグ報告](https://github.com/kai-kou/slack-fast-mcp/issues/new)
- [機能リクエスト](https://github.com/kai-kou/slack-fast-mcp/issues/new)
- ドキュメント改善

ガイドラインは [CONTRIBUTING.md](./CONTRIBUTING.md) を参照してください。

### 開発

```bash
git clone https://github.com/kai-kou/slack-fast-mcp.git
cd slack-fast-mcp

make build         # バイナリをビルド
make test          # テスト実行
make test-race     # Race detector 付きテスト
make quality       # 品質ゲート（vet, build, test, coverage, smoke）
make smoke         # スモークテスト
make help          # ヘルプ表示
```

---

## 謝辞

以下の優れたライブラリを活用しています：

- [mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP SDK
- [slack-go/slack](https://github.com/slack-go/slack) — Go Slack API クライアント
- [cobra](https://github.com/spf13/cobra) — Go CLI フレームワーク

## ライセンス

[MIT License](./LICENSE)
