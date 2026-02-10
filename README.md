---
project:
  name: "slack-fast-mcp"
  title: "Slack Fast MCP Server"
  status: active
  priority: high
  created: "2026-02-10"
  updated: "2026-02-10"
  owner: "kai.ko"
  tags: [slack, mcp, go, cli, cursor]
  summary: "実行速度最優先のSlack投稿・閲覧MCP Server（Go実装）"
  next_action: "要件定義・技術設計"
---

# Slack Fast MCP Server

## 概要

Cursorをはじめとする MCP 対応クライアントから、Slackワークスペースの特定チャンネルに対して高速に投稿・閲覧・スレッド投稿を行うためのMCPサーバー。Go言語で実装し、実行速度・起動速度を最優先とする。CLIとしても利用可能で、チームメンバーやOSSコミュニティへの公開を視野に入れる。

## ゴール

- [ ] Slack APIを利用した高速なメッセージ投稿・閲覧・スレッド投稿機能
- [ ] MCP Server として Cursor から seamless に利用可能
- [ ] CLI ツールとしてユーザー（人）も直接利用可能
- [ ] Mac / Windows クロスプラットフォーム対応
- [ ] プロジェクト（ワークディレクトリ）ごとのワークスペース・チャンネル設定
- [ ] 明快な初期設定ガイド（ユーザーが迷わない手順案内）
- [ ] GitHub パブリック公開に耐えうる品質・ドキュメント

## スコープ

### 含むもの
- Slack Web API を利用したメッセージ投稿（chat.postMessage）
- チャンネルのメッセージ履歴取得（conversations.history）
- スレッドへの返信投稿（chat.postMessage with thread_ts）
- パラメータによるチャンネル指定
- MCP Server プロトコル実装（stdio transport）
- CLI モード（人間が直接コマンドで利用）
- プロジェクトローカル設定ファイル（.slack-mcp.json 等）
- セットアップウィザード / ガイド機能
- クロスプラットフォームバイナリ配布（GitHub Releases）

### 含まないもの
- Slack Bot としてのリアルタイム受信（Events API / Socket Mode）
- ファイルアップロード機能（初期スコープ外）
- Slack App の管理画面 UI
- Web UI ダッシュボード

## 技術スタック（想定）

| 要素 | 選定 | 理由 |
|------|------|------|
| 言語 | Go | 実行速度最優先、シングルバイナリ、クロスプラットフォーム |
| MCP SDK | mcp-go | Go製MCP SDK |
| Slack API | slack-go/slack | Go製Slack APIクライアント |
| 設定管理 | viper / 独自JSON | プロジェクトローカル設定対応 |
| ビルド・配布 | GoReleaser + GitHub Actions | クロスコンパイル・自動リリース |

## 関連リソース

- マイルストーン: [milestones.md](./milestones.md)
- タスク一覧: [tasks.md](./tasks.md)
- 要件定義: [docs/requirements.md](./docs/requirements.md)
- 技術設計書: [docs/architecture.md](./docs/architecture.md)
- Slack App セットアップ: [docs/slack-app-setup.md](./docs/slack-app-setup.md)
- ライセンス: [MIT License](./LICENSE)

## メモ

- 実行速度を最優先とするため、Go言語でのシングルバイナリ実装を選択
- `npx` 等のランタイム依存を排除し、ダウンロード即実行を実現
- Cursor の MCP 設定（`.cursor/mcp.json`）に追加するだけで利用開始可能にする
- プロジェクトルートに `.slack-mcp.json` を置くことで、プロジェクトごとのワークスペース・チャンネルをデフォルト指定可能にする
