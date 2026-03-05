import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { DashboardsManager } from "./dashboards-manager";

type Dashboard = components["schemas"]["Dashboard"];

export default async function DashboardsPage() {
  const { token, currentTenantId } = await getAuthContext();

  let dashboards: Dashboard[] = [];
  const res = await backendFetch("/api/v1/dashboards", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });
  if (res.ok) {
    const data: { items: Dashboard[] } = await res.json();
    dashboards = data.items ?? [];
  }

  return <DashboardsManager initialDashboards={dashboards} />;
}
