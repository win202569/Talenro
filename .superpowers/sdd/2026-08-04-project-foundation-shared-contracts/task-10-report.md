# Task 10 report: Privacy-safe Prometheus observability

## Status

Implemented the private Prometheus surface with `github.com/prometheus/client_golang v1.23.2`, bounded request labels, a loopback-safe metrics address, and coordinated lifecycle for the public and metrics HTTP servers.

## Metric and privacy design

- Private registry: a new `prometheus.Registry`; no default Go/process collectors are registered.
- Metrics and exact help text:
  - `talenro_control_http_requests_total`: `Completed control API HTTP requests.`
  - `talenro_control_http_request_duration_seconds`: `Control API HTTP request duration.`
  - `talenro_control_build_info`: `Talenro control API build identity.`
- Request label names are exactly `method`, `route`, and `status_class`. Build information label names are exactly `version` and `commit`.
- Route labels allow only the generated templates `/livez` and `/readyz`. The controller clarified that every empty, unknown, or unmatched route must use the fixed literal `unmatched` (no leading slash).
- The middleware never uses raw request path, query, URL, address, domain, error text, credentials, or protocol/traffic contents as a label.
- The public handler is wrapped outside the generated `http.ServeMux`; after routing, `Request.Pattern` is normalized to a known template. Unknown requests have an empty/unrecognized pattern and become `unmatched`.
- `/metrics` is mounted on a dedicated mux/server only. The public mux does not expose it, and the metrics mux does not expose control API routes.
- `TALENRO_METRICS_ADDRESS` defaults to `127.0.0.1:9090`. It uses `net.SplitHostPort` validation and rejects unspecified/public binds unless the existing explicit public-bind opt-in is set.

## TDD evidence

### Route cardinality RED/GREEN

- RED: `go test ./internal/observability` failed to compile with `undefined: NewRegistry` before observability production code existed.
- GREEN: `go test ./internal/observability` passed after adding the private registry and middleware.
- RED: `TestMiddlewareCollapsesUnknownRouteToUnmatched` showed the raw label `route=/device/private-node-123`.
- GREEN: the same test passed after allowlisting generated route templates and collapsing all other inputs to `unmatched`.
- Tests reject leaks of `example.invalid`, `198.51.100.1`, `random-device-123`, and `private-node-123`; the structured label assertion requires exactly `method=GET`, `route=unmatched`, and `status_class=4xx`.

### Configuration RED/GREEN

- RED: default test failed because `Config.MetricsAddress` did not exist.
- GREEN: default `127.0.0.1:9090` added.
- RED: custom-address test observed the default instead of `127.0.0.1:19090`.
- GREEN: `TALENRO_METRICS_ADDRESS` parsing added.
- RED: public `0.0.0.0:9090` and malformed `127.0.0.1` metrics addresses were accepted.
- GREEN: both unsafe configurations are rejected using the same host/port and unspecified-address policy as the public server.

### Bind and lifecycle RED/GREEN

- RED: `TestRunServesMetricsOnlyOnPrivateListener` brought up public `/livez`, then timed out waiting for private `/metrics` because only the public server existed.
- GREEN: private `/metrics` and public `/livez` became available on separate listeners; public `/metrics` and private `/livez` both returned 404; gathered metrics contained the normalized `/livez` route.
- RED: with a forced public bind failure, the private metrics listener remained reachable.
- GREEN: either listener result now runs metrics-first then public shutdown under one shared bounded context, and the sibling listener is no longer reachable.
- Existing forced-close tests remain green. Shutdown still emits only the Task 9 stable event names/categories and retains the exact server resource limits through `platform.NewHTTPServer` for both listeners.

## Verification evidence

All commands used process-scoped `GOOS=windows`, `GOARCH=amd64`, fresh `%TEMP%\talenro-task10-gocache`, and shared `%TEMP%\talenro-task2-gomodcache`.

- Dependency: `go list -m github.com/prometheus/client_golang` -> `github.com/prometheus/client_golang v1.23.2`.
- Focused race: `go test -race ./internal/observability ./internal/platform/... ./internal/config ./cmd/control-api` -> PASS.
- Build: `go build ./...` -> PASS. Process-scoped Git safe-directory entries covered the worktree and parent repository for VCS stamping.
- Full suite: `go test ./...` -> PASS across all repository packages.
- Formatting/diff: `gofmt` applied; `git diff --check` clean.

An initial race run exposed a brittle protobuf text-spacing assertion while the correct `route=unmatched` value was present. The test was changed to inspect structured labels and the complete focused race command then passed.

## Files

- Added `internal/observability/metrics.go`.
- Added `internal/observability/http.go`.
- Added `internal/observability/metrics_test.go`.
- Modified `cmd/control-api/main.go` and `cmd/control-api/main_test.go` for listener wiring and lifecycle tests.
- Modified `internal/config/config.go` and `internal/config/config_test.go`.
- Modified `.env.example`, `go.mod`, and `go.sum`.
- `internal/platform/server.go` was deliberately reused without edits so Task 9 limits and sanitized `ErrorLog` behavior remain identical on both servers.
- Generated OpenAPI files were not hand-edited.

## Self-review

- Confirmed exact metric names, help strings, label names, Prometheus client version, and default histogram buckets against the brief.
- Confirmed raw path/query values are absent from gathered families in both normalized and unknown-route tests.
- Confirmed only `/metrics` is mounted on the private mux and it is absent from the public mux.
- Confirmed metrics-first/public-second shutdown ordering shares a single timeout context.
- Confirmed a non-`http.ErrServerClosed` listener error retains its internal cause but reports only the existing stable `http_listen_or_serve` category.
- Confirmed health handlers/payloads, generated routing, public error sanitization, forced close, and server time/header limits remain covered by the passing full suite.
- Mutation check: removing request observation, changing any allowed label, preserving an unknown route, mounting `/metrics` publicly, omitting the second listener, or failing to stop its sibling is covered by a test.

## Concerns

- No implementation concerns. Public metrics binds rely on the existing explicit `TALENRO_ALLOW_PUBLIC_HTTP` deployment opt-in; deployments using it must place the listener behind authenticated infrastructure as required by the brief.
