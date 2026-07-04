package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultEnvironment           = "local"
	defaultLogLevel              = "info"
	defaultShutdownTimeout       = 10 * time.Second
	defaultHTTPAddress           = ":8080"
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPReadTimeout       = 10 * time.Second
	defaultHTTPWriteTimeout      = 10 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second

	envEnvironment           = "PULSEWARDEN_ENV"
	envLogLevel              = "PULSEWARDEN_LOG_LEVEL"
	envShutdownTimeout       = "PULSEWARDEN_SHUTDOWN_TIMEOUT"
	envHTTPAddress           = "PULSEWARDEN_HTTP_ADDRESS"
	envHTTPReadHeaderTimeout = "PULSEWARDEN_HTTP_READ_HEADER_TIMEOUT"
	envHTTPReadTimeout       = "PULSEWARDEN_HTTP_READ_TIMEOUT"
	envHTTPWriteTimeout      = "PULSEWARDEN_HTTP_WRITE_TIMEOUT"
	envHTTPIdleTimeout       = "PULSEWARDEN_HTTP_IDLE_TIMEOUT"

	minShutdownTimeout = time.Second
	maxShutdownTimeout = 2 * time.Minute
	minHTTPTimeout     = 100 * time.Millisecond
	maxHTTPTimeout     = 10 * time.Minute
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
	Environment           string
	LogLevel              string
	ShutdownTimeout       time.Duration
	HTTPAddress           string
	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
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

	httpReadHeaderTimeout, err := parseDurationSetting(
		envHTTPReadHeaderTimeout,
		defaultHTTPReadHeaderTimeout,
		minHTTPTimeout,
		maxHTTPTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	httpReadTimeout, err := parseDurationSetting(
		envHTTPReadTimeout,
		defaultHTTPReadTimeout,
		minHTTPTimeout,
		maxHTTPTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	httpWriteTimeout, err := parseDurationSetting(
		envHTTPWriteTimeout,
		defaultHTTPWriteTimeout,
		minHTTPTimeout,
		maxHTTPTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	httpIdleTimeout, err := parseDurationSetting(
		envHTTPIdleTimeout,
		defaultHTTPIdleTimeout,
		minHTTPTimeout,
		maxHTTPTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:           environment,
		LogLevel:              logLevel,
		ShutdownTimeout:       shutdownTimeout,
		HTTPAddress:           httpAddress,
		HTTPReadTimeout:       httpReadTimeout,
		HTTPReadHeaderTimeout: httpReadHeaderTimeout,
		HTTPWriteTimeout:      httpWriteTimeout,
		HTTPIdleTimeout:       httpIdleTimeout,
	}, nil
}

func envOrDefault(name, defaultValue string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return defaultValue
	}

	return value
}

func parseDurationSetting(
	name string,
	defaultValue time.Duration,
	minValue time.Duration,
	maxValue time.Duration,
) (time.Duration, error) {
	rawValue := strings.TrimSpace(
		envOrDefault(name, defaultValue.String()),
	)

	value, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}

	if value < minValue {
		return 0, fmt.Errorf(
			"%s must be at least %s",
			name,
			minValue,
		)
	}

	if value > maxValue {
		return 0, fmt.Errorf(
			"%s must not exceed %s",
			name,
			maxValue,
		)
	}

	return value, nil
}
