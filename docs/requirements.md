# 要件定義（詳細版）

**作成日**: 2026-02-10
**最終更新**: 2026-02-10（レビュー指摘反映）
**ステータス**: 確定

> **対象読者**: 本プロジェクトの開発者・コントリビューター。技術選定の背景から詳細仕様まで、実装に必要な情報をすべて含む。
>
> **ドキュメント構成**:
> - **本ドキュメント（requirements.md）**: 要件定義・仕様の正（Single Source of Truth）
> - **[architecture.md](./architecture.md)**: 技術設計・実装詳細
> - **[slack-app-setup.md](./slack-app-setup.md)**: エンドユーザー向けセットアップ手順

---

## 0. 背景と動機

### なぜ slack-fast-mcp を作るのか

CursorなどのAI搭載エディタからSlackに投稿・確認するニーズは日常的に発生する。既存のSlack MCP実装（Node.js/Python製）はインタプリタ型言語の起動オーバーヘッドがあり、MCPのstdioトランスポートでは毎回プロセスが起動されるため、体感速度に直結する問題があった。

本プロジェクトは「**実行速度を最優先**」という明確な設計方針のもと、Go言語のシングルバイナリとして実装することで、既存ツールでは実現できなかった高速なSlack連携MCP Serverを提供する。

### 既存ツールとの差別化

| 比較項目 | slack-fast-mcp（本ツール） | 既存 Node.js 製 MCP | 既存 Python 製 MCP |
|---------|--------------------------|---------------------|-------------------|
| 起動速度 | ~10ms（ネイティブバイナリ） | ~200-500ms（Node.js起動） | ~300-800ms（Python起動） |
| インストール | バイナリ配置のみ | npm install 必要 | pip install 必要 |
| ランタイム依存 | なし | Node.js 必要 | Python 必要 |
| バイナリサイズ | ~10-15MB | N/A（ソース配布） | N/A（ソース配布） |
| クロスプラットフォーム | ビルド済みバイナリ提供 | 要ランタイム | 要ランタイム |

### OSS公開方針

- **公開判断基準**: Phase 3（品質・配布）完了後、README・LICENSE・CI/CDが整備された段階で公開
- **ライセンス**: MIT License
- **メンテナンス体制**: 個人プロジェクトとして開始。Issue対応・PR受け入れは可能な範囲で対応
- **期待するコミュニティ参加**: バグ報告、機能提案、ドキュメント改善、多言語対応

---

## 1. ユーザー要件（原文）

> ※ 以下はプロジェクト開始時のユーザー要件原文であり、変更不可のトレーサビリティ記録として保持する。

### 機能要件

- 特定のSlackワークスペースの
    - 特定のチャンネルに
        - 投稿できる
        - 投稿内容を確認できる
        - 既存の投稿に対してスレッドで投稿できる
- パラメータで投稿するチャンネルを指定できる

### 非機能要件・品質要件

- Cursorが扱いやすい
- 初期設定がわかりやすい
    - ユーザーが操作すべき手順を案内してくれる
- インストールコストが低い
    - AI Agentが自動でインストールできる
- Mac、Windowsでも実行できる
- **実行速度を最優先にしたい**
- ユーザー(人)も利用することを想定したい

### 展開・公開要件

- いい感じのツールになったらチームメンバーやGitHubでパブリックに公開して広く利用してほしい
- プロジェクト(ワークディレクトリ)ごとにSlackのワークスペース、チャンネルを指定できるようにしたい

---

## 2. 要件分析

### Cursorが扱いやすい → MCP Server

- CursorはMCP（Model Context Protocol）をネイティブサポート
- MCP Serverとして実装すれば、Cursorから直接ツールとして呼び出し可能
- `.cursor/mcp.json` に設定を追加するだけで利用開始

### 実行速度を最優先 → Go言語

