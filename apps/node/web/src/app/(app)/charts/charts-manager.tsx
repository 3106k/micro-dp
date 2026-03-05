"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { readApiErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";

type Chart = components["schemas"]["Chart"];
type Dataset = components["schemas"]["Dataset"];
type DatasetColumn = components["schemas"]["DatasetColumn"];
type ChartType = components["schemas"]["ChartType"];

const CHART_TYPES: { value: ChartType; label: string }[] = [
  { value: "line", label: "Line" },
  { value: "bar", label: "Bar" },
  { value: "pie", label: "Pie" },
];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function ChartsManager({
  initialCharts,
  datasets,
}: {
  initialCharts: Chart[];
  datasets: Dataset[];
}) {
  const router = useRouter();
  const { pushToast } = useToast();
  const [charts, setCharts] = useState(initialCharts);

  // Create form state
  const [name, setName] = useState("");
  const [chartType, setChartType] = useState<ChartType>("line");
  const [datasetId, setDatasetId] = useState("");
  const [measure, setMeasure] = useState("");
  const [dimension, setDimension] = useState("");
  const [columns, setColumns] = useState<DatasetColumn[]>([]);
  const [loadingColumns, setLoadingColumns] = useState(false);
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  async function refreshList() {
    try {
      const res = await fetch("/api/charts", { cache: "no-store" });
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to fetch charts");
        throw new Error(msg);
      }
      const data = await res.json();
      setCharts(data.items ?? []);
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to fetch charts",
      });
    }
  }

  async function handleDatasetChange(nextDatasetId: string) {
    setDatasetId(nextDatasetId);
    setMeasure("");
    setDimension("");
    setColumns([]);

    if (!nextDatasetId) return;

    setLoadingColumns(true);
    try {
      const res = await fetch(`/api/datasets/${nextDatasetId}`, {
        cache: "no-store",
      });
      if (!res.ok) {
        const msg = await readApiErrorMessage(
          res,
          "Failed to fetch dataset details"
        );
        throw new Error(msg);
      }
      const data = await res.json();
      setColumns(data.columns ?? []);
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to fetch dataset columns",
      });
    } finally {
      setLoadingColumns(false);
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || !datasetId || !measure || !dimension) return;

    setCreating(true);
    try {
      const res = await fetch("/api/charts", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          chart_type: chartType,
          dataset_id: datasetId,
          measure,
          dimension,
        }),
      });
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to create chart");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Chart created" });
      setName("");
      setChartType("line");
      setDatasetId("");
      setMeasure("");
      setDimension("");
      setColumns([]);
      await refreshList();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to create chart",
      });
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    setDeleting(id);
    try {
      const res = await fetch(`/api/charts/${id}`, { method: "DELETE" });
      if (!res.ok && res.status !== 204) {
        const msg = await readApiErrorMessage(res, "Failed to delete chart");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Chart deleted" });
      await refreshList();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to delete chart",
      });
    } finally {
      setDeleting(null);
    }
  }

  const datasetMap = new Map(datasets.map((d) => [d.id, d]));

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Charts</h1>

      {/* Create form */}
      <form
        onSubmit={handleCreate}
        className="rounded-lg border p-4 space-y-4"
      >
        <h2 className="text-lg font-semibold">Create Chart</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <label htmlFor="chart-name" className="text-sm font-medium">
              Name <span className="text-destructive">*</span>
            </label>
            <input
              id="chart-name"
              type="text"
              className="flex h-9 w-full rounded-md border bg-background px-3 py-1 text-sm"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Chart"
              required
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="chart-type" className="text-sm font-medium">
              Chart Type <span className="text-destructive">*</span>
            </label>
            <select
              id="chart-type"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={chartType}
              onChange={(e) => setChartType(e.target.value as ChartType)}
            >
              {CHART_TYPES.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-1">
            <label htmlFor="chart-dataset" className="text-sm font-medium">
              Dataset <span className="text-destructive">*</span>
            </label>
            <select
              id="chart-dataset"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={datasetId}
              onChange={(e) => handleDatasetChange(e.target.value)}
            >
              <option value="">Select a dataset...</option>
              {datasets.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-1">
            <label htmlFor="chart-measure" className="text-sm font-medium">
              Measure <span className="text-destructive">*</span>
            </label>
            <select
              id="chart-measure"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={measure}
              onChange={(e) => setMeasure(e.target.value)}
              disabled={!datasetId || loadingColumns}
            >
              <option value="">
                {loadingColumns ? "Loading columns..." : "Select measure..."}
              </option>
              {columns.map((c) => (
                <option key={c.name} value={c.name}>
                  {c.name} ({c.type})
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-1">
            <label htmlFor="chart-dimension" className="text-sm font-medium">
              Dimension <span className="text-destructive">*</span>
            </label>
            <select
              id="chart-dimension"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={dimension}
              onChange={(e) => setDimension(e.target.value)}
              disabled={!datasetId || loadingColumns}
            >
              <option value="">
                {loadingColumns ? "Loading columns..." : "Select dimension..."}
              </option>
              {columns.map((c) => (
                <option key={c.name} value={c.name}>
                  {c.name} ({c.type})
                </option>
              ))}
            </select>
          </div>
        </div>
        <Button
          type="submit"
          disabled={
            creating || !name.trim() || !datasetId || !measure || !dimension
          }
        >
          {creating ? "Creating..." : "Create"}
        </Button>
      </form>

      {/* Table */}
      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">Type</th>
              <th className="px-4 py-3 text-left font-medium">Dataset</th>
              <th className="px-4 py-3 text-left font-medium">
                Measure / Dimension
              </th>
              <th className="px-4 py-3 text-left font-medium">Created</th>
              <th className="px-4 py-3 text-right font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {charts.map((c) => {
              const ds = datasetMap.get(c.dataset_id);
              return (
                <tr
                  key={c.id}
                  className="border-b last:border-0 cursor-pointer hover:bg-muted/50"
                  onClick={() => router.push(`/charts/${c.id}`)}
                >
                  <td className="px-4 py-3 font-medium">{c.name}</td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {c.chart_type}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {ds?.name ?? c.dataset_id}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {c.measure} / {c.dimension}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {formatDate(c.created_at)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Button
                      variant="destructive"
                      size="sm"
                      disabled={deleting === c.id}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(c.id);
                      }}
                    >
                      {deleting === c.id ? "Deleting..." : "Delete"}
                    </Button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>

        {charts.length === 0 ? (
          <div className="px-4 py-8 text-center text-sm text-muted-foreground">
            No charts yet. Create one to get started.
          </div>
        ) : null}
      </div>
    </div>
  );
}
