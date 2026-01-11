package api

import (
	"github.com/gorilla/mux"

	"github.com/daveKoala/go-crt-filter/internal/scanner"
)

// SetupRoutes registers all application routes
func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/status", StatusHandler).Methods("GET")
	router.HandleFunc("/scan", scanner.ScanHandler).Methods("POST")
	router.HandleFunc("/test", TestHandler).Methods("POST")
}
