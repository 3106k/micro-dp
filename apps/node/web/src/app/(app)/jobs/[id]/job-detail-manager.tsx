"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast-provider";
import type { components } from "@/lib/api/generated";

type Job = components["schemas"]["Job"];
type JobVersionDetail = components["schemas"]["JobVersionDetail"];
type JobModule = components["schemas"]["JobModule"];
type ModuleType = components["schemas"]["ModuleType"];

const kindStyles: Record<string, string> = {
  pipeline: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
  transform:
    "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400",
  import:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  export:
    "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400",
};

const categoryStyles: Record<string, string> = {
  source: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
  transform:
    "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400",
  destination:
    "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function renderModuleConfig(
  mod: JobModule,
  moduleType: ModuleType | undefined,
) {
  if (!mod.config_json) {
    return <p className="text-sm text-muted-foreground">No configuration</p>;
  }

  let parsed: Record<string, unknown>;
  try {
    parsed = JSON.parse(mod.config_json);
  } catch {
    return (
      <pre className="overflow-x-auto rounded bg-muted p-3 text-xs">
        {mod.config_json}
      </pre>
    );
  }

  const isTransform = moduleType?.category === "transform";
  const sqlValue = parsed["sql"];

  return (
    <div className="space-y-3">
      {isTransform && typeof sqlValue === "string" && (
        <pre className="overflow-x-auto rounded bg-muted p-3 text-xs">
          <code>{sqlValue}</code>
        </pre>
      )}
      {Object.entries(parsed)
        .filter(([key]) => !(isTransform && key === "sql"))
        .map(([key, value]) => (
          <div key={key} className="flex gap-2 text-sm">
            <span className="shrink-0 font-medium text-muted-foreground">
              {key}:
            </span>
            <span className="break-all">
              {typeof value === "string" ? value : JSON.stringify(value)}
            </span>
          </div>
        ))}
    </div>
  );
}

export function JobDetailManager({
  initialJob,
  publishedVersionDetail,
  moduleTypeMap,
}: {
  initialJob: Job;
  publishedVersionDetail: JobVersionDetail | null;
  moduleTypeMap: Record<string, ModuleType>;
}) {
  const { pushToast } = useToast();
  const router = useRouter();
  const [job, setJob] = useState(initialJob);
  const [name, setName] = useState(initialJob.name);
  const [slug, setSlug] = useState(initialJob.slug);
  const [description, setDescription] = useState(
    initialJob.description ?? "",
  );
  const [isActive, setIsActive] = useState(initialJob.is_active);
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);

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

  async function handleRun() {
    setRunning(true);
    try {
      const res = await fetch("/api/job-runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ job_id: job.id }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "Failed to create run");
      }
      pushToast({ variant: "success", message: "Run created" });
      router.push("/job-runs");
    } catch (err) {
      pushToast({
        variant: "error",
        message: err instanceof Error ? err.message : "Request failed",
      });
    } finally {
      setRunning(false);
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

      {/* Configuration */}
      <div className="rounded-lg border p-4">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-semibold">Configuration</h2>
          {publishedVersionDetail && (
            <span className="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800 dark:bg-green-900/30 dark:text-green-400">
              v{publishedVersionDetail.version.version} · published
            </span>
          )}
        </div>

        {!publishedVersionDetail ? (
          <div className="mt-3">
            <p className="text-sm text-muted-foreground">
              No published version.
            </p>
            <Button variant="outline" size="sm" className="mt-2" asChild>
              <Link href={`/jobs/${job.id}/versions`}>Manage Versions</Link>
            </Button>
          </div>
        ) : publishedVersionDetail.modules.length === 0 ? (
          <p className="mt-3 text-sm text-muted-foreground">
            No modules configured.
          </p>
        ) : (
          <div className="mt-3 space-y-4">
            {publishedVersionDetail.modules.map((mod) => {
              const mt = moduleTypeMap[mod.module_type_id];
              return (
                <div
                  key={mod.id}
                  className="rounded-md border bg-muted/30 p-3"
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-medium text-sm">{mod.name}</span>
                    {mt && (
                      <>
                        <span
                          className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                            categoryStyles[mt.category] ??
                            "bg-secondary text-secondary-foreground"
                          }`}
                        >
                          {mt.category}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {mt.name}
                        </span>
                      </>
                    )}
                  </div>
                  {mod.connection_id && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      Connection: <span className="font-mono">{mod.connection_id}</span>
                    </p>
                  )}
                  <div className="mt-2">
                    {renderModuleConfig(mod, mt)}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Run Job */}
      <div className="rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Run Job</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Create a new run for this job using the latest published version.
        </p>
        <div className="mt-4">
          <Button onClick={handleRun} disabled={running}>
            {running ? "Starting..." : "Run Now"}
          </Button>
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
