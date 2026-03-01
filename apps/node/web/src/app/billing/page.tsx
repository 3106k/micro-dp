import { cookies } from "next/headers";

import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import type { components } from "@/lib/api/generated";
import { BillingActions } from "./billing-actions";

type BillingSubscriptionResponse =
  components["schemas"]["BillingSubscriptionResponse"];

export default async function BillingPage() {
  const jar = await cookies();
  const token = jar.get(TOKEN_COOKIE)?.value;
  const tenantId = jar.get(TENANT_COOKIE)?.value;

  let subscription: BillingSubscriptionResponse | null = null;
  if (token && tenantId) {
    const res = await backendFetch("/api/v1/billing/subscription", {
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Tenant-ID": tenantId,
      },
    });
    if (res.ok) {
      subscription = await res.json();
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Billing</h1>
      <div className="rounded-lg border bg-card p-4 text-sm space-y-2">
        <p>
          <span className="font-medium">Status:</span>{" "}
          {subscription?.status ?? "inactive"}
        </p>
        <p>
          <span className="font-medium">Current plan:</span>{" "}
          {subscription?.plan?.display_name ?? "N/A"}
        </p>
        <p>
          <span className="font-medium">Current period end:</span>{" "}
          {subscription?.current_period_end ?? "N/A"}
        </p>
      </div>
      <BillingActions />
    </div>
  );
}
