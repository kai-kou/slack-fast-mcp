# 修正計画書: slack-fast-mcp プロジェクトドキュメント Ver.1

**作成日**: 2026-02-10
**対象ドキュメント**: requirements.md / architecture.md / slack-app-setup.md
**レビュー手法**: 7軸専門レビュー（戦略/論理・MECE/実行設計/ポジション/可読性/ヒューマンライズ/リスク）
**レビュー結果**: `reviews/` ディレクトリ参照

---

## 修正方針

本修正計画書は、7つの専門レビュー結果を統合し、**18件の修正項目**を3つの優先度に分類したものです。

- **Priority 1（実装開始前に対応）**: 設計に影響する重大な問題。これを修正しないと実装時に手戻りが発生する
- **Priority 2（開発中に対応）**: 実装品質に影響する問題。開発フェーズの中で順次対応
- **Priority 3（公開前に対応）**: ドキュメント品質・OSS公開品質に関する問題。公開時までに対応

---

## Priority 1: 実装開始前に対応（5件）

### P1-1. チャンネルID判定ロジックの修正

| 項目 | 内容 |
|---|---|
| **指摘元** | 論理(L-7), リスク(RK-12) |
| **対象** | architecture.md §4.2 チャンネル名解決ロジック |
| **問題** | 現在「"C" で始まる → チャンネルIDとしてそのまま返す」としているが、SlackのチャンネルIDは `C`（パブリック）、`G`（プライベート/グループ）、`D`（DM）の3種類がある。また、`ci-notifications` のようなチャンネル名が誤判定されないための考慮も不足 |
| **影響** | プライベートチャンネルのIDが `G` 始まりの場合、チャンネル名として `conversations.list` 検索が走り、不要なAPI呼び出しが発生。最悪の場合 `channel_not_found` エラーになる |

**修正内容:**

architecture.md のチャンネル名解決ロジックを以下に修正:

```
入力: channel string
  ├── 正規表現 ^[CGD][A-Z0-9]{8,}$ にマッチ → チャンネルIDとしてそのまま返す
  ├── "#" で始まる → "#" を除去してチャンネル名として検索
  └── その他 → チャンネル名として conversations.list で検索
```

requirements.md §6.4 の記述も同様に修正。

---

### P1-2. トークン保護の設計強化

| 項目 | 内容 |
|---|---|
| **指摘元** | リスク(RK-1, RK-2), 論理(L-9) |
| **対象** | requirements.md §4.5, architecture.md §4.1, slack-app-setup.md §5.3/§6 |
| **問題** | トークンの直書きが技術的に可能であり、「推奨しない」の記述のみ。OSS公開後にユーザーがトークンをGitにコミットする事故が確実に発生する。ドキュメント内にもトークン直書きのサンプルがある（setup.md §5.3、§6.1、§6.2） |
| **影響** | Slackワークスペースへの不正アクセスリスク。OSS公開後の信頼性低下 |

**修正内容:**

1. **requirements.md §4.5 に技術的防御の追加:**
   ```
   - 設定ファイル内のトークンが `xoxb-` 等の実トークン形式に一致する場合:
     - stderr に警告メッセージを出力（「環境変数参照 ${SLACK_BOT_TOKEN} の使用を推奨します」）
     - 将来的に --strict モードでエラー終了も検討
   ```

2. **slack-app-setup.md の修正:**
   - §5.3 グローバル設定のサンプルを `"token": "${SLACK_BOT_TOKEN}"` 形式に変更
   - §6.1, §6.2 のCursor MCP設定内 `"SLACK_BOT_TOKEN": "xoxb-xxxx-xxxx-xxxx"` に「※ここに実際のトークンを設定」のコメントと、「環境変数で設定する場合の代替方法」を追記

3. **architecture.md §7.3 のサンプルにも同様の修正:**
   - 環境変数での設定方法を最初に提示

---

### P1-3. LLM向けツール description の設計指針

