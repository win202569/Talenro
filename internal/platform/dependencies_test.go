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
	secret := "credential@private.example:6543"
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
			cause := errors.New(secret)
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
			if strings.Contains(err.Error(), secret) {
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
		DatabaseURL:       "postgres://ignored:ignored@127.0.0.1:1/ignored",
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
