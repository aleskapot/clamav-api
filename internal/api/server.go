package api

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clamav-api/internal/clamscan"
	"github.com/clamav-api/internal/config"
	"github.com/clamav-api/internal/handler"
	"github.com/clamav-api/internal/logger"
	"github.com/clamav-api/internal/middleware"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const shutdownTimeout = 30 * time.Second

type Server struct {
	app          *fiber.App
	config       *config.Config
	clamavClient *clamscan.Client
}

func NewServer(cfg *config.Config) *Server {
	clamavClient := clamscan.NewClient(&cfg.ClamAV)

	filesHandler := handler.NewFilesHandler(
		clamavClient,
		&cfg.Webhook,
		&cfg.Storage,
		cfg.App.MaxFileSize,
	)

	healthHandler := handler.NewHealthHandler(clamavClient)

	app := fiber.New(fiber.Config{
		BodyLimit:     cfg.App.MaxFileSize * 1024 * 1024,
		ServerHeader:  "ClamAV API",
		AppName:       "ClamAV API",
		StrictRouting: false,
		CaseSensitive: false,
	})

	app.Use(recover.New())
	app.Use(requestid.New(requestid.Config{
		Header: "Request-ID",
	}))
	app.Use(middleware.RequestLogger())

	app.Use(cors.New())
	app.Use(middleware.PrometheusMiddleware())

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/health",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
			defer cancel()
			return clamavClient.Ping(ctx) == nil
		},
		ReadinessEndpoint: "/ready",
	}))

	app.Get("/info", healthHandler.Info)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/swagger.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/openapi.yaml")
	})

	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/openapi.yaml",
		Path:     "swagger",
		Title:    "ClamAV API Documentation",
	}))

	app.Use(middleware.NewAuthMiddleware(&cfg.Auth))

	files := app.Group("/files")
	files.Post("/scan", filesHandler.Scan)
	files.Post("/upload", filesHandler.Upload)

	logger.Log.Info("API server configured",
		zap.String("address", cfg.App.Address()),
		zap.Int("max_file_size_mb", cfg.App.MaxFileSize),
	)

	return &Server{
		app:          app,
		config:       cfg,
		clamavClient: clamavClient,
	}
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) Start() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.app.Listen(s.config.App.Address()); err != nil {
			logger.Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Log.Info("Server started", zap.String("address", s.config.App.Address()))

	<-quit

	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server stopped")
	return nil
}
