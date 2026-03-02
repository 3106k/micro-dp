import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { TransformJobForm } from "./transform-job-form";

type Dataset = components["schemas"]["Dataset"];

export default async function NewTransformJobPage() {
  const { token, currentTenantId } = await getAuthContext();

  let datasets: Dataset[] = [];
  const datasetsRes = await backendFetch("/api/v1/datasets?limit=100", {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": currentTenantId,
    },
    cache: "no-store",
  });
  if (datasetsRes.ok) {
    const data: { items: Dataset[] } = await datasetsRes.json();
    datasets = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">
        Create Transform Job
      </h1>
      <TransformJobForm datasets={datasets} />
    </div>
  );
}
