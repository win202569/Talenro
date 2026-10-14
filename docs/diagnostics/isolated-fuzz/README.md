# Isolated fuzz diagnosis

Use PowerShell on Windows. Read the approved2026-10-14 spec/plan and AGENTS.md.
This directory currently contains the verified isolation utility, not a complete
framework reproducer. Never run scripts against a project/cache root as target.

Checkpoint: context gate and trace templates passed7 Go tests; materialization
passed8 assertions. coordinator.patch now contains both context and actual
internal/fuzz tracing changes. The runner is implemented but its real Go build
handoff is blocked: baseline compilation exited0 while job accounting reported1
remaining process at root exit, cleaned to0 afterward. No actual fuzz arm ran.
Do not retry baseline or run later arms without new bounded authorization.
Do not treat these component tests as a full reproduction.
The diagnostic gate deliberately pauses propagation and is never an official
toolchain patch. Trace IDs preserve equality for comparable errors without
printing payloads; non-comparable errors are explicitly unsupported, not silently
classified. Observation mode releases the propagation gate immediately.
Final return marks child/old state unread and stoppingRead=false; stop captures
the original operands. Source parent-done names the actual ctx.Done branch;
incoming deadline/canceled distinguishes the cause without extra Err reads.

run.tests.ps1 currently passes16 assertions, including actual descendant cleanup,
wrong-mode/missing-worker evidence rejection and wrapped child-only environment.
process-native.cs launches suspended into an owned, non-breakaway Windows job,
whitelists inherited std handles and records group accounting. Its baseline
remaining-count interpretation still needs investigation; unit passes are not
proof of real Go integration. Current run.ps1 also has a known pending compile
log path-escaping incompatibility (raw doubled backslashes); do not bypass its
provenance guard. Keep the failed launch/build receipts unchanged.

Windows API references: [CreateProcessW](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-createprocessw),
[AssignProcessToJobObject](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-assignprocesstojobobject),
[job limits](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_extended_limit_information).

After verifying a new independent copy and the pinned context.go hash
765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7,
materialize context-diag.go.txt/context-diag_test.go.txt as diagnostic_fuzz.go/
diagnostic_fuzz_test.go inside copied src/context, and apply the context patch
only there. This hash-pinned patch uses zero-context hunks: require the exact
source hash and Git's --unidiff-zero flag (never apply it to drifted sources).
The reverse check against the tested copy succeeds. Build context's test executable with the copy's explicit GOROOT,
independent GOCACHE and offline process environment from the approved plan.
Run with quoted native arguments '-test.run=^TestDiagnosticFuzz', '-test.count=1',
'-test.timeout=30s'. In PowerShell, unquoted dotted arguments can be misparsed.
Registration rejects unsupported/canceled/duplicate contexts; a forgotten gate
releases after1s and records watchdog expiry. These5 context tests are not
coordinator/worker/RPC or original project acceptance results.

Run `powershell -NoProfile -File docs/diagnostics/isolated-fuzz/isolation.tests.ps1`
from the development worktree. Tests generate unique tiny fixture directories
inside the owned diagnostic root and retain them. Real link tests report UNRUN
if privileges prevent creation; UNRUN is not PASS. No external test library needed.
Run materialize.tests.ps1 for actual patch replay/source-drift guard tests using
tiny source-only fixtures. Install-DiagnosticPatch requires both exact source
pins, rejects official/outside targets and existing template files, then installs
all four templates and the patch only into the validated copy.

Dot-source isolation.ps1, then use New-IsolatedToolchain with an existing source
directory and a nonexistent unique RunRoot beneath
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/. The function checks
paths, links and independent source/copy manifests before returning. It does not
run or patch the copy. Do not substitute another toolchain or download dependencies.
The fixed version is Go1.26.5 windows/amd64, independently prepared on a fresh PC.

Generated copies, caches, fixtures and receipts are not recovery artifacts and
must not be committed. Rebuild from tracked scripts/templates; source hashes,
actual executable/GOROOT and runtime results still require verification. Original
matrix3 failure and Pair-unrun remain open; no product acceptance or main merge.
