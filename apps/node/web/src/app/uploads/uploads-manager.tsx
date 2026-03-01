"use client";

import { FormEvent, useMemo, useRef, useState } from "react";
import { Upload } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import type { components } from "@/lib/api/generated";

type CreateUploadPresignResponse =
  components["schemas"]["CreateUploadPresignResponse"];
type Upload = components["schemas"]["Upload"];

type UploadItemStatus =
  | "pending"
  | "presigning"
  | "uploading"
  | "uploaded"
  | "failed";

type UploadItem = {
  id: string;
  file: File;
  status: UploadItemStatus;
  progress: number;
  error: string;
};

const MAX_FILE_SIZE_BYTES = 100 * 1024 * 1024;
const MAX_FILES_PER_REQUEST = 10;
const ACCEPTED_EXTENSIONS = [
  ".csv",
  ".json",
  ".parquet",
  ".xlsx",
  ".txt",
  ".tsv",
  ".gz",
  ".zip",
] as const;

function getFileExtension(name: string): string {
  const index = name.lastIndexOf(".");
  if (index < 0) {
    return "";
  }
  return name.slice(index).toLowerCase();
}

function formatErrorMessage(message: string): string {
  const lower = message.toLowerCase();
  if (lower.includes("exceeds max size") || lower.includes("invalid size")) {
    return "ファイルサイズが上限を超えているか不正です（上限100MB）。";
  }
  if (lower.includes("extension") && lower.includes("not allowed")) {
    return "許可されていない拡張子です。対応形式を選択してください。";
  }
  if (lower.includes("too many files")) {
    return "1回でアップロードできるファイル数は最大10件です。";
  }
  if (lower.includes("not authenticated") || lower.includes("unauthorized")) {
    return "認証が無効です。再ログインしてください。";
  }
  return message;
}

function getClientValidationError(file: File): string | null {
  if (file.size <= 0) {
    return "空ファイルはアップロードできません。";
  }
  if (file.size > MAX_FILE_SIZE_BYTES) {
    return "ファイルサイズが100MBを超えています。";
  }
  const ext = getFileExtension(file.name);
  if (!ext || !ACCEPTED_EXTENSIONS.includes(ext as (typeof ACCEPTED_EXTENSIONS)[number])) {
    return `拡張子 ${ext || "(なし)"} は許可されていません。`;
  }
  return null;
}

function uploadWithProgress(
  file: File,
  presignedUrl: string,
  onProgress: (progress: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", presignedUrl, true);
    xhr.timeout = 120_000;

    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable) {
        return;
      }
      const progress = Math.min(
        100,
        Math.round((event.loaded / event.total) * 100)
      );
      onProgress(progress);
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        onProgress(100);
        resolve();
        return;
      }
      reject(new Error(`direct upload failed (${xhr.status})`));
    };

    xhr.onerror = () => {
      reject(new Error("network failed during direct upload"));
    };

    xhr.ontimeout = () => {
      reject(new Error("direct upload timed out"));
    };

    xhr.setRequestHeader(
      "Content-Type",
      file.type || "application/octet-stream"
    );
    xhr.send(file);
  });
}

function fileId(file: File): string {
  return `${file.name}:${file.size}:${file.lastModified}`;
}

function statusTone(status: UploadItemStatus): string {
  if (status === "uploaded") {
    return "text-green-700";
  }
  if (status === "failed") {
    return "text-destructive";
  }
  if (status === "uploading" || status === "presigning") {
    return "text-blue-700";
  }
  return "text-muted-foreground";
}

