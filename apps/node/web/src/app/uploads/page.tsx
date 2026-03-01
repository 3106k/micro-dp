import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { UploadsManager } from "./uploads-manager";

type MeResponse = components["schemas"]["MeResponse"];

export default async function UploadsPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;
  if (!token || !tenantId) {
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
    tenantId && tenantIds.has(tenantId) ? tenantId : me.tenants[0]?.id ?? "";

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
        <h1 className="text-2xl font-semibold tracking-tight">Uploads</h1>
        <UploadsManager />
      </main>
    </div>
  );
}
