import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { backendFetch } from "@/lib/api/server";
import { TOKEN_COOKIE } from "@/lib/auth/constants";
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
  const tenantId = me.tenants.length > 0 ? me.tenants[0].id : "";

  return (
    <div className="min-h-screen">
      <DashboardHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
      />
      <TrackerProvider tenantId={tenantId} userId={me.user_id}>
        <main className="container py-8">{children}</main>
      </TrackerProvider>
    </div>
  );
}
