package middleware

import (
	"github.com/clamav-api/internal/config"
	"github.com/gofiber/fiber/v2"
)

func NewAuthMiddleware(cfg *config.AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("API-Key")

		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"code":    fiber.StatusUnauthorized,
				"message": "API-Key header is required",
			})
		}

		if apiKey != cfg.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"code":    fiber.StatusUnauthorized,
				"message": "Invalid API-Key",
			})
		}

		return c.Next()
	}
}
