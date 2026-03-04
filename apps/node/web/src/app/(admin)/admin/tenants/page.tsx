import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { TenantsManager } from "./tenants-manager";

type Tenant = components["schemas"]["Tenant"];

export default async function AdminTenantsPage() {
  const { token } = await getAuthContext();

  let initialTenants: Tenant[] = [];
  const tenantsRes = await backendFetch("/api/v1/admin/tenants", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (tenantsRes.ok) {
    const data: { items: Tenant[] } = await tenantsRes.json();
    initialTenants = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Admin Tenants</h1>
      <TenantsManager initialTenants={initialTenants} />
    </div>
  );
}
