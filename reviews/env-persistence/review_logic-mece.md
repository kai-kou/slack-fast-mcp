# ② 論理・MECEレビュー: 環境変数の永続化手順追記

**レビュー日**: 2026-02-10
**レビュー対象**: README.md, README_ja.md, docs/slack-app-setup.md, docs/architecture.md
**レビュー範囲**: E-03指摘対応（環境変数の永続化手順追記）

---

## 評価サマリー

| 観点 | 評価 | コメント |
|---|---|---|
| 論理的整合性 | ⭕ 良好 | 4ファイル間の記述に矛盾なし |
| MECE検証 | ⚠️ 一部不足 | fish shell等の未対応シェルへの言及なし |
| 根拠の妥当性 | ⭕ 良好 | シェルプロファイルの選定理由が明示されている |
| 英語版・日本語版の整合性 | ⚠️ 一部不一致 | 複数箇所で軽微な不一致あり |

**総合評価: B+（概ね良好、軽微な問題あり）**

---

## 詳細レビュー

### L-E01: 英語版・日本語版の構造的整合性 ✅

README.md と README_ja.md の修正部分は、以下の点で整合が取れている：
- `> Important` / `> 重要` の格上げが両方で実施
- シェル別テーブル構造が同一
- コマンド例が同一
- §5.1 への参照リンクが同一アンカー

### L-E02: リンク先の正確性検証 ⚠️ 要確認

README.md L152:
```markdown
[Slack App Setup Guide §5.1](./docs/slack-app-setup.md#51-方法a-環境変数で設定推奨)
```

README_ja.md L152:
```markdown
[Slack App セットアップガイド §5.1](./docs/slack-app-setup.md#51-方法a-環境変数で設定推奨)
```

**問題**: アンカー `#51-方法a-環境変数で設定推奨` は日本語を含むアンカーリンクである。GitHubのMarkdownレンダリングでは、見出し `### 5.1 方法A: 環境変数で設定（推奨）` から自動生成されるアンカーは:
- 全角文字のエンコーディングに依存
- GitHub上では `#51-方法a-環境変数で設定推奨` で機能する可能性が高いが、**括弧の処理**が環境依存

**推奨**: GitHub上でのプレビューでリンクの動作を実際に確認すること。

### L-E03: シェル対応のMECE性 ⚠️ 軽微

現在の対応:
| シェル | 対応状況 |
|---|---|
| zsh | ✅ `~/.zprofile` |
| bash | ✅ `~/.bash_profile` |
| fish | ❌ 未記載 |
| PowerShell (Windows) | ✅ `[Environment]::SetEnvironmentVariable` |

**fish shell** はAIエディタ（特にCursor）のユーザー層で一定の利用者がいる。完全なMECEを目指すなら `~/.config/fish/config.fish` への言及があると良いが、zsh/bashで95%以上をカバーしているため**優先度は低い**。

### L-E04: `~/.zprofile` vs `~/.zshrc` の説明の論理性 ✅

Setup Guide §5.1 ステップ1の注記:
> `~/.zprofile` はログインシェル起動時に一度だけ読み込まれるファイルで、環境変数の設定に適しています。`~/.zshrc` はインタラクティブシェル起動時に毎回読み込まれます。

この説明は技術的に正確であり、「どちらに書いても動作する」と明言した上で推奨を示している点が良い。ただし、README側のテーブルでは:

```
| **zsh** (macOS default) | `~/.zprofile` or `~/.zshrc` | `echo $SHELL` shows `/bin/zsh` |
```

と「or」で併記しているのに対し、Setup Guide のステップ1テーブルでは:

```
| `/bin/zsh` | zsh（macOS デフォルト） | `~/.zprofile` |
```

と `~/.zprofile` のみを記載している。

**問題**: README では「~/.zprofile or ~/.zshrc」と選択肢を提示しているのに、Setup Guide では `~/.zprofile` に絞っている。初見ユーザーが README → Setup Guide と遷移した際に、「~/.zshrc は使えないの？」と混乱する可能性がある。

**推奨**: Setup Guide のテーブルにも `~/.zprofile`（推奨）の形で記載し、下の注記で `.zshrc` でも動作する旨を示す現在の構成は実は問題ない。ただし、README 側のテーブルでも `~/.zprofile`（推奨）と明示するとより一貫性が出る。

### L-E05: Configuration details セクションの参照追加 ✅

README.md L305:
```markdown
> To persist these across terminal sessions, add `export` lines to your shell profile (`~/.zprofile` for zsh, `~/.bash_profile` for bash). See [Quick Start §3](#3-configure) for details.
```

README_ja.md L305:
```markdown
> これらの環境変数をターミナル再起動後も維持するには、シェルプロファイル（zsh: `~/.zprofile`、bash: `~/.bash_profile`）に `export` 行を追記してください。詳しくは[クイックスタート §3](#3-設定) を参照。
```

内部リンクのアンカーが異なる点に注意:
- 英語版: `#3-configure`
- 日本語版: `#3-設定`

これは各版の見出しに基づいており**正しい**。

### L-E06: architecture.md の参照リンク ✅

architecture.md L434:
```markdown
あらかじめ環境変数 `SLACK_BOT_TOKEN` をシェルプロファイルに設定した上で（設定手順は [slack-app-setup.md §5.1](./slack-app-setup.md#51-方法a-環境変数で設定推奨) を参照）:
```

これは `docs/` ディレクトリ内からの相対パスであり、`./slack-app-setup.md` で正しい。

### L-E07: README の Claude Desktop 設定でのトークン直書き ⚠️ 既存の問題

README.md L184:
```json
"SLACK_BOT_TOKEN": "your-token-here"
```

Claude Desktop の設定例でトークン直書きになっている。これは今回の修正範囲外の既存の問題だが、環境変数の永続化を強調する文脈と矛盾する。

**推奨**: 今回のスコープ外だが、将来的に Claude Desktop でも `${SLACK_BOT_TOKEN}` 参照が可能か調査し、可能であれば統一する。

---

## 指摘一覧

| ID | 重要度 | 内容 | 対象ファイル |
|---|---|---|---|
| L-E01 | — | 構造的整合性は良好 | — |
| L-E02 | 中 | 日本語アンカーリンクの動作確認が必要 | README.md, README_ja.md |
| L-E03 | 低 | fish shell の未記載（カバレッジ95%以上のため低優先） | docs/slack-app-setup.md |
| L-E04 | 低 | README のテーブルで `~/.zprofile`（推奨）と明示すると一貫性向上 | README.md, README_ja.md |
| L-E05 | — | Configuration details の参照は正しい | — |
| L-E06 | — | architecture.md の参照リンクは正しい | — |
| L-E07 | 低 | Claude Desktop 設定のトークン直書き（既存の問題、今回スコープ外） | README.md, README_ja.md |
