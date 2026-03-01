import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { MembersManager } from "./members-manager";

type MeResponse = components["schemas"]["MeResponse"];
type TenantMember = components["schemas"]["TenantMember"];

export default async function MembersPage() {
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

  let initialMembers: TenantMember[] = [];
  const membersRes = await backendFetch("/api/v1/tenants/current/members", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
  });
  if (membersRes.ok) {
    const data: { items: TenantMember[] } = await membersRes.json();
    initialMembers = data.items ?? [];
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
        <h1 className="text-2xl font-semibold tracking-tight">Members</h1>
        <MembersManager
          initialMembers={initialMembers}
          currentUserId={me.user_id}
        />
      </main>
    </div>
  );
}
