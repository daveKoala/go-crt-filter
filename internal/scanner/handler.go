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

	// Example: Show first Google log
	fmt.Println(cfg.Google.Logs[0].ID)

	// TODO: Start scanner workers
	// TODO: Stream results back

	// Echo back the request as JSON response (placeholder)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}
