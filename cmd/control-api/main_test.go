package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/platform"
)

func TestRunReportsSanitizedDependencyError(t *testing.T) {
	logs := captureLogs(t)
	private := "credential@private.example:6543"
	cause := errors.New(private)
	lookup := testLookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://unused",
	})
	factory := func(context.Context, config.Config) (*openedRuntime, error) {
		return nil, cause
	}

	err := runWithFactory(context.Background(), lookup, factory)
	reportStopped(err)
	if got := err.Error(); got != string(platform.CategoryDependencies) {
		t.Fatalf("error = %q, want %q", got, platform.CategoryDependencies)
	}
	if !errors.Is(err, cause) {
		t.Fatal("categorized error did not retain internal cause")
	}
	assertSanitizedLog(t, logs.String(), "control_api_stopped", string(platform.CategoryDependencies), private)
}

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
	logs := captureLogs(t)
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
	reportStopped(err)
	if err == nil || err.Error() != string(platform.CategoryHTTPListenOrServe) {
		t.Fatalf("runWithFactory error = %v, want %q", err, platform.CategoryHTTPListenOrServe)
	}
	if got := closeCalls.Load(); got != 1 {
		t.Fatalf("runtime close calls = %d, want 1", got)
	}
	assertSanitizedLog(
		t,
		logs.String(),
		"control_api_stopped",
		string(platform.CategoryHTTPListenOrServe),
		listener.Addr().String(),
	)
}

func TestShutdownForcesClosedBlockingHandler(t *testing.T) {
	logs := captureLogs(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	requestStarted := make(chan struct{})
	handlerFinished := make(chan struct{})
	handler := http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(requestStarted)
		<-request.Context().Done()
		close(handlerFinished)
	})
	server := platform.NewHTTPServer(listener.Addr().String(), handler)
	serveDone := make(chan struct{})
	go func() {
		_ = server.Serve(listener)
		close(serveDone)
	}()

	clientDone := make(chan struct{})
	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		response, requestErr := client.Get("http://" + listener.Addr().String())
		if response != nil {
			_ = response.Body.Close()
		}
		_ = requestErr
		close(clientDone)
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("blocking handler did not start within 1s")
	}

	started := time.Now()
	err = shutdownHTTPServer(server, 50*time.Millisecond)
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("forced shutdown took %s, want at most 1s", elapsed)
	}
	if err == nil || err.Error() != string(platform.CategoryHTTPShutdown) {
		t.Fatalf("shutdown error = %v, want %q", err, platform.CategoryHTTPShutdown)
	}

	for name, done := range map[string]<-chan struct{}{
		"handler": handlerFinished,
		"client":  clientDone,
		"server":  serveDone,
	} {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("%s remained active after forced close", name)
		}
	}
	assertSanitizedLog(
		t,
		logs.String(),
		"control_api_forced_close",
		string(platform.CategoryHTTPForcedClose),
		listener.Addr().String(),
	)
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &logs
}

func assertSanitizedLog(t *testing.T, output, event, category, private string) {
	t.Helper()
	if !strings.Contains(output, "msg="+event) || !strings.Contains(output, "category="+category) {
		t.Fatalf("event/category missing from log: %q", output)
	}
	if strings.Contains(output, private) {
		t.Fatalf("private detail %q leaked in log: %q", private, output)
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
