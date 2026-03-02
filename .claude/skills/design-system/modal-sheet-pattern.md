# Modal & Sheet Pattern

## 使い分け

| コンポーネント | 用途 | 幅 |
|--------------|------|-----|
| **Dialog** | 確認・警告・短いフォーム | `max-w-lg` (固定中央) |
| **Sheet** | 詳細表示・編集フォーム・プレビュー | `sm:max-w-sm` 〜 カスタム幅 (右スライド) |

### いつ Dialog を使うか

- 削除確認 ("Are you sure?")
- 短い入力 (rename, invite)
- 重要な警告・エラー詳細
- ユーザーの明示的な判断が必要なとき

### いつ Sheet を使うか

- 一覧から選択した項目の詳細・編集
- フィルター/設定パネル
- フォームが複数フィールドある編集
- メインコンテンツを見ながら操作したいとき

---

## Dialog (確認ダイアログ)

### 基本構造

```tsx
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";

<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Delete connection</DialogTitle>
      <DialogDescription>
        This action cannot be undone. The connection "Production DB" will be
        permanently deleted.
      </DialogDescription>
    </DialogHeader>
    <DialogFooter>
      <DialogClose asChild>
        <Button variant="outline">Cancel</Button>
      </DialogClose>
      <Button variant="destructive" onClick={handleDelete} disabled={loading}>
        {loading ? "Deleting..." : "Delete"}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### 確認ダイアログの共通パターン

Manager 内で再利用しやすい state パターン：

```tsx
// State
const [deleteTarget, setDeleteTarget] = useState<Item | null>(null);

// テーブル行のボタンで target をセット
<Button size="sm" variant="destructive" onClick={() => setDeleteTarget(item)}>
  Delete
</Button>

// Dialog (open は target の有無で制御)
<Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Delete item</DialogTitle>
      <DialogDescription>
        Are you sure you want to delete &quot;{deleteTarget?.name}&quot;?
        This action cannot be undone.
      </DialogDescription>
    </DialogHeader>
    <DialogFooter>
      <DialogClose asChild>
        <Button variant="outline">Cancel</Button>
      </DialogClose>
      <Button
        variant="destructive"
        disabled={loading}
        onClick={async () => {
          setLoading(true);
          try {
            await handleDelete(deleteTarget!.id);
            setDeleteTarget(null);
          } finally {
            setLoading(false);
          }
        }}
      >
        {loading ? "Deleting..." : "Delete"}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### Dialog 内フォーム (短い入力)

```tsx
<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Invite member</DialogTitle>
      <DialogDescription>
        Send an invitation email to add a new team member.
      </DialogDescription>
    </DialogHeader>
    <form onSubmit={handleInvite} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="invite-email">Email</Label>
        <Input id="invite-email" name="email" type="email" required />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <Button type="submit" disabled={loading}>
          {loading ? "Sending..." : "Send Invite"}
        </Button>
      </DialogFooter>
    </form>
  </DialogContent>
</Dialog>
```

---

## Sheet (サイドパネル)

### 基本構造

```tsx
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
  SheetClose,
} from "@/components/ui/sheet";

<Sheet open={open} onOpenChange={setOpen}>
  <SheetContent>   {/* デフォルト: side="right" */}
    <SheetHeader>
      <SheetTitle>Edit Connection</SheetTitle>
      <SheetDescription>Update the connection settings.</SheetDescription>
    </SheetHeader>

    <div className="mt-6 space-y-4">
      {/* コンテンツ */}
    </div>

    <SheetFooter className="mt-6">
      <SheetClose asChild>
        <Button variant="outline">Cancel</Button>
      </SheetClose>
      <Button onClick={handleSave} disabled={loading}>
        {loading ? "Saving..." : "Save"}
      </Button>
    </SheetFooter>
  </SheetContent>
</Sheet>
```

### 幅のカスタマイズ

デフォルトは `sm:max-w-sm`。広くしたい場合：

```tsx
// 中幅 (フォーム向き)
<SheetContent className="sm:max-w-md">

// 広め (詳細表示 + フォーム)
<SheetContent className="sm:max-w-lg">

// ワイド (プレビュー、コード表示)
<SheetContent className="sm:max-w-xl">
```

### 詳細表示パターン

テーブル行クリック → Sheet で詳細を表示：

```tsx
const [selectedItem, setSelectedItem] = useState<Item | null>(null);

// テーブル行
<tr className="cursor-pointer hover:bg-muted/50" onClick={() => setSelectedItem(item)}>

// Sheet
<Sheet open={!!selectedItem} onOpenChange={(open) => !open && setSelectedItem(null)}>
  <SheetContent className="sm:max-w-md">
    <SheetHeader>
      <SheetTitle>{selectedItem?.name}</SheetTitle>
    </SheetHeader>
    <div className="mt-6 space-y-4">
      <dl className="space-y-3 text-sm">
        <div>
          <dt className="text-muted-foreground">Status</dt>
          <dd><StatusBadge status={selectedItem?.status} /></dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Created</dt>
          <dd>{formatDate(selectedItem?.created_at)}</dd>
        </div>
      </dl>
    </div>
    <SheetFooter className="mt-6">
      <Button variant="outline" onClick={() => handleEdit(selectedItem)}>
        Edit
      </Button>
      <Button variant="destructive" onClick={() => setDeleteTarget(selectedItem)}>
        Delete
      </Button>
    </SheetFooter>
  </SheetContent>
</Sheet>
```

### 編集フォーム Sheet

```tsx
<Sheet open={!!editingItem} onOpenChange={(open) => !open && setEditingItem(null)}>
  <SheetContent className="sm:max-w-md">
    <SheetHeader>
      <SheetTitle>Edit {editingItem?.name}</SheetTitle>
    </SheetHeader>
    <form
      onSubmit={async (e) => {
        e.preventDefault();
        await handleUpdate(editingItem!.id, new FormData(e.currentTarget));
        setEditingItem(null);
      }}
      className="mt-6 space-y-4"
    >
      <div className="space-y-2">
        <Label htmlFor="edit-name">Name</Label>
        <Input id="edit-name" name="name" defaultValue={editingItem?.name} required />
      </div>
      <SheetFooter>
        <SheetClose asChild>
          <Button variant="outline">Cancel</Button>
        </SheetClose>
        <Button type="submit" disabled={loading}>
          {loading ? "Saving..." : "Save"}
        </Button>
      </SheetFooter>
    </form>
  </SheetContent>
</Sheet>
```

---

## 設計ルール

1. **open 制御**: `useState` で管理。target オブジェクトの有無 (`!!target`) で開閉
2. **閉じる**: `onOpenChange` + `DialogClose`/`SheetClose` の両方を使う
3. **Loading**: ボタンに `disabled={loading}` + テキスト切り替え
4. **フィードバック**: 操作完了後は `pushToast()` で通知し、モーダル/シートを閉じる
5. **フォーカス**: Radix が自動管理するので手動制御不要
6. **アクセシビリティ**: `DialogTitle`/`SheetTitle` は必須（スクリーンリーダー対応）
