import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { ConnectionsManager } from "./connections-manager";

type MeResponse = components["schemas"]["MeResponse"];
type Connection = components["schemas"]["Connection"];
type ConnectorDefinition = components["schemas"]["ConnectorDefinition"];

export default async function ConnectionsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;
  if (!token || !tenantId) {
    redirect("/signin");
  }

  const authHeaders = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": tenantId,
  };

  const meRes = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!meRes.ok) {
    redirect("/signin");
  }
  const me: MeResponse = await meRes.json();

  // Fetch connections and connectors in parallel
  const [connectionsRes, connectorsRes] = await Promise.all([
    backendFetch("/api/v1/connections", { headers: authHeaders }),
    backendFetch("/api/v1/connectors", { headers: authHeaders }),
  ]);

  let initialConnections: Connection[] = [];
  if (connectionsRes.ok) {
    const data: { items: Connection[] } = await connectionsRes.json();
    initialConnections = data.items ?? [];
  }

  let initialConnectors: ConnectorDefinition[] = [];
  if (connectorsRes.ok) {
    const data: { items: ConnectorDefinition[] } = await connectorsRes.json();
    initialConnectors = data.items ?? [];
  }

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
        tenants={me.tenants}
        currentTenantId={tenantId}
      />
      <main className="container space-y-6 py-8">
        <h1 className="text-2xl font-semibold tracking-tight">Connections</h1>
        <ConnectionsManager
          initialConnections={initialConnections}
          initialConnectors={initialConnectors}
        />
      </main>
    </div>
  );
}
