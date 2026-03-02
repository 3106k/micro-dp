import { UploadsManager } from "./uploads-manager";

export default function UploadsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Uploads</h1>
      <UploadsManager />
    </div>
  );
}