- Go言語はコンパイル型言語で高速起動・高速実行
- シングルバイナリにコンパイルされるためランタイム不要
- Node.js/Python等のインタプリタ型言語と比較して起動時間が圧倒的に短い
- MCP Serverは毎回プロセス起動するため、起動速度が直接体感速度に影響する

### インストールコストが低い → シングルバイナリ配布

- GoのシングルバイナリはOS別に配布するだけでインストール完了
- `npm install` や `pip install` 等のパッケージマネージャ不要
- GitHub Releasesからダウンロード→パスに配置→即利用可能
- AI Agentが `curl` + `chmod` で自動インストール可能

### プロジェクトごとのワークスペース・チャンネル指定 → ローカル設定ファイル

- プロジェクトルートに `.slack-mcp.json` 等の設定ファイルを配置
- ワークスペースのトークン、デフォルトチャンネルを指定
- 環境変数でもオーバーライド可能にする
- `.gitignore` に追加してトークン漏洩を防止

### ユーザー(人)も利用 → CLIモード併設

- MCP Serverモードに加え、CLIモードを実装
- `slack-fast-mcp post --channel general --message "Hello"` のように直接実行
- `slack-fast-mcp history --channel general --limit 10` で履歴確認
- `slack-fast-mcp reply --channel general --thread-ts 12345 --message "Reply"` でスレッド投稿

### 公開を視野 → OSS品質のドキュメント・CI/CD

- README.md（英語）に利用方法・セットアップ手順を記載
- GitHub Actions で自動テスト・自動リリース
- GoReleaserによるクロスプラットフォームバイナリ自動ビルド
- ライセンス（MIT予定）

---

## 3. MCP ツール仕様（確定）

### 3.0 MCP ツール description 設計方針

MCP Serverの最大のユーザーはLLM（Cursor/Claude等）であるため、ツールの `description` はLLMがツールを正しく選択・利用できるように設計する。

#### description に含めるべき情報

- **目的と使用タイミング**: ツールが何をするのか、いつ使うべきかを明記
- **デフォルト値の挙動**: パラメータ省略時の動作を含める
- **対応フォーマット**: Slack mrkdwn 記法の対応状況を記載
- **制約事項**: 文字数上限、Botのチャンネル参加要件等を含める
- **前提条件**: Botがチャンネルに招待済みである必要がある旨を含める

#### エラー Hint のLLM最適化方針

- LLMが次のアクションを判断できる具体的な指示を含める
- 「Ask the user to ...」形式で、LLMがそのままユーザーに伝達可能な文面にする
- 英語で記述（LLMの処理精度向上のため）

#### 出力 JSON のエンコーディング

- UTF-8そのまま出力（日本語をエスケープしない）。LLMの可読性を優先

### 3.1 `slack_post_message` - メッセージ投稿

**説明**: 指定したSlackチャンネルにメッセージを投稿する

#### 入力パラメータ

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|------|------|-----------|------|
| `channel` | string | No | 設定ファイルのdefault channel | 投稿先チャンネル名 or チャンネルID |
| `message` | string | Yes | - | 投稿するメッセージ本文（Slack mrkdwn対応） |
| `display_name` | string | No | 設定ファイルの display_name | 送信者表示名。メッセージ末尾に `#名前` ハッシュタグを付与 |

> **チャンネル未指定時の挙動**: `channel` パラメータ省略かつ `default_channel` 未設定の場合は、エラーを返却する（メッセージ: 「Set default_channel in config or specify the channel parameter」）

#### 出力

```json
{
  "ok": true,
  "channel": "C01234ABCDE",
  "channel_name": "general",
  "ts": "1234567890.123456",
  "message": "投稿されたメッセージ本文",
  "permalink": "https://workspace.slack.com/archives/C01234/p1234567890123456"
}
```

#### エラーケース

