package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/buildinfo"
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
		slog.Error("control_api_stopped", "category", "startup_or_runtime")
		os.Exit(1)
	}
}

func run(ctx context.Context, lookup config.Lookup) error {
	return runWithFactory(ctx, lookup, openRuntime)
}

func runWithFactory(ctx context.Context, lookup config.Lookup, factory runtimeFactory) error {
	cfg, err := config.Load(lookup)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	runtime, err := factory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open dependencies: %w", err)
	}
	defer runtime.close()

	info := buildinfo.Current()
	slog.Info("control_api_starting",
		"version", info.Version,
		"commit", info.Commit,
		"built_at", info.BuiltAt,
		"listener", cfg.HTTPAddress,
	)

	server := platform.NewHTTPServer(cfg.HTTPAddress, runtime.handler)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	}
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
