# Slack App セットアップガイド

**作成日**: 2026-02-10
**最終更新**: 2026-02-10（レビュー指摘反映）

> **対象読者**: slack-fast-mcp のエンドユーザー（開発者・非開発者）。Slack Appの作成からツールの設定まで、初回セットアップに必要な手順を案内する。
>
> **前提**: Slackワークスペースの管理者権限、またはSlack App作成権限が必要。権限がない場合はワークスペース管理者に依頼してください。
>
> **関連ドキュメント**: 設定仕様の詳細は [requirements.md §4](./requirements.md)、技術設計は [architecture.md](./architecture.md) を参照。

---

## 概要

slack-fast-mcp を利用するには、Slack App を作成し、Bot User OAuth Token を取得する必要があります。このガイドでは、Slack App の作成からトークン取得、slack-fast-mcp への設定までの手順を説明します。

---

## 1. Slack App の作成

### 1.1 Slack API サイトにアクセス

1. [https://api.slack.com/apps](https://api.slack.com/apps) にアクセス
2. Slackワークスペースにログインしていることを確認
3. **「Create New App」** をクリック

### 1.2 アプリの作成方法を選択

1. **「From scratch」** を選択
2. **App Name**: `slack-fast-mcp`（任意の名前）
3. **Pick a workspace**: 利用するワークスペースを選択
4. **「Create App」** をクリック

---

## 2. Bot Token Scopes の設定

### 2.1 OAuth & Permissions ページへ移動

1. 左サイドバーの **「OAuth & Permissions」** をクリック

### 2.2 Scopes の追加

**「Scopes」** セクションの **「Bot Token Scopes」** で以下を追加：

> スコープの詳細な用途は [requirements.md §6.2](./requirements.md) を参照。

#### 必須スコープ

| スコープ | 用途 | 追加手順 |
|---------|------|---------|
| `chat:write` | メッセージの投稿 | 「Add an OAuth Scope」→ `chat:write` を検索して追加 |
| `channels:history` | パブリックチャンネルの履歴取得 | 同上 |
| `channels:read` | チャンネル一覧の取得（名前→ID変換） | 同上 |

#### 推奨スコープ（オプション）

| スコープ | 用途 | 備考 |
|---------|------|------|
| `groups:history` | プライベートチャンネルの履歴取得 | プライベートチャンネルを利用する場合 |
| `groups:read` | プライベートチャンネル一覧の取得 | プライベートチャンネルを利用する場合 |
| `users:read` | ユーザー名の解決 | 履歴取得時にユーザー名を表示する場合 |

---

## 3. アプリのインストール

### 3.1 ワークスペースにインストール

1. **「OAuth & Permissions」** ページの上部にある **「Install to Workspace」** をクリック
2. 権限の確認画面で **「許可する」** をクリック
3. **Bot User OAuth Token** が表示される（`xoxb-` で始まる文字列）

### 3.2 トークンの保存

表示されたトークン（`xoxb-xxxx-xxxx-xxxx`）を安全に保存してください。

> **重要**: このトークンは機密情報です。Gitリポジトリにコミットしないでください。

---

## 4. Bot のチャンネルへの招待

slack-fast-mcp で投稿・閲覧するチャンネルに Bot を招待する必要があります。

### 4.1 チャンネルへの招待方法

1. Slack で対象チャンネルを開く
2. チャンネル名をクリック → **「インテグレーション」** タブ
3. **「アプリを追加する」** をクリック
4. 作成したアプリ（`slack-fast-mcp`）を検索して追加

**または**、チャンネル内で以下のコマンドを入力：

```
/invite @slack-fast-mcp
```

---

## 5. slack-fast-mcp の設定

### 5.1 方法A: 環境変数で設定（推奨）

```bash
export SLACK_BOT_TOKEN="<ここに §3.2 で取得した Bot User OAuth Token を設定>"
export SLACK_DEFAULT_CHANNEL="general"
```

### 5.2 方法B: プロジェクトローカル設定ファイル

プロジェクトルートに `.slack-mcp.json` を作成：

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general"
}
```

> **注意**: `token` フィールドには `${SLACK_BOT_TOKEN}` のように環境変数参照を使うことを強く推奨します。トークンを直書きしないでください。

`.gitignore` に追加：

```
.slack-mcp.json
```

### 5.3 方法C: グローバル設定ファイル

```bash
mkdir -p ~/.config/slack-fast-mcp
```

`~/.config/slack-fast-mcp/config.json` を作成：

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general"
}
```

