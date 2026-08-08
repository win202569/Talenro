# Local development

This runbook operates the foundation only. Run commands from the repository root unless a command says otherwise. The smoke and verification entrypoints locate the repository from their own script paths, so they can also be invoked from another directory.

## Prerequisites and pins

- Windows PowerShell 5.1 or newer, or Bash 5 or newer on Unix.
- Go 1.26.5, selected by the `toolchain go1.26.5` directive in `go.mod`.
- Docker Engine with Docker Compose v2 and permission to start local containers.
- Loopback ports 5432, 6379, 4222, 8222, 8080, and 9090 available.
- Repository-pinned Go tools: Buf 1.72.0, protoc-gen-go 1.36.11, oapi-codegen 2.8.0, sqlc 1.31.1, goose 3.27.1, and golangci-lint 2.12.2.
- Pinned service images: PostgreSQL 18.4, Redis 8.8.1, and NATS Server 2.14.3. Never substitute a floating `latest` tag.

Confirm the toolchain before starting. The check script reports only stable stage names and version failures; it does not print environment values.

### Windows PowerShell

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-tools.ps1
docker compose -f deploy/dev/compose.yaml version
```

### Bash

```bash
./scripts/check-tools.sh
docker compose -f deploy/dev/compose.yaml version
```

## Environment

Use a fresh process environment containing exactly the seven tracked `TALENRO_*` assignments in `.env.example`. Keep the HTTP and metrics listeners on their tracked loopback defaults and keep both public-listener override flags false. Do not source, evaluate, print, paste into tickets, or commit a copy of the file. The smoke scripts enforce this policy with an allowlist parser, clear every ambient `TALENRO_*` variable, and then set only the seven parsed keys before starting the API.

For manual commands below, load those exact local-only values into the current shell by a trusted environment manager, then verify presence without displaying values.

### Windows PowerShell

```powershell
$required = @(
  'TALENRO_HTTP_ADDRESS', 'TALENRO_METRICS_ADDRESS',
  'TALENRO_ALLOW_PUBLIC_METRICS', 'TALENRO_DATABASE_URL',
  'TALENRO_REDIS_ADDRESS', 'TALENRO_NATS_URL',
  'TALENRO_ALLOW_PUBLIC_HTTP'
)
if ($required.Where({ [string]::IsNullOrEmpty([Environment]::GetEnvironmentVariable($_)) }).Count) {
  throw 'Required local environment is incomplete.'
}
```

### Bash

```bash
required=(TALENRO_HTTP_ADDRESS TALENRO_METRICS_ADDRESS TALENRO_ALLOW_PUBLIC_METRICS TALENRO_DATABASE_URL TALENRO_REDIS_ADDRESS TALENRO_NATS_URL TALENRO_ALLOW_PUBLIC_HTTP)
for name in "${required[@]}"; do
  [[ -n "${!name:-}" ]] || { printf '%s\n' 'Required local environment is incomplete.' >&2; exit 1; }
done
```

## Start dependencies and migrate

### Windows PowerShell

```powershell
docker compose -f deploy/dev/compose.yaml up -d --wait
if ($LASTEXITCODE -ne 0) { throw 'Dependency startup failed.' }
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
if ($LASTEXITCODE -ne 0) { throw 'Migration failed.' }
go run ./cmd/control-api
```

Stop `go run` with Ctrl+C. The process shuts down its metrics listener first, then its public listener, drains NATS, and closes Redis and PostgreSQL clients within bounded shutdown contexts.

### Bash

```bash
docker compose -f deploy/dev/compose.yaml up -d --wait
go tool goose -dir db/migrations postgres "${TALENRO_DATABASE_URL}" up
go run ./cmd/control-api
```

Stop `go run` with Ctrl+C and wait for it to exit before stopping dependencies.

## Migrations and generation

Exercise migration reversibility against disposable local data: apply all migrations, roll back exactly one migration, then reapply it.

### Windows PowerShell

```powershell
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL down
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1
git diff --exit-code -- api gen internal/store
```

### Bash

```bash
go tool goose -dir db/migrations postgres "${TALENRO_DATABASE_URL}" up
go tool goose -dir db/migrations postgres "${TALENRO_DATABASE_URL}" down
go tool goose -dir db/migrations postgres "${TALENRO_DATABASE_URL}" up
./scripts/generate.sh
git diff --exit-code -- api gen internal/store
```

Generation runs Buf lint and Protobuf generation, OpenAPI generation, sqlc generation, and formatting with repository-pinned tools. A nonempty diff is drift and must be reviewed; do not discard it automatically.

## Tests and acceptance

Unit tests do not require containers. Integration tests require the healthy compose stack and applied migrations.

### Windows PowerShell

```powershell
go test -race -count=1 ./...
go test -tags=integration -count=1 ./internal/testinfra ./internal/store
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

### Bash

```bash
go test -race -count=1 ./...
go test -tags=integration -count=1 ./internal/testinfra ./internal/store
./scripts/verify.sh
./scripts/smoke.sh
```

The smoke test owns its compose lifecycle. It migrates, builds `cmd/control-api` into a unique task-specific temporary directory, starts that binary from the repository root, verifies liveness and readiness, stops only PostgreSQL, proves readiness fails while liveness remains healthy, restarts PostgreSQL, and proves recovery. Cleanup targets the actual API PID, removes the exact temporary binary, removes only the now-empty temporary directory, and runs compose down even after a partial failure.

## Safe diagnostics

Do not collect `.env.example`, process environments, connection strings, request or response headers, response bodies, Docker inspect output, stack traces, traffic data, account data, or raw error output. The safe first diagnostic is a service-name-only inventory.

### Windows PowerShell

```powershell
docker compose -f deploy/dev/compose.yaml ps --services --filter status=running
```

### Bash

```bash
docker compose -f deploy/dev/compose.yaml ps --services --filter status=running
```

When a maintainer requests application evidence, share only the bounded event name, severity, and error category already emitted by the control API, after manually checking the excerpt. Do not attach raw log files. The allowed foundation event names are `control_api_starting`, `control_api_stopped`, `control_api_shutdown_failed`, and `control_api_forced_close`; allowed categories are the bounded values defined in `internal/platform/errors.go`.

To collect an event-name-only stream, discard every full line and emit only a matched allowlisted event name.

### Windows PowerShell

```powershell
go run ./cmd/control-api 2>&1 | ForEach-Object {
  if ($_ -match '(control_api_(?:starting|stopped|shutdown_failed|forced_close))') { $Matches[1] }
}
```

### Bash

```bash
go run ./cmd/control-api 2>&1 | sed -nE 's/.*(control_api_(starting|stopped|shutdown_failed|forced_close)).*/\1/p'
```

These filters never emit the original line. Stop the process with Ctrl+C; review the resulting event-name-only stream before sharing it.

## Normal shutdown

First interrupt the API and wait for its process to exit. Then stop the local dependency stack without deleting volumes.

### Windows PowerShell

```powershell
docker compose -f deploy/dev/compose.yaml down
```

### Bash

```bash
docker compose -f deploy/dev/compose.yaml down
```

Do not add `--volumes` to routine shutdown. PostgreSQL is authoritative even in the local model; deleting its volume is a separate destructive recovery action governed by the failure runbook.
