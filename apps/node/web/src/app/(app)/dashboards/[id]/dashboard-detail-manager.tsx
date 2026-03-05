"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { ChartPreview } from "@/components/chart-preview";
import { useToast } from "@/components/ui/toast-provider";
import { readApiErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";

type Dashboard = components["schemas"]["Dashboard"];
type DashboardWidget = components["schemas"]["DashboardWidget"];
type Chart = components["schemas"]["Chart"];
type ChartDataResponse = components["schemas"]["ChartDataResponse"];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function DashboardDetailManager({
  dashboard: initial,
  initialWidgets,
  charts,
}: {
  dashboard: Dashboard;
  initialWidgets: DashboardWidget[];
  charts: Chart[];
}) {
  const router = useRouter();
  const { pushToast } = useToast();

  // Edit form state
  const [name, setName] = useState(initial.name);
  const [description, setDescription] = useState(initial.description ?? "");
  const [saving, setSaving] = useState(false);

  // Widget state
  const [widgets, setWidgets] = useState(initialWidgets);
  const [selectedChartId, setSelectedChartId] = useState("");
  const [addingWidget, setAddingWidget] = useState(false);
  const [deletingWidget, setDeletingWidget] = useState<string | null>(null);

  async function handleUpdate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;

    setSaving(true);
    try {
      const res = await fetch(`/api/dashboards/${initial.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          description: description.trim() || undefined,
        }),
      });
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to update dashboard");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Dashboard updated" });
      router.refresh();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to update dashboard",
      });
    } finally {
      setSaving(false);
    }
  }

  async function refreshWidgets() {
    try {
      const res = await fetch(
        `/api/dashboards/${initial.id}/widgets`,
        { cache: "no-store" }
      );
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to fetch widgets");
        throw new Error(msg);
      }
      const data = await res.json();
      setWidgets(data.items ?? []);
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to fetch widgets",
      });
    }
  }

  async function handleAddWidget(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedChartId) return;

    setAddingWidget(true);
    try {
      const nextPosition = widgets.length > 0
        ? Math.max(...widgets.map((w) => w.position)) + 1
        : 0;

      const res = await fetch(
        `/api/dashboards/${initial.id}/widgets`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            chart_id: selectedChartId,
            position: nextPosition,
          }),
        }
      );
      if (!res.ok) {
        const msg = await readApiErrorMessage(res, "Failed to add widget");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Widget added" });
      setSelectedChartId("");
      await refreshWidgets();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to add widget",
      });
    } finally {
      setAddingWidget(false);
    }
  }

  async function handleDeleteWidget(widgetId: string) {
    setDeletingWidget(widgetId);
    try {
      const res = await fetch(
        `/api/dashboards/${initial.id}/widgets/${widgetId}`,
        { method: "DELETE" }
      );
      if (!res.ok && res.status !== 204) {
        const msg = await readApiErrorMessage(res, "Failed to delete widget");
        throw new Error(msg);
      }
      pushToast({ variant: "success", message: "Widget removed" });
      await refreshWidgets();
    } catch (error) {
      pushToast({
        variant: "error",
        message:
          error instanceof Error ? error.message : "Failed to delete widget",
      });
    } finally {
      setDeletingWidget(null);
    }
  }

  // Map chart_id -> chart for display
  const chartMap = new Map(charts.map((c) => [c.id, c]));

  // Sort widgets by position for display
  const sortedWidgets = [...widgets].sort((a, b) => a.position - b.position);

  return (
    <div className="space-y-6">
      {/* Dashboard info */}
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

      {/* Edit form */}
      <form onSubmit={handleUpdate} className="rounded-lg border p-4 space-y-4">
        <h2 className="text-lg font-semibold">Edit Dashboard</h2>
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
            <label htmlFor="edit-desc" className="text-sm font-medium">
              Description
            </label>
            <input
              id="edit-desc"
              type="text"
              className="flex h-9 w-full rounded-md border bg-background px-3 py-1 text-sm"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <Button type="submit" disabled={saving || !name.trim()}>
          {saving ? "Saving..." : "Save"}
        </Button>
      </form>

      {/* Widgets section */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold">Widgets</h2>

        {/* Add widget form */}
        <form
          onSubmit={handleAddWidget}
          className="flex items-end gap-3 rounded-lg border p-4"
        >
          <div className="flex-1 space-y-1">
            <label htmlFor="widget-chart" className="text-sm font-medium">
              Chart
            </label>
            <select
              id="widget-chart"
              className="flex h-9 w-full rounded-md border bg-background px-3 text-sm"
              value={selectedChartId}
              onChange={(e) => setSelectedChartId(e.target.value)}
            >
              <option value="">Select a chart...</option>
              {charts.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} ({c.chart_type})
                </option>
              ))}
            </select>
          </div>
          <Button
            type="submit"
            disabled={addingWidget || !selectedChartId}
          >
            {addingWidget ? "Adding..." : "Add Widget"}
          </Button>
        </form>

        {/* Widget grid with previews */}
        {sortedWidgets.length === 0 ? (
          <div className="rounded-lg border px-4 py-8 text-center text-sm text-muted-foreground">
            No widgets yet. Add a chart above.
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {sortedWidgets.map((w) => {
              const chart = chartMap.get(w.chart_id);
              return (
                <WidgetCard
                  key={w.id}
                  widget={w}
                  chart={chart}
                  deleting={deletingWidget === w.id}
                  onDelete={() => handleDeleteWidget(w.id)}
                />
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

function WidgetCard({
  widget,
  chart,
  deleting,
  onDelete,
}: {
  widget: DashboardWidget;
  chart: Chart | undefined;
  deleting: boolean;
  onDelete: () => void;
}) {
  const [chartData, setChartData] = useState<ChartDataResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function fetchData() {
      try {
        const res = await fetch(
          `/api/charts/${widget.chart_id}/data?period=last_30_days`,
          { cache: "no-store" }
        );
        if (res.ok && !cancelled) {
          const data: ChartDataResponse = await res.json();
          setChartData(data);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    fetchData();
    return () => {
      cancelled = true;
    };
  }, [widget.chart_id]);

  return (
    <div className="rounded-lg border p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div>
          <p className="font-medium">{chart?.name ?? widget.chart_id}</p>
          <p className="text-xs text-muted-foreground">
            {chart?.chart_type ?? "-"}
          </p>
        </div>
        <Button
          variant="destructive"
          size="sm"
          disabled={deleting}
          onClick={onDelete}
        >
          {deleting ? "Removing..." : "Remove"}
        </Button>
      </div>
      <ChartPreview
        chartType={chart?.chart_type ?? "line"}
        labels={chartData?.labels ?? []}
        datasets={chartData?.datasets ?? []}
        loading={loading}
        height={200}
      />
    </div>
  );
}
