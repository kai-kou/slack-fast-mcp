# テスト戦略

**作成日**: 2026-02-10
**最終更新**: 2026-02-10
**ステータス**: 確定

> **対象読者**: 本プロジェクトの開発者・コントリビューター。テスト方針・実装ガイドラインを記載する。
>
> **前提ドキュメント**: [requirements.md](./requirements.md)、[architecture.md](./architecture.md) を先に読むことを推奨。

---

## 1. テスト方針

### 1.1 基本原則

- **外部依存なし**: Go 標準テスト（`testing` + `httptest`）のみ使用。サードパーティテストフレームワーク（testify 等）は導入しない
- **高速実行**: 全テストが 10 秒以内に完了することを目標とする
- **CI 完結**: CI 上ではモックテストのみ実行し、実 Slack API 呼び出しは行わない
- **テスタビリティ**: `SlackClient` インターフェース（ADR-003）により、モック注入でユニットテスト可能

### 1.2 カバレッジ目標

| レイヤー | 目標カバレッジ | 理由 |
|---------|-------------|------|
| `internal/config` | 90%+ | 設定読み込みはバグの温床になりやすい |
| `internal/slack` | 80%+ | API 呼び出しのモックテスト中心 |
| `internal/mcp` | 80%+ | ツールハンドラーのロジック検証 |
| `internal/cli` | 60%+ | CLI 統合はE2E に近い。主要パスを網羅 |
| **全体** | **75%+** | OSS 公開基準として十分な水準 |

> **注意**: カバレッジ数値は指標であり、絶対目標ではない。重要なのはクリティカルパス（設定読み込み、エラーハンドリング、チャンネル名解決）が確実にテストされていること。

---

## 2. テスト種別

### 2.1 テスト種別一覧

| テスト種別 | 対象 | 手法 | CI実行 | 実行タイミング |
|-----------|------|------|--------|-------------|
| ユニットテスト | 各パッケージの関数・メソッド | Go標準テスト + モック | Yes | 毎コミット |
| インテグレーションテスト | Slack API 呼び出し（モック） | `httptest` + `slacktest` | Yes | 毎コミット |
| MCP プロトコルテスト | MCP Server のツール呼び出し | `mcptest` パッケージ | Yes | 毎コミット |
| E2E テスト | 実 Slack ワークスペースでの動作 | 手動 or サンドボックス | No | リリース前 |

### 2.2 テストファイル配置

```
slack-fast-mcp/
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go          # ユニットテスト
│   ├── slack/
│   │   ├── client.go
│   │   ├── client_test.go          # インテグレーションテスト（httptest）
│   │   ├── channel.go
│   │   └── channel_test.go         # ユニットテスト
│   ├── mcp/
│   │   ├── server.go
│   │   ├── tools.go
│   │   └── tools_test.go           # MCP プロトコルテスト（mcptest）
│   └── cli/
│       ├── post.go
│       ├── post_test.go            # CLI テスト
│       └── ...
└── testdata/                       # テスト用固定データ
    ├── config/
    │   ├── valid.json              # 正常な設定ファイル
    │   ├── invalid.json            # 不正なJSON
    │   ├── env_ref.json            # 環境変数参照あり
    │   └── hardcoded_token.json    # トークン直書き（警告テスト用）
    └── slack/
        ├── post_message_ok.json    # Slack API レスポンスモック
        ├── history_ok.json
        └── error_channel_not_found.json
```

---

## 3. レイヤー別テスト詳細

### 3.1 Config Layer テスト（`internal/config`）

設定読み込みはバグの温床になりやすいため、最も手厚くテストする。

#### テストケース一覧

| # | テストケース | 検証内容 |
|---|-----------|---------|
| C01 | 正常なJSON設定ファイル読み込み | 全フィールドが正しくパースされる |
| C02 | 環境変数参照（`${SLACK_BOT_TOKEN}`）の展開 | 環境変数の値に正しく展開される |
| C03 | 環境変数未設定時の展開 | 空文字列になる（エラーではない） |
| C04 | 設定ファイル不在 | エラーにならない（環境変数フォールバック） |
| C05 | 不正なJSON | `config_parse_error` エラーが返る |
| C06 | プロジェクトローカル + グローバル設定のマージ | ローカルが優先される |
| C07 | 環境変数がローカル設定を上書き | 環境変数が最優先 |
| C08 | トークン直書き検出 | `xoxb-` 形式で警告が出力される |
| C09 | 空の設定ファイル | デフォルト値が適用される |
| C10 | `default_channel` 未設定 | `DefaultChannel` が空文字列 |

