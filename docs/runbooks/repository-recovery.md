# Recover the project on a new computer

Clone `https://github.com/win202569/Talenro.git` and select the branch/commit named
in the latest handoff; do not assume `main` contains an unaccepted checkpoint.
For the latest development work, use `codex/c12-b01-task9-coordinator` and verify
its remote SHA against that handoff. The Task 10 checkpoint adds private read
boundaries and isolated tests only; it does not enable production serving or
clear PostgreSQL/PITR/C12 security gates. See [Task 10 scope and evidence](../roadmap/2026-10-08-task10-read-boundary.md).
The following hash identifies the earlier tooling checkpoint, not the latest tip.
The development checkpoint `fed278f3afae1abcd496caeef1ce855e73cc021d`
on `codex/c12-b01-task9-coordinator` was pushed and remotely confirmed on
2026-10-01, including its prerequisite commits. It is not merged into main and
is not full acceptance. Git carries
the source, generated contracts, tests, designs, implementation plans, and
runbooks. Start with [current status](../roadmap/current-status.md) to resume work.

## Standing preservation rule and latest development work

Latest saved-only [cap/chronology analysis](../roadmap/2026-10-16-leaf-retained-cap-analysis.md)
explains the cumulative32-attempt limit was reached at23.42s of a74.94s receipt.
No late new-member queries follow, although retained snapshots still occur.
No new execution/code change/acceptance; next is a separately agreed diagnostic
design, not cap expansion or a warm-cache rerun.

Latest separately approved leaf-retained-compile-001 returned compiler exit0 but
STOPPED on the32-attempt retained sampling cap; no runtime/retry/acceptance.
Read [actual result and next scope](../roadmap/2026-10-16-leaf-retained-compile-001.md).
Recover source/tests and portable progress from this branch, not ignored SDK/
cache/raw evidence. Do not reuse001, increase caps or start a full matrix.

Latest prepared diagnostic entry targets only nodeleafbinding compilation, not
test execution; see [scope and portable checks](../roadmap/2026-10-16-leaf-retained-compile-entry.md).
Real Go invocation is not approved by its fake-only implementation. On another
PC prepare/approve the concrete fixed SDK/reference/cache inputs separately; do
not infer live readiness from cloning this development checkpoint.

Latest controlled observation002 after the phase fix passes sampled diagnostic
checks only; see [evidence and limits](../roadmap/2026-10-16-retained-child-observation-002.md).
It does not clear the historical matrix/error87 or production/main gates.001
remains failed and must not be replayed or relabeled.

Latest2026-10-16 diagnostic preparation adds a dedicated retained-child wrapper
and portable fake-only tests. Start with
[its scope and continuation](../roadmap/2026-10-16-retained-child-entry.md) and
current status. The later separately approved single live observation STOPPED
on a wrapper phase-confusion defect; read
[its evidence and unresolved fix](../roadmap/2026-10-16-retained-child-observation-001.md).
Do not retry that identity or treat native cleanup facts as observation acceptance.
The later [observer phase fix](../roadmap/2026-10-16-retained-child-phase-fix.md)
corrects that wrapper check with portable fake-only tests, not a live rerun.
A provisional result alone is
not completion: later observations require completed.json, matching artifact/
manifest hashes and no stopped record. Recover source/tests from the development
branch; ignored receipts, caches and live process state are not cloned.

From2026-10-13 the user authorizes ordinary commits and pushes of necessary files
at completed-work/handoff checkpoints. The durable instruction is [AGENTS.md](../../AGENTS.md).
It does not authorize main merges, force pushes, branch deletion or clearing
acceptance/security gates. Only necessary reviewed files are saved; secrets,
caches, temporary files and ignored raw receipts are excluded.

The current handoff adds declared issuer mapping consistency plus the M1 test
hardening. Its outcome is declared consistency only, not authenticated registry,
production compatibility, database uniqueness or v7 acceptance. Read its
[execution record](../roadmap/2026-10-13-declared-issuer-mapping-consistency-execution.md)
and latest current-status appendices. The final commit SHA is verified remotely
and reported at the actual preservation handoff, not inferred from a pending push.
This checkpoint is unfinished: pre-save matrix3 stopped on FuzzIssuerEnvelope's
context deadline exceeded after24 successful commands; Pair was not run.
Earlier reviewed matrix2 passed26/26, but does not erase this latest blocker.
Recover the saved source/plans and continue the recorded diagnosis; do not treat
the saved development branch as main-ready or accepted production code.

On a fresh computer, clone into a new empty directory using PowerShell:

```powershell
git clone --branch codex/c12-b01-task9-coordinator --single-branch https://github.com/win202569/Talenro.git Talenro
if ($LASTEXITCODE -ne 0) { throw 'Clone failed.' }
Set-Location -LiteralPath Talenro
git rev-parse HEAD
git ls-remote origin refs/heads/codex/c12-b01-task9-coordinator
if ($LASTEXITCODE -ne 0) { throw 'Remote identity check failed.' }
```

Check the local SHA against the remote branch and the handoff SHA. Then read
AGENTS.md, current-status.md, the latest execution record, spec and plan to find
the exact completed work and open blockers. Do not reuse a nonempty checkout or
overwrite unrelated local changes through these example commands.
These are recovery instructions, not a claim that a fresh-machine clone or
physical acceptance was performed. Continue with the separately authorized
dependency preparation and verification gates below.

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

Consumer execution now enforces the verifier's fixed cache/offline environment.
The empty-Dir metadata issue is repaired in the published checkpoint: an offline
exact-version query resolves metadata while preserving identity/cache binding
and complete integrity checks. The standalone verifier has one 3600s budget;
C11 outer verification/check-tools/generate caps are 3660/4200/4200s. These
approved limits do not clear security gates or prove fresh-machine recovery.
See [current verification evidence](../roadmap/2026-09-28-devtools-metadata-fix.md).
Do not change locks or repeat downloads to bypass a failure; inherited
cache/proxy/workspace overrides remain unsupported.

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
