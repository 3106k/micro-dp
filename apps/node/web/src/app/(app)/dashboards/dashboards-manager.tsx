"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { readApiErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";

type Dashboard = components["schemas"]["Dashboard"];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function DashboardsManager({
  initialDashboards,
}: {
  initialDashboards: Dashboard[];
}) {
  const router = useRouter();
  const { pushToast } = useToast();
  const [dashboards, setDashboards] = useState(initialDashboards);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  async function refreshList() {
    try {
      const res = await fetch("/api/dashboards", { cache: "no-store" });
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to fetch dashboards");
        throw new Error(msg);
      }
      const data = await res.json();
      setDashboards(data.items ?? []);
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to fetch dashboards",
      });
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;

    setCreating(true);
    try {
      const res = await fetch("/api/dashboards", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          description: description.trim() || undefined,
        }),
      });
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to create dashboard");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Dashboard created" });
      setName("");
      setDescription("");
      await refreshList();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to create dashboard",
      });
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    setDeleting(id);
    try {
      const res = await fetch(`/api/dashboards/${id}`, { method: "DELETE" });
      if (!res.ok && res.status !== 204) {
        const msg = await readApiErrorMessage(res, "Failed to delete dashboard");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Dashboard deleted" });
      await refreshList();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to delete dashboard",
      });
    } finally {
      setDeleting(null);
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <h1 className="text-2xl font-semibold tracking-tight">Dashboards</h1>

      {/* Create form */}
      <form
        onSubmit={handleCreate}
        className="rounded-lg border p-4 space-y-4"
      >
        <h2 className="text-lg font-semibold">Create Dashboard</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <label htmlFor="dash-name" className="text-sm font-medium">
              Name <span className="text-destructive">*</span>
            </label>
            <input
              id="dash-name"
              type="text"
              className="flex h-9 w-full rounded-md border bg-background px-3 py-1 text-sm"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Dashboard"
              required
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="dash-desc" className="text-sm font-medium">
              Description
            </label>
            <input
              id="dash-desc"
              type="text"
              className="flex h-9 w-full rounded-md border bg-background px-3 py-1 text-sm"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
            />
          </div>
        </div>
        <Button type="submit" disabled={creating || !name.trim()}>
          {creating ? "Creating..." : "Create"}
        </Button>
      </form>

      {/* Table */}
      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">Description</th>
              <th className="px-4 py-3 text-left font-medium">Created</th>
              <th className="px-4 py-3 text-right font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {dashboards.map((d) => (
              <tr
                key={d.id}
                className="border-b last:border-0 cursor-pointer hover:bg-muted/50"
                onClick={() => router.push(`/dashboards/${d.id}`)}
              >
                <td className="px-4 py-3 font-medium">{d.name}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {d.description || "-"}
                </td>
                <td className="px-4 py-3 text-muted-foreground">
                  {formatDate(d.created_at)}
                </td>
                <td className="px-4 py-3 text-right">
                  <Button
                    variant="destructive"
                    size="sm"
                    disabled={deleting === d.id}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDelete(d.id);
                    }}
                  >
                    {deleting === d.id ? "Deleting..." : "Delete"}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {dashboards.length === 0 ? (
          <div className="px-4 py-8 text-center text-sm text-muted-foreground">
            No dashboards yet. Create one to get started.
          </div>
        ) : null}
      </div>
    </div>
  );
}
