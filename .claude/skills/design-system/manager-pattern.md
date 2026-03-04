# Manager Pattern

`*-manager.tsx` は Client Component で、一覧表示 + CRUD 操作を一つのコンポーネントに集約するパターン。

## 構造

```tsx
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast-provider";

type Item = { id: string; name: string };

export function ItemsManager({ initialItems }: { initialItems: Item[] }) {
  const { pushToast } = useToast();
  const [items, setItems] = useState(initialItems);
  const [loading, setLoading] = useState(false);

  // --- データ再取得 ---
  async function refreshItems() {
    const res = await fetch("/api/items");
    if (res.ok) {
      const data = await res.json();
      setItems(data.items ?? []);
    }
  }

  // --- 作成 ---
  async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    try {
      const form = new FormData(e.currentTarget);
      const res = await fetch("/api/items", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: form.get("name") }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error ?? "Failed to create");
      }
      pushToast({ variant: "success", message: "Item created" });
      e.currentTarget.reset();
      await refreshItems();
    } catch (error) {
      pushToast({
        variant: "error",
        message: error instanceof Error ? error.message : "Request failed",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-8">
      {/* --- Create Form --- */}
      <form onSubmit={handleCreate} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" name="name" required />
          </div>
        </div>
        <Button type="submit" disabled={loading}>
          {loading ? "Creating..." : "Create"}
        </Button>
      </form>

      {/* --- Table --- */}
      <div className="rounded-lg border">
        <table className="w-full text-sm">
          {/* ... see table-data-display.md */}
        </table>
      </div>
    </div>
  );
}
```

## 設計ルール

### State

```tsx
const [items, setItems] = useState(initialItems);   // データ
const [loading, setLoading] = useState(false);       // 送信中
```

- `message` / `error` state は **使わない** → `pushToast()` に統一
- 編集中のアイテムがある場合: `const [editingId, setEditingId] = useState<string | null>(null)`

### データフロー

```
Server Page (backendFetch → /api/v1/*)
  ↓ initialItems
Manager (useState)
  ↓ fetch → /api/* (Next.js API Routes)
  ↓ refreshItems()
  ↓ setState
```

- Server Component: `backendFetch()` で Go API を直接呼ぶ
- Client Component: `/api/*` (Next.js API Routes) 経由でプロキシ

### フィードバック

```tsx
// 成功
pushToast({ variant: "success", message: "Saved successfully" });

// エラー
pushToast({ variant: "error", message: error.message });
```

### 削除確認

Dialog を使って確認する（`modal-sheet-pattern.md` 参照）：

```tsx
const [deleteTarget, setDeleteTarget] = useState<Item | null>(null);

// テーブル行のボタン
<Button size="sm" variant="destructive" onClick={() => setDeleteTarget(item)}>
  Delete
</Button>

// 確認 Dialog
<ConfirmDialog
  open={!!deleteTarget}
  onOpenChange={(open) => !open && setDeleteTarget(null)}
  title="Delete item"
  description={`Are you sure you want to delete "${deleteTarget?.name}"?`}
  onConfirm={() => handleDelete(deleteTarget!.id)}
/>
```

### 編集

- 簡単な編集: Sheet（右パネル）で編集フォームを表示
- 複雑な編集: 別ページ (`/items/[id]/edit`) に遷移

## ファイル命名

| ファイル | 役割 |
|---------|------|
| `page.tsx` | Server Component（データ取得） |
| `*-manager.tsx` | Client Component（CRUD 管理） |
| `*-form.tsx` | Client Component（フォーム専用、Manager から分離する場合） |

## 既存の Manager 一覧

| ファイル | 機能 |
|---------|------|
| `jobs-manager.tsx` | Job 一覧 + 作成 |
| `connections-manager.tsx` | Connection 一覧 + 作成 + テスト |
| `members-manager.tsx` | メンバー一覧 + 招待 + ロール変更 |
| `tenants-manager.tsx` | テナント一覧 + 作成 (Admin) |
| `uploads-manager.tsx` | ファイルアップロード + 一覧 (`(app)/datasets/upload/`) |
| `job-runs-manager.tsx` | Job Run 一覧 |
| `versions-manager.tsx` | Job Version 一覧 + 作成 + Publish |
| `job-detail-manager.tsx` | Job 詳細 + 編集 |
