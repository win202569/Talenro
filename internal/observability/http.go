package observability

import (
	"net/http"
	"strconv"
	"time"
)

const unmatchedRoute = "unmatched"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (r *Registry) Middleware(route string, next http.Handler) http.Handler {
	dynamicRoute := route == ""
	route = boundedRoute(route)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, req)
		statusClass := strconv.Itoa(recorder.status/100) + "xx"
		requestRoute := route
		if dynamicRoute {
			requestRoute = boundedRoute(req.Pattern)
		}
		labels := []string{req.Method, requestRoute, statusClass}
		r.Requests.WithLabelValues(labels...).Inc()
		r.Duration.WithLabelValues(labels...).Observe(time.Since(started).Seconds())
	})
}

func boundedRoute(route string) string {
	switch route {
	case "/livez", "GET /livez":
		return "/livez"
	case "/readyz", "GET /readyz":
		return "/readyz"
	default:
		return unmatchedRoute
	}
}
