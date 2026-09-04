# C12 Secure Execution Architecture Design

## Purpose

Replace the C12 runner's race-prone background-job execution and name-based cleanup with a fail-closed Windows execution controller. The controller must prove process containment, authenticated protocol progression, immutable build inputs, and exact resource ownership before releasing untrusted native work or deleting anything.

The existing assertions in `internal/testinfra/c12_integration_manifest_test.go` are acceptance contracts. The implementation must satisfy them without deleting, skipping, or weakening security checks.

## Scope

The implementation covers five connected boundaries in `scripts/run-c12-integration.ps1`:

1. suspended native process creation and Job Object containment;
2. authenticated, private, message-mode named-pipe coordination;
3. purpose-labelled Go dependency graph receipts;
4. exact direct-leaf and PITR ownership ledgers with retryable cleanup;
5. Docker endpoint, image, container, network, and volume identity custody.

The node-control V7 implementation remains unchanged except for test hashes or integration wiring that must follow the final runner bytes. External Docker volumes are never adopted, mounted, relabelled, or deleted.

## Architecture

### PowerShell orchestration

PowerShell owns policy, state transitions, canonical JSON receipts, deadlines, diagnostics, and top-level convergence. Each mutable resource has a single controller record whose phase moves monotonically. Cleanup consumes that same record; it never reconstructs ownership from names or directory enumeration.

The runner retains one top-level finalizer. Lower layers report exact state and cleanup results but do not erase ownership records that the finalizer may need for retries.

### Embedded C# security primitives

One `Add-Type` block exposes a small Windows-only controller used by PowerShell. It wraps:

- `CreateProcessW` with `CREATE_SUSPENDED`, `CREATE_NO_WINDOW`, and `CREATE_UNICODE_ENVIRONMENT`;
- explicit `bInheritHandles = false`;
- Job Object creation, assignment, membership verification, active-process queries, and completion-port observation;
- `ResumeThread`, `TerminateProcess`, and `TerminateJobObject`;
- message-mode named-pipe creation with `FILE_FLAG_FIRST_PIPE_INSTANCE`, one instance, local-client rejection, and client/server PID verification;
- exact path/file identity, link-count, write-through creation, flush, and deletion primitives.

Native handles remain owned by a disposable controller object until the top-level finalizer has observed process exit and exact cleanup. Disposed handles cannot be reused for verification or cleanup.

## Prepared Worker State Machine

The prepared worker progresses through these states:

1. `CreatedSuspended`
2. `AssignedToJob`
3. `MembershipVerified`
4. `PipeAuthenticated`
5. `CapabilitiesInstalled`
6. `GateWaiting`
7. `Released`
8. `Exited`
9. `CleanIntent`
10. `Removed` or retained retry state

The primary thread cannot resume before `IsProcessInJob` confirms membership in the exact controller Job. Any failure before release terminates the process and Job, waits for `JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO`, and preserves the ownership record if convergence is incomplete.

The prepared-worker path must not invoke `Start-Job`. Other bounded utility paths may retain existing jobs only where their tests permit them and where they cannot create an execution escape window.

## Private Pipe Protocol

The controller creates one duplex, message-mode, first-instance-only named pipe. Remote clients are rejected. The server binds the connected client PID to the suspended process and the client binds the server PID to the controller.

Frames are canonical JSON with a schema, run identifier, worker nonce, sequence, type, payload digest, previous-frame digest, and HMAC. The accepted ordered protocol is:

1. `HELLO`
2. `VERIFICATION_PROJECTION`
3. `JOB_MEMBER_READY`
4. `CAPABILITIES`
5. `CAPABILITIES_INSTALLED`
6. `GATE_WAITING`

Only after all six frames validate does the controller start the release stopwatch and signal the gate. Duplicate, reordered, malformed, oversized, foreign-PID, wrong-nonce, wrong-digest, or replayed frames fail closed.

The worker receives only the minimal immutable projection it needs. Receipt keys and controller cleanup authority remain outside the worker projection.

## Go Graph Receipts

Every selected Go test/build purpose receives its own immutable graph receipt. Graph discovery uses `go list -deps -json -test -tags=integration` under a shared deadline and records:

- exact Go executable and GOROOT identities;
- module root and module-cache identities;
- `GOOS`, `GOARCH`, and `CGO_ENABLED`;
- `go.mod` and `go.sum` digests;
- build arguments and candidate-tree digest;
- packages, imports, dependencies, test imports, external-test imports, and embed patterns;
- every selected source/embed file's origin, relative path, byte length, and SHA-256;
- module path, version, sum, replacement, and directory identity.

The runner verifies `go mod verify`, uses `GOSUMDB` and `-mod=readonly` policy exactly as declared, and binds `cmd/test2json` plus all stream and artifact size limits into the receipt. Validator and initializer receipts remain separately purpose-labelled.

