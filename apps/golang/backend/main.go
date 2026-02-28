package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/user/micro-dp/handler"
)

func main() {
	mode := flag.String("mode", "api", "Run mode: api or worker")
	flag.Parse()

	switch *mode {
	case "api":
		runAPI()
	case "worker":
		runWorker()
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

func runAPI() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Healthz)

	addr := ":8080"
	log.Printf("api server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func runWorker() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Healthz)

	addr := ":8081"
	log.Printf("worker starting (healthcheck on %s)", addr)

	// TODO: start queue consumer goroutine

	fmt.Println("worker: waiting for jobs...")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
