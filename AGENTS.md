# AGENTS.md

## Contract-First API Rules

- `spec/openapi/v1.yaml` is the single source of truth (SSOT) for API contracts.
- Do not update API request/response contracts directly in frontend/backend code first.
- Update the OpenAPI spec first, then generate code.

## Required Flow for API Contract Changes

1. Edit `spec/openapi/v1.yaml`
2. Run:
   - `make openapi-lint`
   - `make openapi-generate`
   - or `make openapi-check` (includes drift check)
3. Commit spec and generated artifacts in the same commit.

## Generated Artifacts

- Frontend:
  - `apps/node/web/src/lib/api/generated.ts`
- Backend:
  - `apps/golang/backend/internal/openapi/types.gen.go`
  - `apps/golang/backend/internal/openapi/server.gen.go`

## Codegen Policy

- Generated files (`*.gen.go`, `generated.ts`) must not be hand-edited.
- If generated output changes, regenerate from `spec/openapi/v1.yaml`.

## CI Expectations

- `openapi-contract` workflow must pass for contract-related PRs.
- CI checks:
  - OpenAPI lint
  - OpenAPI generation
  - drift detection via `git diff --exit-code`

## Backend Image CI

- On changes under `apps/golang/backend/**`, backend image build/push CI is expected to run.
- Workflow: `.github/workflows/backend-image.yml`
- Registry target: `ghcr.io/<owner>/<repo>/backend`
- Tag strategy:
  - `sha-<commit>`
  - branch tag (for branch pushes)
  - `latest` on default branch
  - `staging` on `staging` branch

## Observability Verification Notes

- Traces require reachable OTLP endpoint (`OTEL_EXPORTER_OTLP_ENDPOINT`).
- If observability stack and app stack are on different Docker networks/projects, traces may not be exported.
- For local verification, confirm both:
  - Jaeger has `micro-dp-api` / `micro-dp-worker` services and traces
  - Grafana Prometheus datasource returns query results (for example `up`)

## SDK Tracker Change Checklist

- When changing `apps/node/sdk-tracker`, run:
  - `make sdk-tracker-build`
  - `make sdk-tracker-test`
- Do not commit generated `dist/` or `node_modules/` for `sdk-tracker`.

## Events Ingest Change Checklist

- Worker 側の変更は `worker/` パッケージ内で行う
- Valkey キー名を変更する場合は API / Worker 両方の整合性を確認する
- DuckDB は CGO 必須 — Dockerfile が Debian ベースであることを維持する
- `CGO_ENABLED=1 go build ./...` でビルド確認する
- E2E テスト `events` スイートが通ることを確認する (`make e2e-cli`)

## Issue Closing Flow

- Standard order:
  1. Commit changes
  2. Push branch
  3. Close related GitHub issue with commit reference
