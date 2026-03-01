import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { DashboardHeader } from "@/app/dashboard/dashboard-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";
import { backendFetch } from "@/lib/api/server";
import { TENANT_COOKIE, TOKEN_COOKIE } from "@/lib/auth/constants";

type MeResponse = components["schemas"]["MeResponse"];
type Dataset = components["schemas"]["Dataset"];
type SourceType = components["schemas"]["DatasetSourceType"];

const sourceTypes: SourceType[] = ["tracker", "parquet", "import"];

function toPositiveInt(value: string | undefined, fallback: number): number {
  const parsed = Number.parseInt(value ?? "", 10);
  if (Number.isNaN(parsed) || parsed <= 0) {
    return fallback;
  }
  return parsed;
}

export default async function DatasetsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = await searchParams;

  const q = typeof params.q === "string" ? params.q : "";
  const sourceType =
    typeof params.source_type === "string" &&
    sourceTypes.includes(params.source_type as SourceType)
      ? (params.source_type as SourceType)
      : "";
  const page = toPositiveInt(
    typeof params.page === "string" ? params.page : undefined,
    1
  );
  const limit = Math.min(
    50,
    toPositiveInt(typeof params.limit === "string" ? params.limit : undefined, 10)
  );
  const offset = (page - 1) * limit;

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

  const query = new URLSearchParams();
  if (q) query.set("q", q);
  if (sourceType) query.set("source_type", sourceType);
  query.set("limit", String(limit));
  query.set("offset", String(offset));

  const datasetsRes = await backendFetch(`/api/v1/datasets?${query.toString()}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Tenant-ID": tenantId,
    },
    cache: "no-store",
  });

  let datasets: Dataset[] = [];
  let errorMessage = "";

  if (datasetsRes.ok) {
    const data: { items: Dataset[] } = await datasetsRes.json();
    datasets = data.items ?? [];
  } else {
    const err = (await datasetsRes.json()) as { error?: string };
    errorMessage = err.error ?? `failed to load datasets (${datasetsRes.status})`;
  }

  const hasPrev = page > 1;
  const hasNext = datasets.length === limit;

  function buildPageHref(targetPage: number): string {
    const p = new URLSearchParams();
    if (q) p.set("q", q);
    if (sourceType) p.set("source_type", sourceType);
    p.set("limit", String(limit));
    p.set("page", String(targetPage));
    return `/datasets?${p.toString()}`;
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
        <h1 className="text-2xl font-semibold tracking-tight">Datasets</h1>

        <form method="GET" className="rounded-lg border p-4">
          <div className="grid gap-3 md:grid-cols-4">
            <Input
              name="q"
              defaultValue={q}
              placeholder="Search by name"
              className="md:col-span-2"
            />
            <select
              name="source_type"
              defaultValue={sourceType}
              className="h-10 rounded-md border bg-background px-3 text-sm"
            >
              <option value="">All source types</option>
              {sourceTypes.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
            <Input name="limit" type="number" min={1} max={50} defaultValue={limit} />
          </div>
          <div className="mt-3">
            <Button type="submit">Apply</Button>
          </div>
        </form>

        {errorMessage ? (
          <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
            {errorMessage}
          </div>
        ) : null}

        {!errorMessage ? (
          <div className="rounded-lg border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left font-medium">Name</th>
                  <th className="px-4 py-3 text-left font-medium">Source</th>
                  <th className="px-4 py-3 text-left font-medium">Rows</th>
                  <th className="px-4 py-3 text-left font-medium">Last Updated</th>
                </tr>
              </thead>
              <tbody>
                {datasets.map((d) => (
                  <tr key={d.id} className="border-b last:border-0">
                    <td className="px-4 py-3">
                      <Link
                        href={`/datasets/${d.id}`}
                        className="font-medium underline-offset-2 hover:underline"
                      >
                        {d.name}
                      </Link>
                    </td>
                    <td className="px-4 py-3">{d.source_type}</td>
                    <td className="px-4 py-3">{d.row_count ?? "-"}</td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {d.last_updated_at
                        ? new Date(d.last_updated_at).toLocaleString()
                        : "-"}
                    </td>
                  </tr>
                ))}
                {datasets.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-4 py-8 text-center text-muted-foreground"
                    >
                      No datasets found.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        ) : null}

        <div className="flex items-center justify-between">
          {hasPrev ? (
            <Link href={buildPageHref(page - 1)}>
              <Button variant="outline">Previous</Button>
            </Link>
          ) : (
            <Button variant="outline" disabled>
              Previous
            </Button>
          )}
          <span className="text-sm text-muted-foreground">Page {page}</span>
          {hasNext ? (
            <Link href={buildPageHref(page + 1)}>
              <Button variant="outline">Next</Button>
            </Link>
          ) : (
            <Button variant="outline" disabled>
              Next
            </Button>
          )}
        </div>
      </main>
    </div>
  );
}