#### テスト実装パターン

```go
func TestLoadConfig_ValidJSON(t *testing.T) {
    // testdata/config/valid.json を使用
    dir := t.TempDir()
    writeTestConfig(t, dir, `{"token":"${SLACK_BOT_TOKEN}","default_channel":"general"}`)
    t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")

    cfg, err := config.Load(dir)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if cfg.Token != "xoxb-test-token" {
        t.Errorf("token = %q, want %q", cfg.Token, "xoxb-test-token")
    }
    if cfg.DefaultChannel != "general" {
        t.Errorf("default_channel = %q, want %q", cfg.DefaultChannel, "general")
    }
}
```

### 3.2 Slack Client テスト（`internal/slack`）

#### モック戦略

Slack API クライアントのテストは **2層構造** で行う:

**Layer 1: インターフェースモック（ユニットテスト向け）**
- `SlackClient` インターフェースのモック実装を作成
- MCP ツールハンドラーやCLIコマンドのテストで使用
- API呼び出し結果を自由に制御可能

```go
// mock_client.go（テスト用）
type MockSlackClient struct {
    PostMessageFunc    func(ctx context.Context, channel, message string) (*PostResult, error)
    PostThreadFunc     func(ctx context.Context, channel, threadTS, message string) (*PostResult, error)
    GetHistoryFunc     func(ctx context.Context, channel string, opts HistoryOptions) (*HistoryResult, error)
    ResolveChannelFunc func(ctx context.Context, channel string) (string, error)
}

func (m *MockSlackClient) PostMessage(ctx context.Context, channel, message string) (*PostResult, error) {
    return m.PostMessageFunc(ctx, channel, message)
}
// ... 他メソッドも同様
```

**Layer 2: HTTP レベルモック（インテグレーションテスト向け）**
- `httptest.NewServer()` で Slack API のモックサーバーを構築
- `slack.OptionAPIURL()` でモックサーバーに接続
- 実際の HTTP リクエスト/レスポンスレベルで検証

```go
func TestClient_PostMessage(t *testing.T) {
    handler := http.NewServeMux()
    handler.HandleFunc("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprint(w, `{"ok":true,"channel":"C01234","ts":"1234567890.123456"}`)
    })
    mockServer := httptest.NewServer(handler)
    defer mockServer.Close()

    api := slack.New("xoxb-test", slack.OptionAPIURL(mockServer.URL+"/"))
    client := slackclient.NewClientWithAPI(api)

    result, err := client.PostMessage(context.Background(), "C01234", "Hello")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.TS == "" {
        t.Error("expected non-empty timestamp")
    }
}
```

#### テストケース一覧

| # | テストケース | テスト層 | 検証内容 |
|---|-----------|---------|---------|
| S01 | PostMessage 正常系 | HTTP モック | メッセージ投稿成功・レスポンスパース |
| S02 | PostMessage チャンネル名指定 | HTTP モック | チャンネル名→ID 変換 + 投稿 |
| S03 | PostThread 正常系 | HTTP モック | スレッド返信成功 |
| S04 | GetHistory 正常系 | HTTP モック | 履歴取得・メッセージ配列パース |
| S05 | GetHistory limit パラメータ | HTTP モック | limit=1〜100 の範囲制約 |
| S06 | ResolveChannel: チャンネルID | ユニット | `C01234ABCDE` → そのまま返す |
| S07 | ResolveChannel: `#` 付きチャンネル名 | HTTP モック | `#general` → `general` で検索 |
| S08 | ResolveChannel: チャンネル名 | HTTP モック | conversations.list で検索 |
| S09 | ResolveChannel: 存在しないチャンネル | HTTP モック | `channel_not_found` エラー |
| S10 | エラー: channel_not_found | HTTP モック | Slack API エラーレスポンス処理 |
| S11 | エラー: not_in_channel | HTTP モック | Bot未参加エラー処理 |
| S12 | エラー: invalid_auth | HTTP モック | 認証エラー処理 |
| S13 | エラー: rate_limited + リトライ | HTTP モック | 429 → Retry-After → リトライ成功 |
| S14 | エラー: network_error | HTTP モック | 接続失敗エラー処理 |
| S15 | context キャンセル | HTTP モック | タイムアウト / キャンセル時の中断 |
| S16 | ページネーション（conversations.list） | HTTP モック | cursor を使った複数ページ取得 |

### 3.3 MCP Server テスト（`internal/mcp`）

#### mcptest パッケージの活用

