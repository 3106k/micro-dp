package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
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
