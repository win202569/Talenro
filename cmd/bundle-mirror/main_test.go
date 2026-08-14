package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/platform"
)

// Removing the narrow mirror-specific loader, or consulting any provider,
// Redis, NATS, signer, or key setting, must make this test fail.
func TestLoadMirrorConfigReadsOnlyDatabaseAndListenAddress(t *testing.T) {
	values := map[string]string{
		"TALENRO_DATABASE_URL":              "postgres://mirror-database",
		"TALENRO_HTTP_ADDRESS":              "127.0.0.1:18081",
		"TALENRO_REDIS_ADDRESS":             "UNRELATED-SECRET-CANARY",
		"TALENRO_NATS_URL":                  "UNRELATED-SECRET-CANARY",
		"TALENRO_SIGNER_PROVIDER":           "UNRELATED-SECRET-CANARY",
		"TALENRO_FIELD_PROTECTOR_PROVIDER":  "UNRELATED-SECRET-CANARY",
		"TALENRO_SENSITIVE_LOOKUP_KEY":      "UNRELATED-SECRET-CANARY",
		"TALENRO_SENSITIVE_ENCRYPTION_KEY":  "UNRELATED-SECRET-CANARY",
		"TALENRO_LOCAL_ROOT_SIGNING_SEED":   "UNRELATED-SECRET-CANARY",
		"TALENRO_LOCAL_CONFIG_SIGNING_SEED": "UNRELATED-SECRET-CANARY",
		"TALENRO_EMAIL_PROVIDER":            "UNRELATED-SECRET-CANARY",
		"TALENRO_ERROR_REPORTER_PROVIDER":   "UNRELATED-SECRET-CANARY",
		"TALENRO_DEPENDENCY_TIMEOUT":        "1ns",
		"TALENRO_SHUTDOWN_TIMEOUT":          "1ns",
	}
	var read []string
	lookup := func(key string) (string, bool) {
		read = append(read, key)
		value, ok := values[key]
		return value, ok
	}

	cfg, err := loadMirrorConfig(lookup)
	if err != nil {
		t.Fatalf("loadMirrorConfig: %v", err)
	}
	if want := []string{"TALENRO_DATABASE_URL", "TALENRO_HTTP_ADDRESS"}; !reflect.DeepEqual(read, want) {
		t.Fatalf("environment keys read = %v, want %v", read, want)
	}
	if cfg.databaseURL != "postgres://mirror-database" || cfg.address != "127.0.0.1:18081" {
		t.Fatalf("mirror config = %+v", cfg)
	}
	if cfg.dependencyTimeout != 2*time.Second || cfg.shutdownTimeout != 10*time.Second {
		t.Fatalf("fixed budgets = (%s, %s), want (2s, 10s)", cfg.dependencyTimeout, cfg.shutdownTimeout)
	}
	if strings.Contains(cfg.String(), "UNRELATED-SECRET-CANARY") {
		t.Fatalf("unrelated provider material reached mirror config: %s", cfg.String())
	}
}

func TestLoadMirrorConfigRejectsInvalidInputWithFixedError(t *testing.T) {
	tests := []map[string]string{
		{},
		{"TALENRO_DATABASE_URL": "postgres://mirror", "TALENRO_HTTP_ADDRESS": "not-a-listener"},
		{"TALENRO_DATABASE_URL": "postgres://mirror", "TALENRO_HTTP_ADDRESS": "127.0.0.1:0"},
	}
	for _, values := range tests {
		_, err := loadMirrorConfig(task18MirrorLookup(values))
		if !errors.Is(err, errMirrorConfiguration) || err.Error() != "bundle-mirror: configuration" {
			t.Fatalf("loadMirrorConfig(%v) error = %v", values, err)
		}
	}
}

