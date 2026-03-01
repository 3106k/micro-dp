package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/usecase"
)

func main() {
	email := flag.String("email", "", "superadmin email (required)")
	password := flag.String("password", "", "superadmin password (required)")
	displayName := flag.String("display-name", "Super Admin", "display name")
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("--email and --password are required")
	}

	sqlDB, err := db.Open()
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	userRepo := db.NewUserRepo(sqlDB)
	tenantRepo := db.NewTenantRepo(sqlDB)
	authService := usecase.NewAuthService(userRepo, tenantRepo, "dummy-secret-not-used")

	ctx := context.Background()
	user, err := userRepo.FindByEmail(ctx, *email)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			log.Fatalf("find user by email: %v", err)
		}

		userID, _, err := authService.Register(ctx, *email, *password, *displayName)
		if err != nil {
			log.Fatalf("register superadmin: %v", err)
		}
		if err := userRepo.UpdatePlatformRole(ctx, userID, domain.PlatformRoleSuperadmin); err != nil {
			log.Fatalf("set superadmin: %v", err)
		}
		fmt.Printf("superadmin created: %s\n", userID)
		return
	}

	if err := userRepo.UpdatePlatformRole(ctx, user.ID, domain.PlatformRoleSuperadmin); err != nil {
		log.Fatalf("set superadmin: %v", err)
	}
	fmt.Printf("superadmin updated: %s\n", user.ID)
}
