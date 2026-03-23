package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/clamav-api/internal/clamscan"
	"github.com/clamav-api/internal/config"
	"github.com/clamav-api/internal/logger"
	"github.com/clamav-api/internal/middleware"
	"github.com/clamav-api/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FilesHandler struct {
	clamavClient *clamscan.Client
	webhookCfg   *config.WebhookConfig
	storageCfg   *config.StorageConfig
	maxFileSize  int64
}

func NewFilesHandler(
	clamavClient *clamscan.Client,
	webhookCfg *config.WebhookConfig,
	storageCfg *config.StorageConfig,
	maxFileSize int,
) *FilesHandler {
	return &FilesHandler{
		clamavClient: clamavClient,
		webhookCfg:   webhookCfg,
		storageCfg:   storageCfg,
		maxFileSize:  int64(maxFileSize) * 1024 * 1024,
	}
}

func (h *FilesHandler) Scan(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Log.Warn("No file in request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error:   "bad_request",
			Code:    fiber.StatusBadRequest,
			Message: "No file provided",
		})
	}

	if fileHeader.Size > h.maxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error:   "file_too_large",
			Code:    fiber.StatusBadRequest,
			Message: fmt.Sprintf("File size exceeds maximum allowed size of %d MB", h.maxFileSize/(1024*1024)),
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.Log.Error("Failed to open uploaded file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to read uploaded file",
		})
	}
	defer file.Close()

	middleware.RecordFileSize("sync", fileHeader.Size)

	ctx, cancel := context.WithTimeout(c.Context(), h.clamavClient.GetTimeout())
	defer cancel()

	result, _, err := h.clamavClient.ScanStream(ctx, file)
	if err != nil {
		logger.Log.Error("Scan failed", zap.Error(err), zap.String("filename", fileHeader.Filename))
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error:   "scan_failed",
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to scan file",
		})
	}

	middleware.RecordFileScanned(string(result.Result))

	logger.Log.Info("File scanned",
		zap.String("request_id", c.Locals("requestid").(string)),
		zap.String("file_name", fileHeader.Filename),
		zap.Int64("file_size", fileHeader.Size),
		zap.String("result", string(result.Result)),
		zap.String("threat", result.Threat),
		zap.Int64("duration_ms", result.DurationMs),
	)

	response := model.ScanResponse{
		FileID:     uuid.New().String(),
		FileName:   fileHeader.Filename,
		FileSize:   fileHeader.Size,
		Result:     result.Result,
		Threat:     result.Threat,
		DurationMs: result.DurationMs,
		ScannedAt:  result.ScannedAt,
	}

	return c.JSON(response)
}

func (h *FilesHandler) Upload(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Log.Warn("No file in request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error:   "bad_request",
			Code:    fiber.StatusBadRequest,
			Message: "No file provided",
		})
	}

	if fileHeader.Size > h.maxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error:   "file_too_large",
			Code:    fiber.StatusBadRequest,
			Message: fmt.Sprintf("File size exceeds maximum allowed size of %d MB", h.maxFileSize/(1024*1024)),
		})
	}

	fileID := uuid.New().String()
	ext := filepath.Ext(fileHeader.Filename)
	savedPath := filepath.Join(h.storageCfg.TempDir, fileID+ext)

	if err := os.MkdirAll(h.storageCfg.TempDir, 0755); err != nil {
		logger.Log.Error("Failed to create temp dir", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to save file",
		})
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		logger.Log.Error("Failed to open uploaded file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to read uploaded file",
		})
	}
	defer srcFile.Close()

	dstFile, err := os.Create(savedPath)
	if err != nil {
		logger.Log.Error("Failed to create destination file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to save file",
		})
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		logger.Log.Error("Failed to copy file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Code:    fiber.StatusInternalServerError,
			Message: "Failed to save file",
		})
	}

	middleware.RecordFileSize("async", fileHeader.Size)

	go h.processAsyncFile(fileID, fileHeader.Filename, fileHeader.Size, savedPath)

	return c.Status(fiber.StatusAccepted).JSON(model.UploadResponse{
		FileID:     fileID,
		FileName:   fileHeader.Filename,
		FileSize:   fileHeader.Size,
		Message:    "File uploaded and queued for scanning",
		ReceivedAt: time.Now().Format(time.RFC3339),
	})
}

func (h *FilesHandler) processAsyncFile(fileID, fileName string, fileSize int64, filePath string) {
	defer os.Remove(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Log.Error("Failed to open file for scan", zap.Error(err), zap.String("file_id", fileID))
		h.sendWebhook(model.WebhookPayload{
			FileID:   fileID,
			FileName: fileName,
			FileSize: fileSize,
			Result:   model.ResultError,
			Threat:   "Failed to open file",
		})
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), h.clamavClient.GetTimeout())
	defer cancel()

	result, _, err := h.clamavClient.ScanStream(ctx, file)
	if err != nil {
		logger.Log.Error("Async scan failed", zap.Error(err), zap.String("file_id", fileID))
		h.sendWebhook(model.WebhookPayload{
			FileID:   fileID,
			FileName: fileName,
			FileSize: fileSize,
			Result:   model.ResultError,
			Threat:   "Scan error: " + err.Error(),
		})
		return
	}

	middleware.RecordFileScanned(string(result.Result))

	logger.Log.Info("Async file scanned",
		zap.String("file_id", fileID),
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
		zap.String("result", string(result.Result)),
		zap.String("threat", result.Threat),
		zap.Int64("duration_ms", result.DurationMs),
	)

	payload := model.WebhookPayload{
		FileID:     fileID,
		FileName:   fileName,
		FileSize:   fileSize,
		Result:     result.Result,
		Threat:     result.Threat,
		ScannedAt:  result.ScannedAt,
		DurationMs: result.DurationMs,
	}

	h.sendWebhook(payload)
}

func (h *FilesHandler) sendWebhook(payload model.WebhookPayload) {
	client := &http.Client{
		Timeout: h.webhookCfg.Timeout,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error("Failed to marshal webhook payload", zap.Error(err))
		return
	}

	req, err := http.NewRequest(http.MethodPost, h.webhookCfg.URL, bytes.NewReader(body))
	if err != nil {
		logger.Log.Error("Failed to create webhook request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if h.webhookCfg.APIKey != "" {
		req.Header.Set("API-Key", h.webhookCfg.APIKey)
	}

	var lastErr error
	for i := 0; i < h.webhookCfg.RetryCount; i++ {
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			logger.Log.Info("Webhook sent successfully",
				zap.String("file_id", payload.FileID),
				zap.String("webhook_url", h.webhookCfg.URL),
			)
			return
		}
		lastErr = err
		logger.Log.Warn("Webhook send failed, retrying",
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", h.webhookCfg.RetryCount),
		)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	logger.Log.Error("Webhook send failed after all retries",
		zap.Error(lastErr),
		zap.String("file_id", payload.FileID),
		zap.String("webhook_url", h.webhookCfg.URL),
	)
}
