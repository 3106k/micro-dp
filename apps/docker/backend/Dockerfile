FROM golang:1.26 AS builder

WORKDIR /build
COPY apps/golang/backend/go.mod apps/golang/backend/go.sum ./
RUN go mod download

COPY apps/golang/backend/ .
RUN CGO_ENABLED=1 go build -o /api ./cmd/api && CGO_ENABLED=1 go build -o /worker ./cmd/worker

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends wget ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /api /usr/local/bin/api
COPY --from=builder /worker /usr/local/bin/worker
RUN mkdir -p /data/sqlite