| エラー | 説明 | 対処 |
|--------|------|------|
| `channel_not_found` | チャンネルが見つからない | チャンネル名/IDを確認 |
| `not_in_channel` | Botがチャンネルに参加していない | Botをチャンネルに招待 |
| `invalid_auth` | トークンが無効 | トークンを再設定 |
| `no_text` | メッセージが空 | messageパラメータを指定 |
| `no_default_channel` | チャンネル未指定かつデフォルト未設定 | channel パラメータを指定するか、default_channel を設定 |
| `rate_limited` | レート制限 | リトライ（Retry-Afterヘッダに従う） |

> エラーコードの詳細（Hint・対象ツール等）は §7 エラーコードマスターテーブルを参照。

---

### 3.2 `slack_get_history` - 投稿履歴取得

**説明**: 指定したSlackチャンネルのメッセージ履歴を取得する

#### 入力パラメータ

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|------|------|-----------|------|
| `channel` | string | No | 設定ファイルのdefault channel | 取得対象チャンネル名 or チャンネルID |
| `limit` | integer | No | 10 | 取得するメッセージ数（1〜100） |
| `oldest` | string | No | - | 取得開始時刻（Unix timestamp） |
| `latest` | string | No | - | 取得終了時刻（Unix timestamp） |

> **チャンネル未指定時の挙動**: `slack_post_message` と同様、エラーを返却する。

#### 出力

```json
{
  "ok": true,
  "channel": "C01234ABCDE",
  "channel_name": "general",
  "messages": [
    {
      "user": "U01234ABCDE",
      "user_name": "john.doe",
      "text": "メッセージ本文",
      "ts": "1234567890.123456",
      "thread_ts": "",
      "reply_count": 3,
      "permalink": "https://..."
    }
  ],
  "has_more": false,
  "count": 10
}
```

#### エラーケース

| エラー | 説明 | 対処 |
|--------|------|------|
| `channel_not_found` | チャンネルが見つからない | チャンネル名/IDを確認 |
| `not_in_channel` | Botがチャンネルに参加していない | Botをチャンネルに招待 |
| `invalid_auth` | トークンが無効 | トークンを再設定 |
| `no_default_channel` | チャンネル未指定かつデフォルト未設定 | channel パラメータを指定するか、default_channel を設定 |
| `rate_limited` | レート制限 | リトライ |

> エラーコードの詳細は §7 エラーコードマスターテーブルを参照。

---

### 3.3 `slack_post_thread` - スレッド返信

**説明**: 既存のメッセージのスレッドに返信を投稿する

#### 入力パラメータ

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|------|------|-----------|------|
| `channel` | string | No | 設定ファイルのdefault channel | チャンネル名 or チャンネルID |
| `thread_ts` | string | Yes | - | 返信先メッセージのタイムスタンプ |
| `message` | string | Yes | - | 返信メッセージ本文（Slack mrkdwn対応） |
| `display_name` | string | No | 設定ファイルの display_name | 送信者表示名。メッセージ末尾に `#名前` ハッシュタグを付与 |

> **チャンネル未指定時の挙動**: `slack_post_message` と同様、エラーを返却する。

#### 出力

```json
{
  "ok": true,
  "channel": "C01234ABCDE",
  "channel_name": "general",
  "ts": "1234567890.654321",
  "thread_ts": "1234567890.123456",
  "message": "返信メッセージ本文",
  "permalink": "https://..."
}
```

#### エラーケース

| エラー | 説明 | 対処 |
|--------|------|------|
| `thread_not_found` | スレッド元が見つからない | thread_tsを確認 |
| `channel_not_found` | チャンネルが見つからない | チャンネル名/IDを確認 |
| `not_in_channel` | Botがチャンネルに参加していない | Botをチャンネルに招待 |
| `invalid_auth` | トークンが無効 | トークンを再設定 |
| `no_default_channel` | チャンネル未指定かつデフォルト未設定 | channel パラメータを指定するか、default_channel を設定 |
| `rate_limited` | レート制限 | リトライ |

> エラーコードの詳細は §7 エラーコードマスターテーブルを参照。

---

## 4. 設定ファイル仕様

### 4.1 設定の優先順位（高い順）

