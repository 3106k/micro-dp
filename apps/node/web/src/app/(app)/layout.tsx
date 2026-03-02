import { getAuthContext } from "@/lib/auth/get-auth-context";
import { AppHeader } from "@/components/app-header";
import { TrackerProvider } from "@/components/tracker-provider";

export default async function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { me, currentTenantId } = await getAuthContext();

  return (
    <div className="min-h-screen">
      <AppHeader
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