| 項目 | 内容 |
|---|---|
| **指摘元** | ポジション(P-8, P-9) |
| **対象** | requirements.md §3, architecture.md §4.3 |
| **問題** | MCP Serverの最大のユーザーはLLM（Cursor/Claude）だが、ツールの `description` が簡素すぎる。LLMがツールを「いつ使うべきか」「何ができるか」「制約は何か」を判断するための情報が不足 |
| **影響** | LLMのツール選択精度が低下し、ユーザー体験が悪化する |

**修正内容:**

1. **requirements.md に「MCP ツール description 設計方針」セクションを追加:**
   ```
   ### MCPツール description の設計方針
   - ツールの目的と使用タイミングを明記
   - デフォルト値の挙動を含める
   - Slack mrkdwn フォーマットの対応状況を記載
   - 制約事項（文字数上限等）を含める
   ```

2. **architecture.md のツール定義サンプルを改善:**
   ```go
   // Before:
   mcp.WithDescription("Post a message to a Slack channel")
   
   // After:
   mcp.WithDescription("Post a message to a Slack channel. " +
       "Supports Slack mrkdwn formatting. " +
       "If channel is omitted, posts to the configured default channel. " +
       "The bot must be invited to the target channel first.")
   ```

3. **エラーHintをLLM向けに最適化する方針を追記:**
   ```
   エラーHint方針:
   - LLMが次のアクションを判断できる具体的な指示を含める
   - 「ユーザーに〜を依頼してください」形式で、LLMがそのまま伝達可能な文面にする
   - 英語で記述（LLMの処理精度向上のため）
   ```

---

### P1-4. MVP定義と開発フェーズの追加

| 項目 | 内容 |
|---|---|
| **指摘元** | 戦略(S-7), 実行(E-8, E-9) |
| **対象** | requirements.md（新規セクション追加） |
| **問題** | 全機能が同列に記述されており、開発順序・MVPの定義がない。「確定」ステータスだが、何を先に作るかが不明確で、実装開始時に判断に迷う |
| **影響** | 開発の手戻り、スコープクリープのリスク |

**修正内容:**

requirements.md に「開発フェーズ」セクションを追加:

```markdown
## 開発フェーズ

### Phase 1: MVP（個人利用）
- Config Layer（設定読み込み + 環境変数展開）
- Slack Client（PostMessage + ResolveChannel）
- MCP Server（slack_post_message ツールのみ）
- 最小限のエラーハンドリング
- **完了基準**: Cursor から #general に投稿できる

### Phase 2: コア機能完成
- Slack Client（GetHistory + PostThread）
- MCP Server（slack_get_history + slack_post_thread）
- CLI Layer（post / history / reply サブコマンド）
- レート制限リトライ
- **完了基準**: 3つのMCPツール + 3つのCLIコマンドが動作

### Phase 3: 品質・配布
- setup コマンド（初期設定ウィザード）
- CI/CD（GitHub Actions + GoReleaser）
- テスト（ユニット + インテグレーション）
- ドキュメント整備（README.md 英語版）
- **完了基準**: GitHub Releases でバイナリ配布可能

### Phase 4: OSS公開
- CONTRIBUTING.md, LICENSE
- セキュリティ監査
- 競合との差別化ポイント明記
- **完了基準**: GitHubで public リポジトリとして公開
```

---

### P1-5. 実装時のグレーゾーン判断の確定

| 項目 | 内容 |
|---|---|
| **指摘元** | 実行(G-1〜G-5) |
| **対象** | requirements.md（新規セクション or 該当箇所に追記） |
| **問題** | 実装時に判断に迷う5つのグレーゾーンが存在 |
| **影響** | 実装者が独自判断で進め、後で修正が必要になるリスク |

**修正内容:**

以下の判断を requirements.md の該当セクションに追記:

