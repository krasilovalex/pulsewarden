package logger

import (
	"fmt"
	"io"
	"log/slog"
)

func New(out io.Writer, level, service string) (*slog.Logger, error) {
	slogLevel, err := parseLevel(level)

	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slogLevel,
	})

	return slog.New(handler).With(
		slog.String("service", service),
	), nil
}

func parseLevel(value string) (slog.Level, error) {
	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q", value)
	}
}
