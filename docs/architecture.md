# 技術設計書（Architecture）

**作成日**: 2026-02-10
**最終更新**: 2026-02-10（レビュー指摘反映）
**ステータス**: 確定

> **対象読者**: 本プロジェクトの開発者・コントリビューター。実装に必要な技術設計の詳細を記載する。
>
> **前提ドキュメント**: [requirements.md](./requirements.md)（要件定義・仕様の正）を先に読むことを推奨。

---

## 1. 技術スタック

| 要素 | 選定 | バージョン | 理由 |
|------|------|-----------|------|
| 言語 | Go | 1.23+ | 高速起動・シングルバイナリ・クロスプラットフォーム |
| MCP SDK | mcp-go | v0.43.2+ | Go製 MCP SDK。stdio transport サポート |
| Slack API | slack-go/slack | v0.17.3+ | Go製 Slack API クライアント。REST API サポート |
| CLI フレームワーク | cobra | v1.8+ | Go標準のCLIフレームワーク。サブコマンド対応 |
| 設定管理 | 独自実装（JSON） | - | 軽量・依存最小化。viper不要 |
| ビルド・配布 | GoReleaser + GitHub Actions | - | クロスコンパイル・自動リリース |
| テスト | Go標準 + httptest | - | 外部依存なし |

---

## 2. プロジェクト構成

```
slack-fast-mcp/
├── cmd/
│   └── slack-fast-mcp/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── config/
│   │   ├── config.go            # 設定読み込み・マージロジック
│   │   └── config_test.go
│   ├── slack/
│   │   ├── client.go            # Slack APIクライアント（インターフェース + 実装）
│   │   ├── client_test.go
│   │   ├── channel.go           # チャンネル名解決ロジック
│   │   └── channel_test.go
│   ├── mcp/
│   │   ├── server.go            # MCP Server 定義・ツール登録
│   │   ├── tools.go             # MCP ツールハンドラー実装
│   │   └── tools_test.go
│   └── cli/
│       ├── root.go              # CLI ルートコマンド
│       ├── post.go              # post サブコマンド
│       ├── history.go           # history サブコマンド
│       ├── reply.go             # reply サブコマンド
│       ├── serve.go             # serve サブコマンド（MCPモード）
│       └── setup.go             # setup サブコマンド（初期設定ウィザード）
├── docs/
│   ├── requirements.md          # 要件定義
│   ├── architecture.md          # 本ドキュメント
│   └── slack-app-setup.md       # Slack App セットアップガイド
├── .github/
│   └── workflows/
│       ├── ci.yml               # CI（テスト・lint）
│       └── release.yml          # リリース（GoReleaser）
├── .goreleaser.yml              # GoReleaser 設定
├── go.mod
├── go.sum
├── LICENSE                      # MIT License
├── README.md                    # 公開用 README
├── README_ja.md                 # 日本語 README
└── .gitignore
```

---

## 3. アーキテクチャ概要

### 3.1 レイヤー構成

```
┌─────────────────────────────────────────────┐
│           エントリーポイント (main.go)          │
│  引数なし → MCP Server / サブコマンド → CLI    │
└──────────────┬──────────────┬───────────────┘
               │              │
    ┌──────────▼──────┐  ┌───▼───────────┐
    │   MCP Server    │  │   CLI Layer   │
    │ (internal/mcp)  │  │ (internal/cli)│
    │                 │  │               │
    │ - Tool定義      │  │ - cobra       │
    │ - Handler       │  │ - サブコマンド  │
    │ - stdio通信     │  │ - 出力整形     │
    └────────┬────────┘  └───────┬───────┘
             │                   │
             └─────────┬─────────┘
                       │
              ┌────────▼────────┐
              │   Slack Client  │
              │ (internal/slack)│
              │                 │
              │ - PostMessage   │
              │ - GetHistory    │
              │ - ResolveChannel│
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │  Config Layer   │
              │(internal/config)│
              │                 │
              │ - ファイル読込    │
              │ - 環境変数       │
              │ - マージ         │
              └─────────────────┘
```

### 3.2 データフロー

#### MCP Server モード
```
Cursor → stdio → MCP Server → Tool Handler → Slack Client → Slack API
                                                              ↓
Cursor ← stdio ← MCP Server ← Tool Handler ← Slack Client ← Response
```

#### CLI モード
```
User → CLI → cobra → Command Handler → Slack Client → Slack API
                                                        ↓
User ← stdout ← CLI ← Command Handler ← Slack Client ← Response
```

---

## 4. コンポーネント詳細設計

### 4.1 Config Layer（`internal/config`）

#### 設定構造体

```go
type Config struct {
    Token          string `json:"token"`
    DefaultChannel string `json:"default_channel"`
    LogLevel       string `json:"log_level"`
}
```

