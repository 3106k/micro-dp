package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/user/micro-dp/db"
	"github.com/user/micro-dp/handler"
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

	healthH := handler.NewHealthHandler(sqlDB)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthH.Healthz)

	addr := ":8081"
	log.Printf("worker starting (healthcheck on %s)", addr)

	// TODO: start queue consumer goroutine
	// TODO: DuckDB processing
	// TODO: MinIO/Iceberg export

	fmt.Println("worker: waiting for jobs...")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
