package observability

import (
	"net/http"
	"strconv"

	"github.com/felixge/httpsnoop"
)

const unmatchedRoute = "unmatched"

func (r *Registry) Middleware(route string, next http.Handler) http.Handler {
	dynamicRoute := route == ""
	route = boundedRoute(route)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		captured := httpsnoop.CaptureMetrics(next, w, req)
		statusClass := strconv.Itoa(captured.Code/100) + "xx"
		requestRoute := route
		if dynamicRoute {
			requestRoute = boundedRoute(req.Pattern)
		}
		labels := []string{req.Method, requestRoute, statusClass}
		r.Requests.WithLabelValues(labels...).Inc()
		r.Duration.WithLabelValues(labels...).Observe(captured.Duration.Seconds())
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