mcp-go の `mcptest` パッケージを使用して、MCP プロトコルレベルでツールをテストする。

```go
func TestPostMessageTool(t *testing.T) {
    mockClient := &MockSlackClient{
        PostMessageFunc: func(ctx context.Context, channel, message string) (*PostResult, error) {
            return &PostResult{
                Channel:     "C01234",
                ChannelName: "general",
                TS:          "1234567890.123456",
                Message:     message,
                Permalink:   "https://example.slack.com/archives/C01234/p1234567890123456",
            }, nil
        },
    }

    cfg := &config.Config{DefaultChannel: "general"}
    srv := mcp.NewServerWithClient(cfg, mockClient)

    // mcptest を使用してツール呼び出しをテスト
    ts := mcptest.NewServer(srv)
    defer ts.Close()

    client := ts.Client()
    result, err := client.CallTool(context.Background(), "slack_post_message", map[string]any{
        "message": "Hello from test",
    })
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // result の検証
}
```

#### テストケース一覧

| # | テストケース | 検証内容 |
|---|-----------|---------|
| M01 | slack_post_message 正常系 | ツール呼び出し成功・JSON 応答 |
| M02 | slack_post_message チャンネル指定 | channel パラメータの伝播 |
| M03 | slack_post_message デフォルトチャンネル | channel 省略時のフォールバック |
| M04 | slack_post_message no_default_channel | channel 省略 + デフォルト未設定 → エラー |
| M05 | slack_post_message no_text | message 空 → エラー |
| M06 | slack_get_history 正常系 | 履歴取得成功・JSON 応答 |
| M07 | slack_get_history limit 指定 | limit パラメータの伝播 |
| M08 | slack_post_thread 正常系 | スレッド返信成功・JSON 応答 |
| M09 | slack_post_thread thread_not_found | 存在しない thread_ts → エラー |
| M10 | Slack API エラーのMCPエラー変換 | AppError → MCP ToolResultError への変換 |
| M11 | エラーHintがLLM向けに適切 | Hint が英語で、具体的なアクション指示を含む |
| M12 | ツール一覧取得（ListTools） | 3 ツールが登録されている |
| M13 | ツール description の内容確認 | LLM 向け description が要件通り |

### 3.4 CLI テスト（`internal/cli`）

CLI テストは cobra のテスト機能を利用し、コマンド実行結果を検証する。

```go
func TestPostCommand(t *testing.T) {
    mockClient := &MockSlackClient{
        PostMessageFunc: func(ctx context.Context, channel, message string) (*PostResult, error) {
            return &PostResult{TS: "1234567890.123456"}, nil
        },
    }

    cmd := NewRootCmd(WithSlackClient(mockClient))
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetArgs([]string{"post", "--message", "Hello", "--channel", "general"})

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // buf の出力を検証
}
```

#### テストケース一覧

| # | テストケース | 検証内容 |
|---|-----------|---------|
| L01 | `post` コマンド正常系 | メッセージ投稿・成功出力 |
| L02 | `history` コマンド正常系 | 履歴取得・テーブル出力 |
| L03 | `reply` コマンド正常系 | スレッド返信・成功出力 |
| L04 | `--json` フラグ | JSON 形式で出力 |
| L05 | 引数なし実行 | serve モード（MCP Server）が起動 |
| L06 | `version` コマンド | バージョン文字列が出力される |
| L07 | 必須引数不足 | エラーメッセージ + usage 表示 |
| L08 | `--config` フラグ | 指定パスの設定ファイルを読み込む |

---

## 4. モック実装設計

### 4.1 モック構成

```
internal/
├── slack/
│   ├── client.go          # SlackClient インターフェース定義
│   ├── client_impl.go     # 実装（本番用）
│   └── mock_client.go     # モック実装（テスト用）※ _test.go と同じパッケージ
```

> **注意**: `mock_client.go` はテスト専用だが、`internal/mcp` や `internal/cli` からも利用するため、`_test.go` サフィックスは付けない。`internal/` 配下なので外部パッケージからのアクセスは不可。

### 4.2 テストヘルパー

共通のテストヘルパーを `internal/testutil/` に配置する:

```go
// internal/testutil/helpers.go
package testutil

// NewMockSlackServer は Slack API のモック HTTP サーバーを作成する
func NewMockSlackServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server

// WriteTestConfig はテスト用の設定ファイルを一時ディレクトリに書き出す
func WriteTestConfig(t *testing.T, dir string, content string) string

// AssertJSONContains は JSON 文字列に期待するキー・値が含まれることを検証する
func AssertJSONContains(t *testing.T, jsonStr string, key string, expected any)
```