#### 設定読み込み順序

1. グローバル設定ファイル（`~/.config/slack-fast-mcp/config.json`）を読み込み
2. プロジェクトローカル設定（`.slack-mcp.json`）で上書き
3. 環境変数で上書き（`SLACK_BOT_TOKEN`, `SLACK_DEFAULT_CHANNEL`）
4. CLI引数 / MCPパラメータで上書き

#### 環境変数展開

- `token` フィールドが `${SLACK_BOT_TOKEN}` 形式の場合、環境変数の値に展開
- 正規表現: `\$\{([A-Z_][A-Z0-9_]*)\}`

---

### 4.2 Slack Client（`internal/slack`）

#### インターフェース設計

```go
// SlackClient はSlack API操作のインターフェース
type SlackClient interface {
    // PostMessage はチャンネルにメッセージを投稿する
    PostMessage(ctx context.Context, channel, message string) (*PostResult, error)

    // PostThread はスレッドに返信を投稿する
    PostThread(ctx context.Context, channel, threadTS, message string) (*PostResult, error)

    // GetHistory はチャンネルの投稿履歴を取得する
    GetHistory(ctx context.Context, channel string, opts HistoryOptions) (*HistoryResult, error)

    // ResolveChannel はチャンネル名をチャンネルIDに解決する
    ResolveChannel(ctx context.Context, channel string) (string, error)
}
```

#### 実装クラス

```go
// Client はSlack APIクライアントの実装
type Client struct {
    api          *slack.Client
    channelCache map[string]string  // チャンネル名 → ID キャッシュ
}
```

#### context.Context 利用方針

| モード | タイムアウト | 説明 |
|--------|------------|------|
| MCP Server モード | リクエスト全体: 30秒 | 1つのMCPツール呼び出しに対する全体タイムアウト |
| CLI モード | コマンド全体: 10秒 | 1つのCLIコマンドに対する全体タイムアウト |
| Slack API 個別呼び出し | 10秒 | 各 Slack API 呼び出しに対するタイムアウト |

- `signal.NotifyContext` を使用し、SIGINT/SIGTERM 受信時に context cancel を伝播
- リトライ中もキャンセルシグナルを確認し、受信時は即座に中断

#### チャンネル名解決ロジック

```
入力: channel string
  ├── 正規表現 ^[CGD][A-Z0-9]{8,}$ にマッチ → チャンネルIDとしてそのまま返す
  │   ├── C: パブリックチャンネル
  │   ├── G: プライベートチャンネル / グループ
  │   └── D: ダイレクトメッセージ
  ├── "#" で始まる → "#" を除去してチャンネル名として検索
  └── その他 → チャンネル名として conversations.list で検索
```

> **注意**: `ci-notifications` のように `C` で始まるチャンネル名が誤判定されないよう、正規表現で大文字英数字のみのパターンに限定する。

#### conversations.list ページネーション

チャンネル名→ID変換時の `conversations.list` API 呼び出し:

- **1リクエストあたり**: `limit=200`
- **ページネーション**: `cursor` パラメータを使用して次のページを取得
- **打ち切り条件**: 最大5ページ（1000チャンネル）でループを終了
- **見つからない場合**: `channel_not_found` エラーを返却
- **キャッシュ**: 変換結果はプロセス内のインメモリキャッシュに保存（MCP Serverは毎回起動のため長期キャッシュ不要）

---

### 4.3 MCP Server（`internal/mcp`）

#### サーバー初期化

```go
func NewServer(cfg *config.Config) *server.MCPServer {
    s := server.NewMCPServer(
        "slack-fast-mcp",
        version.Version,
        server.WithToolCapabilities(false),
    )

    slackClient := slack.NewClient(cfg.Token)

    // ツール登録
    s.AddTool(postMessageTool(), postMessageHandler(slackClient, cfg))
    s.AddTool(getHistoryTool(), getHistoryHandler(slackClient, cfg))
    s.AddTool(postThreadTool(), postThreadHandler(slackClient, cfg))

    return s
}
```

#### ツール定義例（slack_post_message）

```go
func postMessageTool() mcp.Tool {
    return mcp.NewTool("slack_post_message",
        mcp.WithDescription("Post a message to a Slack channel. "+
            "Supports Slack mrkdwn formatting (bold, italic, links, code blocks). "+
            "If channel is omitted, posts to the configured default channel. "+
            "The bot must be invited to the target channel first."),
        mcp.WithString("channel",
            mcp.Description("Channel name (e.g. 'general') or channel ID (e.g. 'C01234ABCDE'). "+
                "If omitted, uses the configured default channel."),
        ),
        mcp.WithString("message",
            mcp.Required(),
            mcp.Description("Message text to post. Supports Slack mrkdwn: "+
                "*bold*, _italic_, `code`, ```code block```, <url|text>."),
        ),
    )
}
```

