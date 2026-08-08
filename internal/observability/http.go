package observability

import (
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/felixge/httpsnoop"
)

const unmatchedRoute = "unmatched"

// Middleware records bounded request metrics around next.
func (r *Registry) Middleware(route string, next http.Handler) http.Handler {
	dynamicRoute := route == ""
	route = boundedRoute(route)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		started := time.Now()
		var status atomic.Int64
		commitOK := func() {
			status.CompareAndSwap(0, http.StatusOK)
		}
		wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					next(code)
					if code < 100 || code > 199 {
						status.CompareAndSwap(0, int64(code))
					}
				}
			},
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(body []byte) (int, error) {
					written, err := next(body)
					commitOK()
					return written, err
				}
			},
			WriteString: func(next httpsnoop.WriteStringFunc) httpsnoop.WriteStringFunc {
				return func(body string) (int, error) {
					written, err := next(body)
					commitOK()
					return written, err
				}
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(source io.Reader) (int64, error) {
					written, err := next(source)
					commitOK()
					return written, err
				}
			},
			Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
				return func() {
					next()
					commitOK()
				}
			},
			FlushError: func(next httpsnoop.FlushErrorFunc) httpsnoop.FlushErrorFunc {
				return func() error {
					err := next()
					commitOK()
					return err
				}
			},
		})
		next.ServeHTTP(wrapped, req)
		statusCode := status.Load()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		statusClass := strconv.Itoa(int(statusCode)/100) + "xx"
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
