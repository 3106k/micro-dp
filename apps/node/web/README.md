# micro-dp web (Next.js + shadcn/ui)

## Setup

```bash
cd apps/node/web
npm install
```

## Run

```bash
npm run dev
```

Open `http://localhost:3000` and confirm shadcn `Button` components are rendered on the top page.

## Added shadcn/ui foundation

- `components.json`
- `tailwind.config.ts`
- `postcss.config.js`
- `src/app/globals.css`
- `src/lib/utils.ts` (`cn` utility)
- `src/components/ui/button.tsx`

## Tracker SDK sample

Use `@micro-dp/sdk-tracker` with environment-based endpoint configuration.

```ts
import { init, page, track } from "@micro-dp/sdk-tracker";

init({
  endpoint: process.env.NEXT_PUBLIC_TRACKER_ENDPOINT ?? "",
  enabled: process.env.NEXT_PUBLIC_TRACKER_ENABLED === "true",
  debug: process.env.NEXT_PUBLIC_TRACKER_DEBUG === "true",
  tenantId: "tenant-local"
});

page("home");
track("signup_clicked", { source: "hero" });
```

Example env values:

- local: `NEXT_PUBLIC_TRACKER_ENDPOINT=http://localhost:8080/api/v1/events`
- staging: `NEXT_PUBLIC_TRACKER_ENDPOINT=https://api-stg.example.com/api/v1/events`
- production: `NEXT_PUBLIC_TRACKER_ENDPOINT=https://api.example.com/api/v1/events`

## Upload UI manual verification

1. Sign in and open `http://localhost:3000/uploads`.
2. Select one valid file (`.csv`, `.json`, `.parquet`, `.xlsx`, `.txt`, `.tsv`, `.gz`, `.zip`) and run upload.
3. Confirm progress reaches 100% and completion message shows `uploaded`.
4. Enable `multiple` and upload 2+ valid files; confirm all complete successfully.
5. Failure checks:
   - Select a file larger than 100MB and confirm size error is shown.
   - Select a non-allowed extension and confirm extension error is shown.
   - Stop MinIO/network and confirm direct upload failure is shown.
6. Verify tenant isolation by signing in with a different tenant and confirming uploaded records are not mixed.
