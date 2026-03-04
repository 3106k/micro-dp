import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { JobsManager } from "./jobs-manager";

type Job = components["schemas"]["Job"];

export default async function JobsPage() {
  const { token, currentTenantId } = await getAuthContext();

  let jobs: Job[] = [];
  const jobsRes = await backendFetch("/api/v1/jobs", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });
  if (jobsRes.ok) {
    const data: { items: Job[] } = await jobsRes.json();
    jobs = data.items ?? [];
  }

  return <JobsManager initialJobs={jobs} />;
}
