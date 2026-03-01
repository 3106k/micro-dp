import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";
import { VersionsManager } from "./versions-manager";

type MeResponse = components["schemas"]["MeResponse"];
type JobVersion = components["schemas"]["JobVersion"];

export default async function JobVersionsPage({
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

  let versions: JobVersion[] = [];
  const versionsRes = await backendFetch(`/api/v1/jobs/${id}/versions`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });
  if (versionsRes.ok) {
    const data: { items: JobVersion[] } = await versionsRes.json();
    versions = data.items ?? [];
  }

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
      />
      <main className="container space-y-6 py-8">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold tracking-tight">Job Versions</h1>
          <Link href={`/jobs/${id}`} className="text-sm underline">
            Back to job
          </Link>
        </div>
        <VersionsManager jobId={id} initialVersions={versions} />
      </main>
    </div>
  );
}
