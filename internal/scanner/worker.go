package scanner

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/daveKoala/go-crt-filter/internal/config"
)

// FetchSTH fetches the Signed Tree Head from a CT log
func FetchSTH(baseURL, logID string) (*STHResponse, error) {
	url := fmt.Sprintf("%s/%s/ct/v1/get-sth", baseURL, logID)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STH: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var sth STHResponse
	if err := json.NewDecoder(resp.Body).Decode(&sth); err != nil {
		return nil, fmt.Errorf("failed to decode STH response: %w", err)
	}

	return &sth, nil
}

// RunSTHWorkers fetches STH data from all configured CT logs concurrently
func RunSTHWorkers(cfg *config.Config) []STHResult {
	allLogs := cfg.AllLogs()
	results := make([]STHResult, len(allLogs))
	var wg sync.WaitGroup

	// Create a worker for each log
	for i, logEntry := range allLogs {
		wg.Add(1)
		go func(idx int, provider, baseURL string, ctLog config.CTLog) {
			defer wg.Done()

			log.Printf("[Worker %d] Fetching STH for %s - %s", idx, provider, ctLog.ID)

			sth, err := FetchSTH(baseURL, ctLog.ID)

			results[idx] = STHResult{
				Provider:    provider,
				LogID:       ctLog.ID,
				BaseURL:     baseURL,
				Description: ctLog.Description,
				STH:         sth,
				Error:       err,
			}

			if err != nil {
				log.Printf("[Worker %d] ERROR: %s - %s: %v", idx, provider, ctLog.ID, err)
			} else {
				log.Printf("[Worker %d] SUCCESS: %s - %s | TreeSize: %d | Timestamp: %s",
					idx, provider, ctLog.ID, sth.TreeSize, time.Unix(sth.Timestamp/1000, 0).Format(time.RFC3339))
			}
		}(i, logEntry.Provider, logEntry.BaseURL, logEntry.Log)
	}

	// Wait for all workers to complete
	wg.Wait()

	return results
}

// LogSTHSummary prints a summary table of all STH results to the terminal
func LogSTHSummary(results []STHResult) {
	log.Println("========================================")
	log.Println("CT Log STH Summary")
	log.Println("========================================")

	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Error != nil {
			failureCount++
			log.Printf("❌ [%s] %s - FAILED: %v", result.Provider, result.LogID, result.Error)
		} else {
			successCount++
			timestamp := time.Unix(result.STH.Timestamp/1000, 0)
			log.Printf("✅ [%s] %s", result.Provider, result.LogID)
			log.Printf("   Description: %s", result.Description)
			log.Printf("   Tree Size: %d", result.STH.TreeSize)
			log.Printf("   Timestamp: %s", timestamp.Format(time.RFC3339))
			log.Printf("   Root Hash: %s...", result.STH.SHA256RootHash[:16])
		}
	}

	log.Println("========================================")
	log.Printf("Total: %d | Success: %d | Failed: %d", len(results), successCount, failureCount)
	log.Println("========================================")
}
