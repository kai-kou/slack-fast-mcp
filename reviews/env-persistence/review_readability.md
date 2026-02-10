# ⑤ 可読性レビュー: 環境変数の永続化手順追記

**レビュー日**: 2026-02-10
**レビュー対象**: README.md, README_ja.md, docs/slack-app-setup.md, docs/architecture.md
**レビュー範囲**: E-03指摘対応（環境変数の永続化手順追記）

---

## 評価サマリー

| 観点 | 評価 | コメント |
|---|---|---|
| 情報密度 | ⚠️ 注意 | README の Blockquote 内にやや情報が詰まりすぎ |
| 文長・文構成 | ⭕ 良好 | 各セクションの文は簡潔 |
| 構成・ナビゲーション | ⭕ 良好 | README → Setup Guide の2段階構成が明確 |
| 冗長性 | ⚠️ 軽微 | 一部の説明が README と Setup Guide で重複 |

**総合評価: B+（概ね良好、情報密度の調整で改善可能）**

---

## 詳細レビュー

### R-E01: README §3 の Important ブロックの情報密度 ⚠️ 中

README.md L137-152 の `> Important` ブロックは以下の要素を含む:
1. 「セッション限定」の警告文（1行）
2. シェル別テーブル（4行）
3. コマンド例（3行）
4. `source` での反映手順（1行）
5. Setup Guide §5.1 へのリンク（1行）

合計で**約16行の Blockquote** になっている。README の Quick Start セクションにおいて、1つのBlockquoteにこれだけの情報を詰めると、**視覚的に重たい印象**を与える。

**推奨**: 以下の2パターンのいずれかで軽量化を検討:

**パターンA（推奨）**: テーブルとコマンド例を削除し、Setup Guide への誘導を強化
```markdown
> **Important:** The `export` command above only sets the variable for the current terminal session.
> To persist it across sessions and for AI editors to pick it up, add the export lines to your shell profile.
> See the [Setup Guide §5.1](./docs/slack-app-setup.md#51-方法a-環境変数で設定推奨) for step-by-step instructions.
```

**パターンB**: 現状維持（README だけで完走できるメリットを重視）

**判断**: パターンBの「READMEだけで完走」という設計思想も妥当。ただし、情報が多い分、初見読者が「長い」と感じてスキップするリスクがある。

### R-E02: Setup Guide §5.1 の構成 ✅

4ステップ構成は可読性が高い:
- 各ステップに明確なタイトル
- コマンド例がコピペ可能
- `<details>` で Linux/Windows を折りたたみ

特に macOS をメインに表示し、Linux/Windows を折りたたむ構成は、**大多数のユーザーにとって必要な情報が最初に表示される**点で優れている。

### R-E03: `~/.zprofile` vs `~/.zshrc` の注記の位置 ✅

Setup Guide §5.1 ステップ1のテーブル直後に配置されている。これはユーザーが「どのファイルを編集するか」を決めた直後に読む位置であり、適切。

ただし、この注記は**かなり技術的**（ログインシェル vs インタラクティブシェル）であり、初見ユーザーの一部は理解できない可能性がある。

**評価**: ターゲット読者（開発者）を考慮すると、この技術レベルは許容範囲内。

### R-E04: コマンド例のコメント ✅

```bash
# 例: ~/.zprofile に追記（macOS + zsh の場合）
echo "export SLACK_BOT_TOKEN='xoxb-your-token-here'" >> ~/.zprofile
```

コメントで「何のコマンドか」「どのOSか」を示しているのは良い。特に `xoxb-your-token-here` というプレースホルダーは、**実際のトークンに置き換える必要がある**ことが明確。

### R-E05: architecture.md の変更量 ✅

architecture.md の変更は L434 の1行のみ:
```markdown
あらかじめ環境変数 `SLACK_BOT_TOKEN` をシェルプロファイルに設定した上で（設定手順は [slack-app-setup.md §5.1](...) を参照）:
```

最小限の変更で、開発者ドキュメントから設定手順への導線を追加している。適切。

### R-E06: 英語版と日本語版の表現品質の差 ⚠️ 軽微

英語版 README L137:
> The `export` command above only sets the variable for the current terminal session. To persist it across sessions (and for AI editors like Cursor to pick it up), **add the export lines to your shell profile**:

日本語版 README_ja.md L137:
> 上記の `export` コマンドは現在のターミナルセッションのみ有効です。セッション終了後も維持し、Cursor などの AI エディタから参照できるようにするには、**シェルプロファイルに追記**してください:

日本語版は英語版の忠実な翻訳であり品質は高い。ただし、日本語版のほうが若干冗長（「セッション終了後も維持し、Cursor などの AI エディタから参照できるようにするには」）。

**推奨**: 日本語版を少し圧縮しても良いが、現状でも十分許容範囲。

---

## 指摘一覧

| ID | 重要度 | 内容 | 対象ファイル |
|---|---|---|---|
| R-E01 | 中 | README の Important ブロック内の情報密度が高い | README.md, README_ja.md |
| R-E02 | — | Setup Guide の構成は良好。問題なし | — |
| R-E03 | — | 技術注記の位置は適切。問題なし | — |
| R-E04 | — | コマンド例のコメントは適切。問題なし | — |
| R-E05 | — | architecture.md の変更は最小限で適切。問題なし | — |
| R-E06 | 低 | 日本語版の表現が若干冗長（許容範囲内） | README_ja.md |
