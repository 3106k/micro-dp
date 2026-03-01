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
	"github.com/user/micro-dp/queue"
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

	// Valkey
	valkeyClient, err := queue.NewValkeyClient()
	if err != nil {
		log.Fatalf("valkey connect: %v", err)
	}
	defer valkeyClient.Close()
	eventQueue := queue.NewEventQueue(valkeyClient)

	// Repositories
	userRepo := db.NewUserRepo(sqlDB)
	tenantRepo := db.NewTenantRepo(sqlDB)
	jobRunRepo := db.NewJobRunRepo(sqlDB)
	jobRepo := db.NewJobRepo(sqlDB)
	jobVersionRepo := db.NewJobVersionRepo(sqlDB)
	jobModuleRepo := db.NewJobModuleRepo(sqlDB)
	jobModuleEdgeRepo := db.NewJobModuleEdgeRepo(sqlDB)
	moduleTypeRepo := db.NewModuleTypeRepo(sqlDB)
	moduleTypeSchemaRepo := db.NewModuleTypeSchemaRepo(sqlDB)
	connectionRepo := db.NewConnectionRepo(sqlDB)
	adminAuditLogRepo := db.NewAdminAuditLogRepo(sqlDB)
	txManager := db.NewTxManager(sqlDB)

	// Bootstrap superadmins
	bootstrapCfg := usecase.ParseBootstrapConfig(os.Getenv("BOOTSTRAP_SUPERADMINS"), os.Getenv("SUPERADMIN_EMAILS"))
	if err := usecase.BootstrapSuperadmins(context.Background(), userRepo, bootstrapCfg); err != nil {
		log.Fatalf("bootstrap superadmins: %v", err)
	}

	// Services
	authService := usecase.NewAuthService(userRepo, tenantRepo, jwtSecret)
	jobRunService := usecase.NewJobRunService(jobRunRepo, jobRepo)
	jobService := usecase.NewJobService(jobRepo, jobVersionRepo, jobModuleRepo, jobModuleEdgeRepo, txManager)
	moduleTypeService := usecase.NewModuleTypeService(moduleTypeRepo, moduleTypeSchemaRepo)
	connectionService := usecase.NewConnectionService(connectionRepo)
	eventService := usecase.NewEventService(eventQueue)
	eventMetrics := observability.NewEventMetrics()
	adminTenantService := usecase.NewAdminTenantService(tenantRepo, adminAuditLogRepo)

	// Handlers
	healthH := handler.NewHealthHandler(sqlDB)
	authH := handler.NewAuthHandler(authService)
	jobRunH := handler.NewJobRunHandler(jobRunService)
	jobH := handler.NewJobHandler(jobService)
	moduleTypeH := handler.NewModuleTypeHandler(moduleTypeService)
	connectionH := handler.NewConnectionHandler(connectionService)
	eventH := handler.NewEventHandler(eventService, eventMetrics)
	adminTenantH := handler.NewAdminTenantHandler(adminTenantService)

	// Middleware
	authMW := handler.AuthMiddleware(jwtSecret)
	tenantMW := handler.TenantMiddleware(tenantRepo)
	superadminMW := handler.SuperadminMiddleware(userRepo)

	protected := func(h http.HandlerFunc) http.Handler {
		return authMW(tenantMW(http.HandlerFunc(h)))
	}
	adminProtected := func(h http.HandlerFunc) http.Handler {
		return authMW(superadminMW(http.HandlerFunc(h)))
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

	// Events
	mux.Handle("POST /api/v1/events", protected(eventH.Ingest))
	mux.Handle("GET /api/v1/events/summary", protected(eventH.Summary))

	// Job runs
	mux.Handle("POST /api/v1/job_runs", protected(jobRunH.Create))
	mux.Handle("GET /api/v1/job_runs", protected(jobRunH.List))
	mux.Handle("GET /api/v1/job_runs/{id}", protected(jobRunH.Get))

	// Jobs
	mux.Handle("POST /api/v1/jobs", protected(jobH.Create))
	mux.Handle("GET /api/v1/jobs", protected(jobH.List))
	mux.Handle("GET /api/v1/jobs/{id}", protected(jobH.Get))
	mux.Handle("PUT /api/v1/jobs/{id}", protected(jobH.Update))

	// Job versions
	mux.Handle("POST /api/v1/jobs/{job_id}/versions", protected(jobH.CreateVersion))
	mux.Handle("GET /api/v1/jobs/{job_id}/versions", protected(jobH.ListVersions))
	mux.Handle("GET /api/v1/jobs/{job_id}/versions/{version_id}", protected(jobH.GetVersionDetail))
	mux.Handle("POST /api/v1/jobs/{job_id}/versions/{version_id}/publish", protected(jobH.PublishVersion))

	// Module types
	mux.Handle("POST /api/v1/module_types", protected(moduleTypeH.Create))
	mux.Handle("GET /api/v1/module_types", protected(moduleTypeH.List))
	mux.Handle("GET /api/v1/module_types/{id}", protected(moduleTypeH.Get))
	mux.Handle("POST /api/v1/module_types/{id}/schemas", protected(moduleTypeH.CreateSchema))
	mux.Handle("GET /api/v1/module_types/{id}/schemas", protected(moduleTypeH.ListSchemas))

	// Connections
	mux.Handle("POST /api/v1/connections", protected(connectionH.Create))
	mux.Handle("GET /api/v1/connections", protected(connectionH.List))
	mux.Handle("GET /api/v1/connections/{id}", protected(connectionH.Get))
	mux.Handle("PUT /api/v1/connections/{id}", protected(connectionH.Update))
	mux.Handle("DELETE /api/v1/connections/{id}", protected(connectionH.Delete))

	// Admin tenants
	mux.Handle("POST /api/v1/admin/tenants", adminProtected(adminTenantH.Create))
	mux.Handle("GET /api/v1/admin/tenants", adminProtected(adminTenantH.List))
	mux.Handle("PATCH /api/v1/admin/tenants/{id}", adminProtected(adminTenantH.Patch))

	addr := ":8080"
	log.Printf("api server starting on %s", addr)
	if err := http.ListenAndServe(addr, observability.WrapHTTPHandler(mux, "api-http")); err != nil {
		log.Fatal(err)
	}
}
