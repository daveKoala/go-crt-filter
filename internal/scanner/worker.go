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

// FetchEntries fetches a batch of entries from a CT log
func FetchEntries(baseURL, logID string, start, end int64) (*EntriesResponse, error) {
	url := fmt.Sprintf("%s/%s/ct/v1/get-entries?start=%d&end=%d", baseURL, logID, start, end)

	client := &http.Client{
		Timeout: 60 * time.Second,
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

// SaveResponseToFile saves the entries response to a JSON file for testing
func SaveResponseToFile(provider, logID string, start, end int64, entries *EntriesResponse) error {
	// Create filename: provider_logid_start_end.json
	cleanLogID := strings.ReplaceAll(logID, "/", "_")
	if cleanLogID == "" {
		cleanLogID = "root"
	}

	filename := fmt.Sprintf("%s_%s_%d_%d.json", provider, cleanLogID, start, end)
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

// DecodeCertificate decodes a certificate from a leaf entry
// TODO: Implement real certificate parsing using x509
func DecodeCertificate(entry LeafEntry) (*Certificate, error) {
	// Stub implementation - returns empty cert for now
	// This will be implemented using saved test data
	return &Certificate{
		Subject:   "TODO",
		Issuer:    "TODO",
		NotBefore: time.Now(),
		NotAfter:  time.Now(),
		SANs:      []string{},
	}, nil
}

// ExtractDomains extracts domain names from a certificate
// TODO: Implement real domain extraction from SANs and CN
func ExtractDomains(cert *Certificate) []string {
	// Stub implementation - returns empty for now
	// This will be implemented using saved test data
	return []string{}
}

// LogWorker processes a single CT log independently
// It works backwards from the most recent entries until hitting the cutoff date
func LogWorker(
	provider string,
	baseURL string,
	logConfig config.CTLog,
	cutOffDate time.Time,
	results chan<- Result,
	saveToFile bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	logID := logConfig.ID
	log.Printf("[%s/%s] Starting worker...", provider, logID)

	// 1. Fetch STH to get tree size
	sth, err := FetchSTH(baseURL, logID)
	if err != nil {
		log.Printf("[%s/%s] ERROR fetching STH: %v", provider, logID, err)
		return
	}

	log.Printf("[%s/%s] Tree size: %d, starting scan backwards...", provider, logID, sth.TreeSize)

	// 2. Work backwards through batches
	windowSize := int64(1000)
	currentIndex := sth.TreeSize - 1
	batchNum := 0

	for currentIndex >= 0 {
		batchNum++
		start := currentIndex - windowSize + 1
		if start < 0 {
			start = 0
		}
		end := currentIndex

		log.Printf("[%s/%s] Batch %d: Fetching entries [%d-%d]",
			provider, logID, batchNum, start, end)

		// 3. Fetch batch of entries
		entries, err := FetchEntries(baseURL, logID, start, end)
		if err != nil {
			log.Printf("[%s/%s] ERROR fetching batch: %v", provider, logID, err)
			return
		}

		log.Printf("[%s/%s] Batch %d: Fetched %d entries",
			provider, logID, batchNum, len(entries.Entries))

		// 4. Optionally save for testing
		if saveToFile {
			if err := SaveResponseToFile(provider, logID, start, end, entries); err != nil {
				log.Printf("[%s/%s] WARNING: Failed to save response: %v", provider, logID, err)
			}
		}

		// 5. Process each entry in the batch
		shouldStop := false
		for _, entry := range entries.Entries {
			// Decode certificate (stub for now)
			cert, err := DecodeCertificate(entry)
			if err != nil {
				log.Printf("[%s/%s] WARNING: Failed to decode cert: %v", provider, logID, err)
				continue
			}

			// Simple cutoff check - stop if cert is too old
			if cert.NotBefore.Before(cutOffDate) {
				log.Printf("[%s/%s] Reached cutoff date, stopping worker", provider, logID)
				shouldStop = true
				break
			}

			// Extract domains (stub for now)
			domains := ExtractDomains(cert)

			// Send results
			for _, domain := range domains {
				results <- Result{
					LogID:     logID,
					Provider:  provider,
					Domain:    domain,
					Issuer:    cert.Issuer,
					NotBefore: cert.NotBefore,
					NotAfter:  cert.NotAfter,
				}
			}
		}

		if shouldStop {
			break
		}

		// Move to next batch backwards
		currentIndex -= windowSize

		// Safety limit for testing - stop after a few batches
		if batchNum >= 5 {
			log.Printf("[%s/%s] Reached batch limit (5), stopping for testing", provider, logID)
			break
		}
	}

	log.Printf("[%s/%s] Worker completed", provider, logID)
}
