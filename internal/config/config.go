package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/spf13/viper"
)

const (
	DefaultPort                  = "8080"
	DefaultKrakenAPIBaseEndpoint = "https://api.kraken.com/0/public/Ticker"
	DefaultLogLevel              = "info"
)

// Rate limiter specific configuration
type RateLimit struct {
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	IPTrackingTTL     time.Duration `mapstructure:"ip_tracking_ttl"`
	CacheSize         int           `mapstructure:"cache_size"`
}

// Config holds application configuration values.
type Config struct {
	Port                  string    `mapstructure:"port"`
	KrakenAPIBaseEndpoint string    `mapstructure:"kraken_api_base_endpoint"`
	LogLevel              string    `mapstructure:"log_level"`
	RateLimit             RateLimit `mapstructure:"rate_limit"`
}

// LoadConfig loads configuration from a file and/or environment variables.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("port", DefaultPort)
	v.SetDefault("kraken_api_base_endpoint", DefaultKrakenAPIBaseEndpoint)
	v.SetDefault("log_level", DefaultLogLevel)

	// set defaults for rate limit
	v.SetDefault("rate_limit.requests_per_minute", 60)
	v.SetDefault("rate_limit.burst_size", 5)
	v.SetDefault("rate_limit.ip_tracking_ttl", "1h")
	v.SetDefault("rate_limit.cache_size", 10000)

	v.AutomaticEnv()
	// Set prefix for env vars
	v.SetEnvPrefix("RELAI")
	// Replace dots with underscores in env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load config file if it exists
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				logger.Warnf("Config file not found at %s, using defaults", configPath)
				// Continue with defaults instead of returning error
			} else {
				// Only return error for non-missing file issues
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		} else {
			logger.Infof("Using config file: %s", configPath)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	logger.Debugf("Configuration loaded: %+v", config)

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}
	if config.KrakenAPIBaseEndpoint == "" {
		return fmt.Errorf("kraken API base endpoint is required")
	}
	return nil
}
