package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/usecase"
)

func main() {
	migrate := flag.Bool("migrate", false, "Run migrations and exit")
	flag.Parse()

	sqlDB, err := db.Open()
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	if *migrate {
		log.Println("migration complete, exiting")
		return
	}

	obsCfg := observability.LoadConfig("micro-dp-api")
	obsShutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		log.Fatalf("observability init: %v", err)
	}
	defer observability.ShutdownWithTimeout(obsShutdown, 5*time.Second)
	observability.LogStartup(obsCfg)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	userRepo := db.NewUserRepo(sqlDB)
	tenantRepo := db.NewTenantRepo(sqlDB)
	jobRunRepo := db.NewJobRunRepo(sqlDB)
	authService := usecase.NewAuthService(userRepo, tenantRepo, jwtSecret)
	jobRunService := usecase.NewJobRunService(jobRunRepo)

	healthH := handler.NewHealthHandler(sqlDB)
	authH := handler.NewAuthHandler(authService)
	jobRunH := handler.NewJobRunHandler(jobRunService)
	authMW := handler.AuthMiddleware(jwtSecret)
	tenantMW := handler.TenantMiddleware(tenantRepo)

	protected := func(h http.HandlerFunc) http.Handler {
		return authMW(tenantMW(http.HandlerFunc(h)))
	}

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /healthz", healthH.Healthz)
	mux.Handle("GET /metrics", observability.MetricsHandler())
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)

	// Authenticated routes
	mux.Handle("GET /api/v1/auth/me", authMW(http.HandlerFunc(authH.Me)))

	// Authenticated + tenant-scoped routes
	mux.Handle("POST /api/v1/job_runs", protected(jobRunH.Create))
	mux.Handle("GET /api/v1/job_runs", protected(jobRunH.List))
	mux.Handle("GET /api/v1/job_runs/{id}", protected(jobRunH.Get))

	addr := ":8080"
	log.Printf("api server starting on %s", addr)
	if err := http.ListenAndServe(addr, observability.WrapHTTPHandler(mux, "api-http")); err != nil {
		log.Fatal(err)
	}
}
