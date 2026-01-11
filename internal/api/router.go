package api

import (
	"net/http"

	"github.com/daveKoala/go-crt-filter/internal/scanner"
)

// SetupRoutes registers all application routes using Go 1.22+ enhanced routing
func SetupRoutes(mux *http.ServeMux) {
	// Method-specific routing built into Go 1.22+
	mux.HandleFunc("GET /status", StatusHandler)
	mux.HandleFunc("POST /scan", scanner.ScanHandler)
	mux.HandleFunc("POST /test", TestHandler)
}