1. **CLI引数** / **MCPツールパラメータ**（都度指定）
2. **環境変数**（`SLACK_BOT_TOKEN`, `SLACK_DEFAULT_CHANNEL`）
3. **プロジェクトローカル設定ファイル**（`.slack-mcp.json`）
4. **グローバル設定ファイル**（`~/.config/slack-fast-mcp/config.json`）

### 4.2 `.slack-mcp.json`（プロジェクトローカル）

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general",
  "display_name": "my-agent"
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|------|------|------|
| `token` | string | Yes | Slack Bot User OAuth Token。`${ENV_VAR}` 形式で環境変数参照可 |
| `default_channel` | string | No | デフォルト投稿チャンネル名 or ID |
| `display_name` | string | No | デフォルトの送信者表示名。メッセージ末尾に `#名前` ハッシュタグを付与 |

### 4.3 グローバル設定ファイル（`~/.config/slack-fast-mcp/config.json`）

```json
{
  "token": "${SLACK_BOT_TOKEN}",
  "default_channel": "general",
  "log_level": "info"
}
```

> **注意**: グローバル設定ファイルでも `${SLACK_BOT_TOKEN}` 形式での環境変数参照を推奨する。

| フィールド | 型 | 必須 | 説明 |
|-----------|------|------|------|
| `token` | string | No | デフォルトのSlack Botトークン |
| `default_channel` | string | No | グローバルデフォルトチャンネル |
| `log_level` | string | No | ログレベル（debug/info/warn/error） |

### 4.4 環境変数

| 変数名 | 説明 | 必須 |
|--------|------|------|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token | Yes（設定ファイルがない場合） |
| `SLACK_DEFAULT_CHANNEL` | デフォルトチャンネル | No |
| `SLACK_DISPLAY_NAME` | デフォルトの送信者表示名 | No |
| `SLACK_FAST_MCP_LOG_LEVEL` | ログレベル | No |

### 4.5 セキュリティ考慮

- `.slack-mcp.json` にトークンを直書きしないことを推奨
- `${SLACK_BOT_TOKEN}` 形式で環境変数参照を推奨
- `.gitignore` への `.slack-mcp.json` 追加を初期設定ガイドで案内
- トークンがファイルに直書きされている場合は警告を表示

#### トークン直書き検出（技術的防御）

設定ファイル読み込み時に、`token` フィールドの値が実トークン形式（`xoxb-`、`xoxp-`、`xoxs-` で始まる文字列）に一致する場合:

1. **stderr に警告メッセージを出力**:
   ```
   WARNING: Token appears to be hardcoded in config file.
   Consider using environment variable reference: "${SLACK_BOT_TOKEN}"
   See: https://github.com/xxx/slack-fast-mcp#security
   ```
2. **動作は継続する**（エラーにはしない。ユーザビリティを優先）
3. **将来的に `--strict` モードでエラー終了も検討**

> **設定ファイルのセキュリティ詳細は [slack-app-setup.md §9](./slack-app-setup.md) を参照。**

---

## 5. CLI コマンド仕様

### 5.1 コマンド体系

```
slack-fast-mcp [command] [flags]

Commands:
  serve     MCP Server モードで起動（デフォルト）
  post      チャンネルにメッセージを投稿
  history   チャンネルの投稿履歴を取得
  reply     スレッドに返信
  setup     初期設定ウィザード
  version   バージョン情報を表示
  help      ヘルプを表示

Global Flags:
  --config    設定ファイルパス（デフォルト: .slack-mcp.json）
  --token     Slack Bot Token（環境変数/設定ファイルよりも優先）
  --channel   チャンネル名 or ID
  --verbose   詳細ログを出力
  --json      JSON形式で出力（CLI利用時）
```

### 5.2 サブコマンド詳細

#### `slack-fast-mcp serve`（デフォルト）
- MCP Server モードで起動
- stdio transport で通信
- 引数なしで実行した場合はこのモードで起動

#### `slack-fast-mcp post`
```
slack-fast-mcp post --message "Hello World" [--channel general]
```

