# Layout & Navigation

## ルートグループ

```
src/app/
  layout.tsx              ← Root: ToastProvider のみ
  (app)/layout.tsx        ← AppHeader + TrackerProvider + <main>
  (admin)/layout.tsx      ← superadmin チェック + AppHeader(sectionLabel="Admin") + AdminNav
  (settings)/layout.tsx   ← AppHeader(sectionLabel="Settings") + SettingsNav
```

### どのグループに入れるか

| 用途 | グループ | ヘッダー | サブナビ |
|------|---------|---------|---------|
| メイン機能（Jobs, Datasets, Dashboard） | `(app)` | AppHeader | なし |
| 設定（Connections, Members, Uploads, Billing） | `(settings)` | AppHeader + "Settings" | SettingsNav |
| 管理者（Tenants, Plans, Analytics） | `(admin)` | AppHeader + "Admin" | AdminNav |
| 認証（SignIn, SignUp） | ルート直下 | なし | なし |

## レイアウトの責務

レイアウトは以下を行い、子ページには **データ取得のみ** を任せる：

```tsx
// (app)/layout.tsx — 典型パターン
export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const { me, currentTenantId } = await getAuthContext();  // 認証
  return (
    <div className="min-h-screen">
      <AppHeader                                            // ヘッダー
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
        tenants={me.tenants}
        currentTenantId={currentTenantId}
      />
      <TrackerProvider tenantId={currentTenantId} userId={me.user_id}>
        <main className="container py-8">{children}</main>  // コンテンツ領域
      </TrackerProvider>
    </div>
  );
}
```

## ナビゲーション

### メインナビ (AppHeader)

```tsx
const mainNavItems = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/jobs", label: "Jobs" },
  { href: "/job-runs", label: "Job Runs" },
  { href: "/datasets", label: "Datasets" },
];
```

- Active: `rounded-md bg-secondary px-2 py-1 font-medium`
- Inactive: `rounded-md px-2 py-1 text-muted-foreground hover:bg-muted hover:text-foreground`

### サブナビ (SettingsNav / AdminNav)

- Active: `border-b-2 border-foreground px-3 py-2 font-medium`
- Inactive: `border-b-2 border-transparent px-3 py-2 text-muted-foreground hover:text-foreground`

### Active 判定

```tsx
function isActive(href: string): boolean {
  return pathname === href || pathname.startsWith(href + "/");
}
```

## 新しいナビ項目を追加するとき

1. メインナビ: `app-header.tsx` の `mainNavItems` に追加
2. サブナビ: `settings-nav.tsx` または `admin-nav.tsx` の `items` に追加
3. ルートグループの layout に新しいページのルートを含める
