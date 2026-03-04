"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import type { components } from "@/lib/api/generated";

type Job = components["schemas"]["Job"];

const kindStyles: Record<string, string> = {
  pipeline: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
  transform:
    "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400",
  import:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  export:
    "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400",
};

const statusStyles: Record<string, string> = {
  active:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  inactive: "bg-muted text-muted-foreground",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function JobsManager({ initialJobs }: { initialJobs: Job[] }) {
  const router = useRouter();
  const { pushToast } = useToast();
  const [jobs, setJobs] = useState(initialJobs);

  async function refreshJobs() {
    try {
      const res = await fetch("/api/jobs", { cache: "no-store" });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error ?? "Failed to fetch jobs");
      }
      const data = await res.json();
      setJobs(data.items ?? []);
    } catch (error) {
      pushToast({
        variant: "error",
        message: error instanceof Error ? error.message : "Failed to fetch jobs",
      });
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
        <Button asChild>
          <Link href="/jobs/new">Create Job</Link>
        </Button>
      </div>

      {/* Table */}
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
              <tr
                key={job.id}
                className="border-b last:border-0 cursor-pointer hover:bg-muted/50"
                onClick={() => router.push(`/jobs/${job.id}`)}
              >
                <td className="px-4 py-3 font-medium">{job.name}</td>
                <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                  {job.slug}
                </td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      kindStyles[job.kind] ??
                      "bg-secondary text-secondary-foreground"
                    }`}
                  >
                    {job.kind}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      statusStyles[job.is_active ? "active" : "inactive"]
                    }`}
                  >
                    {job.is_active ? "active" : "inactive"}
                  </span>
                </td>
                <td className="px-4 py-3 text-muted-foreground">
                  {job.updated_at ? formatDate(job.updated_at) : "-"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {jobs.length === 0 ? (
          <div className="px-4 py-8 text-center text-sm text-muted-foreground">
            No jobs found. Create one to get started.
          </div>
        ) : null}
      </div>
    </div>
  );
}
