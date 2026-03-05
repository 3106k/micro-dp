import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { ChartsManager } from "./charts-manager";

type Chart = components["schemas"]["Chart"];
type Dataset = components["schemas"]["Dataset"];

export default async function ChartsPage() {
  const { token, currentTenantId } = await getAuthContext();

  const authHeaders = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": currentTenantId,
  };

  const [chartsRes, datasetsRes] = await Promise.all([
    backendFetch("/api/v1/charts", {
      headers: authHeaders,
      cache: "no-store",
    }),
    backendFetch("/api/v1/datasets", {
      headers: authHeaders,
      cache: "no-store",
    }),
  ]);

  let charts: Chart[] = [];
  if (chartsRes.ok) {
    const data: { items: Chart[] } = await chartsRes.json();
    charts = data.items ?? [];
  }

  let datasets: Dataset[] = [];
  if (datasetsRes.ok) {
    const data: { items: Dataset[] } = await datasetsRes.json();
    datasets = data.items ?? [];
  }

  return <ChartsManager initialCharts={charts} datasets={datasets} />;
}
