// Package main runs the dependency-isolated immutable bundle mirror.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/platform"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/trust"
)

const (
	defaultMirrorAddress  = "127.0.0.1:8081"
	mirrorDependencyLimit = 2 * time.Second
	mirrorShutdownLimit   = 10 * time.Second
)

var (
	errMirrorConfiguration = errors.New("bundle-mirror: configuration")
	errMirrorRuntime       = errors.New("bundle-mirror: runtime")
)

type mirrorLookup func(string) (string, bool)

// mirrorConfig deliberately has no provider, key, Redis, NATS, or control API
// fields. The two durations are fixed process budgets, not environment input.
type mirrorConfig struct {
	databaseURL       string
	address           string
	dependencyTimeout time.Duration
	shutdownTimeout   time.Duration
}

func (mirrorConfig) String() string { return "bundle-mirror.Config([REDACTED])" }

func loadMirrorConfig(lookup mirrorLookup) (mirrorConfig, error) {
	if lookup == nil {
		return mirrorConfig{}, errMirrorConfiguration
	}
	cfg := mirrorConfig{
		address:           defaultMirrorAddress,
		dependencyTimeout: mirrorDependencyLimit,
		shutdownTimeout:   mirrorShutdownLimit,
	}
	var ok bool
	cfg.databaseURL, ok = lookup("TALENRO_DATABASE_URL")
	if !ok || cfg.databaseURL == "" {
		return mirrorConfig{}, errMirrorConfiguration
	}
	if value, exists := lookup("TALENRO_HTTP_ADDRESS"); exists && value != "" {
		cfg.address = value
	}
	if !validMirrorAddress(cfg.address) {
		return mirrorConfig{}, errMirrorConfiguration
	}
	return cfg, nil
}

func validMirrorAddress(address string) bool {
	_, port, err := net.SplitHostPort(address)
	if err != nil || port == "" {
		return false
	}
	value, err := strconv.Atoi(port)
	return err == nil && value >= 1 && value <= 65535 && strconv.Itoa(value) == port
}

type mirrorDatabase interface {
	store.DBTX
	Ping(context.Context) error
	Close()
}

type mirrorDatabaseOperations struct {
	openPostgres func(context.Context, string) (mirrorDatabase, error)
}

type mirrorRuntime struct {
	handler  http.Handler
	database mirrorDatabase
	close    sync.Once
}

func (runtime *mirrorRuntime) Close() {
	if runtime == nil {
		return
	}
	runtime.close.Do(func() {
		if !nilMirrorDatabase(runtime.database) {
			runtime.database.Close()
		}
	})
}

type mirrorRuntimeFactory func(context.Context, mirrorConfig) (*mirrorRuntime, error)

type mirrorClock struct{}

var _ securitykit.Clock = mirrorClock{}

func (mirrorClock) Now() time.Time { return time.Now().UTC() }

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runMirror(ctx, os.LookupEnv); err != nil {
		reportMirrorStopped(err)
		os.Exit(1)
	}
}

func runMirror(ctx context.Context, lookup mirrorLookup) error {
	return runMirrorWithFactory(ctx, lookup, openMirrorRuntime)
}

func runMirrorWithFactory(ctx context.Context, lookup mirrorLookup, factory mirrorRuntimeFactory) error {
	if ctx == nil || factory == nil {
		return platform.NewCategorizedError(platform.CategoryConfiguration, errMirrorConfiguration)
	}
	cfg, err := loadMirrorConfig(lookup)
	if err != nil {
		return platform.NewCategorizedError(platform.CategoryConfiguration, err)
	}
	runtime, err := factory(ctx, cfg)
	if err != nil {
		return platform.NewCategorizedError(platform.CategoryDependencies, err)
	}
	if runtime == nil || runtime.handler == nil || nilMirrorDatabase(runtime.database) {
		if runtime != nil {
			runtime.Close()
		}
		return platform.NewCategorizedError(platform.CategoryDependencies, errMirrorRuntime)
	}
	defer runtime.Close()

	server := newMirrorHTTPServer(cfg.address, runtime.handler)
	slog.Info("bundle_mirror_starting", "category", platform.CategoryStartup)
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.ListenAndServe() }()

	select {
	case <-ctx.Done():
		return shutdownMirrorServer(ctx, server, cfg.shutdownTimeout)
	case serveErr := <-serveResult:
		shutdownErr := shutdownMirrorServer(ctx, server, cfg.shutdownTimeout)
		if errors.Is(serveErr, http.ErrServerClosed) {
			return shutdownErr
		}
		return platform.NewCategorizedError(platform.CategoryHTTPListenOrServe, serveErr)
	}
}

