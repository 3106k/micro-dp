# Design System Skill

管理画面の UI を一貫性をもって構築するためのデザインシステム。
新しいページ・コンポーネントを作る前に、該当パターンのドキュメントを参照すること。

## 参照ガイド

| 作るもの | 読むファイル |
|---------|------------|
| 新しいページ | `page-templates.md` → `layout-navigation.md` |
| 一覧 + CRUD 機能 | `manager-pattern.md` → `table-data-display.md` |
| フォーム | `form-pattern.md` |
| テーブル・バッジ | `table-data-display.md` |
| 確認ダイアログ・編集パネル | `modal-sheet-pattern.md` |
| Loading/Error/Empty 表示 | `state-feedback.md` |
| 色・余白・文字 | `tokens.md` |

## 基本原則

1. **Server/Client 分離**: データ取得は Server Component、インタラクションは Client Component
2. **Manager パターン**: CRUD を持つ画面は `*-manager.tsx` に集約
3. **UI コンポーネント**: `@/components/ui/*` の shadcn コンポーネントを使う。素の HTML 要素は避ける
4. **フィードバック**: ユーザー操作の結果は `pushToast()` で通知（inline message は使わない）
5. **英語統一**: UI テキスト・エラーメッセージは英語
6. **Tailwind only**: インラインスタイルやカスタム CSS は使わない

## 変更後の検証 (必須)

UI の変更を行ったら、必ず **UI Review スキル** (`/ui-review`) で検証すること。
手順は `.claude/skills/ui-review/SKILL.md` を参照。

最低限のチェック:
1. `browser_navigate` → 変更したページを表示
2. `browser_snapshot` → 描画確認
3. `browser_console_messages(level: "error")` → エラーがないこと

フル検証 (新規ページ・大きな変更時): `/ui-review` の全ステップを実行。

## UIコンポーネント一覧

```
components/ui/
  button.tsx      — Button (variant: default|destructive|outline|secondary|ghost|link)
  input.tsx       — Input
  label.tsx       — Label
  card.tsx        — Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter
  form-error.tsx  — FormError (バリデーションエラー表示)
  submit-button.tsx — SubmitButton (loading 状態付きボタン)
  toast-provider.tsx — ToastProvider + useToast hook
  dialog.tsx      — Dialog (中央モーダル)
  sheet.tsx       — Sheet (サイドパネル、デフォルト右)
```
