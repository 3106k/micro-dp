"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";

type JobRun = components["schemas"]["JobRun"];

export function JobRunsManager({ initialRuns }: { initialRuns: JobRun[] }) {
  const [runs, setRuns] = useState(initialRuns);
  const [jobId, setJobId] = useState("");
  const [jobVersionId, setJobVersionId] = useState("");
  const [message, setMessage] = useState("");
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
    setMessage("");
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
      setMessage("Run created.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-8">
      <form onSubmit={handleCreate} className="space-y-4 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Create Job Run</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <Input
            value={jobId}
            onChange={(e) => setJobId(e.target.value)}
            placeholder="job_id"
            required
          />
          <Input
            value={jobVersionId}
            onChange={(e) => setJobVersionId(e.target.value)}
            placeholder="job_version_id (optional)"
          />
        </div>
        <Button type="submit" disabled={loading}>
          Create Run
        </Button>
        {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
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
                  <Link href={`/job-runs/${run.id}`} className="font-mono underline">
                    {run.id.slice(0, 8)}
                  </Link>
                </td>
                <td className="px-4 py-3">{run.job_id}</td>
                <td className="px-4 py-3">{run.status}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {run.started_at ? new Date(run.started_at).toLocaleString() : "-"}
                </td>
              </tr>
            ))}
            {runs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
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
