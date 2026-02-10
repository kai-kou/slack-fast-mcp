# レビュー結果: ② 論理・MECE

**レビュー日**: 2026-02-10
**対象**: slack-fast-mcp プロジェクト全体
**レビュアー軸**: 論理整合性・MECE検証・根拠の妥当性

---

## 1. 総合評価

**評価: B+（概ね良好だが整合性の穴あり）**

ドキュメント群とソースコードは全体的に高い整合性を保っている。ただし、requirements.md と実装の間にいくつかの乖離があり、ドキュメント間のクロスリファレンスにも不備がある。

---

## 2. 指摘事項

### L-01: requirements.md にない display_name パラメータが実装されている（重要度: 高）

- **場所**: requirements.md §3.1/3.3 vs internal/mcp/tools.go L33-37, L158-163
- **内容**: `display_name` パラメータは requirements.md の MCP ツール仕様に定義がないが、実装・README には存在する
- **影響**: 要件定義書が Single Source of Truth として機能していない
- **提案**: requirements.md §3.1, §3.3 に `display_name` パラメータを追加。§4.2 の設定仕様にも `display_name` フィールドを追加

### L-02: architecture.md の Config 構造体に display_name がない（重要度: 中）

- **場所**: architecture.md §4.1 の Config struct vs internal/config/config.go L16-21
- **内容**: architecture.md の Config struct 定義に `DisplayName` フィールドがない。実装には存在する
- **影響**: ドキュメント通りに実装しようとすると不整合が発生
- **提案**: architecture.md の Config struct を実装と同期

### L-03: architecture.md の SlackClient インターフェースが4メソッドだが実装と一致（重要度: 低）

- **場所**: architecture.md §4.2 vs internal/slack/types.go
- **内容**: 一致しているが、architecture.md では `HistoryOptions` 型の定義が省略されている
- **影響**: 軽微（型定義は実装を見れば分かる）

### L-04: CONTRIBUTING.md の Go バージョン要件が不一致（重要度: 高）

- **場所**: CONTRIBUTING.md L9 vs go.mod L3
- **内容**: CONTRIBUTING.md は "Go 1.25+" と記載しているが、go.mod は `go 1.25.0` を指定。architecture.md §1 は "Go 1.23+" と記載
- **影響**: コントリビューターが誤ったGoバージョンで開発を始める可能性
- **提案**: go.mod の値を正とし、全ドキュメントを `Go 1.25+` に統一

### L-05: testing-strategy.md の CI ワークフロー例と実際の ci.yml が不一致（重要度: 中）

- **場所**: testing-strategy.md §5.1 vs .github/workflows/ci.yml
- **内容**: 
  - testing-strategy.md は Go 1.23/1.24 のマトリクスを例示
  - 実際の ci.yml は `go-version-file: go.mod` で go.mod から読み取り
  - testing-strategy.md は lint + test の2ジョブだが、実際は lint + test + build の3ジョブ
- **影響**: テスト戦略書が現状を正確に反映していない
- **提案**: testing-strategy.md の CI例を実際の ci.yml と同期

### L-06: requirements.md の設定構造体に log_level があるがCLI出力に反映されていない（重要度: 低）

- **場所**: requirements.md §4.2/4.3 vs internal/config/config.go
- **内容**: `log_level` フィールドは設定に存在し読み込まれるが、実際のログ出力制御ロジックが未実装（stderr への WARNING 以外のログ出力箇所がない）
- **影響**: log_level を設定しても効果が確認しにくい。debug/info レベルのログが一切出ない
- **提案**: 将来的なログ基盤整備（slog 等）を検討。現時点では README にログ機能が限定的である旨を注記

### L-07: カバレッジ閾値の不一致（重要度: 中）

- **場所**: testing-strategy.md §1.2 vs Makefile L9 vs .github/workflows/ci.yml L38
- **内容**:
  - testing-strategy.md: 全体 75%+
  - Makefile: `COVERAGE_THRESHOLD := 65`
  - ci.yml: `< 65` で失敗
  - milestones.md M3: 67.3%（75%未達だがパス判定）
- **影響**: 品質基準が曖昧。テスト戦略書の基準を満たしていないのに品質ゲートが通過する
- **提案**: テスト戦略書の目標値を65%に修正するか、Makefile/CIの閾値を75%に引き上げ。どちらかに統一

### L-08: OS名の表記が GoReleaser の出力と README のダウンロードURLで大文字小文字が混在（重要度: 低）

- **場所**: .goreleaser.yml archives name_template vs README.md L50
- **内容**: GoReleaserは `{{ .Os }}` テンプレートを使用しており、`Darwin`, `Linux`, `Windows` と頭文字大文字で出力する。README のダウンロード URL もこれに合致しているが、`{{- .Os }}_{{- .Arch }}` のテンプレートにハイフンによるスペース除去が含まれ、意図通りに動作するか要確認
- **影響**: 軽微（現時点では動作している）
