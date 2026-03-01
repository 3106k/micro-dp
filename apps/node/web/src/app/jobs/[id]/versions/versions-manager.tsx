"use client";

import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import type { components } from "@/lib/api/generated";

type JobVersion = components["schemas"]["JobVersion"];
type JobVersionDetail = components["schemas"]["JobVersionDetail"];

const defaultCreatePayload = `{
  "modules": [
    {
      "module_type_id": "replace-with-module-type-id",
      "name": "Module A"
    }
  ],
  "edges": []
}`;

export function VersionsManager({
  jobId,
  initialVersions,
}: {
  jobId: string;
  initialVersions: JobVersion[];
}) {
  const [versions, setVersions] = useState(initialVersions);
  const [createJson, setCreateJson] = useState(defaultCreatePayload);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<JobVersionDetail | null>(null);

  async function refreshVersions() {
    const res = await fetch(`/api/jobs/${jobId}/versions`, { cache: "no-store" });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error ?? "failed to fetch versions");
    }
    setVersions(data.items ?? []);
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      const body = JSON.parse(createJson);
      const res = await fetch(`/api/jobs/${jobId}/versions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "failed to create version");
      }
      await refreshVersions();
      setMessage("Version created.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  async function handlePublish(versionId: string) {
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch(`/api/jobs/${jobId}/versions/${versionId}/publish`, {
        method: "POST",
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "failed to publish version");
      }
      await refreshVersions();
      setMessage("Version published.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleViewDetail(versionId: string) {
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch(`/api/jobs/${jobId}/versions/${versionId}`);
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "failed to get version detail");
      }
      setDetail(data);
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-8">
      <form onSubmit={handleCreate} className="space-y-3 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Create Version</h2>
        <p className="text-xs text-muted-foreground">
          Provide request JSON for modules/edges.
        </p>
        <textarea
          className="h-56 w-full rounded-md border bg-background p-3 font-mono text-xs"
          value={createJson}
          onChange={(e) => setCreateJson(e.target.value)}
        />
        <Button type="submit" disabled={loading}>
          Create Version
        </Button>
        {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
      </form>

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Version</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Published</th>
              <th className="px-4 py-3 text-left font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {versions.map((v) => (
              <tr key={v.id} className="border-b last:border-0">
                <td className="px-4 py-3">{v.version}</td>
                <td className="px-4 py-3">{v.status}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {v.published_at ? new Date(v.published_at).toLocaleString() : "-"}
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleViewDetail(v.id)}
                      disabled={loading}
                    >
                      Detail
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => handlePublish(v.id)}
                      disabled={loading || v.status === "published"}
                    >
                      Publish
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
            {versions.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
                  No versions yet.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>

      {detail ? (
        <div className="rounded-lg border p-4">
          <h3 className="mb-2 text-lg font-semibold">Version Detail</h3>
          <pre className="overflow-auto rounded bg-muted p-3 text-xs">
            {JSON.stringify(detail, null, 2)}
          </pre>
        </div>
      ) : null}
    </div>
  );
}
