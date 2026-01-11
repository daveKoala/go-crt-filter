package scanner

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// FetchEntries fetches a batch of entries from a CT log
func FetchEntries(baseURL, logID string, start, end int64) (*EntriesResponse, error) {
	url := fmt.Sprintf("%s/%s/ct/v1/get-entries?start=%d&end=%d", baseURL, logID, start, end)

	client := &http.Client{
		Timeout: 60 * time.Second, // Longer timeout for batch fetches
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var entries EntriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries response: %w", err)
	}

	return &entries, nil
}

// GenerateBatchJobs creates batch jobs by working backwards from tree size
// Interleaves jobs from different logs to ensure parallel processing across all logs
func GenerateBatchJobs(sthResults []STHResult, windowSize int64, maxBatchesPerLog int64, cutOffDate time.Time) []BatchJob {
	// Create a map of log -> batches
	logBatches := make(map[string][]BatchJob)

	for _, result := range sthResults {
		if result.Error != nil || result.STH == nil {
			continue
		}

		logKey := fmt.Sprintf("%s/%s", result.Provider, result.LogID)
		treeSize := result.STH.TreeSize

		var batches []BatchJob
		batchCount := int64(0)

		// Work backwards in batches
		for end := treeSize - 1; end >= 0 && batchCount < maxBatchesPerLog; end -= windowSize {
			start := end - windowSize + 1
			if start < 0 {
				start = 0
			}

			batches = append(batches, BatchJob{
				Provider:    result.Provider,
				LogID:       result.LogID,
				BaseURL:     result.BaseURL,
				Description: result.Description,
				StartIndex:  start,
				EndIndex:    end,
				CutOffDate:  cutOffDate,
			})

			batchCount++

			// Stop at index 0
			if start == 0 {
				break
			}
		}

		logBatches[logKey] = batches
	}

	// Interleave batches from different logs
	// This ensures workers process multiple logs in parallel
	var jobs []BatchJob
	maxBatches := 0
	for _, batches := range logBatches {
		if len(batches) > maxBatches {
			maxBatches = len(batches)
		}
	}

	for i := 0; i < maxBatches; i++ {
		for logKey := range logBatches {
			if i < len(logBatches[logKey]) {
				jobs = append(jobs, logBatches[logKey][i])
			}
		}
	}

	return jobs
}

// SaveResponseToFile saves the entries response to a JSON file for testing
func SaveResponseToFile(job BatchJob, entries *EntriesResponse) error {
	// Create filename: provider_logid_start_end.json
	// Clean the logID to remove slashes
	cleanLogID := strings.ReplaceAll(job.LogID, "/", "_")
	if cleanLogID == "" {
		cleanLogID = "root"
	}

	filename := fmt.Sprintf("%s_%s_%d_%d.json", job.Provider, cleanLogID, job.StartIndex, job.EndIndex)
	filepath := filepath.Join("testdata", "responses", filename)

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal entries: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// RunBatchWorker processes a single batch job
func RunBatchWorker(job BatchJob, workerID int, results chan<- Result, saveToFile bool) {
	log.Printf("[Batch Worker %d] Fetching %s/%s entries [%d-%d]",
		workerID, job.Provider, job.LogID, job.StartIndex, job.EndIndex)

	entries, err := FetchEntries(job.BaseURL, job.LogID, job.StartIndex, job.EndIndex)
	if err != nil {
		log.Printf("[Batch Worker %d] ERROR: %s/%s [%d-%d]: %v",
			workerID, job.Provider, job.LogID, job.StartIndex, job.EndIndex, err)
		return
	}

	log.Printf("[Batch Worker %d] SUCCESS: %s/%s fetched %d entries",
		workerID, job.Provider, job.LogID, len(entries.Entries))

	// Save to file if requested
	if saveToFile {
		if err := SaveResponseToFile(job, entries); err != nil {
			log.Printf("[Batch Worker %d] WARNING: Failed to save response: %v", workerID, err)
		} else {
			log.Printf("[Batch Worker %d] Saved response to testdata/responses/", workerID)
		}
	}

	// TODO: Parse certificates, filter by cutoff date, extract domains
	// For now, just log the count
}

// RunBatchWorkers processes all batch jobs with a worker pool
func RunBatchWorkers(jobs []BatchJob, maxWorkers int, saveToFile bool) {
	jobsChan := make(chan BatchJob, len(jobs))
	resultsChan := make(chan Result, 1000)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobsChan {
				RunBatchWorker(job, workerID, resultsChan, saveToFile)
			}
		}(i)
	}

	// Send jobs to workers
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	// Wait for all workers to complete
	wg.Wait()
	close(resultsChan)

	// Collect and log results
	var totalResults int
	for range resultsChan {
		totalResults++
	}

	log.Printf("Batch processing complete: %d results collected", totalResults)
}