#### `slack-fast-mcp history`
```
slack-fast-mcp history [--channel general] [--limit 10]
```

#### `slack-fast-mcp reply`
```
slack-fast-mcp reply --thread-ts 1234567890.123456 --message "Reply" [--channel general]
```

#### `slack-fast-mcp setup`
```
slack-fast-mcp setup
```

対話形式の初期設定ウィザード。以下のフローで実行する:

1. **Slack App 作成確認**: 「Slack App は作成済みですか？ (y/N)」
   - `N` の場合: Slack App 作成手順（Slack API サイトのURL、必要なスコープ一覧）を表示し、作成後に再実行を案内
   - `y` の場合: 次のステップへ
2. **Bot Token 入力**: 「Bot User OAuth Token を入力してください:」
   - `xoxb-` で始まる形式のバリデーションを実施
   - 不正な形式の場合: エラーメッセージと正しい形式の案内を表示し、再入力を促す
3. **デフォルトチャンネル入力**: 「デフォルトの投稿チャンネルを入力してください（空欄でスキップ）:」
   - 空欄の場合: スキップ（各コマンドで都度指定が必要になる旨を案内）
4. **`.slack-mcp.json` の生成**: トークンは `${SLACK_BOT_TOKEN}` 形式で書き出し、環境変数の設定方法を案内
5. **`.gitignore` 追記確認**: `.gitignore` に `.slack-mcp.json` が未追加の場合、「追加しますか？ (Y/n)」
6. **Cursor MCP 設定の案内**: `.cursor/mcp.json` の設定例を表示

> **CLI の `--json` 未指定時のデフォルト出力**: テーブル形式（視認性重視）。`--json` フラグでJSON出力に切り替え。

---

## 6. Slack API 利用方針

### 6.1 使用する Slack API メソッド

| メソッド | 用途 | 必要スコープ |
|---------|------|-------------|
| `chat.postMessage` | メッセージ投稿 + スレッド返信 | `chat:write` |
| `conversations.history` | チャンネル履歴取得 | `channels:history`, `groups:history` |
| `conversations.list` | チャンネル名→ID変換 | `channels:read`, `groups:read` |
| `users.info` | ユーザー名解決（履歴表示用） | `users:read` |

### 6.2 必要な Bot Token Scopes

| スコープ | 必須 | 用途 |
|---------|------|------|
| `chat:write` | Yes | メッセージ投稿 |
| `channels:history` | Yes | パブリックチャンネル履歴取得 |
| `groups:history` | No | プライベートチャンネル履歴取得 |
| `channels:read` | Yes | チャンネル名→ID変換 |
| `groups:read` | No | プライベートチャンネル名→ID変換 |
| `users:read` | No | ユーザー名解決（推奨） |

### 6.3 レート制限対応

- Slack APIのレート制限に対応（429レスポンス + Retry-Afterヘッダ）
- 指数バックオフによるリトライ（最大3回）
- **注意**: 2025年5月以降、非Marketplaceアプリの`conversations.history`は Tier 1（1req/min、15件/req）に制限変更
  - ただし内部カスタムアプリ（社内利用）は従来のレート制限が維持される
  - 本ツールは内部カスタムアプリとして利用想定のため影響なし

### 6.4 チャンネル名解決

- ユーザーはチャンネル名（例: `general`）またはチャンネルID（例: `C01234ABCDE`）のどちらでも指定可能
- **チャンネルID判定**: 正規表現 `^[CGD][A-Z0-9]{8,}$` にマッチする場合はチャンネルIDとして直接使用
  - `C`: パブリックチャンネル
  - `G`: プライベートチャンネル / グループ
  - `D`: ダイレクトメッセージ
- `#` で始まる場合は `#` を除去してチャンネル名として検索
- それ以外の場合は `conversations.list` でチャンネル名→IDに変換
  - ページネーション: `cursor` パラメータを使用して全ページを取得（1リクエストあたり200チャンネル、最大5ページ=1000チャンネルで打ち切り）
  - 見つからない場合は `channel_not_found` エラーを返却
