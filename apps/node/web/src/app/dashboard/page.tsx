import { cookies } from "next/headers";

import { backendFetch } from "@/lib/api/server";
import { TOKEN_COOKIE, TENANT_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";
import { JobRunsTable } from "./job-runs-table";

type JobRun = components["schemas"]["JobRun"];

export default async function DashboardPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;

  let jobRuns: JobRun[] = [];

  if (token && tenantId) {
    const headers = {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    };

    const jobRunsRes = await backendFetch("/api/v1/job_runs", { headers });

    if (jobRunsRes.ok) {
      const data: { items: JobRun[] } = await jobRunsRes.json();
      jobRuns = data.items;
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
      <JobRunsTable jobRuns={jobRuns} />
    </div>
  );
}
