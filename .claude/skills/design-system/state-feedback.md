# State & Feedback

## フィードバック手段の統一

| 手段 | 用途 | 使う場面 |
|------|------|---------|
| **Toast** (`pushToast`) | 操作結果の通知 | CRUD 成功/失敗、テナント切替など |
| **FormError** | フォームバリデーション | Auth フォームのみ（インライン表示が必要な場合） |
| **Inline error div** | ページレベルのエラー | Server Component でデータ取得失敗時 |

**原則**: Client Component のフィードバックは **すべて Toast** に統一する。`message` state は使わない。

## Toast

```tsx
import { useToast } from "@/components/ui/toast-provider";

const { pushToast } = useToast();

// 成功
pushToast({ variant: "success", message: "Connection saved" });

// エラー
pushToast({ variant: "error", message: "Failed to delete" });

// 情報
pushToast({ message: "Copied to clipboard" });  // variant 省略 = info
```

### Toast スタイル

| variant | ボーダー | 背景 | テキスト |
|---------|---------|------|---------|
| `success` | `border-emerald-300` | `bg-emerald-50` | `text-emerald-800` |
| `error` | `border-destructive/30` | `bg-destructive/10` | `text-destructive` |
| `info` | `border-border` | `bg-background` | `text-foreground` |

- 自動消去: 4 秒
- 位置: 右上 (`fixed right-4 top-4 z-50`)

## Loading State

### ボタン

```tsx
<Button type="submit" disabled={loading}>
  {loading ? "Saving..." : "Save"}
</Button>
```

ルール:
- `disabled={loading}` で二重送信防止
- テキストを `...ing` 形式に切り替え
- `finally` ブロックで必ず `setLoading(false)`

### テーブル行の操作

```tsx
<Button
  size="sm"
  variant="outline"
  disabled={actionLoading === item.id}
  onClick={() => handleAction(item.id)}
>
  {actionLoading === item.id ? "..." : "Publish"}
</Button>
```

### ページ全体 (将来)

Suspense + loading.tsx で対応:

```tsx
// src/app/(app)/jobs/loading.tsx
export default function Loading() {
  return (
    <div className="space-y-6">
      <div className="h-8 w-48 animate-pulse rounded bg-muted" />
      <div className="h-64 animate-pulse rounded-lg border bg-muted/30" />
    </div>
  );
}
```

## Error State

### Server Component (ページレベル)

```tsx
// データ取得失敗時
if (!res.ok) {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
        Failed to load jobs.
      </div>
    </div>
  );
}
```

### Client Component (操作エラー)

Toast で通知:

```tsx
try {
  const res = await fetch(...);
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error ?? "Request failed");
  }
} catch (error) {
  pushToast({
    variant: "error",
    message: error instanceof Error ? error.message : "An error occurred",
  });
}
```

### エラーメッセージのルール

- ユーザーに見せるメッセージは具体的に: "Failed to save connection" (not "Error")
- API エラーをそのまま表示: `data.error ?? "Request failed"`
- 英語で統一

## Empty State

### テーブル内

```tsx
{items.length === 0 ? (
  <div className="px-4 py-8 text-center text-sm text-muted-foreground">
    No connections found. Create one to get started.
  </div>
) : null}
```

### ページ全体（データが全くない場合）

```tsx
{items.length === 0 ? (
  <div className="flex flex-col items-center gap-4 rounded-lg border border-dashed p-12 text-center">
    <p className="text-sm text-muted-foreground">No datasets yet</p>
    <Button asChild>
      <Link href="/datasets/upload">Upload CSV</Link>
    </Button>
  </div>
) : null}
```

## 非同期操作のテンプレート

```tsx
async function handleAction() {
  setLoading(true);
  try {
    const res = await fetch("/api/...", { method: "POST", ... });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error ?? "Failed");
    }
    pushToast({ variant: "success", message: "Done" });
    await refreshData();  // 一覧更新
  } catch (error) {
    pushToast({
      variant: "error",
      message: error instanceof Error ? error.message : "Request failed",
    });
  } finally {
    setLoading(false);
  }
}
```