- 変換結果はプロセス内でキャッシュ（MCP Serverは毎回プロセス起動のため、長期キャッシュは不要）

> **設計の詳細は [architecture.md §4.2](./architecture.md) を参照。**

---

## 7. エラーハンドリング方針

### 7.1 エラーカテゴリ

| カテゴリ | 例 | 対処 |
|---------|------|------|
| 設定エラー | トークン未設定、設定ファイルパースエラー | 具体的な設定方法を案内 |
| 認証エラー | トークン無効、スコープ不足 | 必要なスコープを明示して案内 |
| チャンネルエラー | チャンネル未参加、チャンネル不在 | Bot招待手順を案内 |
| ネットワークエラー | タイムアウト、接続エラー | リトライ + エラー詳細表示 |
| レート制限 | 429レスポンス | 自動リトライ（Retry-After準拠） |
| 入力エラー | 必須パラメータ不足 | 必要なパラメータを明示 |

#### トークン未設定時の挙動（早期失敗）

トークンが未設定（環境変数・設定ファイルのいずれにも存在しない）の場合:
- **MCP Server モード**: 起動を拒否し、具体的なセットアップ手順を stderr に出力して終了
- **CLI モード**: コマンド実行前にエラーを返却し、`slack-fast-mcp setup` の実行を案内

### 7.2 エラーメッセージ方針

- **具体的**: 何が問題で、何をすればいいかを明確に伝える
- **MCP向け**: LLMが次のアクションを判断できる情報を含める
- **CLI向け**: 人間が読みやすい、色付きのフォーマットされた出力

### 7.3 エラーコードマスターテーブル

本テーブルは全MCP ツール共通のエラーコード定義であり、各ツール仕様（§3）からの正（Single Source of Truth）である。

| Code | Message | Hint（LLM向け・英語） | 対象ツール |
|------|---------|----------------------|-----------|
| `channel_not_found` | 指定されたチャンネルが見つかりません | "The channel was not found. Ask the user to verify the channel name or ID. Do not include the '#' prefix." | 全ツール |
| `not_in_channel` | Botがチャンネルに参加していません | "The bot is not a member of this channel. Ask the user to invite the bot by running: /invite @slack-fast-mcp" | 全ツール |
| `invalid_auth` | トークンが無効です | "The Slack token is invalid or expired. Ask the user to regenerate the token at https://api.slack.com/apps" | 全ツール |
| `missing_scope` | 必要なOAuthスコープが不足しています | "Required OAuth scope is missing. Ask the user to add the missing scope in Slack App settings and reinstall the app." | 全ツール |
| `no_text` | メッセージが空です | "The message parameter is required and cannot be empty." | `slack_post_message`, `slack_post_thread` |
| `no_default_channel` | チャンネル未指定かつデフォルト未設定 | "No channel specified and no default_channel configured. Set default_channel in config or specify the channel parameter." | 全ツール |
| `thread_not_found` | スレッド元メッセージが見つかりません | "The thread_ts does not match any existing message. Ask the user to verify the thread timestamp." | `slack_post_thread` |
| `rate_limited` | レート制限に到達しました | "Slack API rate limit reached. The tool will automatically retry. If this persists, wait a moment and try again." | 全ツール |
| `token_not_configured` | トークンが設定されていません | "No Slack token found. Ask the user to run 'slack-fast-mcp setup' or set the SLACK_BOT_TOKEN environment variable." | 全ツール（起動時） |
| `config_parse_error` | 設定ファイルの解析に失敗しました | "Failed to parse config file. Ask the user to verify the JSON syntax in .slack-mcp.json" | 全ツール（起動時） |
| `network_error` | Slack APIへの接続に失敗しました | "Failed to connect to Slack API. Check network connectivity and try again." | 全ツール |

---

## 8. ログ・デバッグ方針

### 8.1 ログレベル

