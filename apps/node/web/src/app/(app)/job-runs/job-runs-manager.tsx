"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast-provider";
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

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function JobRunsManager({ initialRuns }: { initialRuns: JobRun[] }) {
  const { pushToast } = useToast();
  const [runs, setRuns] = useState(initialRuns);
  const [jobId, setJobId] = useState("");
  const [jobVersionId, setJobVersionId] = useState("");
  const [loading, setLoading] = useState(false);

  async function refreshRuns() {
    const res = await fetch("/api/job-runs", { cache: "no-store" });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error ?? "failed to fetch runs");
    }
    setRuns(data.items ?? []);
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    try {
      const payload: Record<string, string> = { job_id: jobId };
      if (jobVersionId) payload.job_version_id = jobVersionId;

      const res = await fetch("/api/job-runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "failed to create run");
      }
      setJobId("");
      setJobVersionId("");
      await refreshRuns();
      pushToast({ variant: "success", message: "Run created" });
    } catch (err) {
      pushToast({
        variant: "error",
        message: err instanceof Error ? err.message : "request failed",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <h1 className="text-2xl font-semibold tracking-tight">Job Runs</h1>

      <form onSubmit={handleCreate} className="space-y-4 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Create Job Run</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="run-job-id">Job ID</Label>
            <Input
              id="run-job-id"
              value={jobId}
              onChange={(e) => setJobId(e.target.value)}
              placeholder="job_id"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="run-version-id">Job Version ID</Label>
            <Input
              id="run-version-id"
              value={jobVersionId}
              onChange={(e) => setJobVersionId(e.target.value)}
              placeholder="optional"
            />
          </div>
        </div>
        <Button type="submit" disabled={loading}>
          {loading ? "Creating..." : "Create Run"}
        </Button>
      </form>

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">ID</th>
              <th className="px-4 py-3 text-left font-medium">Job</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Started</th>
            </tr>
          </thead>
          <tbody>
            {runs.map((run) => (
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
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusStyles[run.status]}`}
                  >
                    {run.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-muted-foreground" suppressHydrationWarning>
                  {run.started_at ? formatDateTime(run.started_at) : "-"}
                </td>
              </tr>
            ))}
            {runs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-sm text-muted-foreground">
                  No runs yet.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </div>
  );
}
