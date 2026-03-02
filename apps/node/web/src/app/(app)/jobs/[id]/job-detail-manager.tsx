"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function JobDetailManager({ initialJob }: { initialJob: Job }) {
  const { pushToast } = useToast();
  const [job, setJob] = useState(initialJob);
  const [name, setName] = useState(initialJob.name);
  const [slug, setSlug] = useState(initialJob.slug);
  const [description, setDescription] = useState(
    initialJob.description ?? "",
  );
  const [isActive, setIsActive] = useState(initialJob.is_active);
  const [loading, setLoading] = useState(false);

  async function handleUpdate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await fetch(`/api/jobs/${job.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          slug,
          description,
          is_active: isActive,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "Failed to update job");
      }
      setJob(data);
      pushToast({ variant: "success", message: "Job updated" });
    } catch (err) {
      pushToast({
        variant: "error",
        message: err instanceof Error ? err.message : "Request failed",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      {/* Metadata */}
      <div className="grid gap-4 rounded-lg border p-4 md:grid-cols-2">
        <div>
          <p className="text-xs text-muted-foreground">ID</p>
          <p className="font-mono text-sm">{job.id}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Kind</p>
          <p className="mt-1">
            <span
              className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                kindStyles[job.kind] ?? "bg-secondary text-secondary-foreground"
              }`}
            >
              {job.kind}
            </span>
          </p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Status</p>
          <p className="mt-1">
            <span
              className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                job.is_active
                  ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                  : "bg-muted text-muted-foreground"
              }`}
            >
              {job.is_active ? "active" : "inactive"}
            </span>
          </p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Created</p>
          <p className="text-sm">
            {job.created_at ? formatDate(job.created_at) : "-"}
          </p>
        </div>
      </div>

      {/* Edit Form */}
      <form
        onSubmit={handleUpdate}
        className="space-y-4 rounded-lg border p-4"
      >
        <h2 className="text-lg font-semibold">Edit Job</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="job-name">Name</Label>
            <Input
              id="job-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="job-slug">Slug</Label>
            <Input
              id="job-slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="job-description">Description</Label>
            <Input
              id="job-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
            />
          </div>
        </div>
        <label className="inline-flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isActive}
            onChange={(e) => setIsActive(e.target.checked)}
            className="rounded border-input"
          />
          Active
        </label>
        <div className="flex gap-2">
          <Button type="submit" disabled={loading}>
            {loading ? "Updating..." : "Update"}
          </Button>
        </div>
      </form>

      {/* Actions */}
      <div className="rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Versions</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Manage modules and edges for this job.
        </p>
        <div className="mt-4">
          <Button variant="outline" asChild>
            <Link href={`/jobs/${job.id}/versions`}>Manage Versions</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
