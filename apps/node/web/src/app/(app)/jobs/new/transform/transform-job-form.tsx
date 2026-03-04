"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";

type Dataset = components["schemas"]["Dataset"];
type ValidateResponse = components["schemas"]["TransformValidateResponse"];
type PreviewResponse = components["schemas"]["TransformPreviewResponse"];
type CreateResponse = components["schemas"]["CreateTransformJobResponse"];

function toSlug(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

const sourceTypeBadge: Record<string, string> = {
  tracker: "bg-blue-100 text-blue-800",
  import: "bg-green-100 text-green-800",
  transform: "bg-purple-100 text-purple-800",
  parquet: "bg-gray-100 text-gray-800",
};

export function TransformJobForm({ datasets }: { datasets: Dataset[] }) {
  const router = useRouter();

  // Job info
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugEdited, setSlugEdited] = useState(false);
  const [description, setDescription] = useState("");

  // Dataset selection
  const [selectedDatasetIds, setSelectedDatasetIds] = useState<string[]>([]);

  // SQL
  const [sql, setSql] = useState("");

  // Validation
  const [validateResult, setValidateResult] =
    useState<ValidateResponse | null>(null);
  const [validating, setValidating] = useState(false);

  // Preview
  const [previewResult, setPreviewResult] = useState<PreviewResponse | null>(
    null
  );
  const [previewing, setPreviewing] = useState(false);

  // Execution
  const [execution, setExecution] = useState<
    "save_only" | "immediate" | "scheduled"
  >("save_only");

  // Submit
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  function handleNameChange(value: string) {
    setName(value);
    if (!slugEdited) {
      setSlug(toSlug(value));
    }
  }

  function toggleDataset(id: string) {
    setSelectedDatasetIds((prev) =>
      prev.includes(id) ? prev.filter((d) => d !== id) : [...prev, id]
    );
  }

  const selectedDatasetNames = datasets
    .filter((d) => selectedDatasetIds.includes(d.id))
    .map((d) => d.name);

  async function handleValidate() {
    setValidating(true);
    setValidateResult(null);
    setPreviewResult(null);
    setError("");
    try {
      const res = await fetch("/api/transform/validate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ sql, dataset_ids: selectedDatasetIds }),
      });
      const data: ValidateResponse | { error: string } = await res.json();
      if (!res.ok) {
        setError(
          "error" in data && data.error
            ? data.error
            : `validation failed (${res.status})`
        );
        return;
      }
      setValidateResult(data as ValidateResponse);
    } catch {
      setError("validation request failed");
    } finally {
      setValidating(false);
    }
  }

  async function handlePreview() {
    setPreviewing(true);
    setPreviewResult(null);
    setError("");
    try {
      const res = await fetch("/api/transform/preview", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          sql,
          dataset_ids: selectedDatasetIds,
          limit: 20,
        }),
      });
      const data: PreviewResponse | { error: string } = await res.json();
      if (!res.ok) {
        setError(
          "error" in data ? data.error : `preview failed (${res.status})`
        );
        return;
      }
      setPreviewResult(data as PreviewResponse);
    } catch {
      setError("preview request failed");
    } finally {
      setPreviewing(false);
    }
  }

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    setMessage("");
    setError("");
    try {
      const res = await fetch("/api/transform/jobs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          slug,
          description: description || undefined,
          sql,
          dataset_ids: selectedDatasetIds,
          execution,
        }),
      });
      const data: CreateResponse | { error: string } = await res.json();
      if (!res.ok) {
        setError(
          "error" in data ? data.error : `creation failed (${res.status})`
        );
        return;
      }
      setMessage("Transform job created successfully!");
      setTimeout(() => router.push("/jobs"), 1500);
    } catch {
      setError("creation request failed");
    } finally {
      setSubmitting(false);
    }
  }

  const canValidate = sql.trim() !== "" && selectedDatasetIds.length > 0;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid gap-6 lg:grid-cols-[1fr_2fr]">
        {/* ── Left Column: Settings ── */}
        <div className="space-y-6 lg:sticky lg:top-8 lg:self-start">
          {/* Job Info */}
          <div className="rounded-lg border p-4 space-y-4">
            <h2 className="text-lg font-semibold">Job Information</h2>
            <div className="grid gap-3">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <Input
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  placeholder="My Transform Job"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Slug</label>
                <Input
                  value={slug}
                  onChange={(e) => {
                    setSlug(e.target.value);
                    setSlugEdited(true);
                  }}
                  placeholder="my-transform-job"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  Description
                </label>
                <Input
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Optional description"
                />
              </div>
            </div>
          </div>

          {/* Dataset Selection */}
          <div className="rounded-lg border p-4 space-y-4">
            <h2 className="text-lg font-semibold">Input Datasets</h2>
            {datasets.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No datasets available. Upload data first.
              </p>
            ) : (
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {datasets.map((ds) => (
                  <label
                    key={ds.id}
                    className="flex items-center gap-3 rounded-md border p-3 cursor-pointer hover:bg-muted/50"
                  >
                    <input
                      type="checkbox"
                      checked={selectedDatasetIds.includes(ds.id)}
                      onChange={() => toggleDataset(ds.id)}
                      className="h-4 w-4"
                    />
                    <span className="font-medium">{ds.name}</span>
                    <span
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        sourceTypeBadge[ds.source_type] ??
                        "bg-gray-100 text-gray-800"
                      }`}
                    >
                      {ds.source_type}
                    </span>
                    {ds.row_count != null ? (
                      <span className="text-xs text-muted-foreground">
                        {ds.row_count.toLocaleString()} rows
                      </span>
                    ) : null}
                  </label>
                ))}
              </div>
            )}
          </div>

          {/* Execution Timing */}
          <div className="rounded-lg border p-4 space-y-4">
            <h2 className="text-lg font-semibold">Execution</h2>
            <div className="space-y-2">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="radio"
                  name="execution"
                  value="save_only"
                  checked={execution === "save_only"}
                  onChange={() => setExecution("save_only")}
                  className="h-4 w-4"
                />
                <div>
                  <span className="font-medium">Save only</span>
                  <p className="text-sm text-muted-foreground">
                    Save the job without running it
                  </p>
                </div>
              </label>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="radio"
                  name="execution"
                  value="immediate"
                  checked={execution === "immediate"}
                  onChange={() => setExecution("immediate")}
                  className="h-4 w-4"
                />
                <div>
                  <span className="font-medium">Run immediately</span>
                  <p className="text-sm text-muted-foreground">
                    Create and start the job right away
                  </p>
                </div>
              </label>
              <label className="flex items-center gap-3 cursor-pointer opacity-50">
                <input
                  type="radio"
                  name="execution"
                  value="scheduled"
                  disabled
                  className="h-4 w-4"
                />
                <div>
                  <span className="font-medium">Schedule (Coming Soon)</span>
                  <p className="text-sm text-muted-foreground">
                    Run at a specified time
                  </p>
                </div>
              </label>
            </div>
          </div>

          {/* Submit */}
          <div className="flex gap-3">
            <Button
              type="submit"
              disabled={
                submitting ||
                !name ||
                !slug ||
                !sql ||
                selectedDatasetIds.length === 0
              }
            >
              {submitting ? "Creating..." : "Create Transform Job"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => router.back()}
            >
              Cancel
            </Button>
          </div>
        </div>

        {/* ── Right Column: SQL & Results ── */}
        <div className="space-y-6">
          {/* SQL Editor */}
          <div className="rounded-lg border p-4 space-y-4">
            <h2 className="text-lg font-semibold">SQL Query</h2>
            {selectedDatasetNames.length > 0 ? (
              <p className="text-sm text-muted-foreground">
                Available tables:{" "}
                {selectedDatasetNames.map((n) => (
                  <code
                    key={n}
                    className="mx-1 rounded bg-muted px-1.5 py-0.5 text-xs font-mono"
                  >
                    {n}
                  </code>
                ))}
              </p>
            ) : null}
            <textarea
              value={sql}
              onChange={(e) => setSql(e.target.value)}
              placeholder={
                selectedDatasetNames.length > 0
                  ? `SELECT * FROM "${selectedDatasetNames[0]}"`
                  : "SELECT ..."
              }
              className="w-full min-h-[160px] rounded-md border bg-background p-3 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              required
            />
            <div className="flex gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={handleValidate}
                disabled={!canValidate || validating}
              >
                {validating ? "Validating..." : "Validate SQL"}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={handlePreview}
                disabled={!canValidate || previewing}
              >
                {previewing ? "Loading..." : "Preview Results"}
              </Button>
            </div>
          </div>

          {/* Validation Result */}
          {validateResult ? (
            <div
              className={`rounded-lg border p-4 ${
                validateResult.valid
                  ? "border-green-200 bg-green-50"
                  : "border-red-200 bg-red-50"
              }`}
            >
              {validateResult.valid ? (
                <div>
                  <p className="font-medium text-green-800">SQL is valid</p>
                  {validateResult.columns &&
                  validateResult.columns.length > 0 ? (
                    <div className="mt-2">
                      <p className="text-sm text-green-700 mb-1">
                        Output columns:
                      </p>
                      <div className="flex flex-wrap gap-2">
                        {validateResult.columns.map((col) => (
                          <span
                            key={col.name}
                            className="inline-flex items-center rounded bg-green-100 px-2 py-0.5 text-xs font-mono"
                          >
                            {col.name}{" "}
                            <span className="ml-1 text-green-600">
                              ({col.type})
                            </span>
                          </span>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : (
                <div>
                  <p className="font-medium text-red-800">
                    SQL validation failed
                  </p>
                  {validateResult.error ? (
                    <p className="mt-1 text-sm text-red-700 font-mono">
                      {validateResult.error}
                    </p>
                  ) : null}
                </div>
              )}
            </div>
          ) : null}

          {/* Preview Table */}
          {previewResult ? (
            <div className="rounded-lg border p-4 space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold">
                  Preview ({previewResult.row_count} rows)
                </h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/50">
                      {previewResult.columns.map((col) => (
                        <th
                          key={col.name}
                          className="px-3 py-2 text-left font-medium whitespace-nowrap"
                        >
                          {col.name}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {previewResult.rows.map((row, idx) => (
                      <tr key={idx} className="border-b last:border-0">
                        {previewResult.columns.map((col) => (
                          <td
                            key={col.name}
                            className="px-3 py-2 whitespace-nowrap font-mono text-xs"
                          >
                            {row[col.name] != null
                              ? String(row[col.name])
                              : "null"}
                          </td>
                        ))}
                      </tr>
                    ))}
                    {previewResult.rows.length === 0 ? (
                      <tr>
                        <td
                          colSpan={previewResult.columns.length}
                          className="px-3 py-4 text-center text-muted-foreground"
                        >
                          No rows returned.
                        </td>
                      </tr>
                    ) : null}
                  </tbody>
                </table>
              </div>
            </div>
          ) : null}
        </div>
      </div>

      {/* Error / Message (full width, outside grid) */}
      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : null}
      {message ? (
        <div className="rounded-lg border border-green-200 bg-green-50 p-4 text-sm text-green-800">
          {message}
        </div>
      ) : null}
    </form>
  );
}
