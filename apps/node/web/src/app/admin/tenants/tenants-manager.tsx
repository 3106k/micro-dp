"use client";

import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { components } from "@/lib/api/generated";

type Tenant = components["schemas"]["Tenant"];

export function TenantsManager({ initialTenants }: { initialTenants: Tenant[] }) {
  const [tenants, setTenants] = useState(initialTenants);
  const [newName, setNewName] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  async function refreshTenants() {
    const res = await fetch("/api/admin/tenants", { cache: "no-store" });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error ?? "failed to load tenants");
    }
    const data = (await res.json()) as { items: Tenant[] };
    setTenants(data.items ?? []);
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch("/api/admin/tenants", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error ?? "create failed");
      }
      setNewName("");
      await refreshTenants();
      setMessage("Tenant created.");
    } catch (e) {
      setMessage(e instanceof Error ? e.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  async function toggleActive(tenant: Tenant) {
    setLoading(true);
    setMessage("");
    try {
      const res = await fetch(`/api/admin/tenants/${tenant.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_active: !tenant.is_active }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error ?? "update failed");
      }
      await refreshTenants();
      setMessage("Tenant updated.");
    } catch (e) {
      setMessage(e instanceof Error ? e.message : "request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-8">
      <form onSubmit={handleCreate} className="space-y-4 rounded-lg border p-4">
        <h2 className="text-lg font-semibold">Create Tenant</h2>
        <div className="space-y-2">
          <Label htmlFor="tenant_name">Tenant Name</Label>
          <Input
            id="tenant_name"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
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
              <th className="px-4 py-3 text-left font-medium">ID</th>
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {tenants.map((tenant) => (
              <tr key={tenant.id} className="border-b last:border-0">
                <td className="px-4 py-3 font-mono text-xs">{tenant.id.slice(0, 8)}</td>
                <td className="px-4 py-3">{tenant.name}</td>
                <td className="px-4 py-3">
                  {tenant.is_active ? (
                    <span className="text-green-700">active</span>
                  ) : (
                    <span className="text-red-700">inactive</span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={loading}
                    onClick={() => toggleActive(tenant)}
                  >
                    {tenant.is_active ? "Deactivate" : "Activate"}
                  </Button>
                </td>
              </tr>
            ))}
            {tenants.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-muted-foreground">
                  No tenants yet.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </div>
  );
}
