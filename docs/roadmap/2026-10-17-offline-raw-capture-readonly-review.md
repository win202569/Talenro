# Offline raw capture — independent readonly review, historical blocked disposition

Status at review: **REVIEW1/1 COMPLETE; Important1 OPEN, Minor1 DEFERRED; INCREMENT UNFINISHED.**
No Critical issue confirmed. All10 test slots and the sole review seat are spent.
No automatic repair, reproduction, retest or additional review is authorized.

This is the original review disposition. Later explicit user approval and the
separate [bounded repair record](2026-10-17-offline-raw-capture-empty-ownership-fix.md)
govern subsequent I1 work. They do not rewrite this review or grant a new review.

Later M1 follow-up after81afee10: the user approved a tests-only design and
four NEW hosts. Eight literal complete event sequences plus alteration-rejection
cases now pass targeted and full30-label suites on both engines, product and
dependencies unchanged. Parent marks M1 addressed within that bounded evidence;
the independent reviewer did not rereview this delivery. See the separate
[M1 record](2026-10-17-offline-raw-capture-event-oracle.md). Original findings,
severity and limitations below remain the historical review, not rewritten green.

One fresh reviewer, offline_raw_capture_review, read actual source/tests,
approved spec/plan, continuation and saved evidence. Range:
5915bc732a658973b368b40a681250388911854e..a54fe96f7eb6c9573918f4c878df1aa67de95d2f,
limited to this increment/necessary dependencies. No tests/imports/product
execution, writes, network/security queries, trust changes or delegation.
Parent independently checked cited locations and agrees with the severities.

## I1 — Important: empty Actions exposes shared runtime state

New module offline-raw-capture.ps1:83 assigns List<object>.ToArray() directly
to returned Actions. For an empty list on PS7/Core this is a cached array;
the object[] cast does not create an independently owned snapshot. The same
cache is used by internal Events at127 and contract Read-ContractStoredEvents
at173. Adding an ETS Length/Rank NoteProperty to returned empty Actions can
contaminate later uses of that shared array.

Every coordinator call performs zero-event header verification at54–55. Its
public verifier reconstructs empty events through contract173, then rejects
the decorated bounds. A fresh valid scenario, including nonempty actions, can
return invalid-header because an earlier result was modified. Existing
empty-event record verification can also be affected; calls are not isolated.

The unchanged contract acknowledges this cache hazard at129 and clones its
public event arrays at130; packet construction clones at97. The new Actions
path lacks that protection. The new test's CopyOwnership at136–143 uses nonempty
actions and misses this case, despite catching the nonempty alias mutation.

Evidence qualification: **static defect, not dynamically reproduced during
review or parent reconciliation.** Receipts identify PS7.6.5 Core; current
corresponding pwsh.runtimeconfig.json declares net10.0/Microsoft.NETCore.App
10.0.11, also read by parent. PS5.1 manifestation is not established and must
not be generalized. Parent checked the public verification chain at contract
152–173/190 onward, accepts Important severity and stops. Source is unchanged.

## M1 — Minor: incomplete independent mapped-event expectations

New test DerivedConsistency at158 checks four status literals, public record
verification and selected intent/index values, not every complete mapped event
array against independent expected fields as promised by the plan. Other cases
check individual fields but do not supply the full oracle. A changed mapping
could remain internally consistent and escape those assertions.

Parent retains Minor: this is an unfulfilled test-strengthening promise, not
another confirmed product mapping defect. It is deferred, not claimed covered,
and not automatically repaired with I1.28 labels prove narrower coverage.

## Strengths and evidence reconciled

Reviewer confirmed actual owned MemoryStream reads, exact returned-byte copying,
short-read/EOF distinction, pre-read stop precedence, prefix/index retention,
no implicit drain/text reconstruction, passive full preflight and strict bounds.
Nonempty snapshots, reversed-UTC refusal, classification versus completion,
conservative flags, import direction and lack of live/persistence path are
implemented. Static finally tracing covers partial construction, Dispose and
clearing of source/argument copies and the reusable buffer, not injected faults.

