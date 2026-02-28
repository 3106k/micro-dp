# @micro-dp/sdk-tracker

JavaScript/TypeScript tracker SDK for `POST /api/v1/events`.

## Install (local monorepo)

```bash
cd apps/node/sdk-tracker
npm install
npm run build
```

## API

```ts
import { init, track, identify, page, flush } from "@micro-dp/sdk-tracker";

init({
  endpoint: "http://localhost:8080/api/v1/events",
  tenantId: "tenant-1",
  enabled: true,
  debug: true
});

track("signup_clicked", { plan: "pro" });
identify("user-123", { email: "user@example.com" });
page("home", { referrer: "ad" });
await flush();
```

## Config

- `endpoint`: event endpoint (required)
- `enabled`: enable/disable sending
- `debug`: logs payload and retry information
- `flushIntervalMs`: periodic flush interval
- `maxQueueSize`: flush when queue size reaches threshold
- `retryMaxAttempts`: max retries for fetch transport
- `retryBaseDelayMs`: retry base delay (exponential backoff)

## Environment-based endpoint example

```ts
init({
  endpoint: process.env.NEXT_PUBLIC_TRACKER_ENDPOINT ?? "http://localhost:8080/api/v1/events",
  enabled: process.env.NEXT_PUBLIC_TRACKER_ENABLED === "true",
  debug: process.env.NODE_ENV !== "production"
});
```

- local: `http://localhost:8080/api/v1/events`
- staging: `https://api-stg.example.com/api/v1/events`
- production: `https://api.example.com/api/v1/events`

## Transport behavior

- Prefer `navigator.sendBeacon` on page hide / unload
- Fallback to `fetch` (`keepalive: true`)
- Failed sends are retried with exponential backoff
