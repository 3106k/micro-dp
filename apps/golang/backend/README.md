# micro-dp backend development (air hot reload)

## Local development

Install `air` first:

```bash
go install github.com/air-verse/air@latest
```

Run API mode hot reload:

```bash
make dev-api
```

Run Worker mode hot reload:

```bash
make dev-worker
```

Both commands use `apps/golang/backend/.air.toml` and pass `--mode` to the backend binary.

## Docker development

Use the normal development stack:

```bash
make up
```

In development, `apps/docker/backend/Dockerfile.dev` starts:

```bash
air -c .air.toml -- --mode=<api|worker>
```

`docker-compose.override.yaml` sets:

- API container: `--mode=api`
- Worker container: `--mode=worker`

## Verification

1. Start API (`make dev-api`) or stack (`make up`)
2. Edit a Go file in `apps/golang/backend` (for example `handler/health.go`)
3. Confirm air logs show rebuild/restart
4. Confirm health endpoints:
   - API: `http://localhost:8080/healthz`
   - Worker: `http://localhost:8081/healthz`