---

## 5. CI/CD テスト連携

### 5.1 GitHub Actions ワークフロー（`.github/workflows/ci.yml`）

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.23', '1.24']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out
      - name: Check coverage
        run: |
          go tool cover -func=coverage.out
          # 全体カバレッジが 75% 未満の場合は失敗
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 75" | bc -l) )); then
            echo "Coverage $COVERAGE% is below threshold 75%"
            exit 1
          fi

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

### 5.2 テスト実行コマンド

```bash
# 全テスト実行（レースコンディション検出あり）
go test ./... -v -race

# カバレッジ付き実行
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 特定パッケージのみ
go test ./internal/config/... -v
go test ./internal/slack/... -v
go test ./internal/mcp/... -v

# 短縮テスト（CI高速化用）
go test ./... -short
```

### 5.3 `-short` フラグの活用

時間のかかるテスト（リトライ待ちなど）は `-short` フラグで制御する:

```go
func TestRateLimitRetry(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping retry test in short mode")
    }
    // リトライを実際に待つテスト
}
```

---

## 6. E2E テスト方針

### 6.1 手動 E2E テスト

リリース前に以下の手動テストを実施する:

| # | テストシナリオ | 検証内容 |
|---|-------------|---------|
| E01 | Cursor MCP 接続 | `.cursor/mcp.json` 設定 → Cursor からツール呼び出し可能 |
| E02 | メッセージ投稿 | Cursor → slack_post_message → Slack にメッセージが表示 |
| E03 | 履歴取得 | Cursor → slack_get_history → 最新メッセージが取得できる |
| E04 | スレッド返信 | Cursor → slack_post_thread → スレッドに返信が表示 |
| E05 | CLI post | `slack-fast-mcp post --message "test"` → Slack に投稿 |
| E06 | CLI history | `slack-fast-mcp history` → 履歴がテーブル表示 |
| E07 | CLI reply | `slack-fast-mcp reply --thread-ts xxx --message "reply"` → スレッド返信 |
| E08 | setup ウィザード | `slack-fast-mcp setup` → 対話的に設定ファイル生成 |
| E09 | エラー: トークン未設定 | トークンなしで起動 → 分かりやすいエラーメッセージ |
| E10 | エラー: Bot 未参加 | 未参加チャンネルに投稿 → 招待手順が案内される |

### 6.2 E2E テスト用 Slack ワークスペース

- **テスト用ワークスペース**: 開発者個人の Slack ワークスペースを使用
- **テスト用チャンネル**: `#mcp-test` チャンネルを作成してテスト
- **CI 連携**: E2E テストは CI には含めない（Slack トークンのシークレット管理が煩雑なため）

---

## 7. テスト品質基準

### 7.1 テストコードの品質ルール

1. **テストは独立している**: 各テストケースは他のテストに依存しない
2. **テストは冪等**: 何度実行しても同じ結果になる
3. **テストは高速**: 個別のテストは 1 秒以内に完了する
4. **テスト名は説明的**: `TestPostMessage_WithValidChannel_ReturnsSuccess` のように状況と期待結果を含む
5. **テーブル駆動テスト**: 入力パターンが多い場合はテーブル駆動テストを使用

### 7.2 テーブル駆動テストの例

```go
func TestResolveChannel(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantID   string
        wantErr  bool
    }{
        {"channel ID", "C01234ABCDE", "C01234ABCDE", false},
        {"channel name", "general", "C01234ABCDE", false},
        {"with hash prefix", "#general", "C01234ABCDE", false},
        {"not found", "nonexistent", "", true},
        {"private channel ID", "G01234ABCDE", "G01234ABCDE", false},
        {"DM channel ID", "D01234ABCDE", "D01234ABCDE", false},
        {"C-prefixed name", "ci-notifications", "C09876ZZZZZ", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実装
        })
    }
}
```

---

## 8. テスト実装の優先順位

Phase 2（コア実装）では、以下の順序でテストを実装する:

| 順序 | 対象 | 理由 |
|------|------|------|
| 1 | `internal/config` テスト | 全レイヤーの基盤。設定が正しく読めないと全機能が動かない |
| 2 | `internal/slack` チャンネル名解決テスト | ビジネスロジックの中核。誤判定のリスクが高い |
| 3 | `internal/slack` API クライアントテスト | Slack API 呼び出しの正常系・エラー系 |
| 4 | `internal/mcp` ツールテスト | MCP プロトコルレベルの検証 |
| 5 | `internal/cli` コマンドテスト | CLI の統合テスト |

