"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";

type Job = components["schemas"]["Job"];

export function JobsManager({ initialJobs }: { initialJobs: Job[] }) {
  const [jobs, setJobs] = useState(initialJobs);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  async function refreshJobs() {
    const res = await fetch("/api/jobs", { cache: "no-store" });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error ?? "failed to fetch jobs");
    }
    setJobs(data.items ?? []);
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch("/api/jobs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, slug, description }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "failed to create job");
      }
      setName("");
      setSlug("");
      setDescription("");
      await refreshJobs();
      setMessage("Job created.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-8">
      <form onSubmit={handleCreate} className="space-y-4 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Create Job</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Job name"
            required
          />
          <Input
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="job-slug"
            required
          />
          <Input
            className="md:col-span-2"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description"
          />
        </div>
        <Button type="submit" disabled={loading}>
          Create
        </Button>
        {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
      </form>

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">Slug</th>
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
                <td className="px-4 py-3">{job.is_active ? "active" : "inactive"}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {job.updated_at ? new Date(job.updated_at).toLocaleString() : "-"}
                </td>
              </tr>
            ))}
            {jobs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
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
