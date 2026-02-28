package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
	"github.com/user/micro-dp/internal/observability"
)

func main() {
	sqlDB, err := db.Open()
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	obsCfg := observability.LoadConfig("micro-dp-worker")
	obsShutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		log.Fatalf("observability init: %v", err)
	}
	defer observability.ShutdownWithTimeout(obsShutdown, 5*time.Second)
	observability.LogStartup(obsCfg)

	healthH := handler.NewHealthHandler(sqlDB)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthH.Healthz)
	mux.Handle("GET /metrics", observability.MetricsHandler())

	addr := ":8081"
	log.Printf("worker starting (healthcheck on %s)", addr)

	// TODO: start queue consumer goroutine
	// TODO: DuckDB processing
	// TODO: MinIO/Iceberg export

	fmt.Println("worker: waiting for jobs...")
	if err := http.ListenAndServe(addr, observability.WrapHTTPHandler(mux, "worker-http")); err != nil {
		log.Fatal(err)
	}
}