// Replacing the mirror-specific PostgreSQL opener with the control-plane
// dependency bundle, skipping Ping, or omitting the deadline breaks this test.
func TestOpenMirrorRuntimeUsesOnlyBoundedPostgresAndRealMirrorHandler(t *testing.T) {
	database := &task18MirrorDatabase{}
	var openedURL string
	operations := mirrorDatabaseOperations{
		openPostgres: func(ctx context.Context, databaseURL string) (mirrorDatabase, error) {
			openedURL = databaseURL
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("PostgreSQL open context has no deadline")
			}
			remaining := time.Until(deadline)
			if remaining <= 0 || remaining > 2*time.Second {
				t.Fatalf("PostgreSQL open deadline remaining = %s", remaining)
			}
			return database, nil
		},
	}
	cfg := mirrorConfig{
		databaseURL:       "postgres://mirror-only",
		address:           "127.0.0.1:18081",
		dependencyTimeout: 2 * time.Second,
		shutdownTimeout:   10 * time.Second,
	}

	runtime, err := openMirrorRuntimeWithOperations(t.Context(), cfg, operations)
	if err != nil {
		t.Fatalf("openMirrorRuntimeWithOperations: %v", err)
	}
	if openedURL != cfg.databaseURL || database.pingCalls.Load() != 1 {
		t.Fatalf("opened URL = %q, ping calls = %d", openedURL, database.pingCalls.Load())
	}

	locator := bytes.Repeat([]byte{1}, 32)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/"+base64.RawURLEncoding.EncodeToString(locator), nil)
	response := httptest.NewRecorder()
	runtime.handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound || response.Body.Len() != 0 {
		t.Fatalf("mirror response = (%d, %q), want body-free 404", response.Code, response.Body.String())
	}
	if database.queryRowCalls.Load() != 1 {
		t.Fatalf("PostgreSQL immutable lookup calls = %d, want 1", database.queryRowCalls.Load())
	}

	runtime.Close()
	runtime.Close()
	if database.closeCalls.Load() != 1 {
		t.Fatalf("PostgreSQL close calls = %d, want 1", database.closeCalls.Load())
	}
}

func TestOpenMirrorRuntimeClosesPostgresAfterPingFailure(t *testing.T) {
	private := errors.New("DATABASE-PRIVATE-CANARY")
	database := &task18MirrorDatabase{pingErr: private}
	cfg := mirrorConfig{databaseURL: "postgres://mirror", dependencyTimeout: 2 * time.Second}
	_, err := openMirrorRuntimeWithOperations(t.Context(), cfg, mirrorDatabaseOperations{
		openPostgres: func(context.Context, string) (mirrorDatabase, error) { return database, nil },
	})
	if !errors.Is(err, private) {
		t.Fatalf("open error = %v, want retained private cause", err)
	}
	if database.closeCalls.Load() != 1 {
		t.Fatalf("PostgreSQL close calls = %d, want 1", database.closeCalls.Load())
	}
}

func TestNewMirrorHTTPServerUsesFixedResourceBudgets(t *testing.T) {
	handler := http.NewServeMux()
	server := newMirrorHTTPServer("127.0.0.1:18081", handler)
	if server.Addr != "127.0.0.1:18081" || server.Handler != handler {
		t.Fatalf("server address/handler not preserved")
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 10*time.Second ||
		server.WriteTimeout != 10*time.Second || server.IdleTimeout != 60*time.Second ||
		server.MaxHeaderBytes != 16<<10 {
		t.Fatalf("HTTP budgets = header %s read %s write %s idle %s bytes %d",
			server.ReadHeaderTimeout, server.ReadTimeout, server.WriteTimeout, server.IdleTimeout, server.MaxHeaderBytes)
	}
}

// Closing the database before Shutdown stops the listener makes the close
// callback observe a live listener and fail this test.
func TestRunStopsListenerBeforeClosingDatabase(t *testing.T) {
	address := task18UnusedMirrorAddress(t)
	database := &task18MirrorDatabase{}
	database.onClose = func() {
		connection, err := (&net.Dialer{Timeout: 100 * time.Millisecond}).DialContext(t.Context(), "tcp", address)
		if err == nil {
			_ = connection.Close()
			database.listenerLiveOnClose.Store(true)
		}
	}
	factory := func(context.Context, mirrorConfig) (*mirrorRuntime, error) {
		return &mirrorRuntime{
			handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusNoContent)
			}),
			database: database,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runMirrorWithFactory(ctx, task18MirrorLookup(map[string]string{
			"TALENRO_DATABASE_URL": "postgres://unused",
			"TALENRO_HTTP_ADDRESS": address,
		}), factory)
	}()
	task18WaitForMirror(t, "http://"+address)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runMirrorWithFactory: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("mirror did not stop within 2s")
	}
	if database.closeCalls.Load() != 1 || database.listenerLiveOnClose.Load() {
		t.Fatalf("close calls = %d, listener live on close = %t",
			database.closeCalls.Load(), database.listenerLiveOnClose.Load())
	}
}

