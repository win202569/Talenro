# Offline owner-control fixture scope repair

Status: PROPOSED, pending user approval. Planning only; not execution authority.

## Intent and evidence

Restore the four actual functions loaded by the private owner-control fixture
across separate invocations of that same module. Preserve the original I1/I2
contracts and every historical attempt. This is an isolated diagnostic repair,
not a product, CI, Docker or default validation-chain change.

Baseline commit: 8f9f76ede0e1b0719394031e4520ce2b7c33aad8.
Read [actual failure](../../roadmap/2026-11-01-offline-controller-recovery.md),
[original design](2026-11-01-offline-i1-i2-single-owner-design.md) and
[controller amendment](../plans/2026-11-01-offline-controller-recovery.md).
The sole actual writer host reported18PASS/8NotRun, capture exit2/parent exit1.
Invoke-ODProfile was not found at record-boundary.tests.ps1:231. Source inspection
shows all four imported definitions have local lifetime in the first invocation;
later control calls are separate invocations. This explanation has NOT yet been
reproduced by a separate scope regression. No missing outcome is inferred.

## Approaches and proposed choice

1. Recommended: qualify only the four imported function declarations with
   module script scope. Keep each full actual definition, including inline
   parameters and body; reject missing, duplicate or unexpected definitions.
   This limits the lifetime change to the existing private fixture module.
2. Move all controls into the initial invocation: broader restructuring,
   changed lifetimes and source positions without a corresponding requirement.
3. Export/globalize or substitute functions: rejected; breaks private binding
   or stops testing the actual functions.

Choice1 is the agent's proposal, not a previously approved implementation.
Functions: Invoke-ODProfile, Capture-Value, Capture-Once, Capture-ErrorSummary.
Do not use Body.GetScriptBlock alone: inline parameters must survive. Do not
change bodies, error collection, expected output, group inventory or exports.

## Append-only ownership and scope

L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner.
Q=.superpowers/sdd/2026-11-01-offline-controller-recovery.

Only five existing L sources may change: record-boundary.tests.ps1 (scope and
fixed unit/manifest bindings), driver.ps1 (strict recovery/provenance and fixed
attempt paths), capture.ps1 (fixed writer output slot only), source-model.ps1
(fixed writer manifest/unit paths), observation-contract.tests.ps1 (fixed writer
evidence/final inputs). Other26 runtime sources stay byte-identical. A new
owner-scope.tests.ps1 is an auxiliary, not a32nd runtime source.

Canonical grant stays writer-green and keeps its26 original groups. The new,
explicitly authorized attempt uses the fixed private slot writer-scope-green
for its intent/final/streams/evidence, capture and unit roots. This is a new
attempt, NOT a refund or replay of the consumed writer-green attempt. There is
no public attempt/path parameter or fallback. Other four grant slots unchanged.
New source manifest writer-scope-repair.frozen.json keeps Phase writer-green,
parent D8D31B8BE9F38A35F12CF5951DC28B3723D759B96AF3BC3B3D8B3D4D95FAC57C.
Use new snapshots/writer-scope-rebind and writer-scope.stop.json. Keep all old
stops, manifests, outputs and portable archives at original bytes/paths.

## Budget and acceptance

Request additional:2existing-tool-context scope computations (one RED, one
GREEN),1writer capture host,1in-memory fixture compilation and31source snapshots.
Scope computations may create at most2private modules each, no child process,
Add-Type, target, profile body invocation or record I/O. No new target/reviewer.
Snapshot ceiling155=93existing+31scope+31later original contract. Runtime total
ceiling6hosts/7targets/2compilations/conditional1review:1host/1compile already
consumed, new writer1/1, later four hosts/7targets/0compile remain reserved.

Success requires exact auxiliary GREEN, one new31-source freeze, then one full
writer26PASS/exit0 with original transport, recorder, identity and pin gates.
Partial PASS or publication is insufficient. Only then resume the original
contract/core/lifecycle/continuation sequence and its single conditional review.
First unexpected failure stops; no automatic repair, retry, re-freeze or budget
extension. Failed checkpoints remain NOT_ACCEPTED. Native implementation only,
no new implementation agents. Necessary documentation/checkpoints go to the
existing development branch; no main merge or broad acceptance is authorized.
