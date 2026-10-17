# P1 supervision policy — scoped review and stopped handoff

Status: **UNFINISHED / IMPORTANT I1 BLOCKS ACCEPTANCE.** Tests8/8 and the single
review seat1/1 are spent. No repair, reproduction, extra import/test or re-review
is authorized. Preserve all raw. This supersedes earlier review-pending entries.

## Evidence and scope

The fresh readonly reviewer inspected the complete plan/spec/module/tests/report,
controller, manifest/index and all eight stream pairs for c2fd4b5063e97fc1eebb46a6279d3ba24848b184
through e8bb42e70b13a97e52d47d56cfebdc972ef85137, not historical whole-branch correctness.
No import, test execution, edit, security/network query, trust change or delegation
occurred. The reviewer independently matched module/test/manifest/controller/index
hashes, all14 protected pins, all16 stream hashes, two precise API REDs, four exact
ordered35-label inventories and both mutation assertions. All stderr was empty.
See [execution record](2026-10-17-offline-supervision-policy-execution.md) for hashes,
durations and contexts. Saved evidence is not newly observed execution.

Strengths: bounded deadline arithmetic/equality expiry, sticky first failure,
fixed false authority flags/unknown actual cleanup, complete literal comparisons,
and separation of evidence completeness from technical success.

## Important I1 — decorated scalars retain caller-owned mutable references

No Critical or separate Minor finding. Locations in
[module](../diagnostics/isolated-fuzz/offline-supervision-policy.ps1): passive/type
checks3–8/20–31, shallow copy113, direct ExitCode assignment156, NaturalCompletion171
and Classification175. [CallerMutation test](../diagnostics/isolated-fuzz/offline-supervision-policy.tests.ps1)
176–178 covers only undecorated fixture values.

Type inspection uses PSObject.BaseObject, while passive validation permits extra
NoteProperties on scalars. A decorated Int32 containing a nested mutable array
can pass validation; copying p.Value and returning it preserves the wrapper and
attached references. Sequential mutation through the original input can change
metadata reachable through a prior result; two results can share it. This violates
the promised independent snapshots without concurrent mutation or hostile runspaces.
It does not by itself change actual authority flags.

Parent disposition: **retain Important and stop**. Readonly tracing confirms the
permissive checks, shallow copy and unnormalized output assignments against the
plan's no-input-reference requirement. This is a static finding, not dynamic
reproduction or a claim of identical manifestation on both engines. The saved
35-label results do not discharge this missing ownership case.

Recommendation, not authority: prepare a narrow repair plan choosing rejection
of extended scalar metadata or reconstruction from validated primitives. Cover
decorated integer/bool/string inputs, nested mutable metadata, sequential mutation
and two results from the same input with independent expectations. Approve a new
finite allowance before repairs/tests; do not reuse these spent slots/review seat.

## Every declined judgment and parent disposition

All limits are retained, not silently accepted as passing requirements.

|Reviewer declined / reason|Parent disposition and consequence|
|---|---|
|Current executions: import/run forbidden|Saved receipts only, no fresh-execution claim.|
|Exhaustive vectors: finite/static evidence|Bounded coverage only; I1 uncovered.|
|QPC accuracy/conversion/cross-process agreement: synthetic milliseconds|Unproved; separate P2.|
|OS containment/startup races/handle inheritance: no live adapter|Unproved; no runtime authority.|
|Actor/session authenticity: declared labels|No identity-authentication claim.|
|Concurrent pipes/backpressure/byte-event provenance: aggregate metadata only|No real capture proof.|
|Cancellation/termination/physical cleanup: Would* only|Actual cleanup unknown; no live-effect proof.|
|Hard wall-clock/scheduler: model and saved successful hosts insufficient|No physical-time guarantee.|
|Durability/crash atomicity/physical readback: modeled booleans|Unproved; separate P3.|
|Harness timeout/kill/drain: no saved timeout|Dynamically unexercised.|
|Nonempty CLIXML progress: stderr empty|Parser branch dynamically unexercised.|
|Catastrophic failure: ordinary catch insufficient|No catastrophic-recovery guarantee.|
|Concurrent caller mutation/hostile runspaces: excluded and untested|Remain excluded; do not excuse sequential I1.|
|Historical authenticity/mutant chronology: no attestation or retained mutant snapshots|Hashes/receipts/diffs show consistency only.|
|Protected dependency behavior: hash-only, execution forbidden|Integrity matched, no fresh behavior verification.|
|Git base/head/clean/branch/origin/remote: ownership protection blocked reviewer|Parent provenance/final checks only, not independent Git evidence; no trust change.|
|Historical whole-branch/project correctness: outside scope|No whole-project clearance.|
|Security-owner/P2/P3/runtime/main/C12/release acceptance: separate approval/evidence|Remain open and unapproved.|

## Recovery and rulings

Task1 checkpoint e8bb42e7 is unchanged sourceFE23DF3B/test61A3BACF with indexE14CCDCA;
this documentation handoff does not fix I1. Recover latest
codex/c12-b01-task9-coordinator, not main. Git contains portable source/tests/
spec/plan/reports, not ignored raw/controller/session evidence. All local packets
remain retained. Cloning cannot restore them or authenticate past executions.
Prior raw coordinator limited acceptance does not accept this P1 module.

Ruling: receive-code-review implementation defaults yield to the explicit spent
budget/Important-stop rule. Cost is leaving a known gap unresolved until a new
bounded decision; no hidden repair/retest/review. Parent Git verification remains
separate from independent review because ownership protection was not changed.
