import Link from "next/link";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";

type JobRun = components["schemas"]["JobRun"];

const statusStyles: Record<JobRun["status"], string> = {
  queued: "bg-secondary text-secondary-foreground",
  running: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
  success:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  failed: "bg-destructive/10 text-destructive",
  canceled:
    "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400",
};

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default async function JobRunDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { token, currentTenantId } = await getAuthContext();

  const runRes = await backendFetch(`/api/v1/job_runs/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">Job Run Detail</h1>
        <Link href="/job-runs" className="text-sm underline-offset-2 hover:underline">
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
            <p className="mt-1">
              <span
                className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusStyles[run.status]}`}
              >
                {run.status}
              </span>
            </p>
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
              {run.started_at ? formatDateTime(run.started_at) : "-"}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Finished</p>
            <p className="text-sm">
              {run.finished_at ? formatDateTime(run.finished_at) : "-"}
            </p>
          </div>
        </div>
      ) : null}
    </div>
  );
}
