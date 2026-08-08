package platform

import (
	"log"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type sanitizedServerErrorWriter struct {
	logger    *slog.Logger
	reporting atomic.Bool
}

func (w *sanitizedServerErrorWriter) Write(message []byte) (int, error) {
	if w.reporting.CompareAndSwap(false, true) {
		w.logger.Error("http_server_error", "category", CategoryHTTPInternal)
		w.reporting.Store(false)
	}
	return len(message), nil
}

// NewHTTPServer constructs a bounded, privacy-safe HTTP server.
func NewHTTPServer(address string, handler http.Handler) *http.Server {
	errorWriter := &sanitizedServerErrorWriter{logger: slog.Default()}
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
		ErrorLog:          log.New(errorWriter, "", 0),
	}
}