> **注意**: グローバル設定ファイルでも `${SLACK_BOT_TOKEN}` 形式での環境変数参照を推奨します。環境変数 `SLACK_BOT_TOKEN` にトークンを設定してください。

---

## 6. Cursor MCP 設定

### 6.1 プロジェクトローカル設定（推奨）

`.cursor/mcp.json` を作成：

**方法A: 環境変数を参照（推奨 - Gitコミット可能）**

あらかじめシェルの環境変数に `SLACK_BOT_TOKEN` を設定した上で:

```json
{
  "mcpServers": {
    "slack-fast-mcp": {
      "command": "/path/to/slack-fast-mcp",
      "args": [],
      "env": {
        "SLACK_BOT_TOKEN": "${SLACK_BOT_TOKEN}"
      }
    }
  }
}
```

**方法B: トークン直書き（非推奨 - テスト用途のみ）**

> **注意**: この方法を使う場合、`.cursor/mcp.json` を `.gitignore` に追加してください。トークンをGitにコミットしないでください。

```json
{
  "mcpServers": {
    "slack-fast-mcp": {
      "command": "/path/to/slack-fast-mcp",
      "args": [],
      "env": {
        "SLACK_BOT_TOKEN": "<ここに §3.2 で取得した Bot User OAuth Token を設定>"
      }
    }
  }
}
```

### 6.2 グローバル設定

`~/.cursor/mcp.json` に追加：

```json
{
  "mcpServers": {
    "slack-fast-mcp": {
      "command": "slack-fast-mcp",
      "args": [],
      "env": {
        "SLACK_BOT_TOKEN": "${SLACK_BOT_TOKEN}"
      }
    }
  }
}
```

> **注意**: バイナリにパスが通っていない場合はフルパスを指定してください。環境変数 `SLACK_BOT_TOKEN` にトークンを設定する必要があります（[§5.1](#51-方法a-環境変数で設定推奨) 参照）。

---

## 7. 動作確認

### 7.1 CLI で確認

```bash
# メッセージ投稿
slack-fast-mcp post --channel general --message "Hello from slack-fast-mcp!"

# 履歴取得
slack-fast-mcp history --channel general --limit 5
```

### 7.2 Cursor から確認

Cursor のチャットで以下のように依頼：

```
slack-fast-mcp ツールを使って、#general チャンネルに「テスト投稿です」と投稿してください
```

---

## 8. トラブルシューティング

### 8.1 `invalid_auth` エラー

- **原因**: トークンが無効または期限切れ
- **対処**: Slack API サイトでトークンを再生成

### 8.2 `not_in_channel` エラー

- **原因**: Bot がチャンネルに招待されていない
- **対処**: 上記「4. Bot のチャンネルへの招待」を実施

### 8.3 `missing_scope` エラー

- **原因**: 必要な OAuth スコープが不足
- **対処**: Slack API サイトの「OAuth & Permissions」で不足スコープを追加し、アプリを再インストール

### 8.4 `channel_not_found` エラー

- **原因**: チャンネル名が間違っている、またはアーカイブされたチャンネル
- **対処**: チャンネル名を確認。`#` プレフィックスは不要

### 8.5 `rate_limited` エラー

- **原因**: Slack API のレート制限に到達
- **対処**: slack-fast-mcp が自動リトライするので、しばらく待ってから再実行

---

## 9. セキュリティベストプラクティス

1. **トークンをGitにコミットしない**: `.slack-mcp.json` を `.gitignore` に追加
2. **環境変数を利用する**: トークンは環境変数で管理
3. **必要最小限のスコープ**: 使わないスコープは追加しない
4. **トークンのローテーション**: 定期的にトークンを再生成
5. **チャンネルアクセス制御**: Bot に必要なチャンネルのみアクセス権を付与
