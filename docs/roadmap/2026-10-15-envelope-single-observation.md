# One Envelope observation from a validated compile — 2026-10-15

Subsequent readonly comparison confirmed all701 historical/current product hashes
match, but runtime/cache differences and missing old operand records prevent
exact historical attribution. See [old/new comparison](2026-10-15-envelope-evidence-comparison.md).
No additional observation or code change occurred in that follow-up.

## Approved scope

The user approved an independent single-observation entry, tests/review, then
exactly one original Envelope observation from local compiled003 under fresh004.
No rebuild, query/preparation replay, second sample, retry, Pair, broader matrix,
download/install, WPR, security change or main merge. Keep5s fuzz/one parallel
worker,30s test timeout,60s native bound and all ownership/natural cleanup/trace
gates. Neither observation success nor an ordinary checkpoint is product acceptance.

## Implementation and binding

New envelope-sample.ps1 imports the existing driver without modifying it or the
native runner. Get-EnvelopeCompiledSamplePlan requires a fresh owned target and
a distinct owned compile identity. It validates seven fixed, single-link,
non-reparse inputs: compile result, binary binding, build receipt/log, before/after
product inventories and executable. Strict compile-only success/non-execution/
non-acceptance flags, saved root/file paths, copied-source identity, actual binary
hash and original callback pin must match. The successful schema3 build receipt
must pass existing completion gates and match the copied executable, exact compile
arguments and offline environment. Both copied-source and original Envelope
compiler-log parsers run. Saved before/after/current product inventories must match.

Assert-EnvelopeSampleInputs rechecks all seven artifact hashes and copied source
identity before launch and after observation. Final validation reuses the freshly
read final inventory rather than rescanning it again. Invoke-EnvelopeCompiledSample
uses the original execute arguments/environment with only the fresh run's cache/
temporary paths. There is exactly one native invocation and no compilation call.
Raw receipt/stdout/stderr persist through the existing recorder. Trace, worker,
watchdog, exit and completion checks use the existing observe predicate.

Successful sample-result.json and success output publish only after final product,
source and artifact checks. Failed observations save failed results/stopped data
without success output; finalization failures downgrade pending success. All
results retain Acceptance=false. Planning refusal creates no target; any consumed
target must not be retried, including partial setup/launch failures.

These are checkpoint-consistency checks, not authenticated attestations or an
atomic filesystem snapshot. Concurrent adversarial replacement is not solved.
If a native call throws before returning a receipt, SamplingExecuted=false means
no returned receipt, NOT proof no process ran; SamplingAttempted=true and the
consumed identity remain authoritative against replay. A second finalization
failure can replace a primary validation message, while raw evidence is retained.

## Test-first checks and review

New envelope-sample.tests.ps1 first failed at the missing independent entry.
After implementation22 behavioral assertions passed on PowerShell7 and5.1.
Tests create fresh on-disk compile records/logs/synthetic receipts and a never-run
unit binary. External execution and full product inventory are substituted; real
planner, file/link/hash checks, copied-source provenance, compiler parsers,
completion/trace gates, driver, result files and streaming consumption run.
They do not establish actual process ownership or real fuzz behavior.

Coverage: one invocation without rebuilding, bound false-acceptance result and
final inventory; prelaunch refusal for wrong mode/root/source/hash/log/product,
failed build completion or used identity; runtime exit/timeout/watchdog rejection;
post-run binary/compile-record drift, no premature success and preserved failures.
Existing19 compile-only,19 preparation-only and21 planning checks also passed on
each host:81 scoped assertions per host. The final new22 tests rerun after adding
optional Toolchain/ModuleCache parameters and a current-user cache default. A new
machine can pass independently prepared prerequisite paths, not a Lenovo profile.
The tests do not consume old raw receipts; the verified copied toolchain/cache
remain explicit preparation prerequisites. No full runner/product suite was rerun.

Independent read-only review found no Critical/Important defect and supported
one authorized sample after checks, not product/main acceptance. Minor limits
include the receipt-return ambiguity and primary/final error replacement above,
and missing direct malformed-trace/final-inventory-read-failure tests. Actual
binary behavior, acceptance, rebuilds/retries and atomic filesystem replacement
were not judged. No reviewer ran tests/programs or changed files.

## Single actual invocation

