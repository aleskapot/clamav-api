package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clamav-api/internal/config"
	"github.com/clamav-api/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidAPIKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{APIKey: "test-api-key"}

	app.Use(middleware.NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("API-Key", "test-api-key")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_MissingAPIKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{APIKey: "test-api-key"}

	app.Use(middleware.NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{APIKey: "test-api-key"}

	app.Use(middleware.NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("API-Key", "wrong-api-key")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_EmptyAPIKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{APIKey: "test-api-key"}

	app.Use(middleware.NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("API-Key", "")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
