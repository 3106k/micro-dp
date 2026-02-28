package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
	"github.com/user/micro-dp/usecase"
)

func main() {
	mode := flag.String("mode", "api", "Run mode: api, worker, or migrate")
	flag.Parse()

	sqlDB, err := db.Open()
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	if *mode == "migrate" {
		log.Println("migration complete, exiting")
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	userRepo := db.NewUserRepo(sqlDB)
	tenantRepo := db.NewTenantRepo(sqlDB)
	authService := usecase.NewAuthService(userRepo, tenantRepo, jwtSecret)

	healthH := handler.NewHealthHandler(sqlDB)
	authH := handler.NewAuthHandler(authService)
	authMW := handler.AuthMiddleware(jwtSecret)

	switch *mode {
	case "api":
		runAPI(healthH, authH, authMW)
	case "worker":
		runWorker(healthH)
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

func runAPI(healthH *handler.HealthHandler, authH *handler.AuthHandler, authMW func(http.Handler) http.Handler) {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /healthz", healthH.Healthz)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)

	// Authenticated routes
	mux.Handle("GET /api/v1/auth/me", authMW(http.HandlerFunc(authH.Me)))

	addr := ":8080"
	log.Printf("api server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func runWorker(healthH *handler.HealthHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthH.Healthz)

	addr := ":8081"
	log.Printf("worker starting (healthcheck on %s)", addr)

	// TODO: start queue consumer goroutine

	fmt.Println("worker: waiting for jobs...")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
