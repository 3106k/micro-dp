import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { JobRunsManager } from "./job-runs-manager";

type JobRun = components["schemas"]["JobRun"];

export default async function JobRunsPage() {
  const { token, currentTenantId } = await getAuthContext();

  let runs: JobRun[] = [];
  const runsRes = await backendFetch("/api/v1/job_runs", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });
  if (runsRes.ok) {
    const data: { items: JobRun[] } = await runsRes.json();
    runs = data.items ?? [];
  }

  return <JobRunsManager initialRuns={runs} />;
}
