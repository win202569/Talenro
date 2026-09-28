# Recover the project on a new computer

Clone `https://github.com/win202569/Talenro.git` and select the branch/commit named
in the latest handoff; do not assume `main` contains an unaccepted checkpoint.
The development-tool implementation is local commit `a0ca18928f552752c6580929f9b6fff88cd8d58b`
on `codex/c12-b01-task9-coordinator`; it has not been pushed as of this update.
Another computer cannot fetch a local-only commit until a separately authorized
push publishes it. Git carries
the source, generated contracts, tests, designs, implementation plans, and
runbooks. Start with [current status](../roadmap/current-status.md) to resume work.

## Windows test environment

The current full repository suite includes Windows process, named-pipe, and
filesystem tests. Use Windows/amd64 with Git, Go 1.26.5 (selected by `go.mod`),
and Windows PowerShell 5.1 available for the existing legacy test/smoke contracts.
The four development-tool entries additionally require the exact stable
Microsoft PowerShell **Core 7.6.5** runtime; arrange its installation separately.
They reject 5.1, other 7.x versions and prerelease builds. An absolute path to
`pwsh.exe` is supported; no Codex-private installation path is a project prerequisite.
The separate live integration acceptance
also requires Docker and the pinned dependencies described in
[local development](local-development.md).

## Prepare dependencies separately from offline verification

From the selected checkout root, check the prerequisites without changing them:

```powershell
pwsh.exe -NoProfile -Command '$PSVersionTable.PSVersion.ToString(); $PSVersionTable.PSEdition'
$env:GOTOOLCHAIN='local'
go version
```

