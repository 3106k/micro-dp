import type { components } from "@/lib/api/generated";

type JobRun = components["schemas"]["JobRun"];

const statusStyles: Record<JobRun["status"], string> = {
  queued:
    "bg-secondary text-secondary-foreground",
  running:
    "bg-primary/10 text-primary",
  success:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  failed:
    "bg-destructive/10 text-destructive",
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

export function JobRunsTable({ jobRuns }: { jobRuns: JobRun[] }) {
  if (jobRuns.length === 0) {
    return (
      <div className="rounded-lg border p-8 text-center text-muted-foreground">
        No job runs yet.
      </div>
    );
  }

  return (
    <div className="rounded-lg border">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="px-4 py-3 text-left font-medium">ID</th>
            <th className="px-4 py-3 text-left font-medium">Project</th>
            <th className="px-4 py-3 text-left font-medium">Job</th>
            <th className="px-4 py-3 text-left font-medium">Status</th>
            <th className="px-4 py-3 text-left font-medium">Started</th>
            <th className="px-4 py-3 text-left font-medium">Finished</th>
          </tr>
        </thead>
        <tbody>
          {jobRuns.map((run) => (
            <tr key={run.id} className="border-b last:border-0">
              <td className="px-4 py-3 font-mono text-xs">
                {run.id.slice(0, 8)}
              </td>
              <td className="px-4 py-3">{run.project_id}</td>
              <td className="px-4 py-3">{run.job_id}</td>
              <td className="px-4 py-3">
                <StatusBadge status={run.status} />
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {run.started_at
                  ? new Date(run.started_at).toLocaleString()
                  : "-"}
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {run.finished_at
                  ? new Date(run.finished_at).toLocaleString()
                  : "-"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
