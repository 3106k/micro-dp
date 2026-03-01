"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { toErrorMessage, readApiErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";
import { track, flush } from "@micro-dp/sdk-tracker";

type Tenant = components["schemas"]["Tenant"];

export function DashboardHeader({
  displayName,
  email,
  platformRole,
  tenants,
  currentTenantId,
}: {
  displayName: string;
  email: string;
  platformRole: "user" | "superadmin";
  tenants: Tenant[];
  currentTenantId: string;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const { pushToast } = useToast();
  const [tenantId, setTenantId] = useState(currentTenantId);
  const [switchingTenant, setSwitchingTenant] = useState(false);
  const [signingOut, setSigningOut] = useState(false);

  const navItems: Array<{ href: string; label: string }> = [
    { href: "/dashboard", label: "Dashboard" },
    { href: "/jobs", label: "Jobs" },
    { href: "/job-runs", label: "Job Runs" },
    { href: "/datasets", label: "Datasets" },
    { href: "/connections", label: "Connections" },
  ];
  if (platformRole === "superadmin") {
    navItems.push({ href: "/admin/tenants", label: "Admin Tenants" });
  }

  async function handleSignOut() {
    setSigningOut(true);
    try {
      track("sign_out");
      flush({ useBeacon: true });
      await fetch("/api/auth/logout", { method: "POST" });
      router.push("/signin");
      router.refresh();
    } finally {
      setSigningOut(false);
    }
  }

  async function handleTenantSwitch(nextTenantId: string) {
    if (!nextTenantId || nextTenantId === tenantId) {
      return;
    }

    setSwitchingTenant(true);
    try {
      const res = await fetch("/api/auth/tenant", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ tenant_id: nextTenantId }),
      });

      if (!res.ok) {
        const message = await readApiErrorMessage(
          res,
          "Failed to switch tenant"
        );
        throw new Error(message);
      }

      setTenantId(nextTenantId);
      track("tenant_switched", { tenant_id: nextTenantId });
      pushToast({ variant: "success", message: "Tenant switched" });
      router.refresh();
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to switch tenant"),
      });
    } finally {
      setSwitchingTenant(false);
    }
  }

  return (
    <header className="border-b">
      <div className="container flex min-h-14 flex-wrap items-center justify-between gap-3 py-2">
        <div className="flex items-center gap-6">
          <span className="font-semibold">micro-dp</span>
          <nav className="flex items-center gap-2 text-sm">
            {navItems.map((item) => {
              const active = pathname === item.href;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={
                    active
                      ? "rounded-md bg-secondary px-2 py-1 font-medium"
                      : "rounded-md px-2 py-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                  }
                >
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>
        <div className="flex items-center gap-4">
          {tenants.length > 1 && (
            <label className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>Tenant</span>
              <select
                className="h-9 rounded-md border bg-background px-2 text-foreground"
                value={tenantId}
                onChange={(event) => handleTenantSwitch(event.target.value)}
                disabled={switchingTenant}
              >
                {tenants.map((tenant) => (
                  <option key={tenant.id} value={tenant.id}>
                    {tenant.name}
                  </option>
                ))}
              </select>
            </label>
          )}
          <span className="text-sm text-muted-foreground">
            {displayName || email}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={handleSignOut}
            disabled={signingOut}
          >
            Sign out
          </Button>
        </div>
      </div>
    </header>
  );
}
