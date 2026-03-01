# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

micro-dp is a data pipeline platform built as a monorepo. The stack is Next.js (UI) → Go backend (API + Worker) → Valkey (queue) → External APIs / DuckDB / MinIO. Multi-tenant with deferred job pattern for rate limit handling.

## Architecture

**Two binaries (`cmd/api`, `cmd/worker`)**: The Go backend builds two separate binaries from `apps/golang/backend/cmd/api/` and `apps/golang/backend/cmd/worker/`. They share packages (`domain/`, `usecase/`, `handler/`, `db/`) but have independent entry points, dependencies, and lifecycle.

**Monorepo layout by language**:

- `apps/golang/backend/` — Go API + Worker
- `apps/node/web/` — Next.js frontend (App Router, standalone output)
- `apps/docker/` — All Docker/Compose configuration
- `apps/shell/` — Shell scripts (worktree env setup)

**Service dependency chain**: web → api → valkey, minio-init → minio. Worker depends on api (healthy) + valkey.

**SQLite via volume**: `.data/sqlite` volume is shared between api and worker containers. Driver: modernc.org/sqlite (pure Go). CGO is enabled due to DuckDB (go-duckdb) dependency.

**Go layered architecture**:

```
cmd/api/       — API entry point (auth, tenant management)
cmd/worker/    — Worker entry point (queue consumer, DuckDB, MinIO/Iceberg)
domain/        — entities, repository interfaces
usecase/       — application services / business logic
handler/       — HTTP handlers (adapter)
db/            — repository implementations, migrations
queue/         — Valkey queue implementation (go-redis/v9)
worker/        — job processing (EventConsumer, ParquetWriter)
storage/       — MinIO client wrapper (minio-go/v7)
internal/      — private packages:
  observability/   — OpenTelemetry traces + Prometheus metrics
  openapi/         — oapi-codegen generated types/interfaces
  featureflag/     — OpenFeature feature flag infrastructure
  notification/    — Email notification (SendGrid / log provider)
```

依存方向: `handler/` → `usecase/` → `domain/` ← `db/`, `queue/`。`domain/` は他パッケージに依存しない。

## Commands

All commands run from the repo root via Makefile:

```bash
make up                # docker compose up -d --build
make down              # stop all services
make build             # build images only
make logs              # stream all service logs
make ps                # show container status
make health            # curl healthz endpoints + valkey ping
make clean             # down + remove volumes and images
make dev-api           # run API with hot reload (air)
make dev-worker        # run Worker with hot reload (air)
make e2e-cli           # run E2E CLI tests against local API
make e2e-ci-template   # up → e2e-cli → down (CI 用)
make sdk-tracker-build # build tracker SDK
make sdk-tracker-test  # run tracker SDK tests
```

### Worktree workflow

Worktrees are created at `../micro-dp-{branch}/` (outside the repo):

```bash
make worktree BRANCH=feature-x      # git worktree add + auto-generate .env with unique ports
make worktree-rm BRANCH=feature-x   # stop containers + remove worktree
make worktree-ls                    # list all worktrees
```

Port isolation: `apps/shell/setup-worktree-env.sh` hashes the branch name to a slot (1-9), applies offset `slot * 100` to all host ports.

### Running services directly (without Docker)

```bash
# Go backend
cd apps/golang/backend
go run ./cmd/api           # API on :8080
go run ./cmd/worker        # Worker on :8081

# Next.js frontend
cd apps/node/web
npm install
npm run dev              # Dev server on :3000
```

## Environment

All variables are configurable via `.env` (loaded by docker-compose from `apps/docker/`). See `.env.example` for full list.

### Host Ports

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `COMPOSE_PROJECT_NAME` | `micro-dp` | Container name prefix |
| `API_HOST_PORT` | `8080` | Go API host port |
| `WEB_HOST_PORT` | `3000` | Next.js host port |
| `VALKEY_HOST_PORT` | `6379` | Valkey host port |
| `MINIO_API_HOST_PORT` | `9000` | MinIO API host port |
| `MINIO_CONSOLE_HOST_PORT` | `9001` | MinIO console host port |

Internal container ports are fixed; only host-side ports change per environment.

### Backend

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `JWT_SECRET` | `dev-secret-...` | JWT 署名鍵 (必須) |
| `SQLITE_PATH` | `/data/sqlite/micro-dp.db` | SQLite ファイルパス |
| `BOOTSTRAP_SUPERADMINS` | `false` | 起動時に superadmin を作成 |
| `SUPERADMIN_EMAILS` | — | superadmin メールアドレス (カンマ区切り) |