## Artifact and PITR Ledgers

### Prepared artifact direct-leaf ledger

The prepared artifact root has an explicit registry of expected direct children. Each entry records kind, creation phase, identity, link count, lifecycle, and cleanup state:

`NeverAttempted → CreateAttempted → Bound → CleanIntent → Removed → Absent`.

Only `exact_file` and `owned_ephemeral_subtree` kinds are supported. The controller rejects unknown direct siblings and never performs an unbounded recursive deletion. Go cache and temporary directories are owned subtrees with identities captured at creation.

### PITR five-leaf ledger

The PITR run root contains exactly:

- `controller-ownership-v1.wal`
- `ownership.wal`
- `tlsgen.go`
- `server.crt`
- `server.key`

Each leaf has the same explicit lifecycle semantics. Unknown siblings stop cleanup and preserve evidence.

### Controller ownership WAL

The controller WAL is created write-through and flushed after every record. Records use canonical UTC timestamps, monotonically increasing sequence numbers, previous-record digests, payload digests, record digests, and HMAC-SHA256. Events are:

- `BOOTSTRAP`
- `INTENT`
- `ACTUAL`
- `NOT_FOUND`
- `CLEAN_INTENT`
- `CLEAN_RESULT`

The WAL registry is the authority for retrying cleanup after partial failure. A missing, malformed, truncated, reordered, or unauthenticated WAL prevents destructive cleanup.

## Docker Custody

Before creating resources, the runner freezes a Docker endpoint receipt containing context name, endpoint, engine ID, server version, OS type, architecture, and all Docker environment selectors. It binds the exact Docker CLI identity, immutable image reference and image ID, run label, role label, profile, nonce digest, and returned object ID.

Container and volume creation use intent/actual WAL records. An attempted but unconfirmed creation is re-inspected by exact name plus complete identity; it is never assumed absent or adopted. Cleanup operates on verified IDs and checks exact absence using structured `container inspect` and `volume inspect` outcomes. Same-name or partially matching resources are refused.

Dependency probes share one monotonic deadline. Port mappings must be loopback-only and structurally valid. Cleanup retains ownership state across stop, inspect, or removal timeouts.

## Error Handling and Cleanup

Every phase returns a primary failure plus cleanup failures. Cleanup never overwrites the primary failure. State records distinguish creation never attempted, attempted but unconfirmed, created, verified, clean intent, removed, and confirmed absent.

Deadlines are absolute and inherited by child operations. Cleanup has bounded retries and records every result. Before run-root acquisition, PITR candidate cleanup is forbidden. Once ownership is acquired, the top-level finalizer is the only component allowed to converge the complete resource set.

## File Boundaries

- `scripts/run-c12-integration.ps1`: production orchestration and embedded C# primitives. The existing single-file delivery constraint is retained because the manifest tests audit reachable PowerShell/C# call paths.
- `internal/testinfra/c12_integration_manifest_test.go`: behavioral and security acceptance tests. Existing security expectations remain; stale whitespace/count change detectors may be replaced only by stronger behavior assertions.
- `internal/testinfra/c12_dependencies_integration_test.go`: dependency protocol tests and exact health semantics.
- `internal/testinfra/c12_authority_pitr_integration.go`: authority/PITR integration entry points.
- `internal/nodecontrol/authority/v7_migration_grant_test.go`: authenticated overlay bindings and semantic probes.

## Delivery Sequence

1. establish deterministic test environment and overlay propagation;
2. implement suspended process and Job containment;
3. implement authenticated pipe protocol and release gate;
4. implement Go graph receipts and direct-leaf artifact ledger;
5. implement PITR five-leaf ledger and controller WAL;
6. close Docker custody and retryable cleanup state;
7. replace stale source-format assertions with behavior tests where necessary;
8. run targeted RED/GREEN cycles, full elevated Windows tests, and a fresh one-shot authoritative Docker oracle;
9. inspect the final Git index, commit, and push `codex/node-pop-control-plane`.

## Acceptance Criteria

- All existing C12 security tests pass without skips or weakened security expectations.
- No prepared worker executes before exact Job membership and protocol verification.
- No cleanup action targets a resource without exact retained ownership proof.
- All Go inputs and Docker identities are cryptographically bound to purpose-labelled receipts.
- Failure injection preserves the primary error and converges owned resources or retains precise retry state.
- Docker PRE/POST normalized custody is byte-identical except for explicitly permitted built-in bridge identity churn.
- External Docker volume count, bytes, and custody digest remain unchanged.
- `go test ./... -count=1` passes with the approved overlay propagated through `GOFLAGS` under an explicit Windows target.
- A new immutable authoritative oracle reports `AUTHORITATIVE_ORACLE=PASS`.
- Only reviewed source, generated artifacts, tests, and design/plan documents are committed; transient `.superpowers` evidence remains ignored.
