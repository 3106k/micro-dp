"use client";

import { FormEvent, useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { components } from "@/lib/api/generated";

type Connection = components["schemas"]["Connection"];

type FormState = {
  name: string;
  type: string;
  config_json: string;
  secret_ref: string;
};

const emptyForm: FormState = {
  name: "",
  type: "",
  config_json: "{}",
  secret_ref: "",
};

export function ConnectionsManager({
  initialConnections,
}: {
  initialConnections: Connection[];
}) {
  const [connections, setConnections] = useState(initialConnections);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [message, setMessage] = useState<string>("");
  const [loading, setLoading] = useState(false);

  const isEditing = useMemo(() => editingId !== null, [editingId]);

  async function refreshConnections() {
    const res = await fetch("/api/connections", { cache: "no-store" });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error ?? "failed to load connections");
    }
    const data = (await res.json()) as { items: Connection[] };
    setConnections(data.items ?? []);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      const payload = {
        name: form.name,
        type: form.type,
        config_json: form.config_json || "{}",
        secret_ref: form.secret_ref || undefined,
      };

      const res = await fetch(
        isEditing ? `/api/connections/${editingId}` : "/api/connections",
        {
          method: isEditing ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        }
      );

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error ?? "save failed");
      }

      await refreshConnections();
      setForm(emptyForm);
      setEditingId(null);
      setMessage(isEditing ? "Connection updated." : "Connection created.");
    } catch (e) {
      setMessage(e instanceof Error ? e.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  function beginEdit(connection: Connection) {
    setEditingId(connection.id);
    setForm({
      name: connection.name,
      type: connection.type,
      config_json: connection.config_json ?? "{}",
      secret_ref: connection.secret_ref ?? "",
    });
    setMessage("");
  }

  async function handleDelete(id: string) {
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch(`/api/connections/${id}`, { method: "DELETE" });
      if (!res.ok && res.status !== 204) {
        const err = await res.json();
        throw new Error(err.error ?? "delete failed");
      }
      await refreshConnections();
      if (editingId === id) {
        setEditingId(null);
        setForm(emptyForm);
      }
      setMessage("Connection deleted.");
    } catch (e) {
      setMessage(e instanceof Error ? e.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-8">
      <form onSubmit={handleSubmit} className="space-y-4 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">
          {isEditing ? "Edit Connection" : "Create Connection"}
        </h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={form.name}
              onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="type">Type</Label>
            <Input
              id="type"
              value={form.type}
              onChange={(e) => setForm((prev) => ({ ...prev, type: e.target.value }))}
              required
            />
          </div>
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="config_json">Config JSON</Label>
            <Input
              id="config_json"
              value={form.config_json}
              onChange={(e) =>
                setForm((prev) => ({ ...prev, config_json: e.target.value }))
              }
            />
          </div>
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="secret_ref">Secret Ref</Label>
            <Input
              id="secret_ref"
              value={form.secret_ref}
              onChange={(e) =>
                setForm((prev) => ({ ...prev, secret_ref: e.target.value }))
              }
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button type="submit" disabled={loading}>
            {isEditing ? "Update" : "Create"}
          </Button>
          {isEditing ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setEditingId(null);
                setForm(emptyForm);
              }}
            >
              Cancel
            </Button>
          ) : null}
        </div>
        {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
      </form>

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">Type</th>
              <th className="px-4 py-3 text-left font-medium">Secret Ref</th>
              <th className="px-4 py-3 text-left font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {connections.map((connection) => (
              <tr key={connection.id} className="border-b last:border-0">
                <td className="px-4 py-3">{connection.name}</td>
                <td className="px-4 py-3">{connection.type}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {connection.secret_ref ?? "-"}
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => beginEdit(connection)}
                      disabled={loading}
                    >
                      Edit
                    </Button>
                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={() => handleDelete(connection.id)}
                      disabled={loading}
                    >
                      Delete
                    </Button>
                  </div>
                </td>
              </tr>
            ))}
            {connections.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-muted-foreground">
                  No connections yet.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </div>
  );
}