### Infrastructure

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `VALKEY_ADDR` | `valkey:6379` | Valkey 接続先 |
| `MINIO_ENDPOINT` | `minio:9000` | MinIO 接続先 |
| `MINIO_ROOT_USER` | `minioadmin` | MinIO 認証 |
| `MINIO_ROOT_PASSWORD` | `minioadmin` | MinIO 認証 |
| `MINIO_BUCKET` | `micro-dp` | MinIO バケット名 |

### Feature Flags

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `FF_EVENTS_INGEST` | `true` | Events 取り込み |
| `FF_DATASETS_API` | `true` | Datasets API |
| `FF_ADMIN_TENANTS` | `true` | Admin テナント管理 |
| `FF_UPLOADS_API` | `true` | Uploads API |

### Notification

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `NOTIFICATION_PROVIDER` | `log` | メール送信プロバイダ (`sendgrid` or `log`) |
| `SENDGRID_API_KEY` | — | SendGrid API キー (`sendgrid` 時のみ必須) |
| `NOTIFICATION_FROM_ADDRESS` | `noreply@example.com` | 送信元メールアドレス |
| `NOTIFICATION_FROM_NAME` | `micro-dp` | 送信元表示名 |

## Health Checks

- API: `GET /healthz` → `{"status":"ok"}`
- Worker: `GET /healthz` on port 8081
- Web: `GET /api/health` → `{"status":"ok"}`
- Valkey: `valkey-cli ping`
- MinIO: `mc ready local`

## Docker

- `docker-compose.yaml` — production services (context is repo root `../..`)
- `docker-compose.override.yaml` — dev mode with volume mounts + hot reload (air for Go, next dev for Node)
- Production builds use multi-stage Dockerfiles; dev builds mount source directly

## API Endpoints

### Public

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | `/healthz` | Health check |
| GET | `/metrics` | Prometheus metrics |
| POST | `/api/v1/auth/register` | Register user (creates tenant) |
| POST | `/api/v1/auth/login` | Login (returns JWT) |

### Authenticated (Bearer)

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | `/api/v1/auth/me` | Current user + tenants |

### Tenant-scoped (Bearer + X-Tenant-ID)

| Method | Path | Description |
| ------ | ---- | ----------- |
| POST | `/api/v1/events` | Ingest event (202 Accepted) |
| GET | `/api/v1/events/summary` | Event counts summary |
| POST | `/api/v1/job_runs` | Create job run |
| GET | `/api/v1/job_runs` | List job runs |
| GET | `/api/v1/job_runs/{id}` | Get job run |
| GET | `/api/v1/job_runs/{job_run_id}/modules` | List job run modules |
| GET | `/api/v1/job_runs/{job_run_id}/modules/{id}` | Get job run module |
| GET | `/api/v1/job_runs/{job_run_id}/artifacts` | List job run artifacts |
| GET | `/api/v1/job_runs/{job_run_id}/artifacts/{id}` | Get job run artifact |
| POST | `/api/v1/jobs` | Create job |
| GET | `/api/v1/jobs` | List jobs |
| GET | `/api/v1/jobs/{id}` | Get job |
| PUT | `/api/v1/jobs/{id}` | Update job |
| POST | `/api/v1/jobs/{job_id}/versions` | Create job version |
| GET | `/api/v1/jobs/{job_id}/versions` | List job versions |
| GET | `/api/v1/jobs/{job_id}/versions/{version_id}` | Get job version detail |
| POST | `/api/v1/jobs/{job_id}/versions/{version_id}/publish` | Publish job version |
| POST | `/api/v1/module_types` | Create module type |
| GET | `/api/v1/module_types` | List module types |
| GET | `/api/v1/module_types/{id}` | Get module type |
| POST | `/api/v1/module_types/{id}/schemas` | Create module type schema |
| GET | `/api/v1/module_types/{id}/schemas` | List module type schemas |
| POST | `/api/v1/connections` | Create connection |
| GET | `/api/v1/connections` | List connections |
| GET | `/api/v1/connections/{id}` | Get connection |
| PUT | `/api/v1/connections/{id}` | Update connection |
| DELETE | `/api/v1/connections/{id}` | Delete connection |
| GET | `/api/v1/datasets` | List datasets |
| GET | `/api/v1/datasets/{id}` | Get dataset |
| POST | `/api/v1/uploads/presign` | Request presigned upload URLs |
| POST | `/api/v1/uploads/{id}/complete` | Mark upload complete |

### Admin (Bearer + Superadmin)

| Method | Path | Description |
| ------ | ---- | ----------- |
| POST | `/api/v1/admin/tenants` | Create tenant |
| GET | `/api/v1/admin/tenants` | List tenants |
| PATCH | `/api/v1/admin/tenants/{id}` | Update tenant |

## Events Ingest Pipeline

非同期イベント処理パイプライン。tracker SDK からのイベントを受信し、Parquet 形式で MinIO に永続化する。

### アーキテクチャ

