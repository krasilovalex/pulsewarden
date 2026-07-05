package config

import (
	"os"
	"testing"
	"time"
)

const testPostgresDSN = "postgres://test:test@localhost:5432/pulsewarden_test"

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults",
			env: map[string]string{
				envPostgresDSN: testPostgresDSN,
			},
			want: Config{
				Environment:           "local",
				LogLevel:              "info",
				ShutdownTimeout:       10 * time.Second,
				HTTPAddress:           ":8080",
				HTTPReadHeaderTimeout: 5 * time.Second,
				HTTPReadTimeout:       10 * time.Second,
				HTTPWriteTimeout:      10 * time.Second,
				HTTPIdleTimeout:       60 * time.Second,

				PostgresDSN:              testPostgresDSN,
				PostgresMaxConns:         10,
				PostgresMinConns:         1,
				PostgresMaxConnLifetime:  30 * time.Minute,
				PostgresMaxConnIdleTime:  5 * time.Minute,
				PostgresConnectTimeout:   5 * time.Second,
				PostgresReadinessTimeout: 2 * time.Second,
			},
		},
		{
			name: "custom values",
			env: map[string]string{
				envEnvironment:           "production",
				envLogLevel:              "debug",
				envShutdownTimeout:       "30s",
				envHTTPAddress:           "127.0.0.1:9090",
				envHTTPReadHeaderTimeout: "2s",
				envHTTPReadTimeout:       "20s",
				envHTTPWriteTimeout:      "25s",
				envHTTPIdleTimeout:       "2m",

				envPostgresDSN:              testPostgresDSN,
				envPostgresMaxConns:         "20",
				envPostgresMinConns:         "3",
				envPostgresMaxConnLifetime:  "2h",
				envPostgresMaxConnIdleTime:  "15m",
				envPostgresConnectTimeout:   "10s",
				envPostgresReadinessTimeout: "4s",
			},
			want: Config{
				Environment:           "production",
				LogLevel:              "debug",
				ShutdownTimeout:       30 * time.Second,
				HTTPAddress:           "127.0.0.1:9090",
				HTTPReadHeaderTimeout: 2 * time.Second,
				HTTPReadTimeout:       20 * time.Second,
				HTTPWriteTimeout:      25 * time.Second,
				HTTPIdleTimeout:       2 * time.Minute,

				PostgresDSN:              testPostgresDSN,
				PostgresMaxConns:         20,
				PostgresMinConns:         3,
				PostgresMaxConnLifetime:  2 * time.Hour,
				PostgresMaxConnIdleTime:  15 * time.Minute,
				PostgresConnectTimeout:   10 * time.Second,
				PostgresReadinessTimeout: 4 * time.Second,
			},
		},
		{
			name: "values are normalized",
			env: map[string]string{
				envEnvironment:           " Staging ",
				envLogLevel:              " WARN ",
				envShutdownTimeout:       " 15s ",
				envHTTPAddress:           " 127.0.0.1:8081 ",
				envHTTPReadHeaderTimeout: " 3s ",
				envHTTPReadTimeout:       " 12s ",
				envHTTPWriteTimeout:      " 14s ",
				envHTTPIdleTimeout:       " 90s ",

				envPostgresDSN:              " " + testPostgresDSN + " ",
				envPostgresMaxConns:         " 15 ",
				envPostgresMinConns:         " 2 ",
				envPostgresMaxConnLifetime:  " 45m ",
				envPostgresMaxConnIdleTime:  " 10m ",
				envPostgresConnectTimeout:   " 3s ",
				envPostgresReadinessTimeout: " 1s ",
			},
			want: Config{
				Environment:           "staging",
				LogLevel:              "warn",
				ShutdownTimeout:       15 * time.Second,
				HTTPAddress:           "127.0.0.1:8081",
				HTTPReadHeaderTimeout: 3 * time.Second,
				HTTPReadTimeout:       12 * time.Second,
				HTTPWriteTimeout:      14 * time.Second,
				HTTPIdleTimeout:       90 * time.Second,

				PostgresDSN:              testPostgresDSN,
				PostgresMaxConns:         15,
				PostgresMinConns:         2,
				PostgresMaxConnLifetime:  45 * time.Minute,
				PostgresMaxConnIdleTime:  10 * time.Minute,
				PostgresConnectTimeout:   3 * time.Second,
				PostgresReadinessTimeout: time.Second,
			},
		},
		{
			name: "unsupported environment",
			env: map[string]string{
				envEnvironment: "development",
				envPostgresDSN: testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "unsupported log level",
			env: map[string]string{
				envLogLevel:    "trace",
				envPostgresDSN: testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "invalid shutdown timeout",
			env: map[string]string{
				envShutdownTimeout: "soon",
				envPostgresDSN:     testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "shutdown timeout below minimum",
			env: map[string]string{
				envShutdownTimeout: "500ms",
				envPostgresDSN:     testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "shutdown timeout above maximum",
			env: map[string]string{
				envShutdownTimeout: "3m",
				envPostgresDSN:     testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "empty environment is rejected",
			env: map[string]string{
				envEnvironment: "",
				envPostgresDSN: testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "empty log level is rejected",
			env: map[string]string{
				envLogLevel:    "",
				envPostgresDSN: testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "empty shutdown timeout is rejected",
			env: map[string]string{
				envShutdownTimeout: "",
				envPostgresDSN:     testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "empty HTTP address is rejected",
			env: map[string]string{
				envHTTPAddress: "",
				envPostgresDSN: testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "invalid HTTP read header timeout",
			env: map[string]string{
				envHTTPReadHeaderTimeout: "slow",
				envPostgresDSN:           testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "HTTP read timeout below minimum",
			env: map[string]string{
				envHTTPReadTimeout: "50ms",
				envPostgresDSN:     testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "HTTP write timeout above maximum",
			env: map[string]string{
				envHTTPWriteTimeout: "11m",
				envPostgresDSN:      testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "empty HTTP idle timeout is rejected",
			env: map[string]string{
				envHTTPIdleTimeout: "",
				envPostgresDSN:     testPostgresDSN,
			},
			wantErr: true,
		},
		{
			name: "empty PostgreSQL DSN is rejected",
			env: map[string]string{
				envPostgresDSN: "",
			},
			wantErr: true,
		},
		{
			name: "invalid PostgreSQL max connections",
			env: map[string]string{
				envPostgresDSN:      testPostgresDSN,
				envPostgresMaxConns: "many",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL max connections below minimum",
			env: map[string]string{
				envPostgresDSN:      testPostgresDSN,
				envPostgresMaxConns: "0",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL max connections above maximum",
			env: map[string]string{
				envPostgresDSN:      testPostgresDSN,
				envPostgresMaxConns: "101",
			},
			wantErr: true,
		},
		{
			name: "invalid PostgreSQL minimum connections",
			env: map[string]string{
				envPostgresDSN:      testPostgresDSN,
				envPostgresMinConns: "few",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL minimum connections below minimum",
			env: map[string]string{
				envPostgresDSN:      testPostgresDSN,
				envPostgresMinConns: "-1",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL minimum connections exceeds maximum",
			env: map[string]string{
				envPostgresDSN:      testPostgresDSN,
				envPostgresMaxConns: "5",
				envPostgresMinConns: "6",
			},
			wantErr: true,
		},
		{
			name: "invalid PostgreSQL max connection lifetime",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnLifetime: "forever",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL max connection lifetime below minimum",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnLifetime: "30s",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL max connection lifetime above maximum",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnLifetime: "25h",
			},
			wantErr: true,
		},
		{
			name: "empty PostgreSQL max connection lifetime is rejected",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnLifetime: "",
			},
			wantErr: true,
		},
		{
			name: "invalid PostgreSQL max connection idle time",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnIdleTime: "later",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL max connection idle time below minimum",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnIdleTime: "10s",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL max connection idle time above maximum",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnIdleTime: "61m",
			},
			wantErr: true,
		},
		{
			name: "empty PostgreSQL max connection idle time is rejected",
			env: map[string]string{
				envPostgresDSN:             testPostgresDSN,
				envPostgresMaxConnIdleTime: "",
			},
			wantErr: true,
		},
		{
			name: "invalid PostgreSQL connect timeout",
			env: map[string]string{
				envPostgresDSN:            testPostgresDSN,
				envPostgresConnectTimeout: "quickly",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL connect timeout below minimum",
			env: map[string]string{
				envPostgresDSN:            testPostgresDSN,
				envPostgresConnectTimeout: "50ms",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL connect timeout above maximum",
			env: map[string]string{
				envPostgresDSN:            testPostgresDSN,
				envPostgresConnectTimeout: "31s",
			},
			wantErr: true,
		},
		{
			name: "empty PostgreSQL connect timeout is rejected",
			env: map[string]string{
				envPostgresDSN:            testPostgresDSN,
				envPostgresConnectTimeout: "",
			},
			wantErr: true,
		},
		{
			name: "invalid PostgreSQL readiness timeout",
			env: map[string]string{
				envPostgresDSN:              testPostgresDSN,
				envPostgresReadinessTimeout: "eventually",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL readiness timeout below minimum",
			env: map[string]string{
				envPostgresDSN:              testPostgresDSN,
				envPostgresReadinessTimeout: "50ms",
			},
			wantErr: true,
		},
		{
			name: "PostgreSQL readiness timeout above maximum",
			env: map[string]string{
				envPostgresDSN:              testPostgresDSN,
				envPostgresReadinessTimeout: "11s",
			},
			wantErr: true,
		},
		{
			name: "empty PostgreSQL readiness timeout is rejected",
			env: map[string]string{
				envPostgresDSN:              testPostgresDSN,
				envPostgresReadinessTimeout: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnvironment(t)

			for name, value := range tt.env {
				t.Setenv(name, value)
			}

			got, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("Load() returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		envEnvironment,
		envLogLevel,
		envShutdownTimeout,
		envHTTPAddress,
		envHTTPReadHeaderTimeout,
		envHTTPReadTimeout,
		envHTTPWriteTimeout,
		envHTTPIdleTimeout,
		envPostgresDSN,
		envPostgresMaxConns,
		envPostgresMinConns,
		envPostgresMaxConnLifetime,
		envPostgresMaxConnIdleTime,
		envPostgresConnectTimeout,
		envPostgresReadinessTimeout,
	} {
		oldValue, existed := os.LookupEnv(name)

		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("unset %s: %v", name, err)
		}

		t.Cleanup(func() {
			if existed {
				if err := os.Setenv(name, oldValue); err != nil {
					t.Errorf("restore %s: %v", name, err)
				}

				return
			}

			if err := os.Unsetenv(name); err != nil {
				t.Errorf("cleanup %s: %v", name, err)
			}
		})
	}
}
