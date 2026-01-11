package scanner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/daveKoala/go-crt-filter/internal/config"
)

// ScanHandler handles POST /scan requests to start CT log scanning
//
// This handler orchestrates a three-phase process:
//  1. Fetch STH (Signed Tree Head) from all configured CT logs to get current tree sizes
//  2. Generate batch jobs working backwards from tree size in configurable windows
//  3. Process batches concurrently using a worker pool
//
// The scan works backwards from the most recent entries to find certificates
// matching the specified cutoff date.
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest

	// Parse the JSON request body containing scan parameters
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Load CT log configuration (Google, Cloudflare, DigiCert, Let's Encrypt, Sectigo)
	cfg := config.Get()

	fmt.Println("========================================")
	fmt.Println("Starting CT log scan...")
	fmt.Printf("Cut-off date: %s\n", req.CutOffDate)
	fmt.Println("========================================")

	// ===================================================================
	// Phase 1: Fetch STH (Signed Tree Head) from all configured CT logs
	// ===================================================================
	// The STH tells us the current size of each log's Merkle tree, which
	// represents the total number of certificates in the log. We need this
	// to know where to start scanning backwards from.
	fmt.Println("\n[Phase 1] Fetching STH from all CT logs...")
	sthResults := RunSTHWorkers(cfg)
	LogSTHSummary(sthResults)

	// Parse cutoff date (TODO: proper parsing)
	// cutOffDate, _ := time.Parse(time.RFC3339, req.CutOffDate)

	// ===================================================================
	// Phase 2: Generate batch jobs working backwards from tree size
	// ===================================================================
	// We divide the scan into smaller batches (windows) and work backwards
	// from the most recent entries. Jobs are interleaved across logs to
	// ensure all logs are processed in parallel.
	windowSize := int64(1000)      // Number of entries to fetch per batch (e.g., entries 1000-1999)
	maxBatchesPerLog := int64(5)   // Limit batches per log for testing (prevents billions of jobs)
	maxWorkers := 10               // Maximum number of concurrent worker goroutines

	fmt.Printf("\n[Phase 2] Generating batch jobs with window size: %d, max batches per log: %d\n", windowSize, maxBatchesPerLog)

	// GenerateBatchJobs creates interleaved jobs from all logs
	// Example: [log1_batch1, log2_batch1, log3_batch1, log1_batch2, log2_batch2, ...]
	jobs := GenerateBatchJobs(sthResults, windowSize, maxBatchesPerLog, time.Now())

	fmt.Printf("Generated %d batch jobs across %d logs\n", len(jobs), len(sthResults))
	fmt.Printf("Starting %d concurrent workers...\n\n", maxWorkers)

	// ===================================================================
	// Phase 3: Process batches with worker pool
	// ===================================================================
	// Workers fetch batches concurrently and process the certificate entries.
	// When saveToFile is true, responses are saved to testdata/responses/
	// for offline testing and development of certificate parsing logic.
	saveToFile := true // Enable saving responses for testing
	fmt.Println("💾 Saving responses to testdata/responses/ for testing...")
	RunBatchWorkers(jobs, maxWorkers, saveToFile)

	fmt.Println("\n========================================")
	fmt.Println("Scan completed!")
	fmt.Println("========================================")

	// Return summary JSON response to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Scan completed",
		"total_logs":    len(sthResults),
		"total_batches": len(jobs),
		"window_size":   windowSize,
		"max_workers":   maxWorkers,
		"cut_off_date":  req.CutOffDate,
	})
}
