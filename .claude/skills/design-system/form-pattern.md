# Form Pattern

## 基本フォーム

```tsx
<form onSubmit={handleSubmit} className="space-y-4">
  {/* 1カラム or 2カラムグリッド */}
  <div className="grid gap-4 md:grid-cols-2">
    <div className="space-y-2">
      <Label htmlFor="name">Name</Label>
      <Input id="name" name="name" required placeholder="Enter name" />
    </div>
    <div className="space-y-2">
      <Label htmlFor="email">Email</Label>
      <Input id="email" name="email" type="email" required placeholder="you@example.com" />
    </div>
  </div>

  {/* フルワイド項目 */}
  <div className="space-y-2">
    <Label htmlFor="description">Description</Label>
    <textarea
      id="description"
      name="description"
      rows={3}
      className="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
      placeholder="Optional description"
    />
  </div>

  {/* アクション */}
  <div className="flex gap-2">
    <Button type="submit" disabled={loading}>
      {loading ? "Saving..." : "Save"}
    </Button>
    <Button type="button" variant="outline" onClick={onCancel}>
      Cancel
    </Button>
  </div>
</form>
```

## Input/Label ペア

必ずセットで使う：

```tsx
<div className="space-y-2">
  <Label htmlFor="field-id">Field Name</Label>
  <Input id="field-id" name="field-name" ... />
</div>
```

- `id` と `htmlFor` を一致させる
- `name` 属性も必ず付ける（FormData 取得用）
- `space-y-2` でラベルと入力の間隔

## グリッドレイアウト

| パターン | クラス | 用途 |
|---------|--------|------|
| 2 カラム | `grid gap-4 md:grid-cols-2` | 短い入力が並ぶ場合 |
| 1 カラム | `space-y-4` | テキストエリア、説明文 |
| サイドバー + メイン | `grid gap-6 lg:grid-cols-[1fr_2fr]` | 複雑なフォーム |

## フォームの種類

### Manager 内フォーム

Manager コンポーネントの上部に配置。作成後テーブルを refresh:

```tsx
async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
  e.preventDefault();
  setLoading(true);
  try {
    const form = new FormData(e.currentTarget);
    const res = await fetch("/api/items", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(Object.fromEntries(form)),
    });
    if (!res.ok) throw new Error((await res.json()).error ?? "Failed");
    pushToast({ variant: "success", message: "Created" });
    e.currentTarget.reset();
    await refreshItems();
  } catch (error) {
    pushToast({ variant: "error", message: error instanceof Error ? error.message : "Failed" });
  } finally {
    setLoading(false);
  }
}
```

### Sheet 内フォーム

右パネルで編集。保存後に Sheet を閉じて refresh:

```tsx
<SheetContent>
  <SheetHeader>
    <SheetTitle>Edit Item</SheetTitle>
  </SheetHeader>
  <form onSubmit={handleEdit} className="mt-6 space-y-4">
    {/* fields */}
    <SheetFooter>
      <SheetClose asChild>
        <Button variant="outline">Cancel</Button>
      </SheetClose>
      <Button type="submit" disabled={loading}>Save</Button>
    </SheetFooter>
  </form>
</SheetContent>
```

### Auth フォーム

Card でラップし、中央配置:

```tsx
<Card className="w-full max-w-sm">
  <CardHeader>
    <CardTitle>Sign In</CardTitle>
    <CardDescription>Enter your credentials</CardDescription>
  </CardHeader>
  <CardContent>
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* fields */}
      <FormError message={error} />
      <SubmitButton loading={loading} loadingLabel="Signing in...">
        Sign In
      </SubmitButton>
    </form>
  </CardContent>
</Card>
```

## セクション分け

複雑なフォームはボーダー付きセクションで区切る：

```tsx
<div className="rounded-lg border p-4 space-y-4">
  <h3 className="text-sm font-medium">Connection Settings</h3>
  {/* fields */}
</div>
```

## Select (ドロップダウン)

```tsx
<div className="space-y-2">
  <Label htmlFor="type">Type</Label>
  <select
    id="type"
    name="type"
    className="h-10 w-full rounded-md border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
    value={value}
    onChange={(e) => setValue(e.target.value)}
  >
    <option value="">Select type...</option>
    {options.map((opt) => (
      <option key={opt.value} value={opt.value}>{opt.label}</option>
    ))}
  </select>
</div>
```

## バリデーション

- HTML5 バリデーション: `required`, `type="email"`, `minLength` を積極的に使う
- カスタムバリデーション: submit ハンドラ内で行い、エラーは `pushToast()` で通知
- Auth フォームのみ `FormError` コンポーネントを使用可（フォーム内にインラインで表示したい場合）
