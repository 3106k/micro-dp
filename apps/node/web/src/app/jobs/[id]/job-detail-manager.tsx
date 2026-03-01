"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";

type Job = components["schemas"]["Job"];

export function JobDetailManager({ initialJob }: { initialJob: Job }) {
  const [job, setJob] = useState(initialJob);
  const [name, setName] = useState(initialJob.name);
  const [slug, setSlug] = useState(initialJob.slug);
  const [description, setDescription] = useState(initialJob.description ?? "");
  const [isActive, setIsActive] = useState(initialJob.is_active);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleUpdate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setMessage("");
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
        throw new Error(data.error ?? "failed to update job");
      }
      setJob(data);
      setMessage("Job updated.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleUpdate} className="space-y-4 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Job Detail</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <Input value={name} onChange={(e) => setName(e.target.value)} required />
          <Input value={slug} onChange={(e) => setSlug(e.target.value)} required />
          <Input
            className="md:col-span-2"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description"
          />
        </div>
        <label className="inline-flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isActive}
            onChange={(e) => setIsActive(e.target.checked)}
          />
          Active
        </label>
        <div>
          <Button type="submit" disabled={loading}>
            Update
          </Button>
        </div>
        {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
      </form>

      <div className="rounded-lg border p-4">
        <p className="text-sm text-muted-foreground">Job ID</p>
        <p className="font-mono text-sm">{job.id}</p>
        <div className="mt-4">
          <Link href={`/jobs/${job.id}/versions`} className="text-sm underline">
            Manage Versions
          </Link>
        </div>
      </div>
    </div>
  );
}
