# Design Tokens

## Typography

| 用途 | クラス | 例 |
|------|--------|-----|
| ページタイトル | `text-2xl font-semibold tracking-tight` | "Jobs", "Dashboard" |
| セクションタイトル | `text-lg font-semibold` | カード内、セクション見出し |
| サブセクション | `text-sm font-medium` | フォームセクション名 |
| ラベル | `text-sm font-medium` | Label コンポーネント |
| 本文 | `text-sm` | テーブルセル、フォーム入力 |
| 補足テキスト | `text-sm text-muted-foreground` | 説明、日付、ヘルプテキスト |
| 微小テキスト | `text-xs text-muted-foreground` | バッジ内、メタ情報 |
| 等幅 | `font-mono text-sm` or `font-mono text-xs` | ID, SQL, コード |

## Spacing

### 垂直間隔 (space-y)

| クラス | 値 | 用途 |
|--------|-----|------|
| `space-y-1.5` | 6px | Label ↔ 説明テキスト (Card 内) |
| `space-y-2` | 8px | Label ↔ Input ペア |
| `space-y-4` | 16px | フォームフィールド間 |
| `space-y-6` | 24px | ページタイトル ↔ コンテンツ |
| `space-y-8` | 32px | セクション間（フォーム ↔ テーブル） |

### 水平間隔 (gap)

| クラス | 値 | 用途 |
|--------|-----|------|
| `gap-2` | 8px | ボタン間、バッジ間 |
| `gap-3` | 12px | ナビアイテム間 |
| `gap-4` | 16px | グリッドカラム間 |
| `gap-6` | 24px | 大きなグリッド間 |

### パディング

| クラス | 用途 |
|--------|------|
| `p-4` | ボーダー付きセクション内 |
| `p-6` | Card 内 (CardHeader, CardContent) |
| `px-4 py-3` | テーブルセル |
| `px-2 py-1` | ナビリンク |
| `px-2.5 py-0.5` | バッジ |
| `py-8` | メインコンテンツ上下 (layout) |
| `p-12` | 空状態の大きなエリア |

## Color

### セマンティックカラー (CSS 変数ベース)

| トークン | 用途 |
|---------|------|
| `text-foreground` | メインテキスト |
| `text-muted-foreground` | 補足テキスト、プレースホルダー |
| `text-primary` | アクセントテキスト |
| `text-destructive` | エラーテキスト |
| `bg-background` | ページ背景 |
| `bg-muted` | 薄い背景 (hover, ヘッダー) |
| `bg-muted/50` | テーブルヘッダー |
| `bg-secondary` | アクティブナビ、queued バッジ |
| `bg-primary` | プライマリボタン |
| `bg-destructive` | 削除ボタン |
| `border` | デフォルトボーダー |
| `border-input` | 入力フィールドボーダー |

### ステータスカラー

```
成功 (success/active/published):
  Light: bg-green-100 text-green-800
  Dark:  dark:bg-green-900/30 dark:text-green-400

進行中 (running):
  Light: bg-blue-100 text-blue-800
  Dark:  dark:bg-blue-900/30 dark:text-blue-400

待機 (pending):
  Light: bg-yellow-100 text-yellow-800
  Dark:  dark:bg-yellow-900/30 dark:text-yellow-400

失敗 (failed/error):
  bg-destructive/10 text-destructive

非活性 (inactive/draft):
  bg-muted text-muted-foreground
```

### カテゴリカラー

```
pipeline:    bg-blue-100 text-blue-800
transform:   bg-purple-100 text-purple-800
source:      bg-teal-100 text-teal-800
destination: bg-orange-100 text-orange-800
```

## Container

| パターン | クラス | 用途 |
|---------|--------|------|
| ページコンテナ | `container py-8` | レイアウトが提供（ページ側で指定しない） |
| セクションボックス | `rounded-lg border p-4` | フォームセクション |
| テーブルラッパー | `rounded-lg border` | テーブルの外枠 |
| カード | `Card` コンポーネント | Auth フォーム、情報パネル |
| 空状態 | `rounded-lg border border-dashed p-12` | データなし時 |
| エラーボックス | `rounded-lg border border-destructive/40 bg-destructive/5 p-4` | エラー表示 |

## Border Radius

| クラス | 用途 |
|--------|------|
| `rounded-md` | ボタン、入力、ナビリンク |
| `rounded-lg` | カード、テーブルラッパー、セクション |
| `rounded-full` | バッジ |
| `rounded-sm` | 閉じるボタン (Dialog/Sheet 内) |

## Responsive Breakpoints

| プレフィックス | 幅 | 用途 |
|-------------|-----|------|
| (なし) | 全幅 | モバイルファースト |
| `sm:` | 640px+ | Sheet 幅制限 |
| `md:` | 768px+ | 2 カラムグリッド |
| `lg:` | 1024px+ | サイドバーレイアウト |
