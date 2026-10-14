# Isolated fuzz diagnosis

Use PowerShell on Windows. Read the approved2026-10-14 spec/plan and AGENTS.md.
This directory currently contains the verified isolation utility, not a complete
framework reproducer. Never run scripts against a project/cache root as target.

Checkpoint: context gate templates passed5 real-context tests in a verified
test-only toolchain copy. coordinator.patch currently contains ONLY the context
part; actual internal/fuzz tracing and runner are still missing. Do not use it
as a complete reproducer or invoke actual framework arms yet. The diagnostic
gate deliberately pauses propagation and is never an official toolchain patch.

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
