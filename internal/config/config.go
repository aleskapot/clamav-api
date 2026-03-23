package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	ClamAV  ClamAVConfig  `mapstructure:"clamav"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Webhook WebhookConfig `mapstructure:"webhook"`
	Storage StorageConfig `mapstructure:"storage"`
}

type AppConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	MaxFileSize int    `mapstructure:"max_file_size"`
	LogLevel    string `mapstructure:"log_level"`
}

type ClamAVConfig struct {
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type AuthConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type WebhookConfig struct {
	URL        string        `mapstructure:"url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int           `mapstructure:"retry_count"`
}

type StorageConfig struct {
	TempDir string `mapstructure:"temp_dir"`
}

func (c *ClamAVConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *AppConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	return c.validate()
}

func (c *Config) validate() error {
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("invalid app.port: %d", c.App.Port)
	}

	if c.App.MaxFileSize <= 0 {
		return fmt.Errorf("invalid app.max_file_size: %d", c.App.MaxFileSize)
	}

	if c.Auth.APIKey == "" {
		return fmt.Errorf("auth.api_key is required")
	}

	if c.ClamAV.Host == "" {
		return fmt.Errorf("clamav.host is required")
	}

	if c.ClamAV.Port <= 0 || c.ClamAV.Port > 65535 {
		return fmt.Errorf("invalid clamav.port: %d", c.ClamAV.Port)
	}

	if c.Storage.TempDir == "" {
		c.Storage.TempDir = "/tmp/clamav-api"
	}

	if c.ClamAV.Timeout == 0 {
		c.ClamAV.Timeout = 60 * time.Second
	}

	if c.Webhook.Timeout == 0 {
		c.Webhook.Timeout = 30 * time.Second
	}

	if c.Webhook.RetryCount == 0 {
		c.Webhook.RetryCount = 3
	}

	return nil
}
