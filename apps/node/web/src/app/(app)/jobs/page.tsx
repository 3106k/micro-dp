import { cookies } from "next/headers";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { JobsManager } from "./jobs-manager";

type Job = components["schemas"]["Job"];

export default async function JobsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value!;
  const tenantId = jar.get(TENANT_COOKIE)?.value!;

  let jobs: Job[] = [];
  const jobsRes = await backendFetch("/api/v1/jobs", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });
  if (jobsRes.ok) {
    const data: { items: Job[] } = await jobsRes.json();
    jobs = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
      <JobsManager initialJobs={jobs} />
    </div>
  );
}
