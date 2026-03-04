import { cookies } from "next/headers";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { ConnectionsManager } from "./connections-manager";

type Connection = components["schemas"]["Connection"];
type ConnectorDefinition = components["schemas"]["ConnectorDefinition"];
type Credential = components["schemas"]["Credential"];

export default async function ConnectionsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value!;
  const tenantId = jar.get(TENANT_COOKIE)?.value!;

  const authHeaders = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": tenantId,
  };

  const [connectionsRes, connectorsRes, credentialsRes] = await Promise.all([
    backendFetch("/api/v1/connections", { headers: authHeaders }),
    backendFetch("/api/v1/connectors", { headers: authHeaders }),
    backendFetch("/api/v1/credentials", { headers: authHeaders }),
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

  let initialCredentials: Credential[] = [];
  if (credentialsRes.ok) {
    const data: { items: Credential[] } = await credentialsRes.json();
    initialCredentials = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Connections</h1>
      <ConnectionsManager
        initialConnections={initialConnections}
        initialConnectors={initialConnectors}
        initialCredentials={initialCredentials}
      />
    </div>
  );
}
