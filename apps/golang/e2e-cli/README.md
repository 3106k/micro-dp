# micro-dp e2e-cli

API E2E verification CLI for `micro-dp` (API First workflow).

## Scope (initial)

- `health` suite: `GET /healthz`
- `auth` suite: register/login/me happy path
- `job_runs` suite: create/list/get happy path
- `tenant` suite:
  - cross-tenant forbidden
  - admin multi-tenant creation and membership check (skipped when admin credentials are not provided)

## Prerequisites

- API is running (for local Docker: `make up`)
- `JWT_SECRET` is configured for backend
- Go 1.26+

## Run

```bash
make e2e-cli
```

Direct execution:

```bash
cd apps/golang/e2e-cli
go run ./cmd/run --base-url=http://localhost:8080
```

## Suite selection

Run only specific suites:

```bash
go run ./cmd/run --suites=health,auth
```

You can also use env vars:

```bash
E2E_BASE_URL=http://localhost:8080 E2E_SUITES=health,auth go run ./cmd/run
```

## Auth suite settings

Optional inputs:

- `--auth-email` / `E2E_AUTH_EMAIL`
- `--auth-password` / `E2E_AUTH_PASSWORD` (default: `Passw0rd!123`)
- `--display-name` / `E2E_DISPLAY_NAME` (default: `E2E User`)

Tenant admin scenario inputs:

- `--admin-email` / `E2E_ADMIN_EMAIL`
- `--admin-password` / `E2E_ADMIN_PASSWORD`

If `auth-email` is not set, a unique email is generated per run.

## Output

- Console report (pass/fail/skip with scenario id)
- JSON report (default: `e2e-report.json`)

Override json path:

```bash
go run ./cmd/run --json-out=/tmp/e2e-report.json
```

## CI template flow

Use `Makefile` template target:

```bash
make e2e-ci-template
```

This runs:

1. `make up`
2. `make e2e-cli`
3. `make down`
