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
	"github.com/user/micro-dp/internal/featureflag"
	"github.com/user/micro-dp/internal/notification"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/queue"
	"github.com/user/micro-dp/storage"
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

	ffCfg := featureflag.LoadConfig()
	featureflag.Init(ffCfg)
	featureflag.LogStartup(ffCfg)

	notifCfg := notification.LoadConfig()
	emailSender := notification.NewEmailSender(notifCfg)
	notification.LogStartup(notifCfg)

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
	userIdentityRepo := db.NewUserIdentityRepo(sqlDB)
	tenantRepo := db.NewTenantRepo(sqlDB)
	jobRunRepo := db.NewJobRunRepo(sqlDB)
	jobRepo := db.NewJobRepo(sqlDB)
	jobVersionRepo := db.NewJobVersionRepo(sqlDB)
	jobModuleRepo := db.NewJobModuleRepo(sqlDB)
	jobModuleEdgeRepo := db.NewJobModuleEdgeRepo(sqlDB)
	moduleTypeRepo := db.NewModuleTypeRepo(sqlDB)
	moduleTypeSchemaRepo := db.NewModuleTypeSchemaRepo(sqlDB)
	connectionRepo := db.NewConnectionRepo(sqlDB)
	datasetRepo := db.NewDatasetRepo(sqlDB)
	uploadRepo := db.NewUploadRepo(sqlDB)
	adminAuditLogRepo := db.NewAdminAuditLogRepo(sqlDB)
	txManager := db.NewTxManager(sqlDB)

	jobRunModuleRepo := db.NewJobRunModuleRepo(sqlDB)
	jobRunArtifactRepo := db.NewJobRunArtifactRepo(sqlDB)

	planRepo := db.NewPlanRepo(sqlDB)
	tenantPlanRepo := db.NewTenantPlanRepo(sqlDB)
	usageRepo := db.NewUsageRepo(sqlDB)

	// Bootstrap superadmins
	bootstrapCfg := usecase.ParseBootstrapConfig(os.Getenv("BOOTSTRAP_SUPERADMINS"), os.Getenv("SUPERADMIN_EMAILS"))
	if err := usecase.BootstrapSuperadmins(context.Background(), userRepo, bootstrapCfg); err != nil {
		log.Fatalf("bootstrap superadmins: %v", err)
	}

	// Services
	authService := usecase.NewAuthService(
		userRepo,
		userIdentityRepo,
		tenantRepo,
		jwtSecret,
		emailSender,
		usecase.GoogleOAuthConfig{
			ClientID:            os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
			ClientSecret:        os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
			RedirectURL:         os.Getenv("GOOGLE_OAUTH_REDIRECT_URI"),
			PostLoginRedirect:   os.Getenv("GOOGLE_OAUTH_POST_LOGIN_REDIRECT_URI"),
			PostFailureRedirect: os.Getenv("GOOGLE_OAUTH_POST_FAILURE_REDIRECT_URI"),
		},
	)
	jobRunService := usecase.NewJobRunService(jobRunRepo, jobRepo)
	jobService := usecase.NewJobService(jobRepo, jobVersionRepo, jobModuleRepo, jobModuleEdgeRepo, moduleTypeSchemaRepo, txManager)
	moduleTypeService := usecase.NewModuleTypeService(moduleTypeRepo, moduleTypeSchemaRepo)
	connectionService := usecase.NewConnectionService(connectionRepo)
	datasetService := usecase.NewDatasetService(datasetRepo)
	eventService := usecase.NewEventService(eventQueue)
	eventMetrics := observability.NewEventMetrics()
	planService := usecase.NewPlanService(planRepo, tenantPlanRepo, usageRepo)

	minioPresignClient, err := storage.NewMinIOPresignClient()
	if err != nil {
		log.Fatalf("minio presign client: %v", err)
	}
	uploadQueue := queue.NewUploadQueue(valkeyClient)
	uploadService := usecase.NewUploadService(uploadRepo, minioPresignClient, uploadQueue)
	jobRunModuleService := usecase.NewJobRunModuleService(jobRunModuleRepo)
	jobRunArtifactService := usecase.NewJobRunArtifactService(jobRunArtifactRepo)
	adminTenantService := usecase.NewAdminTenantService(tenantRepo, adminAuditLogRepo)

	// Handlers
	healthH := handler.NewHealthHandler(sqlDB)
	authH := handler.NewAuthHandler(authService)
	jobRunH := handler.NewJobRunHandler(jobRunService)
	jobH := handler.NewJobHandler(jobService)
	moduleTypeH := handler.NewModuleTypeHandler(moduleTypeService)
	connectionH := handler.NewConnectionHandler(connectionService)
	datasetH := handler.NewDatasetHandler(datasetService)
	eventH := handler.NewEventHandler(eventService, planService, eventMetrics)
	uploadH := handler.NewUploadHandler(uploadService, planService)
	jobRunModuleH := handler.NewJobRunModuleHandler(jobRunModuleService)
	jobRunArtifactH := handler.NewJobRunArtifactHandler(jobRunArtifactService)
	adminTenantH := handler.NewAdminTenantHandler(adminTenantService)
	planH := handler.NewPlanHandler(planService)
	adminPlanH := handler.NewAdminPlanHandler(planService)

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
	mux.HandleFunc("GET /api/v1/auth/google/start", authH.GoogleStart)
	mux.HandleFunc("GET /api/v1/auth/google/callback", authH.GoogleCallback)

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

	// Job run modules
	mux.Handle("GET /api/v1/job_runs/{job_run_id}/modules", protected(jobRunModuleH.List))
	mux.Handle("GET /api/v1/job_runs/{job_run_id}/modules/{id}", protected(jobRunModuleH.Get))

	// Job run artifacts
	mux.Handle("GET /api/v1/job_runs/{job_run_id}/artifacts", protected(jobRunArtifactH.List))
	mux.Handle("GET /api/v1/job_runs/{job_run_id}/artifacts/{id}", protected(jobRunArtifactH.Get))

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

	// Datasets
	mux.Handle("GET /api/v1/datasets", protected(datasetH.List))
	mux.Handle("GET /api/v1/datasets/{id}", protected(datasetH.Get))

	// Uploads
	mux.Handle("POST /api/v1/uploads/presign", protected(uploadH.Presign))
	mux.Handle("POST /api/v1/uploads/{id}/complete", protected(uploadH.Complete))

	// Plan & Usage
	mux.Handle("GET /api/v1/plan", protected(planH.GetPlan))
	mux.Handle("GET /api/v1/usage/summary", protected(planH.GetUsageSummary))

	// Admin tenants
	mux.Handle("POST /api/v1/admin/tenants", adminProtected(adminTenantH.Create))
	mux.Handle("GET /api/v1/admin/tenants", adminProtected(adminTenantH.List))
	mux.Handle("PATCH /api/v1/admin/tenants/{id}", adminProtected(adminTenantH.Patch))

	// Admin plans
	mux.Handle("POST /api/v1/admin/plans", adminProtected(adminPlanH.Create))
	mux.Handle("GET /api/v1/admin/plans", adminProtected(adminPlanH.List))
	mux.Handle("PUT /api/v1/admin/plans/{id}", adminProtected(adminPlanH.Update))
	mux.Handle("POST /api/v1/admin/tenants/{tenant_id}/plan", adminProtected(adminPlanH.AssignPlan))

	addr := ":8080"
	log.Printf("api server starting on %s", addr)
	if err := http.ListenAndServe(addr, observability.WrapHTTPHandler(mux, "api-http")); err != nil {
		log.Fatal(err)
	}
}
