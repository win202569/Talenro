# Isolated fuzz diagnosis

Latest bounded diagnostic-only increment: opt-in `CaptureShutdownDiagnostic` on
Invoke-RecordedProcess records a separate DiagnosticShutdown member observation
on the same retained, live-job-verified handle. Only running-after-confirmation
members with missing image metadata qualify; unknown ownership/core queries skip.
One wait uses the remainder of the shared500ms shutdown window, not a fresh window
per member. Core state/exit/image/first/confirm fields stay untouched. Live pre-root
sampling and ordinary arm execution do not enable this mode. Member cap32 and
shared5000ms cleanup remain unchanged; scheduler latency is not hard real-time.

Every opt-in receipt carries ShutdownDiagnosticOnly=true; Assert-ProcessCompletion
refuses it even if the process naturally finished with an empty job. Diagnostic
observations never substitute for core acceptance evidence or launch an arm.
shutdown-observation.tests.ps1 runs real short/long owned process waits plus
ownership/query/window/no-upgrade controls. It reuses11 same-handle assertions;
22 additional assertions give33 in that script and131 in the full scoped suite.
CoreObservedUTC separately captures the core snapshot before any later diagnostic
wait; legacy ObservedUTC still describes inspection completion. A real native
live-child test protects this ordering and owned cleanup.
Literal image-error boundary rows are controlled inputs, not proof that those test
children naturally produced Windows image errors. No product full-suite/fuzz ran.

One authorized copied-Go version probe, shutdown-probe-001, ran2026-10-14 UTC17:51:32.
It did not hit the running+missing-image target: the owned running member had a
readable conhost.exe image, so observation was skipped as not-needed. Exit0,
bounded owned cleanup verified; acceptance deliberately refused. This cannot
identify old failed members or explain their image errors. No second probe,
remaining fuzz execution, guard relaxation or acceptance follows from a no-hit.
Probe receipts/caches remain local ignored data, not recoverable from Git.
The single probe preceded the CoreObservedUTC addition; no probe was repeated to
populate that extra field. Final native/serialization behavior is component-tested.

Current checkpoint2026-10-14: baseline passed once; framework copy now instrumented.
The approved exited-state/image split and one same-handle confirmation are verified.
observe-prep-002 and parent-first actual fuzz passed their diagnostic checks;
synchronized stopped during version preparation with the confirming wait still
running and missing image metadata. synchronized/crash/cancel actual fuzz remain
UNRUN. No retries, old failure reclassification or main merge.
[Review/dispositions](../../roadmap/2026-10-14-isolated-fuzz-checkpoint-review.md).
Entries below preserve historical checkpoints, not the latest state.

The latest continuation passed109 PowerShell assertions. Member receipts retain
InitialState/InitialExitCode/InitialExitCodeRead and ConfirmationState/UTC/attempt;
the producer performs only one zero-wait on the same retained, job-verified handle
before closing it. A signaled confirmation refreshes the exit code; running with
missing image, failed queries and unknown ownership remain failures. No new delay
or budget is introduced. same-handle.tests.ps1 uses real owned process handles.

The explicitly approved call was `Invoke-DiagnosticArm <verified-run-root> observe
-PreparationAttempt prep-002`. Preparation identities must match prep-NNN and
require an existing failed nonbaseline preparation. Original or any prep execute
directory prevents another execution; duplicate identities are rejected. This is
not authorization to generate retries. Failed synchronized version preparation
is immutable; further execution needs a separate bounded investigation/approval.
Fresh-PC recovery reads the execution report before creating any new run root;
never repeat completed baseline/observe/parent-first merely to restore progress.

