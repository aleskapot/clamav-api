package model

import "time"

type WebhookPayload struct {
	FileID     string     `json:"file_id"`
	FileName   string     `json:"file_name"`
	FileSize   int64      `json:"file_size"`
	Result     ScanResult `json:"result"`
	Threat     string     `json:"threat,omitempty"`
	ScannedAt  time.Time  `json:"scanned_at"`
	DurationMs int64      `json:"duration_ms"`
}
