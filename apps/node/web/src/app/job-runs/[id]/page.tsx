import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";

type MeResponse = components["schemas"]["MeResponse"];
type JobRun = components["schemas"]["JobRun"];

export default async function JobRunDetailPage({
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

  const runRes = await backendFetch(`/api/v1/job_runs/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });

  let run: JobRun | null = null;
  let errorMessage = "";
  if (runRes.ok) {
    run = await runRes.json();
  } else {
    const err = (await runRes.json()) as { error?: string };
    errorMessage = err.error ?? `failed to load run (${runRes.status})`;
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
          <h1 className="text-2xl font-semibold tracking-tight">Job Run Detail</h1>
          <Link href="/job-runs" className="text-sm underline">
            Back to runs
          </Link>
        </div>

        {errorMessage ? (
          <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
            {errorMessage}
          </div>
        ) : null}

        {run ? (
          <div className="grid gap-4 rounded-lg border p-4 md:grid-cols-2">
            <div>
              <p className="text-xs text-muted-foreground">ID</p>
              <p className="font-mono text-sm">{run.id}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Status</p>
              <p className="text-sm">{run.status}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Job ID</p>
              <p className="font-mono text-sm">{run.job_id}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Job Version ID</p>
              <p className="font-mono text-sm">{run.job_version_id ?? "-"}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Started</p>
              <p className="text-sm">
                {run.started_at ? new Date(run.started_at).toLocaleString() : "-"}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Finished</p>
              <p className="text-sm">
                {run.finished_at ? new Date(run.finished_at).toLocaleString() : "-"}
              </p>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  );
}