New module4C0E6C10/test7010CBCF, all12 prior pins, manifest/index/controller and
all ten raw receipt hashes independently match. Original rows1..5 equal the
stopped snapshot; slot5 remains ExpectedResult=false. Exact assertions, four
28-label inventories, two110-label inventories and all elapsed/timeout facts
reconcile. Actual stderr is empty in all ten runs. Slot9/10 selection and the
disclosed new-suite TestSHA256 index field agree with the frozen controller.
The slot5 exception does not retrospectively promote its verdict. Full hashes
and per-slot results: [execution report](2026-10-17-offline-raw-capture-execution.md).

Git ownership protection blocked reviewer Git reads; no bypass was attempted.
Base/HEAD, branch/clean status and diff provenance remain supplied parent facts,
not independently verified Git facts. Parent accepts this limit together with
independent filesystem hashing, not as whole-branch approval or a trust change.

## Declined to judge — parent dispositions

Every limit remains explicit, never treated as a pass:

|Reviewer limit|Parent disposition and cost|
|---|---|
|Dynamic reproduction of I1 prohibited|Keep static blocker; an authorized future RED must establish actual symptom before repair|
|PS5.1 manifestation unestablished|No claim it occurs there; future compatibility GREEN is separate|
|Real pipes/draining/scheduling/hanging reads/cancellation/read faults|Memory streams do not prove them; real collector remains open|
|Watchdog/timeout cleanup/termination/hard wall-clock bound|Virtual times and non-timeout receipts do not prove these; timeout path unexercised|
|Runtime Dispose/allocation/interruption/catastrophic cleanup faults|Static tracing only; no fault-recovery guarantee|
|Concurrent input or hostile runspace/type-data mutation|Stable-input scope only; no concurrency/hostile-host assurance|
|Maximum transfer/performance|1MiB admission and4096 ticks are not full-byte transfer/worst-case benchmarks|
|Exhaustive callbacks/tokens/malformed inputs/combinations|Named cases are bounded, not exhaustive proof|
|Nonempty CLIXML stderr branch|All actual stderr empty; parser not dynamically exercised by this packet|
|Execution authentication/retained mutant-source snapshots|Receipts/diffs/hashes are consistency evidence, not attestation or independently inspected mutant files|
|Actions persistence/crash/durability/archive acceptance|Wrapper/Actions not persisted here; old archive scope unchanged|
|Command provenance/cleanup/security/Stage B/I2/I3/Task10/11/full C12/main/release/historical branch|Remain excluded and open; this review grants none|

## Proposed bounded I1 repair — historical proposal, subsequently approved below

Short design: explicitly clone the new module's returned Actions and outbound
Events arrays, following the existing contract pattern, without changing any
of the12 dependencies or interfaces. Empty-only allocation is an alternative
with fewer nonempty copies but another branch; uniform cloning is recommended.
Add RawCaptureEmptyOwnership: annotate returned empty Actions, verify its Record
still verifies and a fresh valid nonempty scenario still creates a record.
Remove only test-added ETS decorations in finally even on RED; check public
behavior, not only ReferenceEquals. M1 stays deferred.

Proposed NEW budget: at most5 sequential synthetic children,120000ms each:
1 PS7 targeted new case against unchanged source, expected exact RED;
2 fixed PS7 targeted GREEN;3 fixed PS5.1 targeted GREEN;
4 fixed PS7 full new suite;5 fixed PS5.1 full new suite.
Full inventory becomes old28 plus EmptyOwnership before Complete,29 total;
targeted group selects ApiRequired/EmptyOwnership/Complete. Freeze amended test
before1, fixed source before2. Unexpected failure stops, no retries/deadline
extension or extra repair. Never reuse the old ten slots.

No new reviewer proposed; the old seat remains consumed. This is only a pending
bounded short design, not an implementation plan or approval. Code/tests may
change only after explicit user approval. Preserve all old records/hashes,
M1 and other limits. No protection changes or real collection.

## Subsequent user approval after f8657855

The user explicitly answered yes to the bounded proposal above after checkpoint
f8657855722927630d17a0b23c04f8e8767dc66a. This authorizes only I1's two owned
array clones, one EmptyOwnership regression and at most five NEW sequential
synthetic test hosts as specified. The old ten slots and reviewer remain spent.
No retries after an unexpected result, no new reviewer, no M1 implementation,
no dependency edits or live/security/main authority. The historical proposal
wording above records the pre-approval state, not the current authorization.
Follow the [separate repair record](2026-10-17-offline-raw-capture-empty-ownership-fix.md)
for actual execution and remaining limits; do not rewrite original receipts.
