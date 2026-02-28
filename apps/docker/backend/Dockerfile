FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY apps/golang/backend/go.mod apps/golang/backend/go.sum ./
RUN go mod download

COPY apps/golang/backend/ .
RUN go build -o /backend .

FROM alpine:3.20
RUN apk add --no-cache wget
COPY --from=builder /backend /usr/local/bin/backend
RUN mkdir -p /data/sqlite
ENTRYPOINT ["backend"]
CMD ["--mode=api"]
