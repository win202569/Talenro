# Fixed nodeleafbinding ordinary verification — proposed design

Status: written design approved by user “确认” after `9fdda6ff`.
The [fake-only implementation plan](../plans/2026-10-16-leaf-ordinary-check.md)
awaits approval and execution-method selection; real execution is not approved.
Baseline development checkpoint: `ffbb842b`.

## 1. Intent and scope

Obtain independent, bounded ordinary verification of the original
`internal/nodecontrol/nodeleafbinding` package: a fresh build with Go's unchanged
default vet policy, followed only after verified compilation by the resulting
original test binary. This addresses missing ordinary package evidence after
[build-only001](../../roadmap/2026-10-16-leaf-stage-build-only-001.md), not the
whole C12 acceptance backlog. Windows PowerShell only; no downloads, security
changes, product edits, matrix replay or diagnostic binary reuse.

The written-spec approval permits an implementation plan only; implementation
requires plan approval and execution-method selection.
A later real invocation requires a separate, concrete single-run approval after
fake-only implementation verification. None of these approvals is implied here.

## 2. Approaches and recommendation

1. **Recommended: fixed dedicated ordinary controller.** Follow the existing
   mapping-single-check pattern, reusing ordinary recorded-process, compile
   evidence and shutdown guards. Add no configurable package/filter or native
   mode. Cost: one small controller and its process-free test surface.
2. Extend the26-check matrix with selection/resume. This broadens a shared
   interface and introduces skipped-check/progress ambiguity; unnecessary here.
3. Repurpose the build-only controller or execute its existing binary. Rejected:
   it requests vet-off and diagnostic sampling, and cannot supply fresh ordinary
   build evidence. Its records and typed false acceptance flags remain intact.

## 3. Components and fixed inputs

New proposed files under `docs/diagnostics/isolated-fuzz`:
`leaf-ordinary-check.ps1`, `leaf-ordinary-check-plan.tests.ps1`,
`leaf-ordinary-check.tests.ps1`. Additional fixture files are permitted only if
the later implementation plan identifies them. Do not modify product Go/tests,
the C# collector, ordinary `Assert-ProcessCompletion`, matrix behavior, mapping
behavior or any diagnostic entry to accommodate this controller.

Public surface: `Get-LeafOrdinaryCheckPlan(RunRoot,Toolchain,ReferenceReceipt,
ModuleCache)` and `Invoke-LeafOrdinaryCheck` with those same four string inputs.
Planner performs file/path/hash checks only, no workload launch or root creation.
The controller independently validates fixed fields; a caller-altered plan is
not an argument override API. An existing-root saved validator, if implemented,
must not invoke the new-root planner or any process seam.

Require64-bit host, the existing linked development worktree, an absent owned
`leaf-ordinary-check-NNN` root beneath the established isolated-fuzz scope,
non-overlapping input/output paths, safe non-reparse paths and safe single-link
input/artifact files using existing helpers. Toolchain must be an owned
`uninstrumented-NNN/toolchain` copy and match its reference receipt. Required
offline module marker must exist. Later execution approval freezes actual
source/helper/input hashes; this proposal does not silently refresh old bindings
or certify another computer's SDK/cache.

Use the same20 fixed environment fields as mapping-single-check: copied GOROOT;
private fresh GOCACHE/TEMP/TMP/GOTMPDIR; existing GOMODCACHE; GOTOOLCHAIN=local;
GOENV/GOPROXY/GOSUMDB/GOAUTH=off; GOVCS=all:off; GO111MODULE=on; GOWORK=off;
GOFLAGS=-mod=readonly; GOOS=windows; GOARCH=amd64; empty diagnostic mode/fixture
and GODEBUG. No environment or argument passthrough.

## 4. Execution and ordinary evidence

Build once from the workspace with exactly:
`test -c -x -p=1 -count=1 -timeout=120s -o=<root>/package.test.exe talenro.local/platform/internal/nodecontrol/nodeleafbinding`.
Native build budget600s. No `-vet=off` or custom vet suite; label VetMode=default,
not all analyzers or independent vet-process identity coverage.

Only after build exit0 AND unchanged ordinary schema3 completion guard AND
original-package/copied-compiler evidence AND safe binary/receipt/stream binding,
execute that new binary once from the original package directory with exactly:
`-test.count=1 -test.timeout=120s -test.v=true -test.run=.`.
Native runtime budget150s. No fuzz/race/benchmark mode or narrowed test selector.
Recheck executable/source bindings immediately before each workload boundary.

Use the existing ordinary recorded-process path, whose native call is
`IsolatedFuzzProcess.Run(..., false, false)`. No retained diagnostics, compile
trigger, live-sampling option or replacement collector. Preserve existing500ms
natural observation and shared5000ms cleanup policy, access rights, bounded
member evidence and all ordinary completion predicates. A repeated error87,
unknown ownership/exit, truncation, forced termination, timeout or diagnostic
receipt stops the check; none becomes an empty-job or successful observation.

