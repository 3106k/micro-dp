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
