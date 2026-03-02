# UI Review Skill

MCP Playwright を使って UI の変更を検証するスキル。
ページ表示・コンソールエラー・レスポンシブ・インタラクション・デザインシステム準拠をチェックする。

## 前提条件

- Next.js dev server が起動済み (`npm run dev` or `make up`)
- 対象ページの URL がわかっている

## 検証フロー

### Step 1: ページ表示 + コンソールエラー

```
1. browser_navigate → 対象 URL
2. browser_snapshot → ページ構造を確認 (要素が正しく描画されているか)
3. browser_console_messages(level: "error") → エラーがないこと
```

**判定基準**:
- スナップショットにページの主要コンテンツが含まれている
- コンソールに `error` レベルのメッセージがない
- (warning は許容するが、明らかな問題は報告)

### Step 2: レスポンシブ確認

3 サイズで確認する:

| Name | Width | Height | 確認ポイント |
|------|-------|--------|------------|
| Mobile | 375 | 812 | ナビゲーション折り畳み、カラム縦積み |
| Tablet | 768 | 1024 | 2カラムグリッドの適用 |
| Desktop | 1280 | 800 | フルレイアウト表示 |

```
各サイズで:
1. browser_resize(width, height)
2. browser_snapshot → レイアウト崩れがないか
3. browser_console_messages(level: "error") → リサイズ起因のエラーがないか
```

**判定基準**:
- コンテンツがはみ出していない (水平スクロール不要)
- テーブルが適切に対応している (スクロール or レスポンシブ)
- ボタン・入力欄がタップ可能なサイズを維持

### Step 3: インタラクション確認

対象ページの主要操作をテストする:

```
1. browser_snapshot → クリック可能な要素を特定
2. browser_click → ボタン・リンク等を操作
3. browser_snapshot → 操作後の状態を確認
4. browser_console_messages(level: "error") → 操作起因のエラーがないか
```

**テスト対象の優先順位**:
1. ページ内のプライマリアクション (作成ボタン, 送信ボタン等)
2. Dialog / Sheet の開閉
3. テーブルの行アクション
4. ナビゲーションリンク

**判定基準**:
- クリック後に期待する UI 変化がある (Dialog 表示, ページ遷移等)
- 操作中にコンソールエラーが出ない
- Loading 状態が適切に表示される

### Step 4: デザインシステム準拠チェック

スナップショットの構造を `.claude/skills/design-system/` のパターンと照合:

| チェック項目 | 参照ドキュメント | 確認ポイント |
|------------|----------------|------------|
| ページ構造 | `page-templates.md` | List/Detail/Form/Auth テンプレートに沿っているか |
| テーブル | `table-data-display.md` | ヘッダー・行・Empty state の構造 |
| フォーム | `form-pattern.md` | Label + Input ペア、バリデーション表示 |
| Modal/Sheet | `modal-sheet-pattern.md` | Dialog vs Sheet の使い分け、フッター構成 |
| フィードバック | `state-feedback.md` | Toast 使用、Loading/Error/Empty 状態 |

**判定基準**:
- shadcn UI コンポーネントが使われている (素の HTML 要素でない)
- デザイントークン (`tokens.md`) に沿ったスタイリング
- UI テキストが英語

## 出力フォーマット

全ステップ完了後、以下の表で結果を報告:

```
## UI Review Result

| Step | Check | Status | Notes |
|------|-------|--------|-------|
| 1 | Page render | Pass/Fail | ... |
| 1 | Console errors | Pass/Fail | ... |
| 2 | Mobile (375px) | Pass/Fail | ... |
| 2 | Tablet (768px) | Pass/Fail | ... |
| 2 | Desktop (1280px) | Pass/Fail | ... |
| 3 | Primary action | Pass/Fail | ... |
| 3 | Dialog/Sheet | Pass/Fail | ... |
| 4 | Page template | Pass/Fail | ... |
| 4 | Components | Pass/Fail | ... |
| 4 | Design tokens | Pass/Fail | ... |

URL: {checked URL}
```

不合格項目がある場合は、具体的な修正提案を添える。

## 認証が必要なページ

ログインが必要なページの場合:

```
1. browser_navigate → /login
2. browser_snapshot → フォーム要素を特定
3. browser_fill_form → email + password を入力
4. browser_click → Login ボタン
5. browser_wait_for → ダッシュボードへの遷移を待機
6. browser_navigate → 対象ページへ移動
```

テスト用認証情報はサービス起動時のシード or 手動登録で事前準備する。
