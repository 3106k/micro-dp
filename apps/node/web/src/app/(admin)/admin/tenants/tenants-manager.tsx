"use client";

import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast-provider";
import type { components } from "@/lib/api/generated";

type Tenant = components["schemas"]["Tenant"];

const statusStyles: Record<string, string> = {
  active:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  inactive: "bg-muted text-muted-foreground",
};

export function TenantsManager({ initialTenants }: { initialTenants: Tenant[] }) {
  const { pushToast } = useToast();
  const [tenants, setTenants] = useState(initialTenants);
  const [newName, setNewName] = useState("");
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
      pushToast({ variant: "success", message: "Tenant created" });
    } catch (e) {
      pushToast({
        variant: "error",
        message: e instanceof Error ? e.message : "request failed",
      });
    } finally {
      setLoading(false);
    }
  }

  async function toggleActive(tenant: Tenant) {
    setLoading(true);
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
      pushToast({ variant: "success", message: "Tenant updated" });
    } catch (e) {
      pushToast({
        variant: "error",
        message: e instanceof Error ? e.message : "request failed",
      });
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
          {loading ? "Creating..." : "Create"}
        </Button>
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
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      statusStyles[tenant.is_active ? "active" : "inactive"]
                    }`}
                  >
                    {tenant.is_active ? "active" : "inactive"}
                  </span>
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
                <td colSpan={4} className="px-4 py-6 text-center text-sm text-muted-foreground">
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