| レベル | 用途 |
|--------|------|
| `error` | エラー発生時のみ |
| `warn` | 非推奨設定の使用、リトライ発生等 |
| `info` | API呼び出し結果のサマリー |
| `debug` | API リクエスト/レスポンス詳細 |

### 8.2 MCP Server モードでのログ

- **stdout**: MCP プロトコル通信専用（ログを混ぜない）
- **stderr**: ログ出力先（MCP仕様に準拠）
- デフォルトログレベル: `warn`

### 8.3 CLI モードでのログ

- **stdout**: コマンド結果出力
- **stderr**: ログ・エラー出力
- デフォルトログレベル: `info`
- `--verbose` フラグで `debug` レベルに切り替え

---

## 9. テスト戦略

### 9.1 テスト方針

| テスト種別 | 対象 | 方法 |
|-----------|------|------|
| ユニットテスト | 設定読み込み、チャンネル名解決ロジック等 | Go標準テスト |
| インテグレーションテスト | Slack API呼び出し | モック（httptest） |
| MCPプロトコルテスト | MCP Server として正しく動作するか | mcp-go のテストユーティリティ |
| E2Eテスト | 実際のSlackワークスペースでの動作 | 手動 or サンドボックスワークスペース |

### 9.2 モック方針

- Slack APIクライアントをインターフェース化してモック可能にする
- `httptest` パッケージでHTTPレベルのモックも活用
- CIではモックテストのみ実行（実Slack API呼び出しなし）

---

## 10. 技術調査結果サマリー

### 10.1 mcp-go（github.com/mark3labs/mcp-go）

| 項目 | 内容 |
|------|------|
| 最新バージョン | v0.43.2（2025-11-28） |
| ライセンス | MIT |
| 利用プロジェクト数 | 1,020+ |
| サポートトランスポート | stdio, HTTP (StreamableHTTP) |
| 主要機能 | Tool定義（型付きパラメータ）、Struct-based Schema、Panic Recovery |
| 本プロジェクトでの利用 | stdio transport で MCP Server 実装 |

### 10.2 slack-go/slack（github.com/slack-go/slack）

| 項目 | 内容 |
|------|------|
| 最新バージョン | v0.17.3（2025-07-04） |
| ライセンス | BSD-2-Clause |
| GitHub Stars | 4,900+ |
| 注意事項 | v1未到達、マイナーバージョンで破壊的変更の可能性あり |
| 本プロジェクトでの利用 | chat.postMessage, conversations.history, conversations.list |

---

## 11. 開発フェーズ

### Phase 1: MVP（個人利用）

- Config Layer（設定読み込み + 環境変数展開）
- Slack Client（PostMessage + ResolveChannel）
- MCP Server（`slack_post_message` ツールのみ）
- 最小限のエラーハンドリング（トークン未設定、チャンネル未参加）
- **完了基準**: Cursor から指定チャンネルにメッセージを投稿できる

### Phase 2: コア機能完成

- Slack Client（GetHistory + PostThread）
- MCP Server（`slack_get_history` + `slack_post_thread`）
- CLI Layer（post / history / reply サブコマンド）
- レート制限リトライ（指数バックオフ）
- チャンネル名解決（conversations.list + キャッシュ）
- **完了基準**: 3つのMCPツール + 3つのCLIコマンドが正常動作する

### Phase 3: 品質・配布

- setup コマンド（初期設定ウィザード）
- CI/CD（GitHub Actions + GoReleaser）
- テスト（ユニットテスト + インテグレーションテスト）
- ドキュメント整備（README.md 英語版 + README_ja.md）
- **完了基準**: GitHub Releases でクロスプラットフォームバイナリを配布可能

### Phase 4: OSS公開

- CONTRIBUTING.md 作成
- LICENSE（MIT）追加
- セキュリティ監査（トークン保護の検証）
- 競合との差別化ポイントをREADMEに明記
- **完了基準**: GitHub で public リポジトリとして公開
