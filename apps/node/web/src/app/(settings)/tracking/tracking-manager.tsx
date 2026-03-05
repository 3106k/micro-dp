"use client";

import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast-provider";
import { readApiErrorMessage, toErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";

type WriteKey = components["schemas"]["WriteKey"];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function TrackingManager({
  initialKeys,
}: {
  initialKeys: WriteKey[];
}) {
  const { pushToast } = useToast();
  const [keys, setKeys] = useState(initialKeys);
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(false);
  const [newRawKey, setNewRawKey] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  async function refreshKeys() {
    const res = await fetch("/api/write-keys", { cache: "no-store" });
    if (!res.ok) {
      const message = await readApiErrorMessage(res, "Failed to load keys");
      throw new Error(message);
    }
    const data = await res.json();
    setKeys(data.items ?? []);
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    setLoading(true);
    try {
      const res = await fetch("/api/write-keys", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: name.trim() }),
      });
      if (!res.ok) {
        const message = await readApiErrorMessage(res, "Failed to create key");
        throw new Error(message);
      }
      const data = await res.json();
      setNewRawKey(data.raw_key);
      setName("");
      await refreshKeys();
      pushToast({ variant: "success", message: "Write key created" });
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to create key"),
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleRegenerate(id: string) {
    setLoading(true);
    try {
      const res = await fetch(`/api/write-keys/${id}/regenerate`, {
        method: "POST",
      });
      if (!res.ok) {
        const message = await readApiErrorMessage(
          res,
          "Failed to regenerate key"
        );
        throw new Error(message);
      }
      const data = await res.json();
      setNewRawKey(data.raw_key);
      await refreshKeys();
      pushToast({ variant: "success", message: "Write key regenerated" });
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to regenerate key"),
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    setLoading(true);
    try {
      const res = await fetch(`/api/write-keys/${deleteTarget}`, {
        method: "DELETE",
      });
      if (!res.ok && res.status !== 204) {
        const message = await readApiErrorMessage(res, "Failed to delete key");
        throw new Error(message);
      }
      setDeleteTarget(null);
      await refreshKeys();
      pushToast({ variant: "success", message: "Write key deleted" });
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to delete key"),
      });
    } finally {
      setLoading(false);
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="space-y-8">
      {/* Create form */}
      <form onSubmit={handleCreate} className="flex items-end gap-3">
        <div className="flex-1 max-w-xs">
          <Label htmlFor="key-name">Key name</Label>
          <Input
            id="key-name"
            placeholder="e.g. Production site"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <Button type="submit" disabled={loading || !name.trim()}>
          Create write key
        </Button>
      </form>

      {/* New key banner */}
      {newRawKey && (
        <div className="rounded-md border bg-muted/50 p-4 space-y-2">
          <p className="text-sm font-medium">
            Your new write key (copy it now &mdash; it won&apos;t be shown
            again):
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 rounded bg-muted px-3 py-2 text-sm font-mono break-all">
              {newRawKey}
            </code>
            <Button
              variant="outline"
              size="sm"
              onClick={() => copyToClipboard(newRawKey)}
            >
              {copied ? "Copied" : "Copy"}
            </Button>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setNewRawKey(null)}
          >
            Dismiss
          </Button>
        </div>
      )}

      {/* Quick Start snippet */}
      {keys.length > 0 && (
        <div className="space-y-2">
          <h2 className="text-lg font-medium">Quick Start</h2>
          <p className="text-sm text-muted-foreground">
            Add this snippet to your website to start collecting events:
          </p>
          <pre className="rounded-md border bg-muted/50 p-4 text-xs overflow-x-auto">
{`<script src="https://your-cdn.com/micro-dp.global.js"></script>
<script>
  MicroDP.init({
    endpoint: "${typeof window !== "undefined" ? window.location.origin : ""}/api/v1/collect",
    writeKey: "${keys[0]?.key_prefix}...",
    collectContext: true,
    debug: false
  });
  MicroDP.page();
</script>`}
          </pre>
        </div>
      )}

      {/* Keys table */}
      {keys.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No write keys yet. Create one to start collecting events from external
          sites.
        </p>
      ) : (
        <div className="rounded-md border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-2 text-left font-medium">Name</th>
                <th className="px-4 py-2 text-left font-medium">Key prefix</th>
                <th className="px-4 py-2 text-left font-medium">Status</th>
                <th className="px-4 py-2 text-left font-medium">Created</th>
                <th className="px-4 py-2 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {keys.map((k) => (
                <tr key={k.id} className="border-b last:border-0">
                  <td className="px-4 py-2">{k.name}</td>
                  <td className="px-4 py-2">
                    <code className="text-xs">{k.key_prefix}...</code>
                  </td>
                  <td className="px-4 py-2">
                    <span
                      className={
                        k.is_active
                          ? "text-green-600 dark:text-green-400"
                          : "text-muted-foreground"
                      }
                    >
                      {k.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-4 py-2">{formatDate(k.created_at)}</td>
                  <td className="px-4 py-2 text-right space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={loading}
                      onClick={() => handleRegenerate(k.id)}
                    >
                      Regenerate
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      disabled={loading}
                      onClick={() => setDeleteTarget(k.id)}
                    >
                      Delete
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Delete confirmation dialog */}
      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete write key</DialogTitle>
            <DialogDescription>
              This will permanently delete the write key. Any sites using it will
              stop collecting events.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={loading}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
