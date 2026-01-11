package main

import (
	"log"
	"net/http"

	"github.com/daveKoala/go-crt-filter/internal/api"
	"github.com/daveKoala/go-crt-filter/internal/config"
	"github.com/daveKoala/go-crt-filter/internal/middleware"
)

func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use Go own router
	mux := http.NewServeMux()

	// Register all routes
	api.SetupRoutes(mux)

	// Apply middleware chain
	// Wrapping order: Logging wraps CustomHeaders wraps CORS wraps mux
	// Execution order: Logging → CustomHeaders → CORS → Router → Handler
	handler := middleware.Logging(
		middleware.CustomHeaders(
			middleware.CORS(mux),
		),
	)

	log.Println("🚀 Server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
