package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/daveKoala/go-crt-filter/internal/api"
	"github.com/daveKoala/go-crt-filter/internal/config"
)

func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatal("Failed to load config: %v", err)
	}
	router := mux.NewRouter()

	// Register all routes
	api.SetupRoutes(router)

	log.Println("🚀 Server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