func reportMirrorStopped(err error) {
	category := platform.ErrorCategoryOf(err, platform.CategoryInternal)
	slog.Error("bundle_mirror_stopped", "category", category)
}

func openMirrorRuntime(ctx context.Context, cfg mirrorConfig) (*mirrorRuntime, error) {
	return openMirrorRuntimeWithOperations(ctx, cfg, mirrorDatabaseOperations{
		openPostgres: func(openCtx context.Context, databaseURL string) (mirrorDatabase, error) {
			return pgxpool.New(openCtx, databaseURL)
		},
	})
}

func openMirrorRuntimeWithOperations(
	ctx context.Context,
	cfg mirrorConfig,
	operations mirrorDatabaseOperations,
) (_ *mirrorRuntime, resultErr error) {
	if ctx == nil || ctx.Err() != nil || cfg.databaseURL == "" || cfg.dependencyTimeout <= 0 ||
		operations.openPostgres == nil {
		return nil, errMirrorConfiguration
	}
	openCtx, cancel := context.WithTimeout(ctx, cfg.dependencyTimeout)
	defer cancel()
	database, err := operations.openPostgres(openCtx, cfg.databaseURL)
	if err != nil {
		return nil, err
	}
	if nilMirrorDatabase(database) {
		return nil, errMirrorRuntime
	}
	defer func() {
		if resultErr != nil {
			database.Close()
		}
	}()
	if err := database.Ping(openCtx); err != nil {
		return nil, err
	}
	byteStore, err := trust.NewPostgresByteStore(database)
	if err != nil {
		return nil, errMirrorRuntime
	}
	handler, err := trust.NewMirror(byteStore, mirrorClock{})
	if err != nil {
		return nil, errMirrorRuntime
	}
	return &mirrorRuntime{handler: handler, database: database}, nil
}

func newMirrorHTTPServer(address string, handler http.Handler) *http.Server {
	return platform.NewHTTPServer(address, handler)
}

func shutdownMirrorServer(lifecycleCtx context.Context, server *http.Server, timeout time.Duration) error {
	if lifecycleCtx == nil {
		return platform.NewCategorizedError(platform.CategoryHTTPShutdown, errMirrorRuntime)
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(lifecycleCtx), timeout)
	defer cancel()
	return shutdownMirrorServerWithContext(shutdownCtx, server)
}

func shutdownMirrorServerWithContext(shutdownCtx context.Context, server *http.Server) error {
	if shutdownCtx == nil || server == nil {
		return platform.NewCategorizedError(platform.CategoryHTTPShutdown, errMirrorRuntime)
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Warn("bundle_mirror_shutdown_failed", "category", platform.CategoryHTTPShutdown)
		slog.Warn("bundle_mirror_forced_close", "category", platform.CategoryHTTPForcedClose)
		if closeErr := server.Close(); closeErr != nil {
			return platform.NewCategorizedError(platform.CategoryHTTPForcedClose, closeErr)
		}
		return platform.NewCategorizedError(platform.CategoryHTTPShutdown, err)
	}
	return nil
}

func nilMirrorDatabase(database mirrorDatabase) bool {
	if database == nil {
		return true
	}
	value := reflect.ValueOf(database)
	kind := value.Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && value.IsNil()
}
