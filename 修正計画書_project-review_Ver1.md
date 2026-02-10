# 修正計画書

**プロジェクト**: slack-fast-mcp
**作成日**: 2026-02-10
**バージョン**: Ver.1
**レビュー手法**: 7軸並列レビュー（戦略/論理・MECE/実行設計/ポジション/可読性/ヒューマンライズ/リスク）

> **レビュー結果ファイル**: `reviews/project-review/` ディレクトリに各軸のレビュー結果を保存

---

## 修正方針

本計画書は、OSS公開済み（v0.1.0）プロジェクトの品質向上を目的とし、以下の優先度基準で修正を計画する。

| 優先度 | 基準 | 対応期限 |
|---|---|---|
| **P0（緊急）** | セキュリティリスク、ドキュメント整合性の致命的な乖離 | 即座（v0.1.1 パッチ） |
| **P1（重要）** | パフォーマンス改善、ユーザビリティ向上、品質基準の信頼性 | v0.2.0 |
| **P2（改善推奨）** | コード品質、ドキュメントの洗練、将来の拡張性 | 適時 |

---

## P0: 緊急修正（v0.1.1 パッチ）

### FIX-01: requirements.md に display_name パラメータを追記

| 項目 | 内容 |
|---|---|
| **関連指摘** | L-01, L-02, S-04 |
| **対象ファイル** | `docs/requirements.md`, `docs/architecture.md` |
| **問題** | `display_name` パラメータが requirements.md（Single Source of Truth）に未定義。architecture.md の Config 構造体にも不在。READMEと実装には存在する |
| **修正内容** | |

1. **requirements.md §3.1** (`slack_post_message`): 入力パラメータに追加
   ```
   | `display_name` | string | No | 設定ファイルの display_name | 送信者表示名。メッセージ末尾に `#名前` ハッシュタグを付与 |
   ```

2. **requirements.md §3.3** (`slack_post_thread`): 同様に追加

3. **requirements.md §4.2**: `.slack-mcp.json` のフィールド定義に追加
   ```
   | `display_name` | string | No | デフォルトの送信者表示名 |
   ```

4. **requirements.md §4.4**: 環境変数テーブルに追加
   ```
   | `SLACK_DISPLAY_NAME` | デフォルトの送信者表示名 | No |
   ```

5. **architecture.md §4.1**: Config struct に `DisplayName` フィールドを追加
   ```go
   type Config struct {
       Token          string `json:"token"`
       DefaultChannel string `json:"default_channel"`
       DisplayName    string `json:"display_name"`
       LogLevel       string `json:"log_level"`
   }
   ```

---

### FIX-02: Go バージョンの統一

| 項目 | 内容 |
|---|---|
| **関連指摘** | L-04, K-08 |
| **対象ファイル** | `CONTRIBUTING.md`, `docs/architecture.md`, (`go.mod` — 必要に応じて) |
| **問題** | go.mod (1.25.0), architecture.md (1.23+), CONTRIBUTING.md (1.25+) で不一致 |
| **修正内容** | |

1. **go.mod の値を正とする**: `go 1.25.0` が実際にリリース済みか確認
   - リリース済みの場合: 全ドキュメントを `Go 1.25+` に統一
   - 未リリースの場合: go.mod を最新安定版（例: `go 1.24.0`）に修正し、全ドキュメントを同期
2. **architecture.md §1**: 技術スタックの Go バージョンを修正
3. **CONTRIBUTING.md**: Prerequisites の Go バージョンを修正

---

### FIX-03: カバレッジ閾値の統一

| 項目 | 内容 |
|---|---|
| **関連指摘** | L-07 |
| **対象ファイル** | `docs/testing-strategy.md`, `Makefile`, `.github/workflows/ci.yml` |
| **問題** | testing-strategy.md は 75%、Makefile/CI は 65%。実カバレッジ 67.3% |
| **修正内容** | |

**方針**: 実態に合わせて testing-strategy.md の目標値を修正する。

1. **testing-strategy.md §1.2**: 全体カバレッジ目標を `65%+` に変更（Makefile/CI と統一）
   - 注記: 「CLI レイヤーのテストカバレッジ向上に伴い、将来的に 75% に引き上げを検討」
2. **Makefile**: `COVERAGE_THRESHOLD` は `65` のまま維持
3. **ci.yml**: 閾値 `65` のまま維持

---

### FIX-04: setup ウィザードのトークンセキュリティ改善

| 項目 | 内容 |
|---|---|
| **関連指摘** | K-01, E-02 |
| **対象ファイル** | `internal/cli/setup.go` |
| **問題** | トークンが平文で stdin/stdout に出力される |
| **修正内容** | |

1. **トークン入力**: `golang.org/x/term` パッケージの `ReadPassword()` を使用してエコーバックなしの入力に変更
   ```go
   import "golang.org/x/term"
   
   fmt.Fprint(out, "Enter your Bot User OAuth Token: ")
   tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
   fmt.Fprintln(out) // 改行を補完
   ```

2. **トークン出力**: マスキングして表示
   ```go
   // 変更前
   fmt.Fprintf(out, "  export SLACK_BOT_TOKEN='%s'\n", token)
   
   // 変更後
   maskedToken := errors.MaskToken(token)
   fmt.Fprintf(out, "  export SLACK_BOT_TOKEN='<your-token>'  # %s\n", maskedToken)
   ```

3. **ユーザー案内**: 「入力したトークンは画面に表示されません。環境変数に設定してください」と案内

---

## P1: 重要修正（v0.2.0）

### FIX-05: GetHistory の N+1 ユーザー名解決問題

| 項目 | 内容 |
|---|---|
| **関連指摘** | E-04 |
| **対象ファイル** | `internal/slack/client.go` |
| **問題** | メッセージごとに `GetUserInfoContext` を呼び出す N+1 問題 |
| **修正内容** | |

1. メッセージループの前にユニークユーザーIDを収集
2. 一括でユーザー情報を取得してキャッシュ
3. メッセージのユーザー名をキャッシュから解決

```go
// ユニークユーザーID収集
userIDs := make(map[string]bool)
for _, msg := range resp.Messages {
    if msg.User != "" {
        userIDs[msg.User] = true
    }
}

