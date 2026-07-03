package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults",
			want: Config{
				Environment:     "local",
				LogLevel:        "info",
				ShutdownTimeout: 10 * time.Second,
			},
		},
		{
			name: "custom values",
			env: map[string]string{
				envEnvironment:     "production",
				envLogLevel:        "debug",
				envShutdownTimeout: "30s",
			},
			want: Config{
				Environment:     "production",
				LogLevel:        "debug",
				ShutdownTimeout: 30 * time.Second,
			},
		},
		{
			name: "values are normalized",
			env: map[string]string{
				envEnvironment:     " Staging ",
				envLogLevel:        " WARN ",
				envShutdownTimeout: " 15s ",
			},
			want: Config{
				Environment:     "staging",
				LogLevel:        "warn",
				ShutdownTimeout: 15 * time.Second,
			},
		},
		{
			name: "unsupported environment",
			env: map[string]string{
				envEnvironment: "development",
			},
			wantErr: true,
		},
		{
			name: "unsupported log level",
			env: map[string]string{
				envLogLevel: "trace",
			},
			wantErr: true,
		},
		{
			name: "invalid shutdown timeout",
			env: map[string]string{
				envShutdownTimeout: "soon",
			},
			wantErr: true,
		},
		{
			name: "shutdown timeout below minimum",
			env: map[string]string{
				envShutdownTimeout: "500ms",
			},
			wantErr: true,
		},
		{
			name: "shutdown timeout above maximum",
			env: map[string]string{
				envShutdownTimeout: "3m",
			},
			wantErr: true,
		},
		{
			name: "empty environment is rejected",
			env: map[string]string{
				envEnvironment: "",
			},
			wantErr: true,
		},
		{
			name: "empty log level is rejected",
			env: map[string]string{
				envLogLevel: "",
			},
			wantErr: true,
		},
		{
			name: "empty shutdown timeout is rejected",
			env: map[string]string{
				envShutdownTimeout: "",
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