func TestRunAndReportHideDependencyCause(t *testing.T) {
	logs := task18CaptureMirrorLogs(t)
	private := errors.New("DATABASE-PRIVATE-CANARY")
	err := runMirrorWithFactory(context.Background(), task18MirrorLookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://unused",
	}), func(context.Context, mirrorConfig) (*mirrorRuntime, error) {
		return nil, private
	})
	reportMirrorStopped(err)
	if err == nil || err.Error() != string(platform.CategoryDependencies) || !errors.Is(err, private) {
		t.Fatalf("run error = %v", err)
	}
	if output := logs.String(); !strings.Contains(output, "bundle_mirror_stopped") ||
		!strings.Contains(output, string(platform.CategoryDependencies)) || strings.Contains(output, private.Error()) {
		t.Fatalf("unsafe stopped log = %q", output)
	}
}

func TestRunHidesListenCauseAndStillClosesDatabase(t *testing.T) {
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	database := &task18MirrorDatabase{}
	err = runMirrorWithFactory(context.Background(), task18MirrorLookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://unused",
		"TALENRO_HTTP_ADDRESS": listener.Addr().String(),
	}), func(context.Context, mirrorConfig) (*mirrorRuntime, error) {
		return &mirrorRuntime{handler: http.NewServeMux(), database: database}, nil
	})
	if err == nil || err.Error() != string(platform.CategoryHTTPListenOrServe) {
		t.Fatalf("listen error = %v", err)
	}
	if database.closeCalls.Load() != 1 {
		t.Fatalf("PostgreSQL close calls = %d, want 1", database.closeCalls.Load())
	}
}

func TestShutdownMirrorServerForcesCloseWithFixedCategory(t *testing.T) {
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	server := newMirrorHTTPServer(listener.Addr().String(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-release
	}))
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		request, requestErr := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+listener.Addr().String(), nil)
		if requestErr != nil {
			return
		}
		response, requestErr := http.DefaultClient.Do(request) // #nosec G107 -- loopback test server.
		if requestErr == nil {
			_ = response.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("blocking handler was not entered")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	err = shutdownMirrorServerWithContext(shutdownCtx, server)
	cancel()
	close(release)
	if err == nil || err.Error() != string(platform.CategoryHTTPShutdown) {
		t.Fatalf("shutdown error = %v", err)
	}
	select {
	case <-serveDone:
	case <-time.After(time.Second):
		t.Fatal("server did not stop after forced close")
	}
	select {
	case <-requestDone:
	case <-time.After(time.Second):
		t.Fatal("request did not stop after forced close")
	}
}

type task18MirrorDatabase struct {
	pingErr             error
	onClose             func()
	pingCalls           atomic.Int32
	queryRowCalls       atomic.Int32
	closeCalls          atomic.Int32
	listenerLiveOnClose atomic.Bool
}

func (database *task18MirrorDatabase) Ping(context.Context) error {
	database.pingCalls.Add(1)
	return database.pingErr
}

func (*task18MirrorDatabase) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (*task18MirrorDatabase) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (database *task18MirrorDatabase) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	database.queryRowCalls.Add(1)
	return task18MirrorRow{}
}

func (database *task18MirrorDatabase) Close() {
	database.closeCalls.Add(1)
	if database.onClose != nil {
		database.onClose()
	}
}

type task18MirrorRow struct{}

func (task18MirrorRow) Scan(...interface{}) error { return pgx.ErrNoRows }

func task18MirrorLookup(values map[string]string) mirrorLookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func task18UnusedMirrorAddress(t *testing.T) string {
	t.Helper()
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func task18WaitForMirror(t *testing.T, target string) {
	t.Helper()
	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("mirror %s did not start", target)
}

func task18CaptureMirrorLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buffer bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buffer, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &buffer
}

var _ mirrorDatabase = (*task18MirrorDatabase)(nil)
