package main

import (
	"context"
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

	server := platform.NewHTTPServer(cfg.HTTPAddress, runtime.handler)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return shutdownHTTPServer(server, cfg.ShutdownTimeout)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return platform.NewCategorizedError(platform.CategoryHTTPListenOrServe, err)
	}
}

func shutdownHTTPServer(server *http.Server, timeout time.Duration) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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
	handler := controlapiv1.HandlerFromMux(controlapi.NewHandler(checker), mux)

	return &openedRuntime{
		handler: handler,
		close:   deps.Close,
	}, nil
}
