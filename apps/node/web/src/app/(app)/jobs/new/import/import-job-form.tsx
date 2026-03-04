"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { components } from "@/lib/api/generated";

type Connection = components["schemas"]["Connection"];
type SchemaItem = components["schemas"]["SchemaItem"];
type ConnectionSchemasResponse =
  components["schemas"]["ConnectionSchemasResponse"];
type CreateResponse = components["schemas"]["CreateImportJobResponse"];

function toSlug(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

function extractSpreadsheetId(input: string): string | null {
  // Accept a raw ID (no slashes) or extract from a Google Sheets URL
  if (!input.includes("/")) return input.trim() || null;
  const match = input.match(/\/spreadsheets\/d\/([a-zA-Z0-9_-]+)/);
  return match ? match[1] : null;
}

export function ImportJobForm({
  connections,
}: {
  connections: Connection[];
}) {
  const router = useRouter();

  // Job info
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugEdited, setSlugEdited] = useState(false);
  const [description, setDescription] = useState("");

  // Connection
  const [connectionId, setConnectionId] = useState("");

  // Spreadsheet
  const [spreadsheetUrl, setSpreadsheetUrl] = useState("");

  // Sheets
  const [sheetsLoading, setSheetsLoading] = useState(false);
  const [spreadsheetTitle, setSpreadsheetTitle] = useState("");
  const [sheets, setSheets] = useState<SchemaItem[]>([]);
  const [sheetName, setSheetName] = useState("");
  const [range, setRange] = useState("");

  // Execution
  const [execution, setExecution] = useState<"save_only" | "immediate">(
    "save_only"
  );

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

  function handleConnectionChange(id: string) {
    setConnectionId(id);
    setSpreadsheetUrl("");
    setSheets([]);
    setSheetName("");
    setRange("");
    setSpreadsheetTitle("");
    setError("");
  }

  const selectedConnection = connections.find((c) => c.id === connectionId);
  const isGoogleSheets = selectedConnection?.type === "source-google-sheets";

  const spreadsheetId = extractSpreadsheetId(spreadsheetUrl);

  async function handleLoadSheets() {
    if (!connectionId || !spreadsheetId) return;
    setSheetsLoading(true);
    setSheets([]);
    setSheetName("");
    setSpreadsheetTitle("");
    setError("");
    try {
      const params = new URLSearchParams({ spreadsheet_id: spreadsheetId });
      const res = await fetch(`/api/connections/${connectionId}/schemas?${params}`);
      const data: ConnectionSchemasResponse | { error: string } =
        await res.json();
      if (!res.ok) {
        const errMsg =
          "error" in data ? data.error : `Failed to load sheets (${res.status})`;
        if (errMsg === "credential_expired") {
          setError(
            "Credential expired. Please re-connect your Google account in the Integrations page."
          );
        } else {
          setError(errMsg);
        }
        return;
      }
      const schemasData = data as ConnectionSchemasResponse;
      setSpreadsheetTitle(schemasData.title);
      setSheets(schemasData.items);
      if (schemasData.items.length > 0) {
        setSheetName(schemasData.items[0].name);
      }
    } catch {
      setError("Failed to load sheets");
    } finally {
      setSheetsLoading(false);
    }
  }

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    setMessage("");
    setError("");
    try {
      const body: Record<string, string | undefined> = {
        name,
        slug,
        description: description || undefined,
        connection_id: connectionId,
        spreadsheet_id: spreadsheetId || undefined,
        execution,
      };
      if (sheetName) body.sheet_name = sheetName;
      if (range) body.range = range;

      const res = await fetch("/api/import/jobs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const data: CreateResponse | { error: string } = await res.json();
      if (!res.ok) {
        setError(
          "error" in data ? data.error : `Creation failed (${res.status})`
        );
        return;
      }
      const created = data as CreateResponse;
      if (execution === "immediate") {
        setMessage("Import job created and queued for execution!");
        setTimeout(() => router.push("/jobs"), 1500);
      } else {
        setMessage("Import job created successfully!");
        setTimeout(() => router.push(`/jobs/${created.job.id}`), 1500);
      }
    } catch {
      setError("Creation request failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-2xl">
      {/* Job Info */}
      <div className="rounded-lg border p-4 space-y-4">
        <h2 className="text-lg font-semibold">Job Information</h2>
        <div className="grid gap-3">
          <div>
            <label className="block text-sm font-medium mb-1">Name</label>
            <Input
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="My Import Job"
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
              placeholder="my-import-job"
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

      {/* Connection Selection */}
      <div className="rounded-lg border p-4 space-y-4">
        <h2 className="text-lg font-semibold">Source Connection</h2>
        {connections.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No source connections available. Create a connection first in the
            Connections page.
          </p>
        ) : (
          <div>
            <label className="block text-sm font-medium mb-1">
              Connection
            </label>
            <select
              value={connectionId}
              onChange={(e) => handleConnectionChange(e.target.value)}
              className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              required
            >
              <option value="">Select a connection...</option>
              {connections.map((conn) => (
                <option key={conn.id} value={conn.id}>
                  {conn.name} ({conn.type})
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      {/* Google Sheets Config */}
      {isGoogleSheets && connectionId ? (
        <div className="rounded-lg border p-4 space-y-4">
          <h2 className="text-lg font-semibold">Spreadsheet</h2>
          <div>
            <label className="block text-sm font-medium mb-1">
              Spreadsheet URL
            </label>
            <div className="flex gap-2">
              <Input
                value={spreadsheetUrl}
                onChange={(e) => {
                  setSpreadsheetUrl(e.target.value);
                  setSheets([]);
                  setSheetName("");
                  setSpreadsheetTitle("");
                }}
                placeholder="https://docs.google.com/spreadsheets/d/..."
                className="flex-1"
              />
              <Button
                type="button"
                variant="outline"
                onClick={handleLoadSheets}
                disabled={sheetsLoading || !spreadsheetId}
              >
                {sheetsLoading ? "Loading..." : "Load Sheets"}
              </Button>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Paste the full Google Sheets URL or the spreadsheet ID.
            </p>
          </div>

          {spreadsheetTitle ? (
            <p className="text-sm text-muted-foreground">
              Spreadsheet: <strong>{spreadsheetTitle}</strong>
            </p>
          ) : null}

          {sheets.length > 0 ? (
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium mb-1">Sheet</label>
                <select
                  value={sheetName}
                  onChange={(e) => setSheetName(e.target.value)}
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  {sheets.map((s) => (
                    <option key={s.name} value={s.name}>
                      {s.name}
                      {s.metadata
                        ? ` (${(s.metadata as Record<string, number>).row_count ?? "?"} rows, ${(s.metadata as Record<string, number>).column_count ?? "?"} cols)`
                        : ""}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  Range (optional)
                </label>
                <Input
                  value={range}
                  onChange={(e) => setRange(e.target.value)}
                  placeholder="A1:Z100"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  A1 notation. Leave empty to import the entire sheet.
                </p>
              </div>
            </div>
          ) : null}
        </div>
      ) : null}

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
        </div>
      </div>

      {/* Submit */}
      <div className="flex gap-3">
        <Button
          type="submit"
          disabled={submitting || !name || !slug || !connectionId || (isGoogleSheets && !spreadsheetId)}
        >
          {submitting ? "Creating..." : "Create Import Job"}
        </Button>
        <Button type="button" variant="outline" onClick={() => router.back()}>
          Cancel
        </Button>
      </div>

      {/* Error / Message */}
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
