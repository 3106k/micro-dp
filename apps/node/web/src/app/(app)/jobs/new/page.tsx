import Link from "next/link";

export default function NewJobPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Create New Job</h1>
      <p className="text-muted-foreground">
        Select the type of job you want to create.
      </p>

      <div className="grid gap-4 sm:grid-cols-2 max-w-2xl">
        <Link
          href="/jobs/new/transform"
          className="group rounded-lg border p-6 hover:border-primary hover:bg-muted/50 transition-colors"
        >
          <div className="space-y-2">
            <h2 className="text-lg font-semibold group-hover:text-primary">
              Transform
            </h2>
            <p className="text-sm text-muted-foreground">
              Write SQL to transform existing datasets into new ones.
            </p>
          </div>
        </Link>

        <Link
          href="/jobs/new/import"
          className="group rounded-lg border p-6 hover:border-primary hover:bg-muted/50 transition-colors"
        >
          <div className="space-y-2">
            <h2 className="text-lg font-semibold group-hover:text-primary">
              Import
            </h2>
            <p className="text-sm text-muted-foreground">
              Import data from external sources like Google Sheets.
            </p>
          </div>
        </Link>
      </div>
    </div>
  );
}
