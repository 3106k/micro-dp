import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { TransformJobForm } from "./transform-job-form";

type MeResponse = components["schemas"]["MeResponse"];
type Dataset = components["schemas"]["Dataset"];

export default async function NewTransformJobPage() {
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

  let datasets: Dataset[] = [];
  const datasetsRes = await backendFetch("/api/v1/datasets?limit=100", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });
  if (datasetsRes.ok) {
    const data: { items: Dataset[] } = await datasetsRes.json();
    datasets = data.items ?? [];
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
        <h1 className="text-2xl font-semibold tracking-tight">
          Create Transform Job
        </h1>
        <TransformJobForm datasets={datasets} />
      </main>
    </div>
  );
}