#### ハンドラー設計方針

- ハンドラーはクロージャで `SlackClient` と `Config` を注入
- チャンネル未指定時は `Config.DefaultChannel` を使用
- チャンネル未指定かつデフォルト未設定の場合は `no_default_channel` エラーを返却
- エラーは `mcp.NewToolResultError()` で返す（Go error は返さない）
- 結果は JSON テキストとして返す（UTF-8そのまま、日本語をエスケープしない）
- エラーの Hint は英語で記述し、LLMがそのまま伝達可能な文面にする（詳細は [requirements.md §7.3](./requirements.md) 参照）

---

### 4.4 CLI Layer（`internal/cli`）

#### コマンド構造

```go
// rootCmd は slack-fast-mcp のルートコマンド
var rootCmd = &cobra.Command{
    Use:   "slack-fast-mcp",
    Short: "Fast Slack MCP Server & CLI",
    // 引数なしの場合は serve（MCP Server モード）
    RunE: serveMCP,
}
```

#### サブコマンド設計

- 各サブコマンドは `internal/slack.SlackClient` を利用
- 出力は `--json` フラグで JSON / テキスト を切り替え
- エラー出力はカラー付きで分かりやすく

---

## 5. 起動フロー

### 5.1 MCP Server モード（引数なし or `serve`）

```
1. main() 開始
2. 設定読み込み（Config Layer）
3. トークン検証（空チェック + 形式チェック）
4. Slack Client 初期化
5. MCP Server 初期化 + ツール登録
6. server.ServeStdio(s) で stdio 待ち受け開始
7. Cursor/Claude からのリクエストを処理
```

### 5.2 CLI モード（サブコマンド指定）

```
1. main() 開始
2. cobra がサブコマンドをパース
3. 設定読み込み（Config Layer）
4. トークン検証
5. Slack Client 初期化
6. サブコマンド実行
7. 結果を stdout に出力
```

### 5.3 Graceful Shutdown

MCP Server モードの終了処理:

```
1. SIGTERM または SIGINT を受信
2. signal.NotifyContext により context cancel を発火
3. 進行中の Slack API 呼び出しの完了を待機（最大5秒）
4. タイムアウト後、強制的にプロセス終了
```

```go
// Graceful shutdown の実装方針
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

// ctx を全ての Slack API 呼び出しに伝播
// ctx.Done() 時に進行中のリクエストが完了するまで最大5秒待機
```

CLI モードでは、コマンド実行中に SIGINT を受信した場合、context cancel によりSlack API呼び出しを中断し、即座に終了する。

---

## 6. エラーハンドリング設計

### 6.1 エラー型

```go
// AppError はアプリケーションエラー
type AppError struct {
    Code    string // エラーコード（channel_not_found 等）
    Message string // 人間向けメッセージ
    Hint    string // 解決のヒント
    Err     error  // 元のエラー
}
```

### 6.2 MCP Server でのエラー返却

```go
// MCP ツールエラーは Go error ではなく CallToolResult で返す
func handleError(appErr *AppError) (*mcp.CallToolResult, error) {
    errMsg := fmt.Sprintf("Error: %s\n%s\nHint: %s", appErr.Code, appErr.Message, appErr.Hint)
    return mcp.NewToolResultError(errMsg), nil
}
```

### 6.3 レート制限リトライ

```go
// 指数バックオフリトライ（最大3回）
// Retry-After ヘッダがあればそれに従う
// なければ 1s → 2s → 4s でリトライ
```

### 6.4 ログのトークンマスキング

debug レベルのログ出力時に、機密情報が露出しないようマスキングを実施する。

#### マスキング対象

| パターン | マスキング後 | 例 |
|---------|------------|---|
| `xoxb-` で始まる文字列 | `xoxb-****` | Slack Bot Token |
| `xoxp-` で始まる文字列 | `xoxp-****` | Slack User Token |
| `xoxs-` で始まる文字列 | `xoxs-****` | Slack Session Token |
| HTTP `Authorization` ヘッダ | `Bearer ****` | APIリクエストヘッダ |

#### 実装方針

- ログ出力前にマスキング関数を適用
- 正規表現: `(xox[bps]-)[A-Za-z0-9-]+` → `$1****`
- HTTP リクエスト/レスポンスのログ出力時にヘッダのマスキングも実施

---

## 7. ビルド・配布設計

### 7.1 GoReleaser 設定

- **対象プラットフォーム**: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- **バイナリ名**: `slack-fast-mcp`
- **バージョン埋め込み**: `ldflags` で `-X main.version={{.Version}}`
- **チェックサム**: SHA256

### 7.2 GitHub Actions CI/CD

#### CI（プルリクエスト時）
- Go lint（golangci-lint）
- Go test
- Go build（コンパイル確認）

