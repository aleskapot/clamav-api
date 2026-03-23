package middleware

import (
	"time"

	"github.com/clamav-api/internal/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		logger.Log.Info("HTTP Request",
			zap.String("request_id", c.Locals("requestid").(string)),
			zap.Int("status", c.Response().StatusCode()),
			zap.String("latency", duration.String()),
			zap.String("method", c.Method()),
			zap.String("url", c.Path()),
			zap.String("ip", c.IP()),
			zap.Int("size", len(c.Response().Body())),
		)

		return err
	}
}
