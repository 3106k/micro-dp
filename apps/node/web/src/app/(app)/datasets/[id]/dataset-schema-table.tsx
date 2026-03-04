"use client";

import { useCallback, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";

type DatasetColumn = components["schemas"]["DatasetColumn"];
type SemanticType = components["schemas"]["DatasetColumnSemanticType"];

const SEMANTIC_TYPES: SemanticType[] = [
  "dimension",
  "measure",
  "timestamp",
  "identifier",
];

interface Props {
  datasetId: string;
  columns: DatasetColumn[];
}

interface EditState {
  description: string;
  semantic_type: SemanticType | "";
  tags: string;
}

export function DatasetSchemaTable({ datasetId, columns }: Props) {
  const [edits, setEdits] = useState<Record<string, EditState>>(() => {
    const init: Record<string, EditState> = {};
    for (const col of columns) {
      init[col.name] = {
        description: col.description ?? "",
        semantic_type: col.semantic_type ?? "",
        tags: col.tags?.join(", ") ?? "",
      };
    }
    return init;
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  const updateEdit = useCallback(
    (name: string, field: keyof EditState, value: string) => {
      setEdits((prev) => ({
        ...prev,
        [name]: { ...prev[name], [field]: value },
      }));
    },
    []
  );

  const handleSave = useCallback(async () => {
    setSaving(true);
    setMessage("");
    try {
      const payload = {
        columns: columns.map((col) => {
          const edit = edits[col.name];
          return {
            name: col.name,
            description: edit.description || undefined,
            semantic_type: edit.semantic_type || undefined,
            tags: edit.tags
              ? edit.tags
                  .split(",")
                  .map((t) => t.trim())
                  .filter(Boolean)
              : undefined,
          };
        }),
      };

      const res = await fetch(`/api/datasets/${datasetId}/columns`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        setMessage(err.error ?? `Save failed (${res.status})`);
        return;
      }
      setMessage("Saved successfully");
    } catch {
      setMessage("Network error");
    } finally {
      setSaving(false);
    }
  }, [datasetId, columns, edits]);

  return (
    <div className="rounded-lg border p-4">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-lg font-semibold">Schema</h2>
        <div className="flex items-center gap-2">
          {message && (
            <span
              className={`text-xs ${message.includes("failed") || message.includes("error") ? "text-destructive" : "text-muted-foreground"}`}
            >
              {message}
            </span>
          )}
          <Button size="sm" onClick={handleSave} disabled={saving}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Name
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Type
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Nullable
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Description
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Semantic Type
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Tags
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Sample Values
              </th>
              <th className="whitespace-nowrap px-3 py-2 text-left font-medium">
                Statistics
              </th>
            </tr>
          </thead>
          <tbody>
            {columns.map((col) => {
              const edit = edits[col.name];
              return (
                <tr key={col.name} className="border-b last:border-b-0">
                  <td className="whitespace-nowrap px-3 py-1.5 font-mono text-xs font-medium">
                    {col.name}
                  </td>
                  <td className="whitespace-nowrap px-3 py-1.5 font-mono text-xs">
                    {col.type}
                  </td>
                  <td className="whitespace-nowrap px-3 py-1.5 text-xs">
                    {col.nullable ? "Yes" : "No"}
                  </td>
                  <td className="px-3 py-1.5">
                    <Input
                      className="h-7 text-xs"
                      value={edit?.description ?? ""}
                      onChange={(e) =>
                        updateEdit(col.name, "description", e.target.value)
                      }
                      placeholder="Add description..."
                    />
                  </td>
                  <td className="px-3 py-1.5">
                    <select
                      className="h-7 rounded border bg-background px-2 text-xs"
                      value={edit?.semantic_type ?? ""}
                      onChange={(e) =>
                        updateEdit(col.name, "semantic_type", e.target.value)
                      }
                    >
                      <option value="">-</option>
                      {SEMANTIC_TYPES.map((st) => (
                        <option key={st} value={st}>
                          {st}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="px-3 py-1.5">
                    <Input
                      className="h-7 text-xs"
                      value={edit?.tags ?? ""}
                      onChange={(e) =>
                        updateEdit(col.name, "tags", e.target.value)
                      }
                      placeholder="tag1, tag2"
                    />
                  </td>
                  <td className="whitespace-nowrap px-3 py-1.5 font-mono text-xs text-muted-foreground">
                    {col.sample_values?.slice(0, 3).join(", ") ?? "-"}
                  </td>
                  <td className="whitespace-nowrap px-3 py-1.5 text-xs text-muted-foreground">
                    {col.statistics ? (
                      <span>
                        {col.statistics.min != null && (
                          <>min: {col.statistics.min} </>
                        )}
                        {col.statistics.max != null && (
                          <>max: {col.statistics.max} </>
                        )}
                        {col.statistics.distinct_count != null && (
                          <>distinct: {col.statistics.distinct_count} </>
                        )}
                        {col.statistics.null_rate != null && (
                          <>null: {(col.statistics.null_rate * 100).toFixed(1)}%</>
                        )}
                      </span>
                    ) : (
                      "-"
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
