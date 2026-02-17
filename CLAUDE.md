# slack-fast-mcp

最速のSlack MCPサーバー（Go言語、約10ms起動）。Claude Desktop / Claude Code から高速にSlackを操作可能。OSSプロジェクト（MIT License）

## 技術スタック

- Go 1.25+
- MCP (Model Context Protocol) - `github.com/mark3labs/mcp-go v0.43.2`
- Slack API - `github.com/slack-go/slack v0.17.3`
- Makefile（ビルド自動化）
- GitHub Actions（CI/CD）

## ディレクトリ構造

```
/Users/kai.ko/dev/01_active/slack-fast-mcp/
├── cmd/                     # エントリーポイント
│   └── slack-fast-mcp/
│       └── main.go
├── internal/                # 内部パッケージ
│   ├── server/              # MCPサーバー実装
│   ├── slack/               # Slack API連携
│   └── config/              # 設定管理
├── docs/                    # ドキュメント
│   ├── ARCHITECTURE.md      # アーキテクチャ設計
│   ├── DEVELOPMENT.md       # 開発ガイド
│   └── API.md               # API仕様
├── scripts/                 # ビルド・テストスクリプト
├── testdata/                # テストデータ
├── .github/workflows/       # CI/CD定義
├── Makefile                 # ビルドタスク定義
├── .golangci.yml            # Linter設定
├── .goreleaser.yml          # リリース自動化設定
├── go.mod / go.sum          # Go依存管理
├── README.md                # 英語版README
└── README_ja.md             # 日本語版README
```

## 開発ルール

### ビルド・テスト

- `make build`: バイナリビルド（`./slack-fast-mcp`）
- `make test`: テスト実行
- `make lint`: golangci-lint実行（CI相当）
- `make coverage`: カバレッジレポート生成
- `make install`: ローカルインストール（`~/.local/bin/`）

### コード品質

- golangci-lint の警告はすべて解消してからコミットする
- テストカバレッジは80%以上を維持する（`.testcoverage.yml`）
- 新規機能追加時は必ずテストを書く
- `internal/` パッケージは外部公開しない設計を維持する

### Git運用

- ブランチ戦略: `main` のみ（小規模OSSプロジェクト）
- コミットメッセージ: Conventional Commits形式推奨（`feat:`, `fix:`, `docs:` 等）
- CI: GitHub Actions で自動テスト・Lint・ビルド検証
- リリース: Git tag push で自動リリース（GoReleaser）

### パフォーマンス要件

- 起動時間: 10ms以下を維持する
- メモリ使用量: 常駐時10MB以下
- Slack API呼び出しレイテンシ: 100ms以下（ネットワーク除く）

## MCP Tools（提供機能）

- `slack_post_message`: メッセージ投稿
- `slack_list_channels`: チャンネル一覧取得
- `slack_get_channel_history`: メッセージ履歴取得
- `slack_search_messages`: メッセージ検索
- `slack_add_reaction`: リアクション追加
- `slack_get_user_info`: ユーザー情報取得

## ローカル開発セットアップ

1. Slack App作成・Bot Token取得（`xoxb-...`）
2. 環境変数設定: `export SLACK_BOT_TOKEN="xoxb-..."`
3. `make build && make test`
4. Claude Desktop の `claude_desktop_config.json` に登録

## コントリビューション

- Issue・PRは英語でも日本語でも可
- セキュリティ関連の問題は非公開で報告（GitHub Security Advisory）
- OSSライセンス（MIT）に同意した上でコントリビュートする
