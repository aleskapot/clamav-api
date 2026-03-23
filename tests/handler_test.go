package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clamav-api/internal/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	app := fiber.New()
	h := handler.NewHealthHandler(nil)

	app.Get("/health", h.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadyHandler_ReadyWithNilClient(t *testing.T) {
	app := fiber.New()
	h := handler.NewHealthHandler(nil)

	app.Get("/ready", h.Ready)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}