Latest bounded revision: exact verified ownership + signaled exited state +
successful exit-code query permits missing image metadata. Unknown ownership,
state/query/unread exit still fails; missing-image running members do not pass.
ShutdownImageEvidenceComplete is recorded separately from ownership/liveness
completeness, with policy tag owned-signaled-exit-image-optional-v1. Budgets and
forced-cleanup refusal unchanged.9 state regressions RED→GREEN,90 PS assertions PASS.
New component gate8 logic tests and completion passed, including an exited owned
member with ImageError31 whose metadata stayed missing. Old failures not rewritten.
The first observe preparation stopped at go-env origin: initial wait state running,
later exit code0, image error5. Non-atomic reads do not prove a signaled exit; current
guard refused/owned cleanup verified. No fixture build or actual fuzz execution.
Five actual arms remain UNRUN; observe now has a failed preparation directory.

Use PowerShell on Windows. Read the approved2026-10-14 spec/plan and AGENTS.md.
This directory currently contains the verified isolation utility, not a complete
framework reproducer. Never run scripts against a project/cache root as target.

Checkpoint: context gate and trace templates passed7 Go tests; materialization
passed8 assertions. coordinator.patch now contains both context and actual
internal/fuzz tracing changes. The runner is implemented but its real Go build
handoff is blocked: baseline compilation exited0 while job accounting reported1
remaining process at root exit, cleaned to0 afterward. No actual fuzz arm ran.
Do not retry baseline or run later arms without new bounded authorization.
The user subsequently approved bounded runner repair and one new baseline
preparation, identity retry-001. Historical receipts remain unchanged. Use
`Invoke-DiagnosticArm <verified-run-root> baseline retry-001` only under that
approval: it creates arms/baseline-retry-001 and binaries/plain-retry-001 with
fresh caches, refuses duplicate identities, and cannot retry other fuzz arms.
Before reusing an unmodified copy, revalidate its complete manifest and links
against the original copy receipt; never adopt the failed build's binary.
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
proof of real Go integration. That checkpoint also recorded a compile-log
path-escaping incompatibility (raw doubled backslashes), repaired below without
bypassing provenance guards. Keep the failed launch/build receipts unchanged.

Current repair: live owned-job membership is recorded at root wait and after
cleanup. Descendants exclude only the exact root creation ID, while raw active
accounting is retained separately. Cleanup still requires both accounting and
member list empty, within the existing5s cleanup bound; query failure is fatal.
Observed genuine descendants remain a failed gate even when cleanup succeeds.
This does not establish what the historical remaining1 represented.
Compile evidence now requires exact decoded Windows tokens for the copied
compiler, correct package and source in the same command, not loose substrings.
Run runner-regression.tests.ps1 (13 assertions) and run.tests.ps1 (18 assertions,
including a root that exits while an owned child remains) to verify the repairs.
Raw membership uses [JobObjectBasicProcessIdList](https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-jobobject_basic_process_id_list).

retry-001 outcome: pristine copy revalidated11536 files, baseline build and
actual5s fuzz passed,238579 executions, no remaining members and no termination
needed. Original full11536-file manifest and701 product hashes remained unchanged.
A fresh controls-only model run printed11 preserved controls but failed the
process gate: after the root exited, the owned job still listed one different
member. Owned cleanup reached empty/zero. Member image/liveness was not captured;
do not query recycled receipt PIDs or treat cleanup as acceptance. No further
arms ran. Task3 remains blocked and final review pending. Once any baseline
retry has launched fuzz, the retry helper refuses additional attempts.

Bounded live-identity investigation now records MemberEvidence before cleanup:
open only IDs from the current job list with limited query/synchronize rights,
verify exact-handle job membership, then read image and zero-wait state/exit code.
Query failures and ownership loss remain unknown; they never clear existing gates.
At most32 identity records are collected. Optional final boolean argument to
Invoke-RecordedProcess enables MembersDuringRun with10ms polling; default remains
off/100ms. This timing-perturbed observation is not a natural fuzz/acceptance arm.
State and exit code are separate reads, not an atomic snapshot; exit code0 does
not prove the prior wait observation was exited. No archived PID is queried.