```
POST /api/v1/events → Valkey dedup (SET NX TTL 24h) → Valkey LIST (LPUSH)
                                                          ↓
Worker (BRPOP) → batch buffer (1000件 or 30秒) → DuckDB Parquet → MinIO
                                                          ↓ (失敗時)
                                                        DLQ LIST
```

### Valkey キー設計

| Key | Type | Purpose |
|-----|------|---------|
| `micro-dp:events:ingest` | LIST | メインキュー |
| `micro-dp:events:dlq` | LIST | Dead Letter Queue |
| `micro-dp:events:seen:{tenant_id}:{event_id}` | STRING | 重複チェック (TTL 24h) |

### MinIO オブジェクトキー

```
events/{tenant_id}/dt={YYYY-MM-DD}/{timestamp}_{batch_id}.parquet
```

### ローカル検証手順

```bash
# 1. サービス起動
make down && make up && make health

# 2. ユーザー登録 + ログイン
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Passw0rd!123","display_name":"Test"}' | jq .

curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Passw0rd!123"}' | jq .
# → token と tenant_id を控える

# 3. イベント送信 (202 Accepted)
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -H "Content-Type: application/json" \
  -d '{"event_id":"test-1","event_name":"page_view","properties":{"page":"/home"},"event_time":"2026-02-28T10:00:00Z"}'
# → {"event_id":"test-1","status":"accepted"}

# 4. 重複チェック (409 Conflict)
# 同じ event_id で再送すると 409 が返る

# 5. E2E テスト
make e2e-cli

# 6. Worker flush 確認 (30秒待機後)
cd apps/docker && docker compose logs worker | grep "flushed batch"

# 7. MinIO 確認
cd apps/docker && docker compose exec minio sh -c \
  'mc alias set m http://localhost:9000 minioadmin minioadmin && mc ls m/micro-dp/events/ --recursive'

# 8. Prometheus メトリクス確認
curl -s http://localhost:8080/metrics | grep events_
```

### メトリクス

| Metric | Type | Location | Description |
|--------|------|----------|-------------|
| `events_received_total` | counter | API | 受信イベント数 |
| `events_enqueued_total` | counter | API | enqueue 成功数 |
| `events_duplicate_total` | counter | API | 重複スキップ数 |
| `events_processed_total` | counter | Worker | Parquet 書き込み成功数 |
| `events_failed_total` | counter | Worker | DLQ 退避数 |
| `events_batch_size` | histogram | Worker | バッチあたりイベント数 |
| `events_batch_duration_seconds` | histogram | Worker | バッチ処理時間 |

## CSV Import Pipeline

upload complete 時に Valkey queue へジョブを投入し、Worker が非同期で CSV→Parquet 変換 + dataset catalog 登録を行う。

### アーキテクチャ

```
POST /api/v1/uploads/{id}/complete → Valkey LIST (LPUSH)
                                         ↓
Worker (BRPOP) → Valkey dedup (SET NX TTL 24h)
                                         ↓
for each .csv file:
  MinIO download → DuckDB read_csv_auto → COPY TO Parquet → MinIO upload
                                         ↓
  Dataset upsert (SQLite ON CONFLICT)
                                         ↓ (失敗時)
                                       DLQ LIST
```

### Valkey キー設計 (uploads)

| Key | Type | Purpose |
|-----|------|---------|
| `micro-dp:uploads:ingest` | LIST | メインキュー |
| `micro-dp:uploads:dlq` | LIST | Dead Letter Queue |
| `micro-dp:uploads:seen:{upload_id}` | STRING | 重複チェック (TTL 24h) |

### MinIO オブジェクトキー (imports)

```
imports/{tenant_id}/dt={YYYY-MM-DD}/{file_id}.parquet
```

### メトリクス (uploads)

| Metric | Type | Location | Description |
|--------|------|----------|-------------|
| `uploads_processed_total` | counter | Worker | 処理完了アップロード数 |
| `uploads_failed_total` | counter | Worker | DLQ 退避数 |
| `uploads_files_converted_total` | counter | Worker | Parquet 変換ファイル数 |
| `uploads_rows_total` | counter | Worker | インポート行数 |
| `uploads_duplicate_total` | counter | Worker | 重複スキップ数 |
| `uploads_processing_duration_seconds` | histogram | Worker | アップロード処理時間 |

## Tracker SDK Integration

フロントエンドからのイベント計測。`@micro-dp/sdk-tracker` を Next.js アプリに統合。

### アーキテクチャ

```
Browser (SDK) → POST /api/events → Next.js API route (proxy) → POST /api/v1/events (Go API)
                                     ↑ cookie → Bearer token + X-Tenant-ID 付与
                                     ↑ バッチ分解 (SDK batch → 個別 event)
```

### 主要ファイル

