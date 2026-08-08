// Package observability provides bounded, privacy-safe service metrics.
package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"talenro.local/platform/internal/buildinfo"
)

type Registry struct {
	prometheus.Registerer
	Gatherer  prometheus.Gatherer
	Requests  *prometheus.CounterVec
	Duration  *prometheus.HistogramVec
	BuildInfo *prometheus.GaugeVec
}

func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "talenro_control_http_requests_total",
		Help: "Completed control API HTTP requests.",
	}, []string{"method", "route", "status_class"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "talenro_control_http_request_duration_seconds",
		Help:    "Control API HTTP request duration.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status_class"})
	build := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "talenro_control_build_info",
		Help: "Talenro control API build identity.",
	}, []string{"version", "commit"})
	info := buildinfo.Current()
	build.WithLabelValues(info.Version, info.Commit).Set(1)
	registry.MustRegister(requests, duration, build)
	return &Registry{
		Registerer: registry,
		Gatherer:   registry,
		Requests:   requests,
		Duration:   duration,
		BuildInfo:  build,
	}
}

func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.Gatherer, promhttp.HandlerOpts{})
}
