"use client";

import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

import type { components } from "@/lib/api/generated";

type ChartType = components["schemas"]["ChartType"];
type ChartDataset = components["schemas"]["ChartDataset"];

const COLORS = [
  "#2563eb",
  "#dc2626",
  "#16a34a",
  "#ca8a04",
  "#9333ea",
  "#0891b2",
  "#e11d48",
  "#65a30d",
];

export function ChartPreview({
  chartType,
  labels,
  datasets,
  loading,
  height = 300,
}: {
  chartType: ChartType;
  labels: string[];
  datasets: ChartDataset[];
  loading?: boolean;
  height?: number;
}) {
  if (loading) {
    return (
      <div
        className="flex items-center justify-center rounded-lg border bg-muted/20"
        style={{ height }}
      >
        <span className="text-sm text-muted-foreground">Loading...</span>
      </div>
    );
  }

  if (labels.length === 0 || datasets.length === 0) {
    return (
      <div
        className="flex items-center justify-center rounded-lg border bg-muted/20"
        style={{ height }}
      >
        <span className="text-sm text-muted-foreground">No data available</span>
      </div>
    );
  }

  if (chartType === "pie") {
    const pieData = labels.map((label, i) => ({
      name: label,
      value: datasets[0]?.data[i] ?? 0,
    }));

    return (
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={pieData}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            outerRadius={height / 3}
            label={({ name, percent }) =>
              `${name} (${((percent ?? 0) * 100).toFixed(0)}%)`
            }
          >
            {pieData.map((_, i) => (
              <Cell key={i} fill={COLORS[i % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    );
  }

  // Build data array for line/bar charts
  const data = labels.map((label, i) => {
    const point: Record<string, string | number> = { label };
    for (const ds of datasets) {
      point[ds.label] = ds.data[i] ?? 0;
    }
    return point;
  });

  if (chartType === "bar") {
    return (
      <ResponsiveContainer width="100%" height={height}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="label" tick={{ fontSize: 12 }} />
          <YAxis tick={{ fontSize: 12 }} />
          <Tooltip />
          <Legend />
          {datasets.map((ds, i) => (
            <Bar
              key={ds.label}
              dataKey={ds.label}
              fill={COLORS[i % COLORS.length]}
            />
          ))}
        </BarChart>
      </ResponsiveContainer>
    );
  }

  // Default: line
  return (
    <ResponsiveContainer width="100%" height={height}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="label" tick={{ fontSize: 12 }} />
        <YAxis tick={{ fontSize: 12 }} />
        <Tooltip />
        <Legend />
        {datasets.map((ds, i) => (
          <Line
            key={ds.label}
            type="monotone"
            dataKey={ds.label}
            stroke={COLORS[i % COLORS.length]}
            strokeWidth={2}
            dot={{ r: 3 }}
          />
        ))}
      </LineChart>
    </ResponsiveContainer>
  );
}
