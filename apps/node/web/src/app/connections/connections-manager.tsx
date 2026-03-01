"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { components } from "@/lib/api/generated";
import { ConnectorSchemaForm } from "./connector-schema-form";

type Connection = components["schemas"]["Connection"];
type ConnectorDefinition = components["schemas"]["ConnectorDefinition"];

type FormState = {
  name: string;
  type: string;
  configValues: Record<string, unknown>;
  secret_ref: string;
};

const emptyForm: FormState = {
  name: "",
  type: "",
  configValues: {},
  secret_ref: "",
};

export function ConnectionsManager({
  initialConnections,
  initialConnectors,
}: {
  initialConnections: Connection[];
  initialConnectors: ConnectorDefinition[];
}) {
  const [connections, setConnections] = useState(initialConnections);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [message, setMessage] = useState<string>("");
  const [loading, setLoading] = useState(false);

  // Spec fetched for the currently selected connector type
  const [spec, setSpec] = useState<Record<string, unknown> | null>(null);
  const [specLoading, setSpecLoading] = useState(false);
  const [testMessage, setTestMessage] = useState<string>("");
  const [testLoading, setTestLoading] = useState(false);

  const isEditing = useMemo(() => editingId !== null, [editingId]);

  const fetchSpec = useCallback(async (connectorId: string) => {
    setSpecLoading(true);
    setSpec(null);
    try {
      const res = await fetch(`/api/connectors/${encodeURIComponent(connectorId)}`);
      if (!res.ok) {
        setSpec(null);
        return;
      }
      const data = await res.json();
      setSpec(data.spec ?? null);
      // Apply default values from spec
      const properties = (data.spec?.properties ?? {}) as Record<
        string,
        { default?: unknown }
      >;
      const defaults: Record<string, unknown> = {};
      for (const [key, prop] of Object.entries(properties)) {
        if (prop.default !== undefined) {
          defaults[key] = prop.default;
        }
      }
      setForm((prev) => ({
        ...prev,
        configValues: { ...defaults, ...prev.configValues },
      }));
    } finally {
      setSpecLoading(false);
    }
  }, []);

  async function refreshConnections() {
    const res = await fetch("/api/connections", { cache: "no-store" });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error ?? "failed to load connections");
    }
    const data = (await res.json()) as { items: Connection[] };
    setConnections(data.items ?? []);
  }

  function handleTypeChange(connectorId: string) {
    setForm((prev) => ({ ...prev, type: connectorId, configValues: {} }));
    setSpec(null);
    setTestMessage("");
    if (connectorId) {
      fetchSpec(connectorId);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      const payload = {
        name: form.name,
        type: form.type,
        config_json: JSON.stringify(form.configValues),
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
      setSpec(null);
      setTestMessage("");
      setMessage(isEditing ? "Connection updated." : "Connection created.");
    } catch (e) {
      setMessage(e instanceof Error ? e.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleTestConnection() {
    setTestLoading(true);
    setTestMessage("");
    try {
      const res = await fetch("/api/connections/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          type: form.type,
          config_json: JSON.stringify(form.configValues),
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setTestMessage(data.error ?? "Test failed");
        return;
      }
      setTestMessage(
        data.status === "ok"
          ? "Connection test passed."
          : `Test failed: ${data.message ?? "unknown error"}`
      );
    } catch (e) {
      setTestMessage(e instanceof Error ? e.message : "request failed");
    } finally {
      setTestLoading(false);
    }
  }

  function beginEdit(connection: Connection) {
    setEditingId(connection.id);
    let configValues: Record<string, unknown> = {};
    try {
      configValues = JSON.parse(connection.config_json ?? "{}");
    } catch {
      configValues = {};
    }
    setForm({
      name: connection.name,
      type: connection.type,
      configValues,
      secret_ref: connection.secret_ref ?? "",
    });
    setMessage("");
    setTestMessage("");
    fetchSpec(connection.type);
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
        setSpec(null);
        setTestMessage("");
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
            <Label htmlFor="type">
              Type <span className="text-destructive">*</span>
            </Label>
            <select
              id="type"
              value={form.type}
              onChange={(e) => handleTypeChange(e.target.value)}
              required
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <option value="">Select connector...</option>
              {initialConnectors.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {specLoading ? (
          <p className="text-sm text-muted-foreground">Loading connector spec...</p>
        ) : null}

        {spec ? (
          <ConnectorSchemaForm
            spec={spec}
            values={form.configValues}
            onChange={(configValues) =>
              setForm((prev) => ({ ...prev, configValues }))
            }
          />
        ) : null}

        <div className="space-y-2">
          <Label htmlFor="secret_ref">Secret Ref</Label>
          <Input
            id="secret_ref"
            value={form.secret_ref}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, secret_ref: e.target.value }))
            }
          />
        </div>

        <div className="flex items-center gap-2">
          <Button type="submit" disabled={loading}>
            {isEditing ? "Update" : "Create"}
          </Button>
          {form.type ? (
            <Button
              type="button"
              variant="outline"
              disabled={testLoading || !form.type}
              onClick={handleTestConnection}
            >
              {testLoading ? "Testing..." : "Test Connection"}
            </Button>
          ) : null}
          {isEditing ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setEditingId(null);
                setForm(emptyForm);
                setSpec(null);
                setTestMessage("");
              }}
            >
              Cancel
            </Button>
          ) : null}
        </div>
        {testMessage ? (
          <p className="text-sm text-muted-foreground">{testMessage}</p>
        ) : null}
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
