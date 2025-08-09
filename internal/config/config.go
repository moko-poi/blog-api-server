package config

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

// Config holds the application configuration
// Following Mat Ryer's pattern of using environment variables for configuration
type Config struct {
	Host            string
	Port            int
	LogLevel        slog.Level
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Load creates a new Config from environment variables
// Following Mat Ryer's pattern of accepting getenv function for testability
func Load(getenv func(string) string) (*Config, error) {
	cfg := &Config{
		// Default values
		Host:            "localhost",
		Port:            8080,
		LogLevel:        slog.LevelInfo,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 15 * time.Second,
	}

	// Override with environment variables if provided
	if host := getenv("HOST"); host != "" {
		cfg.Host = host
	}

	if portStr := getenv("PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
	}

	if logLevel := getenv("LOG_LEVEL"); logLevel != "" {
		level, err := parseLogLevel(logLevel)
		if err != nil {
			return nil, fmt.Errorf("invalid LOG_LEVEL: %w", err)
		}
		cfg.LogLevel = level
	}

	if readTimeoutStr := getenv("READ_TIMEOUT"); readTimeoutStr != "" {
		timeout, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
		}
		cfg.ReadTimeout = timeout
	}

	if writeTimeoutStr := getenv("WRITE_TIMEOUT"); writeTimeoutStr != "" {
		timeout, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
		}
		cfg.WriteTimeout = timeout
	}

	if shutdownTimeoutStr := getenv("SHUTDOWN_TIMEOUT"); shutdownTimeoutStr != "" {
		timeout, err := time.ParseDuration(shutdownTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = timeout
	}

	return cfg, nil
}

// Address returns the full address string for the server
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// parseLogLevel converts a string to slog.Level
func parseLogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug", "DEBUG":
		return slog.LevelDebug, nil
	case "info", "INFO":
		return slog.LevelInfo, nil
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn, nil
	case "error", "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown level: %s", level)
	}
}