// 一括ユーザー名解決
userNames := make(map[string]string)
for userID := range userIDs {
    if user, err := c.api.GetUserInfoContext(ctx, userID); err == nil {
        userNames[userID] = user.Name
    }
}

// メッセージにユーザー名をマッピング
for _, msg := range resp.Messages {
    hm.UserName = userNames[msg.User]
}
```

---

### FIX-06: Claude Desktop 設定例のセキュリティ注記追加

| 項目 | 内容 |
|---|---|
| **関連指摘** | P-02 |
| **対象ファイル** | `README.md`, `README_ja.md` |
| **問題** | Claude Desktop 設定例でトークン直書きが推奨されているように見える |
| **修正内容** | |

Claude Desktop の設定例の直前に以下の注記を追加:

```markdown
> **Security Note:** Claude Desktop may not support `${VAR}` environment variable expansion.
> If you must set the token directly, ensure this config file is NOT committed to Git.
> Consider adding `claude_desktop_config.json` to `.gitignore`.
```

---

### FIX-07: testing-strategy.md と CI ワークフローの同期

| 項目 | 内容 |
|---|---|
| **関連指摘** | L-05 |
| **対象ファイル** | `docs/testing-strategy.md` |
| **問題** | CI ワークフロー例が実際の ci.yml と不一致 |
| **修正内容** | |

1. §5.1 の YAML 例を実際の `.github/workflows/ci.yml` の内容に更新
2. ジョブ数を lint + test + build の3ジョブに修正
3. Go バージョンマトリクスを `go-version-file: go.mod` に更新

---

### FIX-08: architecture.md のファイル構成を実態に同期

| 項目 | 内容 |
|---|---|
| **関連指摘** | E-08, L-02 |
| **対象ファイル** | `docs/architecture.md` |
| **問題** | `channel_test.go` が記載されているが存在しない。`version.go`, `errors/` パッケージのパス等 |
| **修正内容** | |

1. §2 のプロジェクト構成ツリーを実際のファイル構成に更新
   - `channel_test.go` を削除
   - `errors/` パッケージを追加
   - `mock_client.go` を追加
   - `types.go` を追加
   - `cli/version.go`, `cli/setup.go` を追加
   - `Makefile`, `scripts/` を追加

---

### FIX-09: メッセージ長バリデーション追加

| 項目 | 内容 |
|---|---|
| **関連指摘** | K-02 |
| **対象ファイル** | `internal/mcp/tools.go`, `internal/cli/post.go`, `internal/cli/reply.go` |
| **問題** | メッセージ長のバリデーションがない（Slack上限: 40,000文字） |
| **修正内容** | |

1. MCP ツールハンドラーにメッセージ長チェックを追加
2. CLI サブコマンドにもチェックを追加
3. 新しいエラーコード `message_too_long` を errors.go に追加

```go
const maxMessageLength = 40000

