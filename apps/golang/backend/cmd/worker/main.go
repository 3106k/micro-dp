package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
	"github.com/user/micro-dp/internal/connector"
	"github.com/user/micro-dp/internal/connector/executors"
	"github.com/user/micro-dp/internal/credential"
	"github.com/user/micro-dp/internal/featureflag"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/queue"
	"github.com/user/micro-dp/storage"
	"github.com/user/micro-dp/usecase"
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

	// Metering
	usageRepo := db.NewUsageRepo(sqlDB)
	meteringService := usecase.NewMeteringService(usageRepo)

	// Aggregation queue (shared: EventConsumer enqueues, AggregationConsumer dequeues)
	aggregationQueue := queue.NewAggregationQueue(valkeyClient)

	// Event consumer
	eventMetrics := observability.NewEventMetrics()
	parquetWriter := worker.NewParquetWriter(minioClient)
	consumer := worker.NewEventConsumer(eventQueue, parquetWriter, eventMetrics, meteringService, aggregationQueue)

	go consumer.Run(ctx)

	// Upload consumer (CSV→Parquet)
	datasetRepo := db.NewDatasetRepo(sqlDB)
	uploadQueue := queue.NewUploadQueue(valkeyClient)
	uploadMetrics := observability.NewUploadMetrics()
	csvImportWriter := worker.NewCSVImportWriter(minioClient, datasetRepo)
	uploadConsumer := worker.NewUploadConsumer(uploadQueue, csvImportWriter, uploadMetrics, meteringService)

	go uploadConsumer.Run(ctx)

	// Transform consumer (SQL→Parquet)
	jobRunRepo := db.NewJobRunRepo(sqlDB)
	transformQueue := queue.NewTransformQueue(valkeyClient)
	transformMetrics := observability.NewTransformMetrics()
	transformWriter := worker.NewTransformWriter(minioClient, datasetRepo)
	transformConsumer := worker.NewTransformConsumer(
		transformQueue, transformWriter, transformMetrics, meteringService, jobRunRepo,
	)

	go transformConsumer.Run(ctx)

	// Credential + Connection (for import jobs)
	credentialRepo := db.NewCredentialRepo(sqlDB)
	connectionRepo := db.NewConnectionRepo(sqlDB)
	googleCredProvider := credential.NewGoogleProvider(credential.GoogleConfig{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	})
	credentialService := usecase.NewCredentialService(
		credentialRepo,
		[]credential.OAuthProvider{googleCredProvider},
		os.Getenv("JWT_SECRET"),
	)

	// Connector registry with import executors
	connectorRegistry := connector.Global()
	sheetsImportWriter := worker.NewSheetsImportWriter(minioClient, datasetRepo)
	connectorRegistry.RegisterExecutor("source-google-sheets",
		executors.NewGoogleSheetsExecutor(sheetsImportWriter))

	// Job Run poller + consumer (generic job execution)
	jobRunQueue := queue.NewJobRunQueue(valkeyClient)
	jobRunMetrics := observability.NewJobRunMetrics()
	jobRunPoller := worker.NewJobRunPoller(jobRunRepo, jobRunQueue, jobRunMetrics, 5*time.Second)
	jobRunConsumer := worker.NewJobRunConsumer(
		jobRunQueue, jobRunRepo, transformWriter,
		connectorRegistry, credentialService, connectionRepo,
		jobRunMetrics, meteringService,
	)

	go jobRunPoller.Run(ctx)
	go jobRunConsumer.Run(ctx)

	// Aggregation consumer (raw → events/visits)
	aggregationWriter := worker.NewAggregationWriter(minioClient)
	aggregationMetrics := observability.NewAggregationMetrics()
	aggregationConsumer := worker.NewAggregationConsumer(aggregationQueue, aggregationWriter, aggregationMetrics)

	go aggregationConsumer.Run(ctx)

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
