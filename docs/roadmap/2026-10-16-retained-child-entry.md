# Retained-child controlled observation entry — 2026-10-16

## Scope and approval

User approved the bounded in-chat design: prepare a dedicated retained-handle
observation wrapper and fake-only tests, then preserve the necessary development
checkpoint. This does NOT authorize a live observation, Go, Docker, SDK audit,
matrix retry, main merge or broader acceptance. Existing short-child helper,
collector, registry and acceptance guard are unchanged.

New portable files under docs/diagnostics/isolated-fuzz:

- retained-child-observation.ps1: read-only planner, explicit native boundary,
  diagnostic evidence checker, controller and local evidence persistence.
- retained-child-observation.tests.ps1: fake native/process boundaries with real
  plan generation, fixture-body execution, converter, validation, files and hashes.

## Entry and interpretation

Dot-sourcing the helper compiles/loads existing C# definitions but launches no
process. Get-RetainedChildObservationPlan accepts only a nonexistent absolute
retained-child-observe-NNN identity beneath the existing owned diagnostic scope.
It returns a nonce, fixed encoded parent/child scripts,20s runner limit,10s marker
readiness/release limit and150ms requested child delay. It does not create files.
The child launch is hidden Windows PowerShell. No general executable/environment
or runtime override is exposed by the controller.

Invoke-RetainedChildObservation is the explicit live entry, NOT executed here.
It consumes an identity once, saves plan.json, checks six production source hashes
and the shell hash, and calls RunWithRetainedHandleDiagnostics once. The existing
500ms natural/shared5s cleanup budgets remain unchanged; native calls cannot be
preempted, so20s is a runner deadline, not a guaranteed wall-clock kill time.
No automatic retry, identity reuse, old-helper mode change or matrix enablement.

Receipt preserves both diagnostic flags and schema3. The independent observation
checker requires complete nonce/PID/time-bounded stage markers, unforced natural
core facts, complete bounded lifetime/query/release evidence, unique retained
entries, ordered phase clocks and the intended child captured running at admission
and exited0 finally on that entry. No PID lookup/reopen or core evidence replacement.
Core shutdown rows are independently checked with the existing pure schema3
same-handle predicate and must cover the original root-wait member list; registry
rows cannot fill gaps in that list. Summary booleans alone are insufficient.
Unknown, missing, ambiguous, budget-limited, forced or release-failed observations
are refused. Image lookup is optional metadata, not an identity repair.

ObservationPassed is sampled diagnostic completeness only. Acceptance and
OriginalTestPassed always remain false. The ordinary acceptance guard still
rejects these receipts; the checker never clears flags to invoke that guard.
Readiness and exit-intent markers are not OS exit timestamps. The child may be
missed;150ms is not an exact observed exit gap. Retaining handles may itself alter
process-object lifetime. None of this establishes error87's cause or repair.

Controller saves process/receipt.json, streams, result.json, and on exception
stopped.json (NoRetry=true); native exceptions also save process/launch-error.json
without inventing a native receipt or proof of cleanup. Finally, a local
evidence-manifest.json binds existing artifacts by relative path/SHA256/length.
result.json has FinalizationRequired=true and is provisional. Only completed.json,
written after the manifest, binds its SHA256 and records durable finalization.
On recovery, require that marker, matching manifest/artifact hashes and no stopped
record; a lone ObservationPassed=true in result.json is NOT durable completion.
Neither marker nor manifest is a signed attestation. The manifest does not include
itself or the later completed.json. JSON writes use CreateNew and reject path
redirection. Persistence/hash failure fails the invocation; partial files do not
prove success. A finalization failure also attempts a stopped record, even when
the earlier result was observationally complete; storage failure can prevent that
record too, so absence of completed.json must remain a stop. Partial native DTOs
on native throw remain unavailable.

## Verification and review

