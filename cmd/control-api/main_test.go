package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"talenro.local/platform/internal/config"
)

func TestRunShutsDownHTTPServerAndClosesRuntime(t *testing.T) {
	address := unusedLocalAddress(t)
	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://unused",
		"TALENRO_HTTP_ADDRESS": address,
	})

	var closeCalls atomic.Int32
	factory := func(context.Context, config.Config) (*openedRuntime, error) {
		return &openedRuntime{
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			close: func() { closeCalls.Add(1) },
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errCh := make(chan error, 1)
	go func() { errCh <- runWithFactory(ctx, lookup, factory) }()
	waitForServer(t, "http://"+address)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runWithFactory returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWithFactory did not return within 2s of cancellation")
	}
	if got := closeCalls.Load(); got != 1 {
		t.Fatalf("runtime close calls = %d, want 1", got)
	}

	client := &http.Client{Timeout: 100 * time.Millisecond}
	if response, err := client.Get("http://" + address); err == nil {
		_ = response.Body.Close()
		t.Fatal("HTTP server still accepted requests after shutdown")
	}
}

func TestRunClosesRuntimeWhenHTTPServerFails(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()

	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://unused",
		"TALENRO_HTTP_ADDRESS": listener.Addr().String(),
	})
	var closeCalls atomic.Int32
	factory := func(context.Context, config.Config) (*openedRuntime, error) {
		return &openedRuntime{
			handler: http.NewServeMux(),
			close:   func() { closeCalls.Add(1) },
		}, nil
	}

	err = runWithFactory(context.Background(), lookup, factory)
	if err == nil || !strings.Contains(err.Error(), "serve HTTP") {
		t.Fatalf("runWithFactory error = %v, want serve HTTP category", err)
	}
	if got := closeCalls.Load(); got != 1 {
		t.Fatalf("runtime close calls = %d, want 1", got)
	}
}

func unusedLocalAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func testLookup(values map[string]string) config.Lookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func waitForServer(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	for time.Now().Before(deadline) {
		response, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("HTTP server did not become ready within 2s")
}
