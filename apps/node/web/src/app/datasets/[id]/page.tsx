import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";

type MeResponse = components["schemas"]["MeResponse"];
type Dataset = components["schemas"]["Dataset"];

function prettyJSON(input?: string): string {
  if (!input) {
    return "{}";
  }
  try {
    return JSON.stringify(JSON.parse(input), null, 2);
  } catch {
    return input;
  }
}

export default async function DatasetDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;
  if (!token || !tenantId) {
    redirect("/signin");
  }

  const meRes = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!meRes.ok) {
    redirect("/signin");
  }
  const me: MeResponse = await meRes.json();

  const datasetRes = await backendFetch(`/api/v1/datasets/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });

  let dataset: Dataset | null = null;
  let errorMessage = "";
  if (datasetRes.ok) {
    dataset = await datasetRes.json();
  } else {
    const err = (await datasetRes.json()) as { error?: string };
    errorMessage = err.error ?? `failed to load dataset (${datasetRes.status})`;
  }

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
        tenants={me.tenants}
        currentTenantId={tenantId}
      />
      <main className="container space-y-6 py-8">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold tracking-tight">Dataset Detail</h1>
          <Link href="/datasets" className="text-sm underline-offset-2 hover:underline">
            Back to list
          </Link>
        </div>

        {errorMessage ? (
          <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
            {errorMessage}
          </div>
        ) : null}

        {dataset ? (
          <div className="space-y-6">
            <div className="grid gap-4 rounded-lg border p-4 md:grid-cols-2">
              <div>
                <p className="text-xs text-muted-foreground">ID</p>
                <p className="font-mono text-sm">{dataset.id}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Name</p>
                <p className="text-sm">{dataset.name}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Source Type</p>
                <p className="text-sm">{dataset.source_type}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Row Count</p>
                <p className="text-sm">{dataset.row_count ?? "-"}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Storage Path</p>
                <p className="font-mono text-sm">{dataset.storage_path}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Last Updated</p>
                <p className="text-sm">
                  {dataset.last_updated_at
                    ? new Date(dataset.last_updated_at).toLocaleString()
                    : "-"}
                </p>
              </div>
            </div>

            <div className="rounded-lg border p-4">
              <h2 className="mb-2 text-lg font-semibold">Schema JSON</h2>
              <pre className="overflow-auto rounded bg-muted p-3 text-xs">
                {prettyJSON(dataset.schema_json)}
              </pre>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  );
}
