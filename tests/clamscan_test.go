package tests

import (
	"testing"
	"time"

	"github.com/clamav-api/internal/config"
	"github.com/clamav-api/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestClamAVClientConfig(t *testing.T) {
	cfg := &config.ClamAVConfig{
		Host:    "localhost",
		Port:    3310,
		Timeout: 60 * time.Second,
	}

	assert.Equal(t, "localhost:3310", cfg.Address())
}

func TestScanResult(t *testing.T) {
	assert.Equal(t, model.ScanResult("clean"), model.ResultClean)
	assert.Equal(t, model.ScanResult("infected"), model.ResultInfected)
	assert.Equal(t, model.ScanResult("error"), model.ResultError)
}

func TestScanResponse(t *testing.T) {
	now := time.Now()
	resp := model.ScanResponse{
		FileID:     "test-uuid",
		FileName:   "test.pdf",
		FileSize:   1024,
		Result:     model.ResultClean,
		Threat:     "",
		DurationMs: 100,
		ScannedAt:  now,
	}

	assert.Equal(t, "test-uuid", resp.FileID)
	assert.Equal(t, "test.pdf", resp.FileName)
	assert.Equal(t, int64(1024), resp.FileSize)
	assert.Equal(t, model.ResultClean, resp.Result)
	assert.Empty(t, resp.Threat)
	assert.Equal(t, int64(100), resp.DurationMs)
	assert.Equal(t, now, resp.ScannedAt)
}

func TestWebhookPayload(t *testing.T) {
	now := time.Now()
	payload := model.WebhookPayload{
		FileID:     "test-uuid",
		FileName:   "test.pdf",
		FileSize:   1024,
		Result:     model.ResultInfected,
		Threat:     "Eicar.Test.File",
		ScannedAt:  now,
		DurationMs: 50,
	}

	assert.Equal(t, "test-uuid", payload.FileID)
	assert.Equal(t, model.ResultInfected, payload.Result)
	assert.Equal(t, "Eicar.Test.File", payload.Threat)
}
