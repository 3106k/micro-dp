package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
	"github.com/user/micro-dp/internal/featureflag"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/queue"
	"github.com/user/micro-dp/storage"
	"github.com/user/micro-dp/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	sqlDB, err := db.Open()
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	obsCfg := observability.LoadConfig("micro-dp-worker")
	obsShutdown, err := observability.Init(ctx, obsCfg)
	if err != nil {
		log.Fatalf("observability init: %v", err)
	}
	defer observability.ShutdownWithTimeout(obsShutdown, 5*time.Second)
	observability.LogStartup(obsCfg)

	ffCfg := featureflag.LoadConfig()
	featureflag.Init(ffCfg)
	featureflag.LogStartup(ffCfg)

	// Valkey
	valkeyClient, err := queue.NewValkeyClient()
	if err != nil {
		log.Fatalf("valkey connect: %v", err)
	}
	defer valkeyClient.Close()
	eventQueue := queue.NewEventQueue(valkeyClient)

	// MinIO
	minioClient, err := storage.NewMinIOClient()
	if err != nil {
		log.Fatalf("minio connect: %v", err)
	}

	// Event consumer
	eventMetrics := observability.NewEventMetrics()
	parquetWriter := worker.NewParquetWriter(minioClient)
	consumer := worker.NewEventConsumer(eventQueue, parquetWriter, eventMetrics)

	go consumer.Run(ctx)

	// Health check server
	healthH := handler.NewHealthHandler(sqlDB)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthH.Healthz)
	mux.Handle("GET /metrics", observability.MetricsHandler())

	addr := ":8081"
	log.Printf("worker starting (healthcheck on %s)", addr)

	srv := &http.Server{Addr: addr, Handler: observability.WrapHTTPHandler(mux, "worker-http")}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("worker shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	log.Println("worker stopped")
}
