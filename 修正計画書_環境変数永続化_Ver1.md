# 修正計画書: 環境変数の永続化手順追記

**作成日**: 2026-02-10
**バージョン**: Ver.1
**対象ドキュメント**: README.md, README_ja.md, docs/slack-app-setup.md, docs/architecture.md
**ステータス**: レビュー完了・修正対応待ち

---

## 修正方針

本修正計画書は、「環境変数の永続化手順追記」（修正計画書 Ver.1 指摘 E-03 対応）に対する7軸レビュー結果に基づき、**追加修正が必要な項目**を定義する。

### 全体評価

**総合: B+ ～ A-（良好。公開品質を満たしている）**

- Must Fix: **0件** — 現状で公開可能
- Should Fix: **3件** — 公開前に対応推奨
- Nice to Have: **8件** — 余裕があれば対応

---

## Should Fix（修正推奨）: 3件

### SF-01: Cursor での環境変数展開メカニズムの説明補足

| 項目 | 内容 |
|---|---|
| **関連指摘** | S-E04, RK-E04 |
| **対象** | docs/slack-app-setup.md §5.1 ステップ4の注記 |
| **問題** | 「Cursor を再起動すると反映されます」とあるが、なぜ再起動が必要かの説明がない。ユーザーが「source したのに Cursor で使えない」と混乱する可能性がある |
| **影響** | 設定後に「動かない」と感じるユーザーが発生し、issue/質問が増える |

**修正内容（docs/slack-app-setup.md §5.1 ステップ4の注記）**:

現在:
```markdown
> **AI エディタ（Cursor / Windsurf）での利用:** Cursor の MCP 設定（`.cursor/mcp.json`）で `"${SLACK_BOT_TOKEN}"` と記述すると、ここで設定した環境変数が自動的に参照されます。新しいターミナルを起動してから Cursor を再起動すると反映されます。詳しくは [§6](#6-cursor-mcp-設定) を参照してください。
```

修正後:
```markdown
> **AI エディタ（Cursor / Windsurf）での利用:** Cursor の MCP 設定（`.cursor/mcp.json`）で `"${SLACK_BOT_TOKEN}"` と記述すると、ここで設定した環境変数が自動的に参照されます。Cursor は起動時にシェルの環境変数を読み込むため、プロファイル変更後は**ターミナルを新しく開いてから Cursor を再起動**してください。詳しくは [§6](#6-cursor-mcp-設定) を参照してください。
```

**日本語版は上記がベースのため、英語版 README.md は対応不要**（この注記は Setup Guide のみに存在）。

---

### SF-02: 日本語アンカーリンクの動作確認

| 項目 | 内容 |
|---|---|
| **関連指摘** | L-E02 |
| **対象** | README.md L152, README_ja.md L152, architecture.md L434, slack-app-setup.md §5.2/§5.3 |
| **問題** | `#51-方法a-環境変数で設定推奨` という日本語を含むアンカーリンクが4箇所で使用されている。GitHub の Markdown レンダリングでは動作するはずだが、括弧 `（）` の処理が環境依存 |
| **影響** | リンク切れの場合、ユーザーが詳細設定手順に到達できない |

**修正内容**:

1. GitHub にプッシュ後、4箇所のリンクが正しく動作するか確認
2. 動作しない場合は、Setup Guide §5.1 の見出しを英語ベースに変更:
   ```markdown
   ### 5.1 Method A: Environment Variables (Recommended)
   ```
   またはアンカーを手動で設定する HTML を挿入:
   ```markdown
   ### 5.1 方法A: 環境変数で設定（推奨）ÿ{#env-var-setup}
   ```

**注**: GitHub上での実際の動作確認が必要なため、プッシュ後のタスクとする。

---

### SF-03: `chmod 600` の適用条件を限定

| 項目 | 内容 |
|---|---|
| **関連指摘** | E-E04 |
| **対象** | docs/slack-app-setup.md §5.1 セキュリティ注意 |
| **問題** | `chmod 600 ~/.zprofile` を一律に推奨しているが、個人PCでは不要な場合が多い。共有サーバー等マルチユーザー環境でのみ必要 |
| **影響** | 不必要なパーミッション変更を行うユーザーが発生する可能性（実害は軽微） |

**修正内容（docs/slack-app-setup.md §5.1 セキュリティ注意）**:

現在:
```markdown
> **セキュリティ上の注意:** トークンをシェルプロファイルに記述する場合、ファイルのパーミッションが適切であることを確認してください（`chmod 600 ~/.zprofile`）。また、dotfiles リポジトリで管理している場合はトークンが含まれないよう注意してください。
```

修正後:
```markdown
> **セキュリティ上の注意:** 共有サーバーなどマルチユーザー環境では、シェルプロファイルのパーミッションを確認してください（`chmod 600 ~/.zprofile`）。また、dotfiles を GitHub 等の公開リポジトリで管理している場合は、トークンが含まれるファイルを除外するよう注意してください。
```

