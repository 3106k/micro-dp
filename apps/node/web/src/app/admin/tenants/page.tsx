import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TOKEN_COOKIE } from "@/lib/auth/constants";
import { TenantsManager } from "./tenants-manager";

type MeResponse = components["schemas"]["MeResponse"];
type Tenant = components["schemas"]["Tenant"];

export default async function AdminTenantsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
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

  if (me.platform_role !== "superadmin") {
    return (
      <div className="min-h-screen">
        <DashboardHeader
          displayName={me.display_name}
          email={me.email}
          platformRole={me.platform_role}
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

  let initialTenants: Tenant[] = [];
  const tenantsRes = await backendFetch("/api/v1/admin/tenants", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (tenantsRes.ok) {
    const data: { items: Tenant[] } = await tenantsRes.json();
    initialTenants = data.items ?? [];
  }

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
      />
      <main className="container space-y-6 py-8">
        <h1 className="text-2xl font-semibold tracking-tight">Admin Tenants</h1>
        <TenantsManager initialTenants={initialTenants} />
      </main>
    </div>
  );
}
