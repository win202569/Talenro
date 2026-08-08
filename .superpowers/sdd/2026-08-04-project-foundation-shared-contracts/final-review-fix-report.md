# Final whole-branch review fix evidence

## HTTP method metric cardinality and privacy

- Starting HEAD: `bde1fa8ed21d109d2e524acb604e175caf1b7996`.
- Finding verified: `Registry.Middleware` used raw `req.Method` in the shared label slice for both the request counter and duration histogram, while the generated server contract registers only `GET /livez` and `GET /readyz`.
- Scope: only `internal/observability/http.go`, its regression test, and this evidence report.

## TDD evidence

- RED test: `TestMiddlewareBoundsMethodLabelsForAllHTTPMetricFamilies` sends one `GET` plus 12 distinct valid HTTP token methods through the production middleware. The non-GET methods include identifier-like values such as `DEVICE-private-node-123`, `TENANT-customer-8472`, and `REQUEST-correlation-id-550e8400`.
- RED execution: a fresh Windows/amd64 test binary failed with `talenro_control_http_request_duration_seconds series count = 13, want 2`.
- GREEN implementation: `boundedMethod` returns only `http.MethodGet` for exact `GET` input and fixed `other` for every other input. The normalized method is placed in the one label slice shared by the counter and histogram.
- GREEN execution: the focused regression passed. The complete `internal/observability` test binary also passed.
- Structural assertions gather both metric families and inspect label protobuf fields directly. For both the counter and histogram, they require exactly one `GET` series with 1 request/observation and one `other` series with 12 requests/observations, and reject every raw method string in every label value. No protobuf `String()` formatting is used by the new regression.

## Verification evidence

- Focused race build/run: `PASS` with Windows/amd64, CGO enabled, and `-race`.
- Repository vet: `go vet ./...` exited 0.
- Windows build: `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...` exited 0.
- Full tests: `go test -count=1 ./...` exited 0 for every package.
- Formatting and diff hygiene: `gofmt -d` produced no output; `git diff --check` exited 0.
- Generated-path inventory/drift against the checked-out tree: worktree diff exit 0, index diff exit 0, and no untracked files under `api`, `gen`, or `internal/store`.

## Bounded limitations

- The combined pinned-tool/version plus pinned formatter stage did not finish within its 180-second hard bound. It was terminated without retry, so neither stage is claimed complete.
- `go tool golangci-lint run ./...` produced no findings before its 120-second hard bound but did not exit. It was terminated without retry, so lint is not claimed complete.
- Full code generation was not rerun after the bounded tool stage; generated paths were verified unchanged against HEAD instead.

## Self-review

- Cardinality is bounded to exactly `GET` and `other`; the normalizer never returns non-GET request input.
- Method privacy is identical for the request counter and duration histogram because both consume the same normalized label slice.
- Metric names, help, buckets, label names/order, route bounding, and status-class logic are unchanged.
- The response-writer hooks, Flush/FlushError behavior, optional capabilities, and controller traversal are unchanged and remain covered by the passing observability suite.
- `boundedMethod` is pure and adds no shared state; concurrent middleware calls retain the existing collector synchronization behavior.
