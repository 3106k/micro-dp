import Link from "next/link";

import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { getAuthContext } from "@/lib/auth/get-auth-context";
import { ChartDetailManager } from "./chart-detail-manager";

type Chart = components["schemas"]["Chart"];
type Dataset = components["schemas"]["Dataset"];

export default async function ChartDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { token, currentTenantId } = await getAuthContext();

  const authHeaders = {
    Authorization: `Bearer ${token}`,
    "X-Tenant-ID": currentTenantId,
  };

  const [chartRes, datasetsRes] = await Promise.all([
    backendFetch(`/api/v1/charts/${id}`, {
      headers: authHeaders,
      cache: "no-store",
    }),
    backendFetch("/api/v1/datasets", {
      headers: authHeaders,
      cache: "no-store",
    }),
  ]);

  let chart: Chart | null = null;
  let errorMessage = "";
  if (chartRes.ok) {
    chart = await chartRes.json();
  } else {
    const err = (await chartRes.json()) as { error?: string };
    errorMessage = err.error ?? `failed to load chart (${chartRes.status})`;
  }

  let datasets: Dataset[] = [];
  if (datasetsRes.ok) {
    const data: { items: Dataset[] } = await datasetsRes.json();
    datasets = data.items ?? [];
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">
          {chart?.name ?? "Chart Detail"}
        </h1>
        <Link
          href="/charts"
          className="text-sm underline-offset-2 hover:underline"
        >
          Back to list
        </Link>
      </div>

      {errorMessage ? (
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
          {errorMessage}
        </div>
      ) : null}

      {chart ? (
        <ChartDetailManager chart={chart} datasets={datasets} />
      ) : null}
    </div>
  );
}