> **実装方針**: 各コンポーネントの実装と同時にテストを書く（TDD は強制しないが、実装と同じ PR にテストを含めること）。

---

## 9. ローカル品質保証システム

### 9.1 概要

「デグレしない」「ユーザーが使った時に不具合が出ない」を **ローカル完結** で担保する仕組みを構築している。CI に依存せず、開発者のマシン上で全チェックが完結する。

```
品質保証フロー:
  開発者がコード変更
    ↓
  make test        → 高速テスト実行（日常開発用）
    ↓
  make quality     → 品質ゲート（push前に自動実行）
    ├── [1/6] go vet          … 静的解析
    ├── [2/6] build           … コンパイル成功確認
    ├── [3/6] test + race     … 全テスト + レースコンディション検出
    ├── [4/6] coverage >= 75% … カバレッジ閾値チェック
    ├── [5/6] smoke test      … バイナリ起動・応答確認
    └── [6/6] report save     … テストレポート保存
    ↓
  git push         → pre-push hook が make quality を自動実行
    ↓（NG なら push 拒否）
```

### 9.2 Makefile ターゲット一覧

| コマンド | 用途 | 実行タイミング |
|---------|------|-------------|
| `make test` | 高速テスト（日常開発） | コード変更のたび |
| `make test-verbose` | 詳細出力テスト | デバッグ時 |
| `make test-race` | race detector 付きテスト | 並行処理変更時 |
| `make test-cover` | カバレッジ計測 | テスト追加・変更時 |
| `make test-cover-html` | カバレッジ HTML レポート | 詳細分析時 |
| `make quality` | **品質ゲート（全チェック）** | **push前（自動）** |
| `make smoke` | スモークテスト単体 | バイナリ動作確認 |
| `make build` | バイナリビルド | 配布前 |
| `make build-verify` | クロスプラットフォームコンパイル確認 | リリース前 |
| `make setup-hooks` | Git hooks セットアップ | 初回のみ |
| `make clean` | 成果物削除 | 必要時 |

### 9.3 Git Pre-Push Hook

**セットアップ**: `make setup-hooks` で自動インストール

- push 前に `make quality` が自動実行される
- 品質ゲートが NG なら push が拒否される
- 緊急時のスキップ: `git push --no-verify`（非推奨）

### 9.4 スモークテスト

「ユーザーが実際に使った時に動くか？」を検証する2段構えのスモークテスト:

**レイヤー1: Go テスト（`internal/mcp/smoke_test.go`）**
- MCPプロトコルレベルでの全3ツール呼び出し検証
- エラー時のユーザーフレンドリーメッセージ検証
- `go test` に含まれるので品質ゲートで毎回実行

**レイヤー2: シェルスクリプト（`scripts/smoke-test.sh`）**
- ビルド済みバイナリの存在・実行権限確認
- トークン未設定時のエラーメッセージ確認
- バイナリサイズの妥当性確認（5〜50MB）

### 9.5 テストレポート

品質ゲート実行時に `reports/` ディレクトリにレポートが自動保存される:

```
reports/
├── latest-report.txt          # 最新レポート（常に上書き）
├── report-YYYY-MM-DD_HHMMSS.txt  # タイムスタンプ付きレポート（履歴）
├── coverage.out               # カバレッジプロファイル
└── coverage.html              # カバレッジHTMLレポート（make test-cover-html 時）
```

> レポートファイルは `.gitignore` に含まれ、Git にはコミットされない（ローカル参照用）。

### 9.6 カバレッジ閾値

`.testcoverage.yml` で定義:

| レベル | 閾値 |
|--------|------|
| 全体 | 75% |
| パッケージごと | 60% |

### 9.7 デグレ防止のメカニズム

| 脅威 | 防止策 | 実装 |
|------|--------|------|
| テストが落ちるコードを push | pre-push hook で全テスト自動実行 | `scripts/pre-push` |
| カバレッジが低下 | 閾値未満で push 拒否 | `make _quality_coverage` |
| レースコンディション | `-race` フラグ付きテスト | `make test-race` |
| バイナリがビルドできない | ビルド検証 | `make _quality_build` |
| バイナリが起動しない | スモークテスト | `scripts/smoke-test.sh` |
| エラーメッセージが不親切 | Smoke テストでHint検証 | `smoke_test.go` |
| 静的解析エラー | go vet | `make _quality_vet` |
