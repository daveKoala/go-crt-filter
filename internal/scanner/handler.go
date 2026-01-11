package scanner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/daveKoala/go-crt-filter/internal/config"
)

// ScanHandler handles POST /scan requests to start CT log scanning
//
// Simple design:
//  - One worker per CT log (16 workers total)
//  - Each worker independently fetches its STH and works backwards
//  - Workers save responses to testdata/responses/ for offline testing
//  - Stub decode/extract functions return empty for now
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest

	// Parse the JSON request body containing scan parameters
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Load CT log configuration
	cfg := config.Get()
	allLogs := cfg.AllLogs()

	fmt.Println("========================================")
	fmt.Println("Starting CT log scan...")
	fmt.Printf("Cut-off date: %s\n", req.CutOffDate)
	fmt.Printf("Total logs to scan: %d\n", len(allLogs))
	fmt.Println("========================================")

	// Parse cutoff date
	// TODO: Proper date parsing and validation
	cutOffDate := time.Now().AddDate(0, -1, 0) // Default: 1 month ago

	// Shared results channel
	resultsChan := make(chan Result, 1000)

	// WaitGroup to track all workers
	var wg sync.WaitGroup

	// Start one worker per log
	fmt.Println("\nLaunching workers (one per log)...")
	for _, logEntry := range allLogs {
		wg.Add(1)
		go LogWorker(
			logEntry.Provider,
			logEntry.BaseURL,
			logEntry.Log,
			cutOffDate,
			resultsChan,
			true, // saveToFile - save responses for testing
			&wg,
		)
	}

	// Wait for all workers in a goroutine and close results channel when done
	go func() {
		wg.Wait()
		close(resultsChan)
		fmt.Println("\n========================================")
		fmt.Println("All workers completed!")
		fmt.Println("========================================")
	}()

	// Collect results (in production, you might stream these)
	var allResults []Result
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	fmt.Printf("Total results collected: %d\n", len(allResults))

	// Return summary JSON response to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Scan completed",
		"total_logs":    len(allLogs),
		"total_results": len(allResults),
		"cut_off_date":  req.CutOffDate,
	})
}
