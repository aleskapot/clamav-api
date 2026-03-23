package model

import "time"

type ScanResult string

const (
	ResultClean    ScanResult = "clean"
	ResultInfected ScanResult = "infected"
	ResultError    ScanResult = "error"
)

type ScanResponse struct {
	FileID     string     `json:"file_id"`
	FileName   string     `json:"file_name"`
	FileSize   int64      `json:"file_size"`
	Result     ScanResult `json:"result"`
	Threat     string     `json:"threat,omitempty"`
	DurationMs int64      `json:"duration_ms"`
	ScannedAt  time.Time  `json:"scanned_at"`
}

type UploadResponse struct {
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	Message    string `json:"message"`
	ReceivedAt string `json:"received_at"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ReadyResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type InfoResponse struct {
	Version     string `json:"version"`
	State       string `json:"state,omitempty"`
	ThreadsLive int    `json:"threads_live,omitempty"`
	ThreadsIdle int    `json:"threads_idle,omitempty"`
	ThreadsMax  int    `json:"threads_max,omitempty"`
	QueueItems  int    `json:"queue_items,omitempty"`
	StatsTime   string `json:"stats_time,omitempty"`
}
