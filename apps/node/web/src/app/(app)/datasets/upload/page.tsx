import Link from "next/link";

import { UploadsManager } from "./uploads-manager";

export default function DatasetUploadPage() {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <Link
          href="/datasets"
          className="text-sm text-muted-foreground underline-offset-2 hover:underline"
        >
          &larr; Back to Datasets
        </Link>
        <h1 className="text-2xl font-semibold tracking-tight">Upload</h1>
      </div>
      <UploadsManager />
    </div>
  );
}
