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

export type NavItem = { href: string; label: string };

const mainNavItems: NavItem[] = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/dashboards", label: "Dashboards" },
  { href: "/jobs", label: "Jobs" },
  { href: "/job-runs", label: "Job Runs" },
  { href: "/datasets", label: "Datasets" },
  { href: "/connections", label: "Connections" },
];

export function AppHeader({
  displayName,
  email,
  platformRole,
  tenants,
  currentTenantId,
  sectionLabel,
}: {
  displayName: string;
  email: string;
  platformRole: "user" | "superadmin";
  tenants: Tenant[];
  currentTenantId: string;
  sectionLabel?: string;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const { pushToast } = useToast();
  const [tenantId, setTenantId] = useState(currentTenantId);
  const [switchingTenant, setSwitchingTenant] = useState(false);
  const [signingOut, setSigningOut] = useState(false);

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

  function isActive(href: string): boolean {
    return pathname === href || pathname.startsWith(href + "/");
  }

  return (
    <header className="border-b">
      <div className="container flex min-h-14 flex-wrap items-center justify-between gap-3 py-2">
        <div className="flex items-center gap-6">
          <span className="font-semibold">
            micro-dp
            {sectionLabel ? (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                / {sectionLabel}
              </span>
            ) : null}
          </span>
          <nav className="flex items-center gap-2 text-sm">
            {mainNavItems.map((item) => {
              const active = isActive(item.href);
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
          <Link
            href="/members"
            className="rounded-md px-2 py-1 text-muted-foreground hover:bg-muted hover:text-foreground"
            title="Settings"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </Link>
          {platformRole === "superadmin" ? (
            <Link
              href="/admin/tenants"
              className="rounded-md px-2 py-1 text-muted-foreground hover:bg-muted hover:text-foreground"
              title="Admin"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" />
              </svg>
            </Link>
          ) : null}
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
