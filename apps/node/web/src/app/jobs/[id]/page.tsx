import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { JobDetailManager } from "./job-detail-manager";

type MeResponse = components["schemas"]["MeResponse"];
type Job = components["schemas"]["Job"];

export default async function JobDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
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

  const jobRes = await backendFetch(`/api/v1/jobs/${id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });

  let job: Job | null = null;
  let errorMessage = "";
  if (jobRes.ok) {
    job = await jobRes.json();
  } else {
    const err = (await jobRes.json()) as { error?: string };
    errorMessage = err.error ?? `failed to load job (${jobRes.status})`;
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
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold tracking-tight">Job</h1>
          <Link href="/jobs" className="text-sm underline">
            Back to jobs
          </Link>
        </div>
        {errorMessage ? (
          <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
            {errorMessage}
          </div>
        ) : null}
        {job ? <JobDetailManager initialJob={job} /> : null}
      </main>
    </div>
  );
}
