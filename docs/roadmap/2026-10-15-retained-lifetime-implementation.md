# Retained-handle diagnostics — implementation checkpoint

Follow-up2026-10-16:the deferred M1 populated nested-receipt test is now covered.
Receipt checks increase9 to16; three portable suites55 PASS per host. No product
implementation change or rerun of the historical298-per-host suite is implied.
See [coverage and recovery](2026-10-16-retained-receipt-coverage.md). The result and
review narrative below describe the original2026-10-15 checkpoint.

## Approved scope and actual changes

User approved the written spec and plan, choosing inline implementation. Base
c4069cc63aef1021851df545f6d0d8f2a812d73a. Task1 bf0c8376 adds the portable
registry/fake fixture/tests; Task2 658571f5 wires the opt-in adapter/entry.
Task3 adds optional receipt serialization and explicit diagnostic-only refusal.
No actual native Run/Go/matrix/real-child probe/PID query or SDK audit was invoked.

RunWithRetainedHandleDiagnostics is a new explicit native entry; existing Run
overloads keep retained mode off. Registry admits only own-job verified handles,
at most32 attempts including failure, with no retries/eviction/reopening. Handles
remain private; scalar query/open/membership/read clocks and errors are sampled
metadata. Initial-open error87 stays unverified. Registry observations never
replace original MemberEvidence/ShutdownMemberEvidence or join identities by PID.

Diagnostic polling uses up-to10ms root slices, existing runtime deadline and the
same shared500ms/5s shutdown clock. Observations are zero-wait only. Post-forced-
cleanup observations are explicitly marked termination-requested, not natural.
Finally runs original cleanup then attempts each retained release once, records
failures and preserves core exceptions. Default mode and matrix do not enable it.
Both diagnostic flags are set in this mode; either flag independently prevents
Assert-ProcessCompletion from accepting even otherwise successful-looking data.

The new source is included in run.ps1's Add-Type load and prerequisite inventory.
No SDK/template/Go source, original callbacks or matrix recipes were modified.
701 recorded product hashes were rechecked unchanged against matrix003 inventory.

## RED/GREEN and verification scope

Task1 RED:missing registry contract; GREEN22 checks on PS7.6.5/PS5.1.
Task2 RED:missing new entry; GREEN11 integration checks on both hosts. Seven are
compile/reflection/fake cleanup behavior; four are explicitly supplementary static
wiring checks, NOT native behavioral tests. Task3 RED:lost optional metadata;
GREEN9 actual converter/guard synthetic-receipt checks on both hosts.

Self-review found CloseHandle's import did not preserve last-error for the new
release diagnostic. A reflection-based RED exposed it; SetLastError=true fixed
the declaration, GREEN12 integration checks/host. No real close API was tested.
Independent whole-increment review found two Important issues: contradictory
same-entry exited-to-running state and enumeration overrun on empty/root/seen-only
lists. Both were reproduced with fake dependencies and fixed in one RED/GREEN
pass:27 registry checks/host. Both raw observations remain preserved with explicit
Ambiguous metadata; query overrun sets incompleteness before any admission work.
An empty-array PowerShell fixture initially became null through pipeline expansion;
after correcting the fixture, the missing overrun check was independently watched
RED again before restoration. This was not a native rerun.

Deferred nonblocking Minor:populate the receipt-roundtrip fixture with all nested
entry/phases/cap fields. Current roundtrip uses an empty report; populated registry
JSON excludes handles in a separate test. This limitation is not hidden as coverage.
Reviewer ran no tests or native/PID queries. Real ownership/release/timing, end-to-
end Run, final regression totals/remote recovery and broader acceptance were
explicitly outside its judgment. The first two remain unverified; final regression
checks below are controller-verified, and remote equality is checked at push time.