Both receipts require schema3, exit0 and unmodified `Assert-ProcessCompletion`.
Validate initial accounting/member-query evidence using the existing mapping
query predicate for both phases. Preserve original receipts and optional image
errors; do not rewrite fields or claim exhaustive process capture.

Runtime output requires final nonempty line exactly PASS, at least one original
top-level Test PASS, no FAIL record and no diagnostic instrumentation marker in
either stream. Report observed RUN/PASS/SKIP names/counts and subtests separately.
Skips retain Go's ordinary semantics and must be disclosed; do not claim every
test body executed merely from exit0 or a final PASS. This is not fuzz evidence.

At most one build and one runtime per approved invocation. Record each attempted
launch durably before entering the native boundary. Ambiguous launch, failed
receipt persistence or any failed gate forbids another call; no automatic retry,
cache warm-up, Go version/env/list command, historical PID lookup or later check.

## 5. Final audits, persistence and result meaning

Inventory actual product files before/after; independently validate full copied
SDK/reference identity before/after; bind frozen helpers/inputs plus all written
receipts, streams and binary. Do not assume a historical product-file count or
walk output cache/tmp. Final product, SDK and artifact checks must each be
attempted independently when their prerequisites exist, preserving primary and
secondary errors rather than allowing one exception to suppress another audit.
Report pre-native/build/runtime/post-native durations separately; native limits
are not a global audit-time, memory or OS-call preemption guarantee.

Persist plan, source/input bindings, launch/progress state, receipts/streams,
product inventories, artifact binding, provisional result and explicit manifest.
Manifest rows use safe root-relative paths, length and SHA256; reject escapes,
duplicates, missing/changed/reparse/hardlinked artifacts. Preserve launch errors
and STOPPED on failure. Persistence failure is never success, even if STOPPED
itself could not be saved; preserve secondary errors where possible and return
failure to the caller. Never repair saved evidence to obtain a pass.

Only a real `completed.json` with the manifest hash and fixed scope may establish
durable completion after all audits. The manifest excludes its own bytes and the
completion marker to avoid circular binding. A pending marker or provisional
result is insufficient; STOPPED/completed contradictions are refused on read.
Publish the marker through an owned pending file only after validation; do not
reuse an existing identity or overwrite an existing completion marker.

Keep distinct facts: CompilePassed, DefaultVetBuildPassed, ExecutionAttempted,
OriginalTestPassed, ProductUnchanged, ToolchainUnchanged, ArtifactsUnchanged and
FinalizationRequired. `DefaultVetBuildPassed` means the fixed default-vet Go build
succeeded, not an exhaustive analyzer/capture claim. `PackagePassed` may be true
only with all required build/runtime/audit facts and a valid completed binding.
Persist provisional result with PackagePassed=false;the completion marker may
record PackagePassed=true only after those checks. Saved evaluation derives
package success from that valid marker/binding,not the provisional result.
Overall Acceptance
remains typed false in every record, including a successful completed marker.
No C11/C12/Task9/Task10/Batch01/security/production/main gate closes automatically.

## 6. Process-free verification required before any real-run proposal

Use controlled fake files/receipts and a replaced native seam; fixture executables
are data only and must never run. Execute scoped tests on PS7 and Windows PS5.1.
The later plan must list exact suites and current counts, not reuse prior totals.

- Positive fixed planner/build/runtime/finalization path; import/planning launch0,
  successful controller launch2 in order, failed build launch1, preflight launch0.
- Wrong target/arguments/environment/cwd/budget, existing/wrong/overlapping root,
  source/toolchain/input drift, and unsafe files fail closed without extra launch.
- Build/runtime exit failure, timeout,87/member ambiguity, forced cleanup,
  diagnostic flags, missing schema/query evidence and receipt write failure.
- Missing binary/source-token evidence; binary or source drift before runtime;
  absent tests/final PASS, injected FAIL/instrumentation; disclosed skip sample.
- Independent final product/SDK/artifact failures, simultaneous failures and
  finalization-write failure; no suppressed audit or false PackagePassed.
- Saved manifest/marker mutation, path escape/duplicate, missing STOPPED evidence,
  pending-only/contradictory markers and provisional-only results; PS5.1 JSON
  roundtrip preserves arrays, typed booleans and nested receipts at chosen depth.
- Compatibility: existing ordinary guards still reject every diagnostic class;
  mapping/matrix/build-only fixed policies and old historical records unchanged.

## 7. Handoff

After written-spec approval, write a fake-only implementation plan with explicit
review focus and test allowlist. Do not include a live invocation in that plan.
After implementation verification/review, propose one separately approved real
run with exact identity, bindings, host and phase budgets. Save necessary source,
tests and progress/recovery docs on the existing development branch. Raw data,
SDK/modules/caches/binaries stay ignored; another computer needs new preparation.
