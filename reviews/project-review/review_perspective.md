# レビュー結果: ④ ポジション（For Whom）

**レビュー日**: 2026-02-10
**対象**: slack-fast-mcp プロジェクト全体
**レビュアー軸**: 上長/同僚/メンバー視点、読者体験

---

## 1. 総合評価

**評価: A（優秀）**

READMEは英語・日本語ともに非常に分かりやすく、Quick Start から実際の利用までの導線が整っている。Slack App セットアップガイドも丁寧。いくつかの読者視点での改善ポイントがある。

---

## 2. 指摘事項

### P-01: 初めて見るOSSコントリビューターにとってプロジェクト構成の把握が難しい（重要度: 低）

- **場所**: CONTRIBUTING.md §Project Structure
- **内容**: ディレクトリ構成はあるが、各パッケージ間の依存関係やデータフローの説明がない
- **影響**: 初回コントリビューターが修正箇所を特定しにくい
- **提案**: architecture.md §3.1 のレイヤー構成図への参照リンクを CONTRIBUTING.md に追加

### P-02: Claude Desktop の設定例でトークン直書きが推奨されているように見える（重要度: 中）

- **場所**: README.md L181-193, README_ja.md L181-193
- **内容**: Claude Desktop の設定例で `"SLACK_BOT_TOKEN": "your-token-here"` と直書き例が示されている。Cursor/Windsurf は `${SLACK_BOT_TOKEN}` 参照だが、Claude Desktop は生トークン
- **影響**: セキュリティ意識の低いユーザーがトークンを直書きし、設定ファイルをGitにコミットするリスク
- **提案**: Claude Desktop でも環境変数参照が使えるか確認。使えない場合は注意喚起を強化

### P-03: setup ウィザードの環境変数設定案内が ~/.zshrc を案内していない（重要度: 低）

- **場所**: internal/cli/setup.go L106
- **内容**: "Add this to your shell profile (~/.zshrc, ~/.bashrc) for persistence." と案内しているが、slack-app-setup.md §5.1 では `~/.zprofile` を推奨
- **影響**: ユーザーによって異なるファイルに設定が分散する可能性
- **提案**: setup ウィザード内の案内も `~/.zprofile`（macOS） を第一候補に記載

### P-04: エラーメッセージが日本語と英語が混在（重要度: 低）

- **場所**: internal/config/config.go L71, internal/errors/errors.go
- **内容**: AppError の Message は日本語（「トークンが設定されていません」）だが、Hint は英語。CLI 出力も混在
- **影響**: 英語圏ユーザーにとって Message 部分が読めない
- **提案**: OSS公開を考慮し、Message も英語化するか、ロケール切替の仕組みを検討。少なくとも Hint（LLM向け）は英語のまま維持

### P-05: version --json の出力が適切にフォーマットされていない（重要度: 低）

- **場所**: internal/cli/version.go L23-24
- **内容**: `fmt.Fprintf` でJSON文字列を手動生成しているため、特殊文字のエスケープ漏れのリスクがある
- **影響**: version/commit/date に特殊文字が含まれた場合にJSONパースエラー
- **提案**: `json.Marshal` を使用して安全にJSON生成
