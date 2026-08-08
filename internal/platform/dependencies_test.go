package platform

import (
	"context"
	"testing"
	"time"

	"talenro.local/platform/internal/config"
)

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
