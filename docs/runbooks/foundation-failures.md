# Foundation dependency failures

The public health contract separates process health from dependency health. `/livez` remains HTTP 200 while the API process can serve requests. `/readyz` is HTTP 200 only when PostgreSQL, Redis, and NATS are usable; it becomes HTTP 503 with bounded dependency names and sanitized states when any required dependency fails. Never capture response bodies, headers, raw errors, environments, connection URLs, Docker inspect output, or traffic data while diagnosing these failures.

Use service-name-only inspection from the repository root. An absent service name means it is not currently running; this command does not print container identifiers, addresses, mounted data, or environment values.

### Windows PowerShell

```powershell
docker compose -f deploy/dev/compose.yaml ps --services --filter status=running
```

### Bash

```bash
docker compose -f deploy/dev/compose.yaml ps --services --filter status=running
```

## PostgreSQL unavailable

Expected behavior is `/readyz` HTTP 503 and `/livez` HTTP 200. Confirm whether `postgres` appears in the service-name-only inventory above. Do not inspect the container or print the database URL.

### Windows PowerShell recovery

```powershell
docker compose -f deploy/dev/compose.yaml start postgres
```

### Bash recovery

```bash
docker compose -f deploy/dev/compose.yaml start postgres
```

Wait up to 10 seconds for `/readyz` to return HTTP 200; use `scripts/smoke.ps1` or `scripts/smoke.sh` to perform this check without printing a body. If PostgreSQL cannot recover, stop and preserve the volume for investigation. PostgreSQL volume deletion destroys the authoritative local database and is prohibited without explicit approval for the exact resolved volume. After approval, first stop the stack, resolve and validate only the compose volume labeled `postgres-data`, remove that one confirmed volume, restart dependencies, and reapply migrations. Never use `docker compose down --volumes` as a shortcut because it targets unrelated volumes too.

## Redis unavailable

Expected behavior is `/readyz` HTTP 503 and `/livez` HTTP 200. Confirm whether `redis` appears in the service-name-only inventory above.

### Windows PowerShell recovery

```powershell
docker compose -f deploy/dev/compose.yaml start redis
```

### Bash recovery

```bash
docker compose -f deploy/dev/compose.yaml start redis
```

Wait for `/readyz` to return HTTP 200. Redis contains only rebuildable, non-authoritative state, so clearing local Redis is recoverable when ordinary restart fails. Obtain confirmation, then clear only the local development database with `docker compose -f deploy/dev/compose.yaml exec -T redis redis-cli FLUSHALL` and restart Redis. Do not use this command against any shared or production environment, and do not delete PostgreSQL while clearing Redis.

## NATS unavailable

Expected behavior is `/readyz` HTTP 503 and `/livez` HTTP 200. Confirm whether `nats` appears in the service-name-only inventory above. Do not query monitoring payloads or capture JetStream messages as diagnostic evidence.

### Windows PowerShell recovery

```powershell
docker compose -f deploy/dev/compose.yaml start nats
```

### Bash recovery

```bash
docker compose -f deploy/dev/compose.yaml start nats
```

Wait for `/readyz` to return HTTP 200. If it does not recover, perform a normal `docker compose -f deploy/dev/compose.yaml down`, then `up -d --wait`; preserve the NATS volume and report only service names plus the bounded control API event/category fields described in the local development runbook.

## Escalation boundary

Restarting a named service and clearing disposable local Redis state are recoverable local actions. Deleting a PostgreSQL volume, using `down --volumes`, removing an unvalidated volume name, or changing service credentials is destructive or scope-expanding and requires explicit approval before execution.
