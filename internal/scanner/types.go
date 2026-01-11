package scanner

import "time"

// ScanRequest represents the request body for the /scan endpoint
type ScanRequest struct {
	CutOffDate string `json:"cut_off_date"`
}

// Entry represents a single CT log entry
type Entry struct {
	Index       int64
	Timestamp   time.Time
	Certificate []byte
	Domain      string
}

// Result represents a processed scan result
type Result struct {
	LogID      string    `json:"log_id"`
	Provider   string    `json:"provider"`
	Domain     string    `json:"domain"`
	Issuer     string    `json:"issuer"`
	NotBefore  time.Time `json:"not_before"`
	NotAfter   time.Time `json:"not_after"`
	Index      int64     `json:"index"`
}

// ScanJob represents a single scanning job for one CT log
type ScanJob struct {
	Provider   string
	BaseURL    string
	LogID      string
	CutOffDate time.Time
}

// STHResponse represents the Signed Tree Head response from a CT log
type STHResponse struct {
	TreeSize          int64  `json:"tree_size"`
	Timestamp         int64  `json:"timestamp"`
	SHA256RootHash    string `json:"sha256_root_hash"`
	TreeHeadSignature string `json:"tree_head_signature"`
}

// STHResult represents the result of fetching STH from a log
type STHResult struct {
	Provider    string
	LogID       string
	BaseURL     string
	Description string
	STH         *STHResponse
	Error       error
}

// BatchJob represents a single batch of entries to fetch
type BatchJob struct {
	Provider    string
	LogID       string
	BaseURL     string
	Description string
	StartIndex  int64
	EndIndex    int64
	CutOffDate  time.Time
}

// LeafEntry represents a single log entry from the CT log
type LeafEntry struct {
	LeafInput string `json:"leaf_input"`
	ExtraData string `json:"extra_data"`
}

// EntriesResponse represents the response from get-entries endpoint
type EntriesResponse struct {
	Entries []LeafEntry `json:"entries"`
}

// Certificate represents a parsed X.509 certificate
type Certificate struct {
	Subject   string
	Issuer    string
	NotBefore time.Time
	NotAfter  time.Time
	SANs      []string // Subject Alternative Names
}
