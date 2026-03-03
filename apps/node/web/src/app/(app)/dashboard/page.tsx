import Link from "next/link";

import { Button } from "@/components/ui/button";
import { backendFetch } from "@/lib/api/server";
import type { components } from "@/lib/api/generated";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { EventSummary } from "@/components/event-summary";
import { JobRunsTable } from "./job-runs-table";

type JobRun = components["schemas"]["JobRun"];
type EventsSummaryResponse = components["schemas"]["EventsSummaryResponse"];

export default async function DashboardPage() {
  const { token, currentTenantId } = await getAuthContext();

  const headers = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": currentTenantId,
  };

  const [jobRunsRes, eventsRes] = await Promise.all([
    backendFetch("/api/v1/job_runs?limit=5", { headers, cache: "no-store" }),
    backendFetch("/api/v1/events/summary", { headers, cache: "no-store" }),
  ]);

  const jobRuns: JobRun[] = jobRunsRes.ok
    ? ((await jobRunsRes.json()).items ?? [])
    : [];

  const eventsSummary: EventsSummaryResponse | null = eventsRes.ok
    ? await eventsRes.json()
    : null;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <div className="flex gap-2">
          <Button variant="outline" asChild>
            <Link href="/datasets/upload">Upload CSV</Link>
          </Button>
          <Button asChild>
            <Link href="/jobs/new">Create Job</Link>
          </Button>
        </div>
      </div>

      <EventSummary summary={eventsSummary} />

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Recent Job Runs</h2>
          {jobRuns.length > 0 ? (
            <Button variant="link" asChild className="px-0">
              <Link href="/job-runs">View all</Link>
            </Button>
          ) : null}
        </div>
        <JobRunsTable jobRuns={jobRuns} />
      </div>
    </div>
  );
}