TDD: new test first failed at the named missing dedicated entry. Existing three
retained suites first passed55 checks onPS7. Initial implementation then passed
the new behavioral cases onPS7 andPS5.1 with no native Run or child creation.
Final scoped checks:13 suites PASS on each of PowerShell7 and Windows
PowerShell5.1,389 assertions per host (778 primary records). Both batch commands
exited0. Breakdown per host:existing eight allowed synthetic suites250,
registry27,integration12,receipt16,historical short-child fixture23,new wrapper61.
Four integration checks are static, not behavioral native proof. Additional
PS5.1 final new-suite61 and initial PS7 retained55 checks are supplemental,
not added to the389-per-host total.

Both batches began before the review fixes, but old suites do not import the new
helper and all existing collector/guard/helper sources are unchanged. The new
suite ran at the end of each batch against the final fixed source;12 final source
pins were rechecked. No claim is made that all commands restarted after every
edit. Live-process/native/Go/Docker/SDK-audit/full-project tests were excluded
by the approved scope, not silently counted as passed. The historical matrix
blocker remains open. Whitespace, PowerShell parsing and local document links
were also checked.

Ignored local evidence:
.superpowers/diagnostics/2026-10-16-retained-child/validation-001 contains26 primary
logs, four review RED/GREEN logs, one supplemental PS5.1 log and context/source
binding/summary. Its34-file evidence-manifest.json SHA256 is
`04DD40EE63AF6BAB763B48E19C457AAE9E701D8D4AEE5720AAB151833F095D68`.
All34 bound lengths/hashes were rechecked. The initial missing-entry RED and
baseline55 command were observed directly; they are not invented manifest logs.
Only necessary code/tests/docs are committed, never these generated artifacts.

One independent read-only source review found two Important issues: summary-only
core validation accepted a fixture missing final member rows; manifest failure
outside catch left a successful-looking result with no stop. Both were reproduced
as named RED assertions, then fixed in one review fix pass. The original positive
fixture was corrected to include actual typed core shutdown evidence. Five core
mutations retain true summary flags but must be refused. Actual manifest-path
directory collision and exclusive stream lock fixtures verify filesystem failure,
stopped persistence and absence of durable completion; no product/native boundary
is invoked. Reviewer did not rerun tests or re-review the fix pass; the coordinator
verifies these fixes and final regressions, not a claimed second review.

Review exclusions accepted: live native/job/timing/disposal behavior and unchanged
dependencies beyond compatibility require later scoped verification; on-disk
hashes do not attest loaded assemblies; error87/matrix/main/product acceptance
are outside this checkpoint. Test execution was intentionally excluded from the
read-only review and is reported separately.

The new suite can be rerun on a fresh Windows host with either `pwsh` or
`powershell -NoProfile -File docs/diagnostics/isolated-fuzz/retained-child-observation.tests.ps1`
(use the same `-NoProfile -File` arguments for pwsh). It has61 assertions after
the review fix pass. Test hosts themselves are started normally; generated
fixture parent/child scripts execute only in-process with fake launch/sleep,
and the actual native Run boundary is replaced. Thus these tests do not prove
the one-line native dispatch shim or OS behavior dynamically; shim selection/
arguments were checked by source review. File link metadata APIs are used for
local evidence safety, never to query a process.

## Fresh-computer continuation

Recover the existing development branch using
[repository recovery](../runbooks/repository-recovery.md) and current-status.md.
New tests need only the repository and suitable Windows PowerShell/.NET hosts;
they do not need ignored receipts, private Go caches or an SDK. Older envelope
regressions have their existing local fixture/cache prerequisites, not restored
by cloning. Raw evidence stays ignored and is not part of Git recovery.

For a later separately approved real observation, review the exact fresh identity,
source/shell bindings, fixed limits and interpretation first. Use a fresh host:
on-disk hashes are provenance, NOT proof of an already-loaded Add-Type assembly's
source identity. Live ownership, adapter correctness, disposal and timing remain
unverified. Do not run a matrix as the next automatic step. Historical matrix003
remains2PASS/1FAIL/23UNRUN; broader C11/C12/Task10/B01/production/main gates stay open.
