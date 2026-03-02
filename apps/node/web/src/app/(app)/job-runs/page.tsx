import { cookies } from "next/headers";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { JobRunsManager } from "./job-runs-manager";

type JobRun = components["schemas"]["JobRun"];

export default async function JobRunsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value!;
  const tenantId = jar.get(TENANT_COOKIE)?.value!;

  let runs: JobRun[] = [];
  const runsRes = await backendFetch("/api/v1/job_runs", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });
  if (runsRes.ok) {
    const data: { items: JobRun[] } = await runsRes.json();
    runs = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Job Runs</h1>
      <JobRunsManager initialRuns={runs} />
    </div>
  );
}
