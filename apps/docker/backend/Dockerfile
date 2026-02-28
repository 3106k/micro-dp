FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY apps/golang/backend/go.mod apps/golang/backend/go.sum ./
RUN go mod download

COPY apps/golang/backend/ .
RUN go build -o /api ./cmd/api && go build -o /worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache wget
COPY --from=builder /api /usr/local/bin/api
COPY --from=builder /worker /usr/local/bin/worker
RUN mkdir -p /data/sqlite
