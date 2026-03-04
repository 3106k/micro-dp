import Link from "next/link";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { JobDetailManager } from "./job-detail-manager";

type Job = components["schemas"]["Job"];
type JobVersion = components["schemas"]["JobVersion"];
type JobVersionDetail = components["schemas"]["JobVersionDetail"];
type ModuleType = components["schemas"]["ModuleType"];

export default async function JobDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { token, currentTenantId } = await getAuthContext();

  const headers = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": currentTenantId,
  };

  const jobRes = await backendFetch(`/api/v1/jobs/${id}`, {
    headers,
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

  // Fetch versions list and module types in parallel
  const [versionsRes, moduleTypesRes] = await Promise.all([
    backendFetch(`/api/v1/jobs/${id}/versions`, { headers, cache: "no-store" }),
    backendFetch(`/api/v1/module_types`, { headers, cache: "no-store" }),
  ]);

  let publishedVersionDetail: JobVersionDetail | null = null;
  let moduleTypeMap: Record<string, ModuleType> = {};

  if (moduleTypesRes.ok) {
    const mtData: { items: ModuleType[] } = await moduleTypesRes.json();
    moduleTypeMap = Object.fromEntries(mtData.items.map((mt) => [mt.id, mt]));
  }

  if (versionsRes.ok) {
    const versionsData: { items: JobVersion[] } = await versionsRes.json();
    const published = versionsData.items
      .filter((v) => v.status === "published")
      .sort((a, b) => b.version - a.version);

    if (published.length > 0) {
      const latestPublished = published[0];
      const detailRes = await backendFetch(
        `/api/v1/jobs/${id}/versions/${latestPublished.id}`,
        { headers, cache: "no-store" },
      );
      if (detailRes.ok) {
        publishedVersionDetail = await detailRes.json();
      }
    }
  }

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
      <JobDetailManager
        initialJob={job}
        publishedVersionDetail={publishedVersionDetail}
        moduleTypeMap={moduleTypeMap}
      />
    </div>
  );
}
