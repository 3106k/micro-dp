"use client";

import Link from "next/link";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import type { components } from "@/lib/api/generated";

type Job = components["schemas"]["Job"];

const kindBadgeColors: Record<string, string> = {
  pipeline: "bg-blue-100 text-blue-800",
  transform: "bg-purple-100 text-purple-800",
  import: "bg-green-100 text-green-800",
  export: "bg-orange-100 text-orange-800",
};

export function JobsManager({ initialJobs }: { initialJobs: Job[] }) {
  const [jobs, setJobs] = useState(initialJobs);

  async function refreshJobs() {
    const res = await fetch("/api/jobs", { cache: "no-store" });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error ?? "failed to fetch jobs");
    }
    setJobs(data.items ?? []);
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-3">
        <Link href="/jobs/new">
          <Button>Create Transform Job</Button>
        </Link>
        <Button variant="outline" disabled title="Coming Soon">
          Import Job
        </Button>
        <Button variant="outline" disabled title="Coming Soon">
          Export Job
        </Button>
      </div>

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">Slug</th>
              <th className="px-4 py-3 text-left font-medium">Kind</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Updated</th>
            </tr>
          </thead>
          <tbody>
            {jobs.map((job) => (
              <tr key={job.id} className="border-b last:border-0">
                <td className="px-4 py-3">
                  <Link
                    className="font-medium underline-offset-2 hover:underline"
                    href={`/jobs/${job.id}`}
                  >
                    {job.name}
                  </Link>
                </td>
                <td className="px-4 py-3">{job.slug}</td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      kindBadgeColors[job.kind] ?? "bg-gray-100 text-gray-800"
                    }`}
                  >
                    {job.kind}
                  </span>
                </td>
                <td className="px-4 py-3">
                  {job.is_active ? "active" : "inactive"}
                </td>
                <td className="px-4 py-3 text-muted-foreground">
                  {job.updated_at
                    ? new Date(job.updated_at).toLocaleString()
                    : "-"}
                </td>
              </tr>
            ))}
            {jobs.length === 0 ? (
              <tr>
                <td
                  colSpan={5}
                  className="px-4 py-8 text-center text-muted-foreground"
                >
                  No jobs yet.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </div>
  );
}