Expect `7.6.5`, `Core`, and `go1.26.5 windows/amd64`. An ordinary Go installation
alone is insufficient: the verifier requires the exact toolchain module under
`%USERPROFILE%/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64`.
Network preparation is a separate manual/authorized action, never an automatic
fallback from validation. The following preparation commands are instructions,
not a claim that fresh-machine recovery has been executed and accepted:

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='https://proxy.golang.org'; $env:GOSUMDB='sum.golang.org'
$env:GOAUTH='off'; $env:GOVCS='all:off'; $env:GOFLAGS=''
$env:GOPRIVATE=''; $env:GONOPROXY=''; $env:GONOSUMDB=''; $env:GOINSECURE=''
$env:GOPATH=Join-Path $env:USERPROFILE 'go'
$env:GOMODCACHE=Join-Path $env:GOPATH 'pkg/mod'
# Separately authorized exact toolchain-cache preparation, if absent:
go mod download golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64
if ($LASTEXITCODE -ne 0) { throw 'Pinned toolchain preparation failed.' }
$fixedGo=Join-Path $env:GOMODCACHE 'golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
if (-not (Test-Path -LiteralPath $fixedGo -PathType Leaf)) { throw 'Pinned toolchain is absent.' }
$fixedVersion=& $fixedGo version
if ($LASTEXITCODE -ne 0 -or $fixedVersion -cne 'go version go1.26.5 windows/amd64') { throw 'Pinned toolchain version mismatch.' }
$env:PATH=(Split-Path $fixedGo)+';'+$env:PATH
go mod download all
if ($LASTEXITCODE -ne 0) { throw 'Root dependency preparation failed.' }
go -C tools/devtools mod download all
if ($LASTEXITCODE -ne 0) { throw 'Tool dependency preparation failed.' }
```

Keep the committed lockfiles unchanged; do not substitute `latest`, add a
workspace/replace, or disable checksum verification. Missing dependencies during
offline checks remain failures requiring separate preparation approval.

Current real verification is blocked by incomplete tool cache, and independent
review found that consumer execution does not yet enforce the verifier's fixed
cache/environment. These preparation instructions do not repair that product
boundary. Do not treat the following commands as accepted recovery until the
[open review finding and real verification](../roadmap/2026-09-28-devtools-real-verification-checkpoint.md)
are resolved; inherited cache/proxy/workspace overrides are not supported.

```powershell
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOFLAGS='-mod=readonly'
go mod verify
if ($LASTEXITCODE -ne 0) { throw 'Root integrity verification failed.' }
pwsh.exe -NoProfile -File scripts/verify-devtools.ps1
if ($LASTEXITCODE -ne 0) { throw 'Tool integrity verification failed.' }
pwsh.exe -NoProfile -File scripts/check-tools.ps1
if ($LASTEXITCODE -ne 0) { throw 'Tool versions failed.' }
pwsh.exe -NoProfile -File scripts/generate.ps1
if ($LASTEXITCODE -ne 0) { throw 'Generation failed.' }
git diff --exit-code -- api gen internal/store
if ($LASTEXITCODE -ne 0) { throw 'Generated content changed; investigate before committing.' }
```

Run one stage at a time using its original bounded execution policy: verification
and tool checks 900 seconds, generation and lint 600 seconds. Preserve failure
results rather than extending a deadline or repeatedly rerunning an unchanged
failure. Root integrity does not replace tool-module integrity, or vice versa.
Root `go mod verify` checks downloaded module content; by itself it is not proof
that a cache is complete. The public tool verifier also checks cache presence.

## Tool ownership and supported entrypoints

The root module retains Goose and runtime libraries. Buf, protoc-gen-go,
oapi-codegen, sqlc and golangci-lint live in `tools/devtools`. A bare root
`go tool buf` (or another moved tool) is no longer supported. For explicit tool
calls from the repository root, first pass the public verifier and use:

```powershell
pwsh.exe -NoProfile -File scripts/verify-devtools.ps1
if ($LASTEXITCODE -ne 0) { throw 'Tool integrity verification failed.' }
go tool "-modfile=$((Join-Path (Get-Location).Path 'tools/devtools/go.mod'))" golangci-lint run ./...
if ($LASTEXITCODE -ne 0) { throw 'Root lint failed.' }
# Goose remains in the root module:
go tool goose -version
```

Do not put `-modfile` into `GOFLAGS`. Lint and generation target the root project,
not the tools directory. Nested devtool entries use the same verified PowerShell
executable, not a PATH fallback.

| Entry family | Windows | Linux/macOS/WSL |
| --- | --- | --- |
| verify-devtools/check-tools/generate/verify-c11 `.ps1` | Core 7.6.5 only; C11 full run remains blocked below | Unsupported |
| Same four `.sh` entries | Reject without tool work | Reject without tool work |
| Existing Bash smoke | Retained under its original prerequisites, including Git Bash | No new acceptance claim |

The withdrawn Unix positive implementation and its four historical failing
cases are preserved in commit `03947e31536a803aaca19c295edd8d9eb4d258f1`.
Rejection tests do not turn those failures into fixes. Do not switch Docker to
Windows-container mode merely because the host tool entrypoints require Windows.

## Ordinary and live tests: current execution gate

Full ordinary tests, full C11, and C12/bootstrap/13 physical gates remain blocked
by the recorded Defender event and open I2/I3 acceptance findings. Tool checks
do not lift that gate. Do not disable protection, add exclusions, restore a
quarantined file, or increase timeouts to obtain a pass. Microsoft-account-based
vendor review is deferred; no security clearance is implied.

Only the approved compile-only probe can run before separate gate restoration:

```powershell
$env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run '^$' -count=1 -timeout=5m
```

This is not test execution or live acceptance. Race testing remains a separate
gate requiring the supported C toolchain and `CGO_ENABLED=1`.

No pre-existing `.superpowers` directory or caller `-overlay` is required.
The former local source replacement has been integrated into
`internal/nodecontrol/authority/postgres_repository.go`. Ordinary child Go
commands clear inherited `GOFLAGS`, `GOWORK`, and persisted Go environment
settings. The integration factory probe generates its own temporary overlay,
restricted to one test source file; it never replaces production sources.
Some semantic probes create temporary packages inside the checkout so that Go's
`internal` import boundary remains valid. Their input is generated by tracked
tests; it is not a file to copy from another computer.

## What Git does not restore

Go caches, local tool binaries, `.superpowers` session state, raw historical
test receipts, Docker volumes, passwords, signing keys, environment secrets,
and Codex conversation history are not part of the repository backup. Restore
any required operational credentials or database backups through their own
approved channels. The tracked plans and status document preserve project
context, but a new checkout does not reproduce old machine-bound evidence.

## Current acceptance boundary

The earlier full-test result used a local overlay and must not be treated as
proof that the old commit was independently reproducible. After the portability
repair, the then-approved full test command passed on 2026-09-08 at code commit
`16c4e8c4e43ae893aad8779639bec0f80ca096cb` in a new independent clone with a
space-containing path and no pre-existing `.superpowers` directory. The
PowerShell runner canonicalizes its pinned embedded Go source to LF before
verifying the digest, so normal CRLF checkout does not require a local patch.

That historical result does not authorize rerunning the currently blocked suite.
After separate security-gate restoration, follow the current stage plan and its
original budgets for verification on the new computer and live acceptance;
C1.2 Batch 01 and the full product remain incomplete, as recorded in
[current status](../roadmap/current-status.md).