New suites are portable and never pass fake handles to real APIs. They exercise
real registry ownership/disposal and real receipt conversion/refusal through fake
dependencies. Final results:11 suites/298 checks PASS on each PS7.6.5 and Windows
PowerShell5.1.26100.9711, plus2 supplemental PS7 checks with the existing matrix
RuntimeEvidenceOnly option; all exit0 (598 assertion records in this scope).

| Suite | PS7 final PASS | PS5.1 final PASS |
| --- | ---: | ---: |
| retained-lifetime |27 |27 |
| retained-lifetime-integration |12 |12 |
| retained-lifetime-receipt |9 |9 |
| matrix-check synthetic |60 |60 |
| pair-sample synthetic |45 |45 |
| envelope-uninstrumented |28 |28 |
| envelope-sample |22 |22 |
| envelope-compile |19 |19 |
| envelope-preparation |19 |19 |
| envelope-observe |21 |21 |
| mapping-single-check synthetic |36 |36 |

Final new-suite logs supersede their initial22/11/9 records. The PS7 full old
matrix suite began before the final registry/close-metadata fixes; its synthetic
default-mode path never consumes that registry or calls native CloseHandle.
After fixes, all three new suites reran on both hosts, the full PS5 old regression
batch used the updated source, and the supplemental PS7 matrix runtime-binding
case reran against frozen source. This is the exact verification sequence, not a
claim that every original PS7 command restarted after every source change.
Some existing Envelope planner/provenance fixtures read the retained local copied
SDK/cache/receipts; they do not launch Go or constitute an SDK audit. Their results
must not be represented as dependency-free fresh-machine verification.

Ignored local logs/context/summary are under
.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/retained-lifetime-validation-001.
They are machine-local, not restored by Git clone. No historical receipt/pin was
rewritten. New source hashes are saved separately with final unit evidence.

Final source binding covers8 changed source/test files without drift. The local
evidence-manifest.json binds34 unit-log/context/summary/source-binding files; hash:
`194A75F1A250A7D0A74AC73841832336C3D16759B9FF9DDAFE4EA906A177E02E`.
final-summary.json contains22 final suite entries; the two supplemental assertions
are in ps7-final-matrix-runtime-binding.log. Final counts/manifests and ordinary
checkpoint remote recovery are controller-verified, not an independent reviewer
rerun or OS/native attestation. Compilation/test hosts are allowed machinery;
"no native run" means no diagnostic Run entry, Go run or target-child probe.

Execution adaptations:project PowerShell-only instructions take precedence over
skill Bash bookkeeping helpers; a manual plan-scoped ledger was used. Full Go/
live-process baseline was deliberately not run under fake-only approval; its
absence leaves broader project/native runtime risk unresolved. The reviewer did
not re-review the fix pass; coordinator RED/GREEN and frozen-source regressions
close the two Important findings under the selected execution workflow. The
nonblocking Minor above remains deferred. No additional scope ruling is implicit.

## Deliberate limits and next authority boundary

Real native ownership, API adapter timing, resource release behavior and observer
effects are not established by fake/static checks. Retaining a handle can change
process-object lifetime; sampled coverage is never complete. If native Run throws,
partial in-memory lifetime observations need not persist under existing launch-
error behavior. Close failures mean release attempted, not guaranteed OS release.
This is not a fix/cause proof for error87 or a matrix/product acceptance claim.

No live PowerShell invocation wrapper or matrix enablement is provided. A future
controlled native experiment needs a separately approved bounded entry, identity,
limits and result interpretation; do not directly invoke the new entry merely to
obtain a PASS or replay matrix003. Matrix003 remains2PASS/1FAIL/23UNRUN. C11/C12/
Task10/B01/database/full acceptance/production/main gates remain open.

Recover the existing development branch and read current-status.md,
docs/runbooks/repository-recovery.md and the approved
[spec](../superpowers/specs/2026-10-15-retained-handle-diagnostics-design.md)/
[plan](../superpowers/plans/2026-10-15-retained-handle-diagnostics.md).
