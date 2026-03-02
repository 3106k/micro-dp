import Link from "next/link";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { JobDetailManager } from "./job-detail-manager";

type Job = components["schemas"]["Job"];

export default async function JobDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { token, currentTenantId } = await getAuthContext();

  const jobRes = await backendFetch(`/api/v1/jobs/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });

  if (!jobRes.ok) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold tracking-tight">Job</h1>
          <Link
            href="/jobs"
            className="text-sm underline-offset-2 hover:underline"
          >
            Back to jobs
          </Link>
        </div>
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
          Failed to load job.
        </div>
      </div>
    );
  }

  const job: Job = await jobRes.json();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">{job.name}</h1>
        <Link
          href="/jobs"
          className="text-sm underline-offset-2 hover:underline"
        >
          Back to jobs
        </Link>
      </div>
      <JobDetailManager initialJob={job} />
    </div>
  );
}
