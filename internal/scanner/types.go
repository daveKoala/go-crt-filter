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
