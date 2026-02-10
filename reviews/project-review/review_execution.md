# レビュー結果: ③ 実行設計（How）

**レビュー日**: 2026-02-10
**対象**: slack-fast-mcp プロジェクト全体
**レビュアー軸**: 実行可能性・リソース・コード品質

---

## 1. 総合評価

**評価: A-（良好）**

Go のベストプラクティスに概ね従った堅実な実装。レイヤー分離、インターフェース設計、エラーハンドリングが適切。CI/CD パイプラインも機能的。いくつかの改善ポイントがある。

---

## 2. 指摘事項

### E-01: CLI グローバル変数によるフラグ管理（重要度: 中）

- **場所**: internal/cli/root.go L22-30, post.go L12, reply.go L12, history.go L14-17
- **内容**: `flagMessage`, `flagThreadTS`, `flagLimit` 等がパッケージレベルのグローバル変数として定義されている。`flagMessage` は post.go と reply.go の両方で使用されるが、同一変数を共有している
- **影響**: テストの並列実行時にデータ競合のリスク。コマンド間の意図しない状態共有
- **提案**: 各コマンドの RunE 内でフラグ値を取得するパターンに変更するか、struct ベースのコマンド設計に移行

### E-02: setup コマンドのトークンが平文で表示される（重要度: 中）

- **場所**: internal/cli/setup.go L104
- **内容**: `fmt.Fprintf(out, "  export SLACK_BOT_TOKEN='%s'\n", token)` でユーザーが入力したトークンがそのまま stdout に出力される
- **影響**: ターミナルログやスクリーンレコーディングでトークンが漏洩する可能性
- **提案**: トークンの最初と最後の数文字のみ表示し、中間をマスキング。または「入力したトークンを環境変数に設定してください」という案内に留める

### E-03: channelCache のスレッドセーフティ（重要度: 低）

- **場所**: internal/slack/client.go L20-21
- **内容**: `channelCache map[string]string` は `sync.Map` やミューテックスなしで使用されている
- **影響**: MCP Server は stdio で逐次処理のため、現時点では問題ない。ただし、将来 HTTP transport 対応時に並行アクセスが発生しうる
- **提案**: 現時点では注意コメントを追加。HTTP transport 対応時に `sync.Map` に変更

### E-04: GetHistory のユーザー名解決が N+1 問題を起こす（重要度: 中）

- **場所**: internal/slack/client.go L158-161
- **内容**: 各メッセージに対して `GetUserInfoContext` を個別に呼び出している。10件のメッセージで10回の API 呼び出し
- **影響**: 履歴取得のレスポンスが遅くなる。レート制限に達しやすくなる
- **提案**: ユニークなユーザーIDを収集し、一括でユーザー情報を取得するか、ユーザーキャッシュを導入

### E-05: GoReleaser の before hooks でテスト全実行（重要度: 低）

- **場所**: .goreleaser.yml L6-8
- **内容**: `before.hooks` に `go test ./...` が含まれている。CI の release ワークフローでは別途テストが実行されるため重複
- **影響**: リリースビルド時間の不必要な増加
- **提案**: CI でテスト済みなら GoReleaser の before hooks からテストを除去

### E-06: cobra が require ではなく indirect 依存になっている（重要度: 低）

- **場所**: go.mod L19
- **内容**: `github.com/spf13/cobra v1.10.2 // indirect` だが、internal/cli/ で直接 import している
- **影響**: `go mod tidy` で意図せず削除される可能性は低いが、依存関係の透明性が低下
- **提案**: `go mod tidy` を再実行して正しい依存分類を確認

### E-07: MCP Server モードのタイムアウト未実装（重要度: 低）

- **場所**: internal/cli/serve.go, architecture.md §4.2
- **内容**: architecture.md では「MCP Server モード: リクエスト全体 30秒タイムアウト」を定義しているが、serve.go では context にタイムアウトを設定していない（CLI モードは10秒タイムアウトあり）
- **影響**: MCP ツール呼び出しが無限にハングする可能性（Slack API がハングした場合）
- **提案**: MCP ハンドラー内で context にタイムアウトを設定

### E-08: テストで channel_test.go が存在しない（重要度: 中）

- **場所**: architecture.md §2 プロジェクト構成 vs 実際のファイル
- **内容**: architecture.md に `channel_test.go` が記載されているが、実際のプロジェクトには存在しない。チャンネル名解決ロジックのテストは client_test.go に含まれている
- **影響**: ドキュメントと実態の乖離。テスト戦略書で定義された S06-S09 のテストは client_test.go に含まれているが、ファイル構成の期待と異なる
- **提案**: architecture.md のファイル構成を実態に合わせて更新
