package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/krasilovalex/pulsewarden/internal/platform/config"
)

func TestOpenRejectsInvalidDSN(t *testing.T) {
	cfg := config.PostgresConfig{
		DSN:              "://invalid",
		MaxConns:         10,
		MinConns:         1,
		MaxConnLifetime:  30 * time.Minute,
		MaxConnIdleTime:  5 * time.Minute,
		ConnectTimeout:   time.Second,
		ReadinessTimeout: time.Second,
	}

	pool, err := Open(context.Background(), cfg)
	if err == nil {
		if pool != nil {
			pool.Close()
		}

		t.Fatal("expected error, got nil")
	}

	if pool != nil {
		pool.Close()
		t.Fatal("pool is not nil after failed initialization")
	}

	if !strings.Contains(err.Error(), "parse PostgreSQL configuration") {
		t.Fatalf("unexpected error: %v", err)
	}
}
