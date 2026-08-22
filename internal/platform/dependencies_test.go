package platform

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"talenro.local/platform/internal/config"
)

func TestOpenReturnsSafeCategoryForDependencyErrors(t *testing.T) {
	privateDetail := "private dependency detail"
	tests := []struct {
		name   string
		mutate func(*dependencyOperations, error)
	}{
		{
			name: "create postgres",
			mutate: func(operations *dependencyOperations, cause error) {
				operations.newPostgres = func(context.Context, string) (*pgxpool.Pool, error) {
					return nil, cause
				}
			},
		},
		{
			name: "ping postgres",
			mutate: func(operations *dependencyOperations, cause error) {
				operations.pingPostgres = func(context.Context, *pgxpool.Pool) error { return cause }
			},
		},
		{
			name: "ping redis",
			mutate: func(operations *dependencyOperations, cause error) {
				operations.pingRedis = func(context.Context, *redis.Client) error { return cause }
			},
		},
		{
			name: "connect nats",
			mutate: func(operations *dependencyOperations, cause error) {
				operations.connectNATS = func(string, ...nats.Option) (*nats.Conn, error) {
					return nil, cause
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := errors.New(privateDetail)
			operations := successfulDependencyOperations()
			tt.mutate(&operations, cause)

			deps, err := openWithOperations(context.Background(), config.Config{
				DependencyTimeout: time.Second,
			}, operations)
			if deps != nil {
				deps.Close()
				t.Fatal("openWithOperations returned dependencies on failure")
			}
			if err == nil {
				t.Fatal("openWithOperations returned no error")
			}
			if got := err.Error(); got != string(CategoryDependencies) {
				t.Fatalf("error = %q, want %q", got, CategoryDependencies)
			}
			if strings.Contains(err.Error(), privateDetail) {
				t.Fatalf("error exposed dependency details: %q", err)
			}
			if !errors.Is(err, cause) {
				t.Fatal("categorized error did not retain internal cause")
			}
		})
	}
}

func successfulDependencyOperations() dependencyOperations {
	return dependencyOperations{
		newPostgres:  func(context.Context, string) (*pgxpool.Pool, error) { return nil, nil },
		pingPostgres: func(context.Context, *pgxpool.Pool) error { return nil },
		closePostgres: func(*pgxpool.Pool) {
		},
		newRedis:   func(*redis.Options) *redis.Client { return nil },
		pingRedis:  func(context.Context, *redis.Client) error { return nil },
		closeRedis: func(*redis.Client) error { return nil },
		connectNATS: func(string, ...nats.Option) (*nats.Conn, error) {
			return nil, nil
		},
		drainNATS: func(*nats.Conn) error { return nil },
		closeNATS: func(*nats.Conn) {
		},
	}
}

func TestOpenConfiguresRedisContextWithSocketBounds(t *testing.T) {
	operations := successfulDependencyOperations()
	var captured redis.Options
	operations.newRedis = func(options *redis.Options) *redis.Client {
		if options == nil {
			t.Fatal("Redis client options missing")
		}
		captured = *options
		return nil
	}

	dependencies, err := openWithOperations(context.Background(), config.Config{DependencyTimeout: time.Second}, operations)
	if err != nil {
		t.Fatal("dependency fixture rejected Redis client options")
	}
	defer dependencies.Close()
	if !captured.ContextTimeoutEnabled {
		t.Fatal("Redis client does not honor caller contexts")
	}
	if captured.DialTimeout != 2*time.Second || captured.ReadTimeout != 2*time.Second || captured.WriteTimeout != 2*time.Second {
		t.Fatal("Redis socket timeout bounds changed")
	}
}

func TestOpenConfiguresNATSForContinuousReconnect(t *testing.T) {
	operations := successfulDependencyOperations()
	captured := nats.GetDefaultOptions()
	operations.connectNATS = func(_ string, options ...nats.Option) (*nats.Conn, error) {
		for _, option := range options {
			if err := option(&captured); err != nil {
				t.Fatal("NATS connection option rejected")
			}
		}
		return nil, nil
	}

	dependencies, err := openWithOperations(context.Background(), config.Config{DependencyTimeout: time.Second}, operations)
	if err != nil {
		t.Fatal("dependency fixture rejected NATS connection options")
	}
	defer dependencies.Close()
	if captured.Timeout != 2*time.Second || captured.ReconnectWait != 500*time.Millisecond ||
		captured.MaxReconnect != -1 || captured.DrainTimeout != 5*time.Second {
		t.Fatal("NATS connection recovery bounds changed")
	}
}

func TestDependenciesCloseUsesReverseOrderOnce(t *testing.T) {
	var closed []string
	deps := &Dependencies{
		closePostgres: func() { closed = append(closed, "postgres") },
		closeRedis:    func() { closed = append(closed, "redis") },
		closeNATS:     func() { closed = append(closed, "nats") },
	}

	deps.Close()
	deps.Close()

	want := []string{"nats", "redis", "postgres"}
	if len(closed) != len(want) {
		t.Fatalf("close calls = %v, want %v", closed, want)
	}
	for i := range want {
		if closed[i] != want[i] {
			t.Fatalf("close calls = %v, want %v", closed, want)
		}
	}
}

func TestOpenHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	started := time.Now()
	deps, err := Open(ctx, config.Config{
		DatabaseURL:       "postgres://127.0.0.1:1/fixture?sslmode=disable",
		DependencyTimeout: time.Minute,
	})
	if err == nil {
		if deps != nil {
			deps.Close()
		}
		t.Fatal("Open returned no error for cancelled context")
	}
	if deps != nil {
		deps.Close()
		t.Fatal("Open returned dependencies on failure")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("Open took %s after cancellation, want at most 1s", elapsed)
	}
}