Two finite model-only observations were recorded, no baseline rerun: first had
no members available at capture; opt-in live observation captured an owned
C:/Windows/System32/conhost.exe during root execution and at root exit. It again
failed DescendantsRemaining1, then owned cleanup reached zero/empty. All11 model
controls were preserved. This identifies this new failure's member, not the old
ones, and does not establish why the console host was created or its exit timing.
See [Microsoft's console-host definition](https://learn.microsoft.com/en-us/windows/console/definitions).
Current tests:20 runner assertions+13 regression assertions passed. Later arms
and final review remain pending. Any change to natural shutdown waiting, launch
flags or member acceptance needs a separately approved bounded design.

Approved bounded shutdown revision (schema2): after root exit, observe for at
most500ms within the existing shared5000ms shutdown clock, including evidence
collection and forced cleanup. No named-process exemption; unknown or truncated
shutdown evidence, timeout, or any requested termination cannot pass. Scheduling
and native-call latency are not hard real-time guarantees; actual elapsed times
are recorded and an over-budget successful receipt is refused.
Initial DescendantsRemaining/ActiveAtRootWait/MembersAtRootWait remain raw facts.
NaturalShutdownVerified plus zero accounting/membership at the natural decision
and final check determine completion; forced CleanupVerified is only cleanup.
Schema-less historical receipts retain the old strict remainder gate, not a
retroactive reinterpretation. Version/origin/build/arm checks share this validator.

Real short-lived child and natural-decision validation RED→GREEN; long-lived
children/timeouts still fail. Final31 runner+13 regression+8 materialization
assertions passed. One new controls-natural-001 run preserved all11 model controls;
initial owned conhost remainder1 became zero/empty naturally in25ms, no termination
requested, schema2 completion passed. It does not rewrite previous failures or
explain console-host creation/historical fuzz failure. Baseline was not rerun;
later five arms and final review still pending. Necessary source/report files
are recovery artifacts; machine receipts/caches/binaries remain excluded.

Fresh-computer static prerequisite check (no toolchain execution, copying,
downloads or generated receipts): dot-source preflight.ps1, then call
`Test-DiagnosticPrerequisites <absolute-source-toolchain-path>`. It checks Windows,
independent pinned go.exe/context.go/fuzz.go, VERSION and portable materials.
Ready only means these prerequisites match; full manifest/link validation during
New-IsolatedToolchain and actual version/GOROOT/compiler evidence remain required.
preflight.tests.ps1 passed4 behavioral assertions, including no generated files.
On another PC, pass its pinned toolchain location with `preflight.tests.ps1 -Source
<absolute-path>`; the test does not require a Lenovo user profile.
The test uses a tiny source fixture; it is not a complete11536-file toolchain.

Current framework continuation stopped before observe: patch installed only in
framework-001. Context5 and trace3 logic tests passed; both compiles passed. The
trace test process exited0 but its shutdown gate failed because a verified owned
member was already signaled/exited with code0, while image lookup returned31.
Current policy treats that missing evidence as failure and forced owned cleanup
then verified0/empty. Image was not captured; do not infer conhost from previous
runs. All five actual diagnostic fuzz arms remain UNRUN; no gate/arm retry or
baseline rerun. Full original11536-file manifest and701 product hashes unchanged.
Resolving known exit state versus missing image identity requires a bounded
policy/design decision, not bypassing the guard. Independent checkpoint review
does not turn these unrun arms into acceptance.

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
It now saves copy-receipt.json in the new RunRoot after verification, via CreateNew.
Do not manually overwrite it. Saved/returned Source/Toolchain/Cache and manifests
match; Invoke-DiagnosticArm consumes that receipt. An incomplete copy is unusable.
The fixed version is Go1.26.5 windows/amd64, independently prepared on a fresh PC.

Generated copies, caches, fixtures and receipts are not recovery artifacts and
must not be committed. Rebuild from tracked scripts/templates; source hashes,
actual executable/GOROOT and runtime results still require verification. Original
matrix3 failure and Pair-unrun remain open; no product acceptance or main merge.

provenance.ps1 requires the exact installed6-file set/current4 templates/patch and
binds binary receipts to canonical source identity. Missing/stale identity refuses
reuse without rebuilding or retrying an existing arm.7 provenance assertions PASS
including template-only drift. Old baseline results remain historical facts, not
permission to reuse unbound binaries.
