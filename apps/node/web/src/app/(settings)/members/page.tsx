import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { MembersManager } from "./members-manager";

type TenantMember = components["schemas"]["TenantMember"];

export default async function MembersPage() {
  const { me, token, currentTenantId } = await getAuthContext();

  let initialMembers: TenantMember[] = [];
  const membersRes = await backendFetch("/api/v1/tenants/current/members", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
  });
  if (membersRes.ok) {
    const data: { items: TenantMember[] } = await membersRes.json();
    initialMembers = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Members</h1>
      <MembersManager
        initialMembers={initialMembers}
        currentUserId={me.user_id}
      />
    </div>
  );
}