**変更ポイント**:
- `chmod 600` の適用条件を「共有サーバーなどマルチユーザー環境」に限定
- dotfiles の注意を「公開リポジトリで管理」と具体化

**英語版 README.md / README_ja.md**: この注記は Setup Guide にのみ存在するため、README の変更は不要。

---

## Nice to Have（余裕があれば対応）: 8件

| # | ID | 内容 | 対象 | 推定工数 |
|---|---|---|---|---|
| 1 | L-E04 | README テーブルで `~/.zprofile`（推奨）を明示 | README.md, README_ja.md | 5分 |
| 2 | L-E03 | fish shell への軽微な言及（`~/.config/fish/config.fish`） | docs/slack-app-setup.md | 10分 |
| 3 | H-E03 | ステップ4で設定確認失敗時のフォールバック案内 | docs/slack-app-setup.md | 5分 |
| 4 | H-E07 | dotfiles リスクの「公開リポジトリに漏洩」という具体的説明 | docs/slack-app-setup.md | SF-03で対応済み |
| 5 | RK-E02 | より安全な代替手段（direnv, 1password CLI等）への簡単な言及 | docs/slack-app-setup.md | 15分 |
| 6 | RK-E03 | `>>` 重複追記への「既に記述済みの場合は不要」の注記 | docs/slack-app-setup.md | 5分 |
| 7 | RK-E05 | トークンローテーション時に「プロファイルも更新」の言及 | docs/slack-app-setup.md | 5分 |
| 8 | R-E01 | README の Important ブロックの軽量化検討 | README.md, README_ja.md | — (現状維持推奨) |

### Nice to Have の修正例

#### NH-01: README テーブルで推奨を明示（L-E04）

現在:
```markdown
| **zsh** (macOS default) | `~/.zprofile` or `~/.zshrc` | ...
```

修正後:
```markdown
| **zsh** (macOS default) | `~/.zprofile` (recommended) or `~/.zshrc` | ...
```

日本語版:
```markdown
| **zsh**（macOS デフォルト） | `~/.zprofile`（推奨）または `~/.zshrc` | ...
```

#### NH-03: ステップ4 のフォールバック案内（H-E03）

Setup Guide §5.1 ステップ4に追加:
```markdown
> 何も表示されない場合は、ステップ2で正しいファイルに追記されているか確認してください。
> ```bash
> grep SLACK_BOT_TOKEN ~/.zprofile  # 追記されているか確認
> ```
```

#### NH-06: 重複追記への注意（RK-E03）

Setup Guide §5.1 ステップ2のコマンド例の後に追加:
```markdown
> ※ 上記コマンドを複数回実行すると同じ行が重複します。エディタで直接編集する方法もおすすめです。
```

---

## 修正対象ファイルと変更箇所の一覧

| ファイル | Should Fix | Nice to Have | 合計 |
|---|---|---|---|
| docs/slack-app-setup.md | 2件 (SF-01, SF-03) | 6件 | 8件 |
| README.md | 0件 | 1件 (NH-01) | 1件 |
| README_ja.md | 0件 | 1件 (NH-01) | 1件 |
| docs/architecture.md | 0件 | 0件 | 0件 |
| **全ファイル（リンク確認）** | 1件 (SF-02) | 0件 | 1件 |

---

## 修正作業見積もり

| 分類 | 件数 | 推定工数 | 対応時期 |
|---|---|---|---|
| Should Fix | 3件 | 30分 | 公開前推奨 |
| Nice to Have | 8件 | 45分（#8除く） | 余裕があれば |
| **合計** | **11件** | **約1.5時間** | — |

---

## 同期チェックリスト

修正実施時に確認すること:

- [ ] SF-01: Setup Guide §5.1 ステップ4の注記が更新されているか
- [ ] SF-02: GitHub プッシュ後に4箇所の日本語アンカーリンクが動作するか
- [ ] SF-03: Setup Guide §5.1 セキュリティ注意の表現が更新されているか
- [ ] README.md と README_ja.md の修正が同期しているか（NH-01 対応時）
- [ ] 修正後、他のセクションとの整合性が保たれているか

---

## 参照レビュー結果

| レビュー結果 | ファイル |
|---|---|
| ① 戦略 | `reviews/env-persistence/review_strategy.md` |
| ② 論理・MECE | `reviews/env-persistence/review_logic-mece.md` |
| ③ 実行設計 | `reviews/env-persistence/review_execution.md` |
| ④ ポジション | `reviews/env-persistence/review_perspective.md` |
| ⑤ 可読性 | `reviews/env-persistence/review_readability.md` |
| ⑥ ヒューマンライズ | `reviews/env-persistence/review_humanize.md` |
| ⑦ リスク | `reviews/env-persistence/review_risk.md` |
| 統合レビュー | `reviews/env-persistence/統合レビュー結果.md` |