#### Release（タグ push 時）
- GoReleaser による自動ビルド・リリース
- GitHub Releases にバイナリアップロード

### 7.3 Cursor MCP 設定例

> 設定の詳細は [slack-app-setup.md §6](./slack-app-setup.md) を参照。

**方法A: 環境変数で設定（推奨）**

あらかじめ環境変数 `SLACK_BOT_TOKEN` を設定した上で:

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

**方法B: 直接指定（非推奨・テスト用途のみ）**

```json
{
  "mcpServers": {
    "slack-fast-mcp": {
      "command": "/path/to/slack-fast-mcp",
      "args": [],
      "env": {
        "SLACK_BOT_TOKEN": "<ここに Bot User OAuth Token を設定>"
      }
    }
  }
}
```

> **注意**: `.cursor/mcp.json` をGitにコミットする場合、トークン直書きは絶対に避けてください。方法Aの環境変数参照を使用してください。直書きする場合は `.cursor/mcp.json` を `.gitignore` に追加してください。

### 7.4 プラットフォーム別の配布・インストール

#### macOS

- **Gatekeeper 警告**: GitHub Releasesからダウンロードしたバイナリは Gatekeeper の警告が表示される場合がある
  - 対処法: `xattr -d com.apple.quarantine /path/to/slack-fast-mcp`
  - READMEに対処手順を記載する
- **Homebrew**: 将来的に Homebrew tap の作成を検討（`brew install xxx/tap/slack-fast-mcp`）
- **コードサイニング**: Apple Developer Program への登録は OSS 公開の初期段階では不要。利用者が増えた段階で検討

#### Windows

- **設定ファイルパス**: `os.UserConfigDir()` を使用してクロスプラットフォーム対応
  - macOS/Linux: `~/.config/slack-fast-mcp/config.json`
  - Windows: `%APPDATA%\slack-fast-mcp\config.json`
- **インストール手順**: PowerShell でのダウンロード・配置手順をREADMEに記載
  ```powershell
  # Windows PowerShell でのインストール例
  Invoke-WebRequest -Uri "https://github.com/xxx/slack-fast-mcp/releases/latest/download/slack-fast-mcp_windows_amd64.exe" -OutFile "$env:USERPROFILE\bin\slack-fast-mcp.exe"
  ```
- **PATH設定**: `$env:USERPROFILE\bin` をPATHに追加する手順を案内

#### Linux

- バイナリダウンロード + `chmod +x` の標準的な手順
- 将来的に `.deb` / `.rpm` パッケージ対応を検討

---

## 8. 将来の拡張ポイント

| 拡張 | 優先度 | 概要 |
|------|--------|------|
| ファイルアップロード | 中 | `files.upload` API を利用 |
| リアクション追加 | 低 | `reactions.add` API を利用 |
| ユーザー検索 | 低 | `users.list` でメンション用 |
| マルチワークスペース | 低 | 複数ワークスペースの切り替え |
| HTTP transport | 低 | リモートからの MCP 接続対応 |

---

## 9. 設計判断記録（ADR）

### ADR-001: CLI フレームワークに cobra を採用

- **背景**: CLIモードのサブコマンド対応が必要
- **選択肢**: cobra / urfave/cli / 標準 flag
- **決定**: cobra を採用
- **理由**: Go CLI のデファクトスタンダード、サブコマンド・フラグ・ヘルプ生成が充実
- **受容したトレードオフ**: 標準 flag に比べてバイナリサイズが約2MB増加する。しかしサブコマンド・ヘルプ生成・フラグバインディングの実装コスト削減のメリットが上回ると判断

### ADR-002: 設定管理に viper ではなく独自 JSON パーサーを採用

- **背景**: 設定ファイルの読み込みと環境変数展開が必要
- **選択肢**: viper / 独自実装
- **決定**: 独自実装（encoding/json + os.ExpandEnv）
- **理由**: viper は依存が大きく、本プロジェクトの設定は単純なJSON。バイナリサイズと起動速度を最優先
- **受容したトレードオフ**: viper の充実した機能（YAML/TOML対応、ホットリロード、環境変数自動バインド、リモート設定取得等）を放棄する。本プロジェクトでは JSON のみ・単純な設定構造のため、これらの機能は不要と判断

### ADR-003: Slack API クライアントをインターフェース化

- **背景**: テスタビリティの確保
- **決定**: `SlackClient` インターフェースを定義し、実装を注入
- **理由**: ユニットテスト時にモック注入可能にする。httptest でのインテグレーションテストも可能
- **受容したトレードオフ**: インターフェース定義と実装の二重管理が発生する。ただし、メソッド数が4つと少ないため管理コストは最小限。テスタビリティの確保による品質向上のメリットが上回る
