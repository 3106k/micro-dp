import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";
import { DashboardHeader } from "./dashboard-header";
import { TrackerProvider } from "@/components/tracker-provider";

type MeResponse = components["schemas"]["MeResponse"];

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantCookie = jar.get(TENANT_COOKIE)?.value;

  if (!token) {
    redirect("/signin");
  }

  const res = await backendFetch("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!res.ok) {
    redirect("/signin");
  }

  const me: MeResponse = await res.json();
  const tenantIds = new Set(me.tenants.map((tenant) => tenant.id));
  const currentTenantId =
    tenantCookie && tenantIds.has(tenantCookie)
      ? tenantCookie
      : me.tenants[0]?.id ?? "";

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
        tenants={me.tenants}
        currentTenantId={currentTenantId}
      />
      <TrackerProvider tenantId={currentTenantId} userId={me.user_id}>
        <main className="container py-8">{children}</main>
      </TrackerProvider>
    </div>
  );
}
