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

**SQLite via volume**: `sqlite-data` volume is shared between api and worker containers. Driver: modernc.org/sqlite (pure Go, no CGO).

**Go layered architecture**:

```
cmd/api/     — API entry point (auth, tenant management)
cmd/worker/  — Worker entry point (queue consumer, DuckDB, MinIO/Iceberg)
domain/      — entities, repository interfaces
usecase/     — application services / business logic
handler/     — HTTP handlers (adapter)
db/          — repository implementations, migrations
queue/       — Valkey queue implementation
worker/      — job processing
```

依存方向: `handler/` → `usecase/` → `domain/` ← `db/`, `queue/`。`domain/` は他パッケージに依存しない。

## Commands

All commands run from the repo root via Makefile:

```bash
make up          # docker compose up -d --build (production Dockerfiles)
make down        # stop all services
make build       # build images only
make logs        # stream all service logs
make ps          # show container status
make health      # curl healthz endpoints + valkey ping
make clean       # down + remove volumes and images
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

All host ports are configurable via `.env` (loaded by docker-compose from `apps/docker/`):

| Variable                  | Default    | Description             |
| ------------------------- | ---------- | ----------------------- |
| `COMPOSE_PROJECT_NAME`    | `micro-dp` | Container name prefix   |
| `API_HOST_PORT`           | `8080`     | Go API host port        |
| `WEB_HOST_PORT`           | `3000`     | Next.js host port       |
| `VALKEY_HOST_PORT`        | `6379`     | Valkey host port        |
| `MINIO_API_HOST_PORT`     | `9000`     | MinIO API host port     |
| `MINIO_CONSOLE_HOST_PORT` | `9001`     | MinIO console host port |

Internal container ports are fixed; only host-side ports change per environment.

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

| Method | Path | Auth | Description |
| ------ | ---- | ---- | ----------- |
| GET | `/healthz` | — | Health check |
| POST | `/api/v1/auth/register` | — | Register user (creates tenant) |
| POST | `/api/v1/auth/login` | — | Login (returns JWT) |
| GET | `/api/v1/auth/me` | Bearer | Current user + tenants |
| POST | `/api/v1/events` | Bearer + X-Tenant-ID | Ingest event (202 Accepted) |
| POST | `/api/v1/job_runs` | Bearer + X-Tenant-ID | Create job run |
| GET | `/api/v1/job_runs` | Bearer + X-Tenant-ID | List job runs (tenant-scoped) |
| GET | `/api/v1/job_runs/{id}` | Bearer + X-Tenant-ID | Get job run (tenant-scoped) |
| POST | `/api/v1/uploads/presign` | Bearer + X-Tenant-ID | Request presigned upload URLs |
| POST | `/api/v1/uploads/{id}/complete` | Bearer + X-Tenant-ID | Mark upload complete |

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
