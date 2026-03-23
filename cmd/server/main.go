package main

import (
	"fmt"
	"os"

	"github.com/clamav-api/internal/api"
	"github.com/clamav-api/internal/config"
	"github.com/clamav-api/internal/logger"
	"go.uber.org/zap"
)

func main() {
	logLevel := os.Getenv("APP_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	if err := logger.InitWithLevel(logLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Log.Fatal("Failed to load configuration", zap.Error(err))
	}

	if err := api.NewServer(cfg).Start(); err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
	}
}
