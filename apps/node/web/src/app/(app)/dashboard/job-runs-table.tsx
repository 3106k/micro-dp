import Link from "next/link";

import type { components } from "@/lib/api/generated";

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

function StatusBadge({ status }: { status: JobRun["status"] }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusStyles[status]}`}
    >
      {status}
    </span>
  );
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function JobRunsTable({ jobRuns }: { jobRuns: JobRun[] }) {
  if (jobRuns.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 rounded-lg border border-dashed p-12 text-center">
        <p className="text-sm text-muted-foreground">No job runs yet.</p>
        <p className="text-xs text-muted-foreground">
          Create a job and run it to see results here.
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="px-4 py-3 text-left font-medium">ID</th>
            <th className="px-4 py-3 text-left font-medium">Job</th>
            <th className="px-4 py-3 text-left font-medium">Status</th>
            <th className="px-4 py-3 text-left font-medium">Started</th>
            <th className="px-4 py-3 text-left font-medium">Finished</th>
          </tr>
        </thead>
        <tbody>
          {jobRuns.map((run) => (
            <tr key={run.id} className="border-b last:border-0">
              <td className="px-4 py-3">
                <Link
                  href={`/job-runs/${run.id}`}
                  className="font-mono text-xs text-primary hover:underline"
                >
                  {run.id.slice(0, 8)}
                </Link>
              </td>
              <td className="px-4 py-3">{run.job_id}</td>
              <td className="px-4 py-3">
                <StatusBadge status={run.status} />
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {run.started_at ? formatDateTime(run.started_at) : "-"}
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {run.finished_at ? formatDateTime(run.finished_at) : "-"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
