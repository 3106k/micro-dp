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
