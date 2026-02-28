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

Commands use `apps/golang/backend/.air.api.toml` and `.air.worker.toml`.

## Docker development

Use the normal development stack:

```bash
make up
```

In development, `apps/docker/backend/Dockerfile.dev` starts:

```bash
air -c .air.<api|worker>.toml
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

## Observability (OpenTelemetry + Prometheus)

### Environment variables

- `OTEL_TRACES_ENABLED` (default: `true`)
- `OTEL_METRICS_ENABLED` (default: `true`)
- `OTEL_EXPORTER_OTLP_ENDPOINT` (default: empty)
- `OTEL_EXPORTER_OTLP_INSECURE` (default: `true`)
- `OTEL_SERVICE_NAME` (set by compose as `micro-dp-api` / `micro-dp-worker`)

Default behavior:

- If `OTEL_EXPORTER_OTLP_ENDPOINT` is empty, trace export is disabled.
- Metrics endpoint (`/metrics`) remains available.

### Local connection example

If you run an external local observability stack (Collector/Prometheus/Grafana):

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 make up
```

### Verification steps

1. Start services with `make up` (or run API/Worker directly)
2. Access metrics:
   - API: `http://localhost:8080/metrics`
   - Worker: `http://localhost:8081/metrics`
3. Generate API traffic (for traces):
   - `curl http://localhost:8080/healthz`
   - `curl -X POST http://localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"x@example.com","password":"x"}'`
4. Confirm in your observability stack:
   - spans are received for HTTP requests
   - Prometheus can scrape `/metrics`