| ファイル | 役割 |
|---------|------|
| `apps/node/sdk-tracker/` | Tracker SDK パッケージ |
| `apps/node/web/src/app/api/events/route.ts` | イベントプロキシ API route (バッチ分解) |
| `apps/node/web/src/app/api/events/summary/route.ts` | サマリ API proxy |
| `apps/node/web/src/components/tracker-provider.tsx` | TrackerProvider (client component) |
| `apps/node/web/src/app/dashboard/event-summary.tsx` | イベントサマリ表示 (server component) |

### 計測イベント

| イベント名 | トリガー |
|-----------|---------|
| `page_view` | pathname 変化時 (TrackerProvider) |
| `login_success` | ログイン成功後 `?event=login_success` で遷移 |
| `sign_out` | サインアウトボタン押下時 (beacon) |

### Valkey カウンター

| Key | Type | Purpose |
|-----|------|---------|
| `micro-dp:events:count:{tenant_id}` | HASH | イベント名ごとの件数 (`HINCRBY`) |

### 環境変数

| Variable | Default | Description |
|----------|---------|-------------|
| `TRACKER_ENABLED` | `true` | トラッカー有効/無効 |
| `TRACKER_DEBUG` | `false` | コンソールログ出力 |

## Contract-First OpenAPI (SSOT)

`spec/openapi/v1.yaml` is the single source of truth for API contracts.

### OpenAPI workflow

Run from repo root:

```bash
make openapi-lint         # Redocly lint for spec/openapi/v1.yaml
make openapi-bundle       # bundle output to spec/openapi/dist/v1.bundle.yaml
make openapi-generate-fe  # generate TS types to apps/node/web/src/lib/api/generated.ts
make openapi-generate-be  # generate Go types/server interfaces to apps/golang/backend/internal/openapi/*.gen.go
make openapi-generate     # run FE + BE generation
make openapi-check        # lint + generate + git diff --exit-code (drift detection)
```

### Rule

When API contract is changed:

1. Update `spec/openapi/v1.yaml`
2. Run `make openapi-generate` (or `make openapi-check`)
3. Commit spec and generated artifacts together

## Feature Flags

OpenFeature ベースの feature flag 基盤。初期は環境変数プロバイダ (`FF_*`)、将来 flagd / Unleash 等に差し替え可能。

### パッケージ構成

| ファイル | 役割 |
|---------|------|
| `internal/featureflag/featureflag.go` | LoadConfig / Init / IsEnabled / LogStartup |
| `internal/featureflag/env_provider.go` | OpenFeature EnvProvider 実装 |

### フラグ一覧

| Flag 定数 | 環境変数 | デフォルト | 対象機能 |
|-----------|---------|-----------|---------|
| `events_ingest` | `FF_EVENTS_INGEST` | `true` | Events 取り込み |
| `datasets_api` | `FF_DATASETS_API` | `true` | Datasets API |
| `admin_tenants` | `FF_ADMIN_TENANTS` | `true` | Admin テナント管理 |
| `uploads_api` | `FF_UPLOADS_API` | `true` | Uploads API |

### 初期化パターン

`cmd/api/main.go` と `cmd/worker/main.go` で observability 初期化の直後に呼び出し:

```go
ffCfg := featureflag.LoadConfig()
featureflag.Init(ffCfg)
featureflag.LogStartup(ffCfg)
```

### フラグ判定

```go
if featureflag.IsEnabled(featureflag.FlagEventsIngest) {
    // feature enabled
}
```

## Notification

メール通知基盤。送信プロバイダを環境変数 (`NOTIFICATION_PROVIDER`) で切替可能。開発環境は `log` (ログ出力のみ)、本番は `sendgrid` を使用。

### アーキテクチャ

```
AuthService.Register() → notification.RenderWelcome() → EmailSender.Send()
                                                            ↓
                                                  logSender (dev) or sendGridSender (prod)
```

- 同期送信: usecase 内で直接送信（シンプルさ優先）
- 送信失敗時はログ出力のみ（リクエスト自体は失敗させない）
- SendGrid 実装: 最大 3 回試行 (500ms → 1s backoff)

### パッケージ構成

| ファイル | 役割 |
|---------|------|
| `internal/notification/notification.go` | EmailSender interface, Config, LoadConfig, NewEmailSender, LogStartup |
| `internal/notification/log_sender.go` | ログ出力のみの開発用実装 |
| `internal/notification/sendgrid_sender.go` | SendGrid API 実装 (リトライ付き) |
| `internal/notification/templates.go` | `//go:embed` テンプレートレンダリング |
| `internal/notification/templates/welcome.html` | ウェルカムメール HTML テンプレート |

### 初期化パターン

`cmd/api/main.go` で featureflag 初期化の直後に呼び出し:

```go
notifCfg := notification.LoadConfig()
emailSender := notification.NewEmailSender(notifCfg)
notification.LogStartup(notifCfg)
```
