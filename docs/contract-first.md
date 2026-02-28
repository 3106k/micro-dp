# Contract-First Development Flow

`spec/openapi/v1.yaml` is the single source of truth (SSOT) for API contracts.

## Workflow

1. Update `spec/openapi/v1.yaml`
2. Run lint and generation
   ```bash
   make openapi-check
   ```
3. Implement backend/frontend based on generated artifacts
4. Commit spec and generated files together

## Commands

- Lint: `make openapi-lint`
- Bundle: `make openapi-bundle`
- Generate frontend types: `make openapi-generate-fe`
- Generate backend types/interfaces: `make openapi-generate-be`
- Generate all: `make openapi-generate`

## Generated Artifacts

- Frontend:
  - `apps/node/web/src/lib/api/generated.ts`
- Backend:
  - `apps/golang/backend/internal/openapi/types.gen.go`
  - `apps/golang/backend/internal/openapi/server.gen.go`

## CI Rule

CI runs:

1. OpenAPI lint
2. OpenAPI generation
3. `git diff --exit-code` for generated drift detection