if len(message) > maxMessageLength {
    appErr := apperr.New("message_too_long",
        fmt.Sprintf("メッセージが長すぎます（%d文字）。上限は%d文字です", len(message), maxMessageLength), nil)
    return mcp.NewToolResultError(appErr.FormatForMCP()), nil
}
```

---

### FIX-10: 設定ファイル権限チェック追加

| 項目 | 内容 |
|---|---|
| **関連指摘** | K-04 |
| **対象ファイル** | `internal/config/config.go` |
| **問題** | トークン含有設定ファイルの権限チェックがない |
| **修正内容** | |

1. `loadFromFile` でファイル権限を確認
2. 0644 より緩い権限（例: 0666, 0777）の場合、stderr に警告を出力
3. 警告のみで動作は継続（ユーザビリティ優先）

```go
func checkFilePermissions(path string) {
    info, err := os.Stat(path)
    if err != nil {
        return
    }
    mode := info.Mode().Perm()
    if mode&0077 != 0 { // group/other にアクセス権がある場合
        fmt.Fprintf(os.Stderr, "WARNING: Config file %s has permissions %o. Consider restricting to 0600.\n", path, mode)
    }
}
```

---

## P2: 改善推奨（適時）

### FIX-11: ベンチマーク根拠の追加

| 項目 | 内容 |
|---|---|
| **関連指摘** | S-01, H-01 |
| **対象ファイル** | 新規 `benchmarks/` ディレクトリ, `README.md`, `README_ja.md` |
| **修正内容** | |

1. `benchmarks/startup-time.sh` にベンチマークスクリプトを作成
2. README の比較表に注釈を追加: 「Benchmark details: see benchmarks/」
3. 比較対象の具体名を記載

---

### FIX-12: コードコメントの英語化（パブリックAPI）

| 項目 | 内容 |
|---|---|
| **関連指摘** | R-03 |
| **対象ファイル** | `internal/` 全体のパブリック関数・型のコメント |
| **修正内容** | |

1. GoDoc に表示されるパブリック API のコメントを英語に統一
2. プライベート関数のコメントは日本語のまま維持（開発者の利便性優先）
3. `internal/errors/errors.go` の AppError 構造体のコメントを英語化

---

### FIX-13: version --json のJSON生成を安全化

| 項目 | 内容 |
|---|---|
| **関連指摘** | P-05 |
| **対象ファイル** | `internal/cli/version.go` |
| **修正内容** | |

`fmt.Fprintf` による手動JSON生成を `json.Marshal` に変更:

```go
if flagJSON {
    out := map[string]string{
        "version":    Version,
        "commit":     Commit,
        "date":       Date,
        "go_version": runtime.Version(),
        "platform":   runtime.GOOS + "/" + runtime.GOARCH,
    }
    encoder := json.NewEncoder(cmd.OutOrStdout())
    encoder.SetEscapeHTML(false)
    return encoder.Encode(out)
}
```

---

### FIX-14: CLI グローバル変数の構造改善

| 項目 | 内容 |
|---|---|
| **関連指摘** | E-01 |
| **対象ファイル** | `internal/cli/root.go`, `post.go`, `reply.go`, `history.go` |
| **修正内容** | |

`flagMessage` が post.go と reply.go で共有されている問題を解消:
- 各コマンドのローカルフラグとして定義
- RunE 内で `cmd.Flags().GetString()` でフラグ値を取得

---

### FIX-15: MCP Server モードのリクエストタイムアウト実装

| 項目 | 内容 |
|---|---|
| **関連指摘** | E-07 |
| **対象ファイル** | `internal/mcp/tools.go` |
| **修正内容** | |

各ツールハンドラーの冒頭で context にタイムアウトを設定:

```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

---

## 修正タスクサマリー

| ID | 修正内容 | 優先度 | 難易度 | 推定時間 |
|---|---|---|---|---|
| FIX-01 | display_name ドキュメント追記 | P0 | 低 | 30分 |
| FIX-02 | Go バージョン統一 | P0 | 低 | 15分 |
| FIX-03 | カバレッジ閾値統一 | P0 | 低 | 15分 |
| FIX-04 | setup トークンセキュリティ | P0 | 中 | 1時間 |
| FIX-05 | GetHistory N+1 解消 | P1 | 中 | 1.5時間 |
| FIX-06 | Claude Desktop セキュリティ注記 | P1 | 低 | 15分 |
| FIX-07 | testing-strategy.md 同期 | P1 | 低 | 30分 |
| FIX-08 | architecture.md ファイル構成更新 | P1 | 低 | 30分 |
| FIX-09 | メッセージ長バリデーション | P1 | 低 | 45分 |
| FIX-10 | 設定ファイル権限チェック | P1 | 低 | 30分 |
| FIX-11 | ベンチマーク根拠追加 | P2 | 中 | 1時間 |
| FIX-12 | コメント英語化 | P2 | 中 | 2時間 |
| FIX-13 | version --json 安全化 | P2 | 低 | 15分 |
| FIX-14 | CLI グローバル変数改善 | P2 | 中 | 1時間 |
| FIX-15 | MCP タイムアウト実装 | P2 | 低 | 30分 |

**合計推定時間**: P0: 約2時間、P1: 約4時間、P2: 約5時間

---

## 推奨実施順序

### Phase A（P0 — 即座）
1. FIX-02: Go バージョン統一（他の全ての作業の前提条件）
2. FIX-01: display_name ドキュメント追記
3. FIX-03: カバレッジ閾値統一
4. FIX-04: setup トークンセキュリティ

### Phase B（P1 — v0.2.0）
5. FIX-05: GetHistory N+1 解消
6. FIX-09: メッセージ長バリデーション
7. FIX-10: 設定ファイル権限チェック
8. FIX-08: architecture.md ファイル構成更新
9. FIX-07: testing-strategy.md 同期
10. FIX-06: Claude Desktop セキュリティ注記

### Phase C（P2 — 適時）
11. FIX-13: version --json 安全化
12. FIX-15: MCP タイムアウト実装
13. FIX-14: CLI グローバル変数改善
14. FIX-12: コメント英語化
15. FIX-11: ベンチマーク根拠追加
