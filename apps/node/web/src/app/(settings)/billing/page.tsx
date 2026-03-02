import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import type { components } from "@/lib/api/generated";
import { BillingActions } from "./billing-actions";

type BillingSubscriptionResponse =
  components["schemas"]["BillingSubscriptionResponse"];

export default async function BillingPage() {
  const { token, currentTenantId } = await getAuthContext();

  let subscription: BillingSubscriptionResponse | null = null;
  const res = await backendFetch("/api/v1/billing/subscription", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
  });
  if (res.ok) {
    subscription = await res.json();
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Billing</h1>
      <div className="grid gap-4 rounded-lg border p-4 md:grid-cols-2">
        <div>
          <p className="text-xs text-muted-foreground">Status</p>
          <p className="text-sm">{subscription?.status ?? "inactive"}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Current Plan</p>
          <p className="text-sm">
            {subscription?.plan?.display_name ?? "N/A"}
          </p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Current Period End</p>
          <p className="text-sm">
            {subscription?.current_period_end ?? "N/A"}
          </p>
        </div>
      </div>
      <BillingActions />
    </div>
  );
}
