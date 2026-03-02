"use client";

import { useCallback, useEffect, useState } from "react";

import type { components } from "@/lib/api/generated";

type DatasetRowsResponse = components["schemas"]["DatasetRowsResponse"];
type DatasetColumn = components["schemas"]["DatasetColumn"];

const PAGE_SIZE = 100;

export function DatasetRowsPreview({ datasetId }: { datasetId: string }) {
  const [data, setData] = useState<DatasetRowsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [offset, setOffset] = useState(0);

  const fetchRows = useCallback(
    async (currentOffset: number) => {
      setLoading(true);
      setError("");
      try {
        const res = await fetch(
          `/api/datasets/${datasetId}/rows?limit=${PAGE_SIZE}&offset=${currentOffset}`
        );
        if (!res.ok) {
          const err = await res.json().catch(() => ({}));
          setError(err.error ?? `Failed to load rows (${res.status})`);
          return;
        }
        const json: DatasetRowsResponse = await res.json();
        setData(json);
      } catch {
        setError("Network error");
      } finally {
        setLoading(false);
      }
    },
    [datasetId]
  );

  useEffect(() => {
    fetchRows(offset);
  }, [fetchRows, offset]);

  const totalRows = data?.total_rows ?? 0;
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(totalRows / PAGE_SIZE));

  return (
    <div className="rounded-lg border p-4">
      <h2 className="mb-3 text-lg font-semibold">Data Preview</h2>

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading...</p>
      ) : error ? (
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      ) : data && data.rows.length > 0 ? (
        <>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  {data.columns.map((col: DatasetColumn) => (
                    <th
                      key={col.name}
                      className="whitespace-nowrap px-3 py-2 text-left font-medium"
                    >
                      <div>{col.name}</div>
                      <div className="text-xs font-normal text-muted-foreground">
                        {col.type}
                      </div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {data.rows.map((row, rowIdx) => (
                  <tr key={rowIdx} className="border-b last:border-b-0">
                    {data.columns.map((col: DatasetColumn) => (
                      <td
                        key={col.name}
                        className="whitespace-nowrap px-3 py-1.5 font-mono text-xs"
                      >
                        {row[col.name] != null ? String(row[col.name]) : ""}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-3 flex items-center justify-between text-sm text-muted-foreground">
            <span>
              {totalRows.toLocaleString()} rows total
            </span>
            <div className="flex items-center gap-2">
              <button
                className="rounded border px-2 py-1 text-xs hover:bg-muted disabled:opacity-50"
                disabled={offset === 0}
                onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
              >
                Previous
              </button>
              <span className="text-xs">
                Page {currentPage} / {totalPages}
              </span>
              <button
                className="rounded border px-2 py-1 text-xs hover:bg-muted disabled:opacity-50"
                disabled={offset + PAGE_SIZE >= totalRows}
                onClick={() => setOffset(offset + PAGE_SIZE)}
              >
                Next
              </button>
            </div>
          </div>
        </>
      ) : (
        <p className="text-sm text-muted-foreground">No data available</p>
      )}
    </div>
  );
}
