package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clamav-api/internal/clamscan"
	"github.com/clamav-api/internal/model"
	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	clamavClient *clamscan.Client
}

func NewHealthHandler(clamavClient *clamscan.Client) *HealthHandler {
	return &HealthHandler{
		clamavClient: clamavClient,
	}
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(model.HealthResponse{
		Status: "ok",
	})
}

func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	services := make(map[string]string)
	allHealthy := true

	if h.clamavClient == nil {
		services["clamav"] = "unavailable: client not configured"
		allHealthy = false
	} else {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		if err := h.clamavClient.Ping(ctx); err != nil {
			services["clamav"] = "unavailable: " + err.Error()
			allHealthy = false
		} else {
			services["clamav"] = "ok"
		}
	}

	status := "ok"
	statusCode := fiber.StatusOK
	if !allHealthy {
		status = "degraded"
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(model.ReadyResponse{
		Status:   status,
		Services: services,
	})
}

func (h *HealthHandler) Info(c *fiber.Ctx) error {
	if h.clamavClient == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.ErrorResponse{
			Error:   "service_unavailable",
			Code:    fiber.StatusServiceUnavailable,
			Message: "ClamAV client not configured",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	version, err := h.clamavClient.Version(ctx)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.ErrorResponse{
			Error:   "info_fetch_failed",
			Code:    fiber.StatusServiceUnavailable,
			Message: err.Error(),
		})
	}

	stats, err := h.clamavClient.Stats(ctx)

	infoResp := model.InfoResponse{
		Version: version,
	}

	if err == nil {
		parseStats(stats, &infoResp)
	}

	return c.JSON(infoResp)
}

func parseStats(raw string, info *model.InfoResponse) {
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "STATE:") {
			info.State = strings.TrimSpace(strings.TrimPrefix(line, "STATE:"))
		}
		if strings.HasPrefix(line, "THREADS:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "live" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &info.ThreadsLive)
				}
				if p == "idle" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &info.ThreadsIdle)
				}
				if p == "max" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &info.ThreadsMax)
				}
			}
		}
		if strings.HasPrefix(line, "QUEUE:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				fmt.Sscanf(parts[1], "%d", &info.QueueItems)
			}
		}
		if strings.HasPrefix(line, "STATS") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.StatsTime = parts[1]
			}
		}
	}
}
