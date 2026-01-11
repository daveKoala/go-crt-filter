package scanner

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daveKoala/go-crt-filter/internal/config"
)

// ScanHandler handles POST /scan requests to start CT log scanning
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest

	// Parse the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	cfg := config.Get()

	fmt.Println("Starting CT log scan...")
	fmt.Printf("Cut-off date: %s\n", req.CutOffDate)

	// Fetch STH (Signed Tree Head) from all configured CT logs
	results := RunSTHWorkers(cfg)

	// Log detailed summary to terminal
	LogSTHSummary(results)

	// Return summary in response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "STH fetch completed",
		"total_logs":    len(results),
		"cut_off_date":  req.CutOffDate,
		"results_count": len(results),
	})
}
