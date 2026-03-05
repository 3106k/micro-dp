"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { ChartPreview } from "@/components/chart-preview";
import { useToast } from "@/components/ui/toast-provider";
import { readApiErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";

type Chart = components["schemas"]["Chart"];
type Dataset = components["schemas"]["Dataset"];
type DatasetColumn = components["schemas"]["DatasetColumn"];
type ChartType = components["schemas"]["ChartType"];
type ChartPeriod = components["schemas"]["ChartPeriod"];
type ChartDataResponse = components["schemas"]["ChartDataResponse"];

const CHART_TYPES: { value: ChartType; label: string }[] = [
  { value: "line", label: "Line" },
  { value: "bar", label: "Bar" },
  { value: "pie", label: "Pie" },
];

const PERIODS: { value: ChartPeriod; label: string }[] = [
  { value: "last_7_days", label: "Last 7 days" },
  { value: "last_30_days", label: "Last 30 days" },
  { value: "last_90_days", label: "Last 90 days" },
  { value: "custom", label: "Custom" },
];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function ChartDetailManager({
  chart: initial,
  datasets,
}: {
  chart: Chart;
  datasets: Dataset[];
}) {
  const router = useRouter();
  const { pushToast } = useToast();

  // Edit form state
  const [name, setName] = useState(initial.name);
  const [chartType, setChartType] = useState<ChartType>(initial.chart_type);
  const [datasetId, setDatasetId] = useState(initial.dataset_id);
  const [measure, setMeasure] = useState(initial.measure);
  const [dimension, setDimension] = useState(initial.dimension);
  const [columns, setColumns] = useState<DatasetColumn[]>([]);
  const [loadingColumns, setLoadingColumns] = useState(false);
  const [saving, setSaving] = useState(false);

  // Preview state
  const [period, setPeriod] = useState<ChartPeriod>("last_30_days");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [chartData, setChartData] = useState<ChartDataResponse | null>(null);
  const [loadingData, setLoadingData] = useState(false);

  // Load columns for the initial dataset
  useEffect(() => {
    if (initial.dataset_id) {
      loadColumns(initial.dataset_id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function loadColumns(dsId: string) {
    setLoadingColumns(true);
    try {
      const res = await fetch(`/api/datasets/${dsId}`, { cache: "no-store" });
      if (res.ok) {
        const data = await res.json();
        setColumns(data.columns ?? []);
      }
    } finally {
      setLoadingColumns(false);
    }
  }

  async function handleDatasetChange(nextDatasetId: string) {
    setDatasetId(nextDatasetId);
    setMeasure("");
    setDimension("");
    setColumns([]);

    if (nextDatasetId) {
      await loadColumns(nextDatasetId);
    }
  }

  const fetchChartData = useCallback(async () => {
    setLoadingData(true);
    try {
      const params = new URLSearchParams();
      params.set("period", period);
      if (period === "custom") {
        if (startDate) params.set("start_date", startDate);
        if (endDate) params.set("end_date", endDate);
      }

      const res = await fetch(
        `/api/charts/${initial.id}/data?${params.toString()}`,
        { cache: "no-store" }
      );
      if (!res.ok) {
        const msg = await readApiErrorMessage(
          res,
          "Failed to fetch chart data"
        );
        throw new Error(msg);
      }
      const data: ChartDataResponse = await res.json();
      setChartData(data);
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to fetch chart data",
      });
    } finally {
      setLoadingData(false);
    }
  }, [initial.id, period, startDate, endDate, pushToast]);

  // Fetch chart data on mount and when period changes
  useEffect(() => {
    fetchChartData();
  }, [fetchChartData]);

  async function handleUpdate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || !datasetId || !measure || !dimension) return;

    setSaving(true);
    try {
      const res = await fetch(`/api/charts/${initial.id}`, {
        method: "PUT",
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
        const msg = await readApiErrorMessage(res, "Failed to update chart");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Chart updated" });
      router.refresh();
      // Refresh preview with new settings
      fetchChartData();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to update chart",
      });
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      {/* Chart info */}
      <div className="grid gap-4 rounded-lg border p-4 md:grid-cols-2">
        <div>
          <p className="text-xs text-muted-foreground">ID</p>
          <p className="font-mono text-sm">{initial.id}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Created</p>
          <p className="text-sm">{formatDate(initial.created_at)}</p>
        </div>
      </div>

      {/* Preview */}
      <div className="rounded-lg border p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Preview</h2>
          <div className="flex items-center gap-2">
            <select
              className="h-9 rounded-md border bg-background px-2 text-sm"
              value={period}
              onChange={(e) => setPeriod(e.target.value as ChartPeriod)}
            >
              {PERIODS.map((p) => (
                <option key={p.value} value={p.value}>
                  {p.label}
                </option>
              ))}
            </select>
            {period === "custom" ? (
              <>
                <input
                  type="date"
                  className="h-9 rounded-md border bg-background px-2 text-sm"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                />
                <input
                  type="date"
                  className="h-9 rounded-md border bg-background px-2 text-sm"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                />
              </>
            ) : null}
          </div>
        </div>
        <ChartPreview
          chartType={initial.chart_type}
          labels={chartData?.labels ?? []}
          datasets={chartData?.datasets ?? []}
          loading={loadingData}
          height={350}
        />
      </div>

      {/* Edit form */}
      <form
        onSubmit={handleUpdate}
        className="rounded-lg border p-4 space-y-4"
      >
        <h2 className="text-lg font-semibold">Edit Chart</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <label htmlFor="edit-name" className="text-sm font-medium">
              Name <span className="text-destructive">*</span>
            </label>
            <input
              id="edit-name"
              type="text"
              className="flex h-9 w-full rounded-md border bg-background px-3 py-1 text-sm"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="edit-type" className="text-sm font-medium">
              Chart Type <span className="text-destructive">*</span>
            </label>
            <select
              id="edit-type"
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
            <label htmlFor="edit-dataset" className="text-sm font-medium">
              Dataset <span className="text-destructive">*</span>
            </label>
            <select
              id="edit-dataset"
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
            <label htmlFor="edit-measure" className="text-sm font-medium">
              Measure <span className="text-destructive">*</span>
            </label>
            <select
              id="edit-measure"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={measure}
              onChange={(e) => setMeasure(e.target.value)}
              disabled={!datasetId || loadingColumns}
            >
              <option value="">
                {loadingColumns ? "Loading..." : "Select measure..."}
              </option>
              {columns.map((c) => (
                <option key={c.name} value={c.name}>
                  {c.name} ({c.type})
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-1">
            <label htmlFor="edit-dimension" className="text-sm font-medium">
              Dimension <span className="text-destructive">*</span>
            </label>
            <select
              id="edit-dimension"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={dimension}
              onChange={(e) => setDimension(e.target.value)}
              disabled={!datasetId || loadingColumns}
            >
              <option value="">
                {loadingColumns ? "Loading..." : "Select dimension..."}
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
            saving || !name.trim() || !datasetId || !measure || !dimension
          }
        >
          {saving ? "Saving..." : "Save"}
        </Button>
      </form>
    </div>
  );
}
