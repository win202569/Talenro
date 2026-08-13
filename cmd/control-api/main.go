// Package main runs the Talenro control API process.
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/config"
	"talenro.local/platform/internal/controlapi"
	"talenro.local/platform/internal/observability"
	"talenro.local/platform/internal/platform"
	natsprobe "talenro.local/platform/internal/platform/nats"
	postgresprobe "talenro.local/platform/internal/platform/postgres"
	redisprobe "talenro.local/platform/internal/platform/redis"
	"talenro.local/platform/internal/readiness"
)

type openedRuntime struct {
	handler http.Handler
	close   func()
}

type runtimeFactory func(context.Context, config.Config) (*openedRuntime, error)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.LookupEnv); err != nil {
		reportStopped(err)
		os.Exit(1)
	}
}

func reportStopped(err error) {
	category := platform.ErrorCategoryOf(err, platform.CategoryInternal)
	slog.Error("control_api_stopped", "category", category)
}

func run(ctx context.Context, lookup config.Lookup) error {
	return runWithFactory(ctx, lookup, openRuntime)
}

func runWithFactory(ctx context.Context, lookup config.Lookup, factory runtimeFactory) error {
	cfg, err := config.Load(lookup)
	if err != nil {
		return platform.NewCategorizedError(platform.CategoryConfiguration, err)
	}

	runtime, err := factory(ctx, cfg)
	if err != nil {
		return platform.NewCategorizedError(platform.CategoryDependencies, err)
	}
	defer runtime.close()

	slog.Info("control_api_starting", "category", platform.CategoryStartup)

	metrics := observability.NewRegistry()
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.Handler())
	metricsServer := platform.NewHTTPServer(cfg.MetricsAddress, metricsMux)
	publicServer := platform.NewHTTPServer(cfg.HTTPAddress, metrics.Middleware("", runtime.handler))

	errCh := make(chan error, 2)
	go func() {
		errCh <- metricsServer.ListenAndServe()
	}()
	go func() {
		errCh <- publicServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return shutdownHTTPServers(ctx, metricsServer, publicServer, cfg.ShutdownTimeout)
	case err := <-errCh:
		shutdownErr := shutdownHTTPServers(ctx, metricsServer, publicServer, cfg.ShutdownTimeout)
		if errors.Is(err, http.ErrServerClosed) {
			return shutdownErr
		}
		return platform.NewCategorizedError(platform.CategoryHTTPListenOrServe, err)
	}
}

func shutdownHTTPServer(server *http.Server, timeout time.Duration) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return shutdownHTTPServerWithContext(shutdownCtx, server)
}

func shutdownHTTPServers(
	lifecycleCtx context.Context,
	metricsServer, publicServer *http.Server,
	timeout time.Duration,
) error {
	shutdownCtx, cancel := deriveShutdownContext(lifecycleCtx, timeout)
	defer cancel()

	metricsErr := shutdownHTTPServerWithContext(shutdownCtx, metricsServer)
	publicErr := shutdownHTTPServerWithContext(shutdownCtx, publicServer)
	if metricsErr != nil {
		return metricsErr
	}
	return publicErr
}

func deriveShutdownContext(lifecycleCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(lifecycleCtx), timeout)
}

func shutdownHTTPServerWithContext(shutdownCtx context.Context, server *http.Server) error {
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Warn("control_api_shutdown_failed", "category", platform.CategoryHTTPShutdown)
		slog.Warn("control_api_forced_close", "category", platform.CategoryHTTPForcedClose)
		if closeErr := server.Close(); closeErr != nil {
			return platform.NewCategorizedError(platform.CategoryHTTPForcedClose, closeErr)
		}
		return platform.NewCategorizedError(platform.CategoryHTTPShutdown, err)
	}
	return nil
}

func openRuntime(ctx context.Context, cfg config.Config) (*openedRuntime, error) {
	deps, err := platform.Open(ctx, cfg)
	if err != nil {
		return nil, err
	}

	checker := readiness.New(cfg.DependencyTimeout,
		postgresprobe.Probe{Pool: deps.Postgres},
		redisprobe.Probe{Client: deps.Redis},
		natsprobe.Probe{Conn: deps.NATS},
	)
	mux := http.NewServeMux()
	apiHandler := controlapi.NewHandler(checker, controlapi.Applications{}, cfg.Security.RequestDeadline, rand.Reader)
	handler := controlapiv1.HandlerWithOptions(apiHandler, controlapiv1.StdHTTPServerOptions{
		BaseRouter: mux, ErrorHandlerFunc: apiHandler.GeneratedParameterError,
	})

	return &openedRuntime{
		handler: handler,
		close:   deps.Close,
	}, nil
}
