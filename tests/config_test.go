package tests

import (
	"os"
	"testing"
	"time"

	"github.com/clamav-api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidation_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Host:        "0.0.0.0",
			Port:        8080,
			MaxFileSize: 100,
		},
		ClamAV: config.ClamAVConfig{
			Host:    "localhost",
			Port:    3310,
			Timeout: 60 * time.Second,
		},
		Auth: config.AuthConfig{
			APIKey: "test-key",
		},
		Webhook: config.WebhookConfig{
			URL:        "http://localhost:8081/webhook",
			Timeout:    30 * time.Second,
			RetryCount: 3,
		},
		Storage: config.StorageConfig{
			TempDir: "/tmp/clamav-api",
		},
	}

	assert.Equal(t, "0.0.0.0:8080", cfg.App.Address())
	assert.Equal(t, "localhost:3310", cfg.ClamAV.Address())
}

func TestConfigValidation_InvalidPort(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Port:        -1,
			MaxFileSize: 100,
		},
		ClamAV: config.ClamAVConfig{
			Host: "localhost",
			Port: 3310,
		},
		Auth: config.AuthConfig{
			APIKey: "test-key",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid app.port")
}

func TestConfigValidation_EmptyAPIKey(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Port:        8080,
			MaxFileSize: 100,
		},
		ClamAV: config.ClamAVConfig{
			Host: "localhost",
			Port: 3310,
		},
		Auth: config.AuthConfig{
			APIKey: "",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth.api_key is required")
}

func TestConfigValidation_EmptyTempDir(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Port:        8080,
			MaxFileSize: 100,
		},
		ClamAV: config.ClamAVConfig{
			Host: "localhost",
			Port: 3310,
		},
		Auth: config.AuthConfig{
			APIKey: "test-key",
		},
		Storage: config.StorageConfig{
			TempDir: "",
		},
	}

	err := cfg.Validate()
	require.NoError(t, err)
	assert.Equal(t, "/tmp/clamav-api", cfg.Storage.TempDir)
}

func TestConfigLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config")
}

func TestConfigLoadFromFile(t *testing.T) {
	content := `
app:
  host: "0.0.0.0"
  port: 8080
  max_file_size: 100
clamav:
  host: "localhost"
  port: 3310
  timeout: 60s
auth:
  api_key: "test-api-key"
webhook:
  url: "http://localhost:8081/webhook"
  timeout: 30s
  retry_count: 3
storage:
  temp_dir: "/tmp/clamav-api"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := config.Load(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.App.Host)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, 100, cfg.App.MaxFileSize)
	assert.Equal(t, "test-api-key", cfg.Auth.APIKey)
	assert.Equal(t, "localhost", cfg.ClamAV.Host)
	assert.Equal(t, 3310, cfg.ClamAV.Port)
}
