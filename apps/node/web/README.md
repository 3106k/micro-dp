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

## Environment variables

- `API_BACKEND_URL`: backend API base URL used by server-side API client.
  - default: `http://localhost:8080`
- `NEXT_PUBLIC_TRACKER_ENDPOINT`: tracker ingest endpoint for browser SDK.
- `NEXT_PUBLIC_TRACKER_ENABLED`: set `true` to enable tracker.
- `NEXT_PUBLIC_TRACKER_DEBUG`: set `true` for tracker debug logs.

Example:

```bash
API_BACKEND_URL=http://localhost:8980
NEXT_PUBLIC_TRACKER_ENDPOINT=http://localhost:8980/api/v1/events
NEXT_PUBLIC_TRACKER_ENABLED=true
NEXT_PUBLIC_TRACKER_DEBUG=false
```

## Auth and tenant foundation

- Login: `POST /api/auth/login` stores auth token cookie and default tenant cookie.
- Logout: `POST /api/auth/logout` clears auth/tenant cookies.
- Me: `GET /api/auth/me` proxies backend `/api/v1/auth/me`.
- Tenant switch: `POST /api/auth/tenant` updates selected tenant cookie after validating membership.
- Protected pages (`/dashboard`, `/jobs`, `/job-runs`, `/datasets`, `/connections`, `/admin`) are guarded by middleware.
- API routes under `src/app/api/**` attach `Authorization` and `X-Tenant-ID` from cookies when proxying to backend APIs.

## UI conventions

- Global toast UI is provided by `ToastProvider` in app root layout.
- Auth forms share common loading/error UI via `SubmitButton` and `FormError`.
- Dashboard header includes tenant switcher (when user belongs to multiple tenants).

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
