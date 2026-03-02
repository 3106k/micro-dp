# Table & Data Display

## テーブル構造

```tsx
<div className="rounded-lg border">
  <table className="w-full text-sm">
    <thead>
      <tr className="border-b bg-muted/50">
        <th className="px-4 py-3 text-left font-medium">Name</th>
        <th className="px-4 py-3 text-left font-medium">Status</th>
        <th className="px-4 py-3 text-left font-medium">Created</th>
        <th className="px-4 py-3 text-right font-medium">Actions</th>
      </tr>
    </thead>
    <tbody>
      {items.map((item) => (
        <tr key={item.id} className="border-b last:border-0">
          <td className="px-4 py-3">{item.name}</td>
          <td className="px-4 py-3"><StatusBadge status={item.status} /></td>
          <td className="px-4 py-3 text-muted-foreground">{formatDate(item.created_at)}</td>
          <td className="px-4 py-3 text-right">
            <div className="flex justify-end gap-2">
              <Button size="sm" variant="outline">Edit</Button>
              <Button size="sm" variant="destructive">Delete</Button>
            </div>
          </td>
        </tr>
      ))}
    </tbody>
  </table>

  {/* Empty state */}
  {items.length === 0 ? (
    <div className="px-4 py-8 text-center text-sm text-muted-foreground">
      No items found.
    </div>
  ) : null}
</div>
```

## テーブルスタイルルール

| 要素 | クラス |
|------|--------|
| ラッパー | `rounded-lg border` |
| table | `w-full text-sm` |
| thead tr | `border-b bg-muted/50` |
| th | `px-4 py-3 text-left font-medium` |
| tbody tr | `border-b last:border-0` |
| td | `px-4 py-3` |
| Actions td | `px-4 py-3 text-right` + `flex justify-end gap-2` |
| 補足テキスト | `text-muted-foreground` |
| 等幅テキスト | `font-mono text-xs` (ID, コード) |

## Empty State

テーブルが空のとき、必ず空メッセージを表示する：

```tsx
{items.length === 0 ? (
  <div className="px-4 py-8 text-center text-sm text-muted-foreground">
    No items found.
  </div>
) : null}
```

- テーブル外の `<div>` で表示（`<tr>` + `colSpan` でも可）
- `py-8` で十分な余白
- メッセージは具体的に: "No jobs found." "No members yet."

## ステータスバッジ

```tsx
const statusStyles: Record<string, string> = {
  // 成功系
  active:    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  success:   "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  published: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",

  // 進行中
  running:   "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
  pending:   "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
  queued:    "bg-secondary text-secondary-foreground",

  // 失敗・非活性
  failed:    "bg-destructive/10 text-destructive",
  error:     "bg-destructive/10 text-destructive",
  inactive:  "bg-muted text-muted-foreground",
  draft:     "bg-muted text-muted-foreground",
};

function StatusBadge({ status }: { status: string }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
        statusStyles[status] ?? "bg-secondary text-secondary-foreground"
      }`}
    >
      {status}
    </span>
  );
}
```

## カテゴリバッジ

ステータス以外の分類用：

```tsx
const kindColors: Record<string, string> = {
  pipeline:  "bg-blue-100 text-blue-800",
  transform: "bg-purple-100 text-purple-800",
  source:    "bg-teal-100 text-teal-800",
  destination: "bg-orange-100 text-orange-800",
};
```

バッジの構造は同じ: `rounded-full px-2.5 py-0.5 text-xs font-medium`

## クリック可能な行

詳細ページへの遷移がある場合:

```tsx
<tr
  key={item.id}
  className="border-b last:border-0 cursor-pointer hover:bg-muted/50"
  onClick={() => router.push(`/jobs/${item.id}`)}
>
```

## 日付フォーマット

```tsx
function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// 日時が必要な場合
function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
```

## アクションボタン

```tsx
// テーブル内
<Button size="sm" variant="outline">Edit</Button>
<Button size="sm" variant="destructive">Delete</Button>

// ページヘッダー横
<div className="flex items-center justify-between">
  <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
  <Button asChild>
    <Link href="/jobs/new">Create Job</Link>
  </Button>
</div>
```