export function UploadsManager() {
  const [allowMultiple, setAllowMultiple] = useState(true);
  const [items, setItems] = useState<UploadItem[]>([]);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [lastCompletedUpload, setLastCompletedUpload] = useState<Upload | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const hasSelectedFiles = useMemo(() => items.length > 0, [items.length]);

  function updateItem(id: string, patch: Partial<UploadItem>) {
    setItems((prev) => prev.map((item) => (item.id === id ? { ...item, ...patch } : item)));
  }

  function handleFileSelect(files: FileList | null) {
    if (!files) {
      setItems([]);
      return;
    }

    const chosen = Array.from(files);
    const normalized = allowMultiple ? chosen : chosen.slice(0, 1);

    const nextItems = normalized.map((file) => ({
      id: fileId(file),
      file,
      status: "pending" as const,
      progress: 0,
      error: "",
    }));

    setItems(nextItems);
    setMessage("");
    setLastCompletedUpload(null);
  }

  async function uploadSelectedFiles(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (loading || items.length === 0) {
      return;
    }

    setMessage("");
    setLastCompletedUpload(null);

    if (items.length > MAX_FILES_PER_REQUEST) {
      setMessage("1回でアップロードできるファイル数は最大10件です。");
      return;
    }

    const invalidById = new Map<string, string>();
    for (const item of items) {
      const validationError = getClientValidationError(item.file);
      if (validationError) {
        invalidById.set(item.id, validationError);
      }
    }

    if (invalidById.size > 0) {
      setItems((prev) =>
        prev.map((item) => {
          const error = invalidById.get(item.id);
          if (!error) {
            return { ...item, status: "pending", progress: 0, error: "" };
          }
          return { ...item, status: "failed", progress: 0, error };
        })
      );
      setMessage("アップロード前の検証でエラーがあります。ファイルを修正してください。");
      return;
    }

    setLoading(true);
    setItems((prev) =>
      prev.map((item) => ({
        ...item,
        status: "presigning",
        progress: 0,
        error: "",
      }))
    );

    try {
      const presignRes = await fetch("/api/uploads/presign", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          files: items.map((item) => ({
            filename: item.file.name,
            content_type: item.file.type || "application/octet-stream",
            size_bytes: item.file.size,
          })),
        }),
      });

      let presignData: CreateUploadPresignResponse | null = null;
      if (presignRes.ok) {
        presignData = (await presignRes.json()) as CreateUploadPresignResponse;
      } else {
        const err = (await presignRes.json()) as { error?: string };
        const mapped = formatErrorMessage(
          err.error ?? `presign failed (${presignRes.status})`
        );
        setItems((prev) =>
          prev.map((item) => ({
            ...item,
            status: "failed",
            error: mapped,
            progress: 0,
          }))
        );
        setMessage(mapped);
        return;
      }

      if (!presignData || presignData.files.length !== items.length) {
        const mismatchMessage =
          "presign 応答の件数が選択ファイル数と一致しませんでした。";
        setItems((prev) =>
          prev.map((item) => ({
            ...item,
            status: "failed",
            error: mismatchMessage,
            progress: 0,
          }))
        );
        setMessage(mismatchMessage);
        return;
      }

      const results = await Promise.allSettled(
        items.map(async (item, index) => {
          const presigned = presignData.files[index];
          updateItem(item.id, { status: "uploading", progress: 1, error: "" });
          await uploadWithProgress(item.file, presigned.presigned_url, (progress) => {
            updateItem(item.id, { progress });
          });
          updateItem(item.id, { status: "uploaded", progress: 100, error: "" });
          return { id: item.id };
        })
      );

      const failed = results
        .map((result, index) => ({ result, item: items[index] }))
        .filter((entry) => entry.result.status === "rejected");

      if (failed.length > 0) {
        for (const entry of failed) {
          const reason =
            entry.result.status === "rejected"
              ? formatErrorMessage(
                  entry.result.reason instanceof Error
                    ? entry.result.reason.message
                    : "direct upload failed"
                )
              : "direct upload failed";
          updateItem(entry.item.id, {
            status: "failed",
            error: reason,
          });
        }
        setMessage(
          `${failed.length}件のファイルアップロードに失敗しました。ネットワークまたはCORS設定を確認してください。`
        );
        return;
      }

      const completeRes = await fetch(
        `/api/uploads/${presignData.upload_id}/complete`,
        {
          method: "POST",
        }
      );
      if (!completeRes.ok) {
        const err = (await completeRes.json()) as { error?: string };
        setMessage(
          formatErrorMessage(
            err.error ?? `complete failed (${completeRes.status})`
          )
        );
        return;
      }

      const completed = (await completeRes.json()) as Upload;
      setLastCompletedUpload(completed);
      setMessage(`アップロード完了: ${items.length}件 (upload_id: ${completed.id})`);
    } catch (error) {
      setMessage(
        formatErrorMessage(
          error instanceof Error ? error.message : "unexpected upload error"
        )
      );
    } finally {
      setLoading(false);
    }
  }

  const accept = ACCEPTED_EXTENSIONS.join(",");

  return (
    <div className="space-y-6">
      <form onSubmit={uploadSelectedFiles} className="space-y-4 rounded-lg border p-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">File Uploader</h2>
            <p className="text-sm text-muted-foreground">
              Presigned URL を使ってブラウザから直接 MinIO へアップロードします。
            </p>
          </div>
          <Label className="flex items-center gap-2 text-sm font-normal">
            <input
              type="checkbox"
              className="h-4 w-4"
              checked={allowMultiple}
              onChange={(event) => {
                setAllowMultiple(event.target.checked);
                setItems([]);
                setMessage("");
                setLastCompletedUpload(null);
              }}
              disabled={loading}
            />
            Enable multiple files
          </Label>
        </div>

        <div className="space-y-2">
          <Label>Files</Label>
          <div
            role="button"
            tabIndex={0}
            className={cn(
              "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed px-4 py-8 text-center transition-colors",
              isDragging
                ? "border-primary bg-primary/5"
                : "border-muted-foreground/25 hover:border-primary/50 hover:bg-muted/50",
              loading && "pointer-events-none opacity-50"
            )}
            onClick={() => fileInputRef.current?.click()}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                fileInputRef.current?.click();
              }
            }}
            onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={(e) => {
              e.preventDefault();
              setIsDragging(false);
              if (!loading) handleFileSelect(e.dataTransfer.files);
            }}
          >
            <Upload className="h-8 w-8 text-muted-foreground" />
            <p className="text-sm font-medium">
              ドラッグ&ドロップ または クリックしてファイルを選択
            </p>
            <p className="text-xs text-muted-foreground">
              最大10ファイル / 各100MB。対応拡張子: {ACCEPTED_EXTENSIONS.join(", ")}
            </p>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            className="sr-only"
            multiple={allowMultiple}
            accept={accept}
            onChange={(e) => { handleFileSelect(e.target.files); e.target.value = ""; }}
            disabled={loading}
          />
        </div>

        <div className="flex items-center gap-2">
          <Button type="submit" disabled={!hasSelectedFiles || loading}>
            {loading ? "Uploading..." : "Start Upload"}
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={loading}
            onClick={() => {
              setItems([]);
              setMessage("");
              setLastCompletedUpload(null);
            }}
          >
            Clear
          </Button>
        </div>

        {message ? <p className="text-sm text-muted-foreground">{message}</p> : null}
        {lastCompletedUpload ? (
          <p className="text-sm text-muted-foreground">
            Complete API status: <span className="font-medium">{lastCompletedUpload.status}</span>
          </p>
        ) : null}
      </form>

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">File</th>
              <th className="px-4 py-3 text-left font-medium">Size</th>
              <th className="px-4 py-3 text-left font-medium">Progress</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Error</th>
            </tr>
          </thead>
          <tbody>
            {items.map((item) => (
              <tr key={item.id} className="border-b last:border-0">
                <td className="px-4 py-3">{item.file.name}</td>
                <td className="px-4 py-3">{Math.ceil(item.file.size / 1024)} KB</td>
                <td className="px-4 py-3">
                  <div className="h-2 w-40 overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full bg-foreground/70 transition-all"
                      style={{ width: `${item.progress}%` }}
                    />
                  </div>
                  <span className="text-xs text-muted-foreground">{item.progress}%</span>
                </td>
                <td className={`px-4 py-3 capitalize ${statusTone(item.status)}`}>
                  {item.status}
                </td>
                <td className="px-4 py-3 text-xs text-destructive">{item.error || "-"}</td>
              </tr>
            ))}
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-6 text-center text-muted-foreground">
                  No files selected.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </div>
  );
}
