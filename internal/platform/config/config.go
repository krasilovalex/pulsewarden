package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultEnvironment     = "local"
	defaultLogLevel        = "info"
	defaultShutdownTimeout = 10 * time.Second
	defaultHTTPAddress     = ":8080"

	envEnvironment     = "PULSEWARDEN_ENV"
	envLogLevel        = "PULSEWARDEN_LOG_LEVEL"
	envShutdownTimeout = "PULSEWARDEN_SHUTDOWN_TIMEOUT"
	envHTTPAddress     = "PULSEWARDEN_HTTP_ADDRESS"

	minShutdownTimeout = time.Second
	maxShutdownTimeout = 2 * time.Minute
)

func isValidEnvironment(value string) bool {
	switch value {
	case "local", "test", "staging", "production":
		return true
	default:
		return false
	}
}

func isValidLogLevel(value string) bool {
	switch value {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

type Config struct {
	Environment     string
	LogLevel        string
	ShutdownTimeout time.Duration
	HTTPAddress     string
}

func Load() (Config, error) {
	environment := envOrDefault(envEnvironment, defaultEnvironment)
	logLevel := envOrDefault(envLogLevel, defaultLogLevel)
	shutdownTimeoutValue := envOrDefault(
		envShutdownTimeout,
		defaultShutdownTimeout.String(),
	)
	httpAddress := strings.TrimSpace(
		envOrDefault(envHTTPAddress, defaultHTTPAddress),
	)

	environment = strings.ToLower(strings.TrimSpace(environment))
	logLevel = strings.ToLower(strings.TrimSpace(logLevel))
	shutdownTimeoutValue = strings.ToLower(strings.TrimSpace(shutdownTimeoutValue))

	if !isValidEnvironment(environment) {
		return Config{}, fmt.Errorf(
			"%s contains unsupported environment %q",
			envEnvironment,
			environment,
		)
	}

	if !isValidLogLevel(logLevel) {
		return Config{}, fmt.Errorf(
			"%s contains unsupported log level %q",
			envLogLevel,
			logLevel,
		)
	}

	shutdownTimeout, err := time.ParseDuration(shutdownTimeoutValue)
	if err != nil {
		return Config{}, fmt.Errorf(
			"parse %s: %w",
			envShutdownTimeout,
			err,
		)
	}

	if shutdownTimeout < minShutdownTimeout {
		return Config{}, fmt.Errorf(

			"%s must be at least %s",
			envShutdownTimeout,
			minShutdownTimeout,
		)
	}

	if shutdownTimeout > maxShutdownTimeout {
		return Config{}, fmt.Errorf(

			"%s must not exceed %s",
			envShutdownTimeout,
			maxShutdownTimeout,
		)
	}

	if httpAddress == "" {
		return Config{}, fmt.Errorf("%s must not be empty", envHTTPAddress)
	}

	return Config{
		Environment:     environment,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
		HTTPAddress:     httpAddress,
	}, nil
}

func envOrDefault(name, defaultValue string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return defaultValue
	}

	return value
}