One invocation on PowerShell7.6.5 used fresh envelope-observe-004 and the003
binary, without recompiling or re-running version/origin. Before invocation the
operator also checked003's previously reported binary pinF422D2CB... and compile
result pinD7C82ACA...; the helper then performed its complete input validation.
Wrapper exit0; SamplingExecuted=true, ObservationValidationPassed=true,
ObservationValidationFailure=null, Outcome=observed-success-not-acceptance and
Acceptance=false. No second attempt, build or wider arm ran.

Native UTC interval2026-10-15T11:32:50.4764449Z–11:32:57.5781448Z
(Asia/Tbilisi15:32),7101.6999ms. Root exited0, no timeout, complete natural shutdown
and cleanup verified, no termination requested. Final accounting0/member list
empty, ShutdownElapsedMs6. TotalProcesses3 is aggregate accounting, not a process
identity claim. A readonly saved-receipt check confirmed these fields.

The original arguments are unchanged: -test.run=^$, -test.fuzz=^FuzzIssuerEnvelope$,
-test.fuzztime=5s, -test.parallel=1, -test.timeout=30s, -test.count=1 and004's fresh
fuzz-cache. Stdout is PASS. Stderr reports five baseline inputs completed, one
worker,73921 executions and six new interesting inputs (total11). Those are
coverage-interesting inputs, not six bugs. Progress reports elapsed6s despite the
5s fuzz setting; native interval includes startup/shutdown and is not an exact
5s wall-clock guarantee. No watchdog or test/native timeout occurred.

Three ordered observe trace records passed the existing gate: parent-done stop
seq1, worker stop seq2 (each activeWorkers1/newError nil), then return seq3 with
activeWorkers0/newError nil. All watchdog flags false. Worker trace and runtime
progress, not aggregate process count alone, support that fuzzing actually ran.

All seven003 artifact hashes remained identical after observation. Binary remains
F422D2CBD6D9F2A6B34B5363CD45CC48E88AA49CDF0FF8DAB0D74DC765A798A6 and copied
source identity E3949C0C3EC7F5D7BB8D239CEF2BBD2B753EC162070C8FB289DC8A60799ED306.
Product inventories match701 before/701 after, zero Path/SHA256 differences.
004 has no build directory or stopped.json. Its fresh cache/raw files stay local.

This sample did not reproduce the older original Envelope deadline failure. One
short observation using the instrumented copied runtime is not proof that the
historical failure is fixed or that official-runtime/product acceptance passes.
Original001/origin87 remains unexplained; Pair and broader matrix revalidation
remain unrun. C11/C12/Task10/B01/production/main gates remain open.

## Evidence and fresh-computer recovery

Ignored worktree-relative evidence root:
`.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/envelope-observe-004/`.
003 is the readonly compiled source artifact;004 is now consumed. Neither may be
replayed as an automatic retry. The lack of stopped.json here is an observation,
not general permission to reuse an execution identity.

| Local evidence under004 | SHA256 |
| --- | --- |
| execute/receipt.json | D0F24957BD9385C1A552467725E9A956864FD73E8061B6EF4F526D7112B192BE |
| execute/stdout.txt | C26DE83ABDC9496CD1301470918EC39ECCA1CF389EF0AE1C6504DA1800D1C431 |
| execute/stderr.txt | 7B97D4BDED33C3CD505F33786342BE2D9A0788B4BD4225FB29ED0157824EC156 |
| sample-result.json | D007A6C8ACCD2A29E1F06916CC8CED1BFAB4DE1E9067B6B8554875C386FF4B08 |
| plan.json | D9CE15EB662E41CFD359F1A833C980A45DCE58E417819D458DBE2E51D646757F |
| before-product.json and after-product.json | 5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |

Save the new helper/tests and necessary public reports to the existing
codex/c12-b01-task9-coordinator branch and verify remote SHA at handoff. On a new
computer restore that latest development checkpoint, not automatically main;
read current-status.md and docs/runbooks/repository-recovery.md. Raw receipts,
compiled binaries, caches and generated interesting inputs are not committed.
Source/progress recovery does not restore the003 binary or authorize another run.

The new helper tests synthesize their own compile records and never-run binary;
they do not read003/004. Supply -Toolchain and -ModuleCache to the test for a
separately prepared verified diagnostic copy and existing offline dependency cache
on another machine. Default cache uses the current user's profile. Missing/drifted
prerequisites fail; there is no download/install fallback. Reconstructing a real
compiled artifact and any new observation require their own scope/authorization.

Next proposed step is a readonly comparison of the saved successful trace and
historical failed evidence to define the next validation scope. Do not infer that
this success authorizes Pair, full matrix reruns, retry-until-pass or main merge.
