import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { TrackingManager } from "./tracking-manager";

type WriteKey = components["schemas"]["WriteKey"];

export default async function TrackingPage() {
  const { token, currentTenantId } = await getAuthContext();

  let initialKeys: WriteKey[] = [];
  const res = await backendFetch("/api/v1/write-keys", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
  });
  if (res.ok) {
    const data: { items: WriteKey[] } = await res.json();
    initialKeys = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Tracking</h1>
      <p className="text-muted-foreground">
        Manage write keys for collecting events from external sites.
      </p>
      <TrackingManager initialKeys={initialKeys} />
    </div>
  );
}
