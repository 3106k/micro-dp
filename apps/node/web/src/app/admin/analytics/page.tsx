import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { EventSummary } from "@/app/dashboard/event-summary";

type MeResponse = components["schemas"]["MeResponse"];
type EventsSummaryResponse = components["schemas"]["EventsSummaryResponse"];

export default async function AdminAnalyticsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantCookie = jar.get(TENANT_COOKIE)?.value;
  if (!token) {
    redirect("/signin");
  }

  const meRes = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!meRes.ok) {
    redirect("/signin");
  }
  const me: MeResponse = await meRes.json();
  const tenantIds = new Set(me.tenants.map((tenant) => tenant.id));
  const currentTenantId =
    tenantCookie && tenantIds.has(tenantCookie)
      ? tenantCookie
      : me.tenants[0]?.id ?? "";

  if (me.platform_role !== "superadmin") {
    return (
      <div className="min-h-screen">
        <DashboardHeader
          displayName={me.display_name}
          email={me.email}
          platformRole={me.platform_role}
          tenants={me.tenants}
          currentTenantId={currentTenantId}
        />
        <main className="container py-8">
          <div className="rounded-lg border p-6">
            <h1 className="text-2xl font-semibold tracking-tight">403 Forbidden</h1>
            <p className="mt-2 text-muted-foreground">
              This page is available only for superadmin users.
            </p>
          </div>
        </main>
      </div>
    );
  }

  let eventsSummary: EventsSummaryResponse | null = null;
  const summaryRes = await backendFetch("/api/v1/tracker/summary", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (summaryRes.ok) {
    eventsSummary = await summaryRes.json();
  }

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
        tenants={me.tenants}
        currentTenantId={currentTenantId}
      />
      <main className="container space-y-6 py-8">
        <h1 className="text-2xl font-semibold tracking-tight">Analytics</h1>
        <EventSummary summary={eventsSummary} />
      </main>
    </div>
  );
}