| # | 判断事項 | 確定方針 | 追記先 |
|---|---|---|---|
| G-1 | チャンネル未指定 & デフォルト未設定時 | **エラー返却**。メッセージ: 「default_channel を設定するか、channel パラメータを指定してください」 | §3 各ツール仕様 |
| G-2 | トークン未設定時のMCP Server起動 | **起動拒否**（早期失敗）。具体的なエラーメッセージでsetup手順を案内 | §7 エラーハンドリング |
| G-3 | conversations.list でチャンネルが見つからない場合 | **全ページ検索後にエラー**。ただし最大1000チャンネル（ページネーション上限5回）で打ち切り | §6.4 チャンネル名解決 |
| G-4 | MCP ツール結果JSONのエンコーディング | **UTF-8そのまま**（日本語をエスケープしない）。LLMの可読性を優先 | §3 ツール仕様 出力 |
| G-5 | CLI の `--json` 未指定時のデフォルト出力 | **テーブル形式**（視認性重視）。`--json` でJSON出力に切り替え | §5 CLI コマンド仕様 |

---

## Priority 2: 開発中に対応（7件）

### P2-1. ログのトークンマスキング設計

| 項目 | 内容 |
|---|---|
| **指摘元** | リスク(RK-4) |
| **対象** | architecture.md §6 エラーハンドリング設計（新規追記） |
| **修正内容** | debug ログ出力時のトークンマスキングルールを追加。`xoxb-` / `xoxp-` / `xoxs-` で始まる文字列を `xoxb-****` 形式にマスキング。HTTP リクエストヘッダの `Authorization` ヘッダも同様にマスキング |

### P2-2. context.Context の利用方針

| 項目 | 内容 |
|---|---|
| **指摘元** | 実行(E-2) |
| **対象** | architecture.md §4.2 Slack Client（追記） |
| **修正内容** | MCP Serverモード: リクエストタイムアウト30秒、CLIモード: コマンドタイムアウト10秒。Slack API個別呼び出しタイムアウト: 10秒。SIGINT/SIGTERM受信時にcontext cancelを伝播 |

### P2-3. エラーコード一覧の一元化

| 項目 | 内容 |
|---|---|
| **指摘元** | 実行(E-3), 可読性(R-8) |
| **対象** | requirements.md §7（再構成） |
| **修正内容** | §3 の各ツール仕様に散在するエラーケースを §7 のエラーコードマスターテーブルに集約。各ツール仕様からは「詳細は §7 参照」で参照。マスターテーブルには Code / Message / Hint / 対象ツール の列を含める |

### P2-4. Windows対応の設計詳細

| 項目 | 内容 |
|---|---|
| **指摘元** | 論理(L-5, L-6) |
| **対象** | architecture.md（新規セクション or §4.1 追記） |
| **修正内容** | グローバル設定ファイルパス: Windows では `%APPDATA%\slack-fast-mcp\config.json`。Goの `os.UserConfigDir()` を使用してクロスプラットフォーム対応。インストール手順: Windows向けに PowerShell でのダウンロード・配置手順を追加 |

### P2-5. conversations.list ページネーション仕様

| 項目 | 内容 |
|---|---|
| **指摘元** | 論理(L-4) |
| **対象** | architecture.md §4.2 チャンネル名解決ロジック（追記） |
| **修正内容** | conversations.list のページネーション処理: `cursor` パラメータを使用して全ページを取得。1リクエストあたり200チャンネル。最大5ページ（1000チャンネル）で打ち切り。見つからない場合はエラー返却 |

### P2-6. setup コマンドの対話フロー設計

| 項目 | 内容 |
|---|---|
| **指摘元** | 実行(E-1) |
| **対象** | requirements.md §5.2 setup コマンド（拡充） |
| **修正内容** | 以下の対話フローを定義: (1) Slack App作成済みか確認 → 未作成なら作成手順を表示 → (2) Bot Tokenの入力（`xoxb-` 形式のバリデーション） → (3) デフォルトチャンネルの入力（空欄可） → (4) `.slack-mcp.json` の生成 → (5) `.gitignore` への追記確認 → (6) Cursor MCP設定の案内表示 |

### P2-7. graceful shutdown の設計

| 項目 | 内容 |
|---|---|
| **指摘元** | リスク(BL-5) |
| **対象** | architecture.md §5 起動フロー（追記） |
| **修正内容** | MCP Serverモードの終了処理: SIGTERM/SIGINT受信 → context cancel → 進行中のSlack API呼び出し完了待ち（最大5秒） → プロセス終了。`signal.NotifyContext` を使用 |

---

## Priority 3: 公開前に対応（6件）

