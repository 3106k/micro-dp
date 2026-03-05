import Link from "next/link";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { DashboardDetailManager } from "./dashboard-detail-manager";

type Dashboard = components["schemas"]["Dashboard"];
type DashboardWidget = components["schemas"]["DashboardWidget"];
type Chart = components["schemas"]["Chart"];

export default async function DashboardDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { token, currentTenantId } = await getAuthContext();

  const authHeaders = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": currentTenantId,
  };

  const [dashboardRes, widgetsRes, chartsRes] = await Promise.all([
    backendFetch(`/api/v1/dashboards/${id}`, {
      headers: authHeaders,
      cache: "no-store",
    }),
    backendFetch(`/api/v1/dashboards/${id}/widgets`, {
      headers: authHeaders,
      cache: "no-store",
    }),
    backendFetch("/api/v1/charts", {
      headers: authHeaders,
      cache: "no-store",
    }),
  ]);

  let dashboard: Dashboard | null = null;
  let errorMessage = "";
  if (dashboardRes.ok) {
    dashboard = await dashboardRes.json();
  } else {
    const err = (await dashboardRes.json()) as { error?: string };
    errorMessage =
      err.error ?? `failed to load dashboard (${dashboardRes.status})`;
  }

  let widgets: DashboardWidget[] = [];
  if (widgetsRes.ok) {
    const data: { items: DashboardWidget[] } = await widgetsRes.json();
    widgets = data.items ?? [];
  }

  let charts: Chart[] = [];
  if (chartsRes.ok) {
    const data: { items: Chart[] } = await chartsRes.json();
    charts = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">
          {dashboard?.name ?? "Dashboard Detail"}
        </h1>
        <Link
          href="/dashboards"
          className="text-sm underline-offset-2 hover:underline"
        >
          Back to list
        </Link>
      </div>

      {errorMessage ? (
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
          {errorMessage}
        </div>
      ) : null}

      {dashboard ? (
        <DashboardDetailManager
          dashboard={dashboard}
          initialWidgets={widgets}
          charts={charts}
        />
      ) : null}
    </div>
  );
}
