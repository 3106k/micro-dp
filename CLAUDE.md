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
| POST | `/api/v1/job_runs` | Bearer + X-Tenant-ID | Create job run |
| GET | `/api/v1/job_runs` | Bearer + X-Tenant-ID | List job runs (tenant-scoped) |
| GET | `/api/v1/job_runs/{id}` | Bearer + X-Tenant-ID | Get job run (tenant-scoped) |

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
