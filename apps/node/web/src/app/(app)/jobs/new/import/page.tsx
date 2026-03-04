import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { ImportJobForm } from "./import-job-form";

type Connection = components["schemas"]["Connection"];

export default async function NewImportJobPage() {
  const { token, currentTenantId } = await getAuthContext();

  let connections: Connection[] = [];
  const connRes = await backendFetch("/api/v1/connections", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });
  if (connRes.ok) {
    const data: { items: Connection[] } = await connRes.json();
    connections = (data.items ?? []).filter((c) =>
      c.type.startsWith("source-")
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">
        Create Import Job
      </h1>
      <ImportJobForm connections={connections} />
    </div>
  );
}
