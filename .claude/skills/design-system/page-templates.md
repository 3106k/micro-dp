# Page Templates

全ページは 4 つのテンプレートのいずれかに該当する。

## 1. List Page（一覧ページ）

Server Component がデータ取得し、Client Manager に渡す。

```tsx
// src/app/(app)/jobs/page.tsx
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { backendFetch } from "@/lib/api/server";
import { JobsManager } from "./jobs-manager";

export default async function JobsPage() {
  const { token, currentTenantId } = await getAuthContext();

  const res = await backendFetch("/api/v1/jobs", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });

  const data = res.ok ? await res.json() : { items: [] };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
      <JobsManager initialJobs={data.items ?? []} />
    </div>
  );
}
```

**ルール**:
- `getAuthContext()` で認証情報取得（cookie 直接触らない）
- `backendFetch()` でバックエンド API 呼び出し
- fetch 失敗時は空配列でフォールバック
- Manager コンポーネントに `initial*` props で渡す

## 2. Detail Page（詳細ページ）

```tsx
// src/app/(app)/jobs/[id]/page.tsx
export default async function JobDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { token, currentTenantId } = await getAuthContext();

  const res = await backendFetch(`/api/v1/jobs/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
  });

  if (!res.ok) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-semibold tracking-tight">Job</h1>
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
          Failed to load job.
        </div>
      </div>
    );
  }

  const job = await res.json();

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">{job.name}</h1>
      <JobDetailManager initialJob={job} />
    </div>
  );
}
```

**ルール**:
- `params` は `Promise` — 必ず `await` する
- エラー時は destructive スタイルのエラーメッセージを表示
- 正常時は Detail Manager に渡す

## 3. Form Page（フォームページ）

```tsx
// src/app/(app)/jobs/new/page.tsx
export default async function NewJobPage() {
  const { token, currentTenantId } = await getAuthContext();

  // フォームに必要な選択肢データを取得
  const datasetsRes = await backendFetch("/api/v1/datasets?limit=100", { ... });
  const datasets = datasetsRes.ok ? (await datasetsRes.json()).items ?? [] : [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Create Job</h1>
      <JobForm datasets={datasets} />
    </div>
  );
}
```

**ルール**:
- フォームに必要なマスタデータは Server Component で取得
- フォーム自体は Client Component (`*-form.tsx`)

## 4. Auth Page（認証ページ）

```tsx
// src/app/signin/page.tsx — ルートグループ外
export default function SignInPage() {
  return (
    <main className="flex min-h-screen items-center justify-center">
      <Suspense>
        <SignInForm />
      </Suspense>
    </main>
  );
}
```

**ルール**:
- ルートグループ外（`src/app/` 直下）
- 中央揃えレイアウト
- `<Suspense>` で囲む（useSearchParams 対応）
- Card コンポーネントでラップ

## 共通ルール

- ページタイトル: `<h1 className="text-2xl font-semibold tracking-tight">`
- コンテンツ間隔: `<div className="space-y-6">`
- レイアウトの `<main className="container py-8">` がラップ済みなので、ページ側で `container` や `py-8` は不要