### P3-1. OSS公開戦略・競合分析

| 項目 | 内容 |
|---|---|
| **指摘元** | 戦略(S-1, S-2), ヒューマンライズ(H-5) |
| **対象** | requirements.md（新規セクション） |
| **修正内容** | 「背景と動機」セクション: なぜ自作するのか（既存ツールの速度問題等）。「競合分析」セクション: 既存Slack MCPとの比較表（速度、依存、機能）。「OSS公開方針」セクション: 公開判断基準、メンテナンス体制、期待するコミュニティ参加の形 |

### P3-2. 重複コンテンツの整理

| 項目 | 内容 |
|---|---|
| **指摘元** | 可読性(R-7〜R-10), 論理(L-10, L-11) |
| **対象** | 全ドキュメント |
| **修正内容** | Single Source of Truth の定義: 設定優先順位 → requirements.md が正、エラーコード → requirements.md §7 が正、Cursor MCP設定例 → slack-app-setup.md が正、Bot Token Scopes → requirements.md §6.2 が正。他のドキュメントからは「詳細は〇〇参照」の形式で参照 |

### P3-3. ドキュメントの読者定義・ナビゲーション

| 項目 | 内容 |
|---|---|
| **指摘元** | ポジション(P-10), 可読性(R-6) |
| **対象** | docs/ ディレクトリ全体 |
| **修正内容** | 各ドキュメントの冒頭に「対象読者」を追記。docs/README.md（またはrequirements.mdの冒頭）にドキュメント構成マップと推奨読み順を追加。ドキュメントステータスの定義（ドラフト/レビュー中/確定/更新中） |

### P3-4. プロジェクト動機・背景の追記

| 項目 | 内容 |
|---|---|
| **指摘元** | ヒューマンライズ(H-5, H-8) |
| **対象** | requirements.md §1 の前に新規セクション |
| **修正内容** | 「背景」セクションを追加。プロジェクトの動機（例: 既存ツールの起動速度への不満、Cursor + Slack連携の日常的ニーズ等）を1-2段落で記述。技術的な背景だけでなく、個人的な体験・課題感を含める |

### P3-5. ADRのトレードオフ補強

| 項目 | 内容 |
|---|---|
| **指摘元** | ヒューマンライズ(H-4, H-6) |
| **対象** | architecture.md §9 ADR |
| **修正内容** | 各ADRに「デメリット・受容したトレードオフ」フィールドを追加。例: ADR-002「独自JSON → viperの充実した機能（YAML/TOML対応、ホットリロード等）を放棄するトレードオフ」。ADR-001「cobra → バイナリサイズが増加するトレードオフ。標準flagに比べて約2MB増」 |

### P3-6. macOS Gatekeeper / パッケージマネージャ対応

| 項目 | 内容 |
|---|---|
| **指摘元** | リスク(BL-2, BL-4) |
| **対象** | architecture.md §7 ビルド・配布設計（追記） |
| **修正内容** | macOS での Gatekeeper 警告への対応方針（READMEに `xattr -d com.apple.quarantine` 手順を記載）。Homebrew tap の作成検討。将来的なコードサイニングの方針 |

---

## 修正作業見積もり

| 優先度 | 件数 | 推定工数 | 対応時期 |
|---|---|---|---|
| Priority 1 | 5件 | 約2-3時間 | 実装開始前（即時） |
| Priority 2 | 7件 | 約3-4時間 | 開発フェーズ中（Phase 1〜2と並行） |
| Priority 3 | 6件 | 約2-3時間 | Phase 3 完了後〜公開前 |
| **合計** | **18件** | **約7-10時間** | - |

---

## 参考: レビュー結果ファイル一覧

```
reviews/
├── review_strategy.md        ← ① 戦略
├── review_logic-mece.md      ← ② 論理・MECE
├── review_execution.md       ← ③ 実行設計
├── review_perspective.md     ← ④ ポジション視点
├── review_readability.md     ← ⑤ 可読性
├── review_humanize.md        ← ⑥ ヒューマンライズ
├── review_risk.md            ← ⑦ リスク
└── 統合レビュー結果.md
```
