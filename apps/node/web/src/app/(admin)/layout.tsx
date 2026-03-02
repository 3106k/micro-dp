import { redirect } from "next/navigation";

import { getAuthContext } from "@/lib/auth/get-auth-context";
import { AppHeader } from "@/components/app-header";
import { AdminNav } from "@/components/admin-nav";

export default async function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { me, currentTenantId } = await getAuthContext();

  if (me.platform_role !== "superadmin") {
    redirect("/dashboard");
  }

  return (
    <div className="min-h-screen">
      <AppHeader
        displayName={me.display_name}
        email={me.email}
        platformRole={me.platform_role}
        tenants={me.tenants}
        currentTenantId={currentTenantId}
        sectionLabel="Admin"
      />
      <AdminNav />
      <main className="container py-8">{children}</main>
    </div>
  );
}
