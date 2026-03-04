import { cookies } from "next/headers";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { IntegrationsManager } from "./integrations-manager";

type Credential = components["schemas"]["Credential"];

export default async function IntegrationsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value!;
  const tenantId = jar.get(TENANT_COOKIE)?.value!;

  const authHeaders = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": tenantId,
  };

  const credentialsRes = await backendFetch("/api/v1/credentials", {
    headers: authHeaders,
  });

  let initialCredentials: Credential[] = [];
  if (credentialsRes.ok) {
    const data: { items: Credential[] } = await credentialsRes.json();
    initialCredentials = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Integrations</h1>
      <IntegrationsManager initialCredentials={initialCredentials} />
    </div>
  );
}
