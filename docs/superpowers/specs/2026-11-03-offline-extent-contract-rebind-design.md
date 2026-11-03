# Extent repair — fresh contract epoch design

Status: DRAFT / PENDING_APPROVAL. User selected a new source freeze and one full
contract validation, and authorized drafting only. Baseline development checkpoint
6c37dcf7e922377fe7d0fa8dee91c37e962592e6. No implementation or execution yet.

## Intent and current evidence

[Local repair](../../roadmap/2026-11-02-offline-command-extent-repair.md) passed
RED16/GREEN17, including10local GREEN groups. It does not prove the52group full
contract passes. The original contract-green attempt remains failed and consumed;
nine stops remain immutable. Its repair-green manifest still pins old helper bytes.

Inspection found four binding sites: driver selects repair-green and the spent
contract-green output slot; capture derives its output slot from semantic Grant;
the test entry explicitly opens repair-green; runtime/unit context constructors
also select repair-green and assume its direct parent is the accepted writer.
Changing only a filename would leave these consumers on different epochs.

## Approaches and bounded choice

Selected direction: one explicit new epoch with coordinated fixed-name binding
and unchanged semantic contract suite. Reusing/overwriting repair-green or
contract-green would destroy provenance and is excluded. A separate copied
runner/suite would duplicate ownership and acceptance logic; prefer the existing
implementation with a narrowly enumerated epoch selector. Automatically resuming
core/lifecycle/continuation is outside the selected scope even after success.

This is a bounded recovery-binding change, not a redesign of runtime behavior.
The four-file window and two preparatory auxiliary runs below are new proposals
revealed by inspection; they are not covered by the previous local-fix allowance.

## New identities, no old-slot reuse

P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner;
L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
Use fixed epoch `extent-contract`, with no caller-provided filename or path:

- Prepare grant extent-contract-rebind; new P/extent-contract.intent.json.
- New manifest P/extent-contract.frozen.json; new snapshots/extent-contract.
- Semantic test Grant stays contract-green; new attempt slot extent-contract-green.
- All host intent/final/stdout/stderr/evidence names use that new slot; capture
  journal root is P/capture/extent-contract-green, also CreateNew/no-retry.
- New recovery stop P/extent-contract.stop.json; old stops are required historical
  evidence, never deleted, renamed, ignored by a generic switch or regraded.

Old routes remain blocked. The new route requires its own explicit approval,
exact parent/history bindings, bounded budget, expected auxiliary receipts and
absence of new-epoch intent/output/stop files. No automatic latest-manifest search,
arbitrary Epoch value, failed-validation fallback or default reopening of old work.

## Four-file operational window

1. L/driver.ps1: add the fixed Prepare/RunOne recovery route, provenance gate and
   semantic-Grant/attempt-slot separation. Only contract-green with explicit
   extent-contract epoch may execute on this route. Preserve the actual owned
   process start/drain/cleanup predicates and exact output/evidence consumption.
2. L/capture.ps1: accept the enumerated explicit epoch, reject it with any other
   Grant, derive the new output/capture slot and pass the epoch to contract tests.
   Keep recorder, publisher, output schema and semantic Grant unchanged.
3. L/observation-contract.tests.ps1: change only entry parameter/binding plumbing
   so the selected manifest and unit context use the explicit epoch. All49suite
   cases, three recorder groups, expected results and negative mutations remain
   identical. No test deletion, reordered group or relaxed assertion.
4. L/source-model.ps1: retain the tested command extent fix; add explicit fixed
   epoch selection through runtime/unit constructors as needed. Historical context
   selection stays historical and never falls back from failed runtime validation.

Other27of31operational files, including pure observation-contract predicates and
expected-output/observations, remain byte-identical. No changes to actual profile,
record ownership, timeouts, child handling, error projection or capability flags.
The implementation plan must identify exact parameter flows and unchanged-region
comparisons before coding; any need for a fifth operational file requires review.

## Preserve both parent and writer lineage

The new manifest's direct parent is original repair-green, SHA
A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556.
Do not pretend the direct parent is writer-projection-repair. Preserve all existing
recovery maps and append a distinct ExtentContractRecovery map, explicitly binding
the accepted writer manifest/final/evidence/streams at their original hashes.

The unit fixture must validate the anchored original repair parent and its writer
lineage before referring to the accepted writer; it must not compare the new
direct-parent hash to the writer hash or overwrite either identity. Never rerun
the accepted writer. Old operational source comparisons use their preserved
snapshots, not the newly edited files.

Inherit2276pins, update exactly four operational entries and append unique new
provenance pins for the approved spec/plan, approvals, diagnosis15, local16/17
receipts, new auxiliary fixtures and receipts. The plan must enumerate the exact
append list/count and protect duplicate-pin consistency before freezing. The
SourcePaths inventory stays31. Do not invent a successful new manifest during
auxiliary tests or backfill old missing evidence hashes.

## Proposed verification and allowances

Additional auxiliary budget: TWO runs, binding RED18 and GREEN19, ceiling19.
These test extracted actual selector/gate/context helpers with fixed data in the
existing tool context; they do not start a host, run the full suite or compile.
Freeze test/input identities before RED and candidate identities before GREEN.
Expected RED must identify missing new binding, not an arbitrary setup exception.

GREEN must cover fixed parent/capture/test/context agreement, unique new slot,
old-slot rejection, absent approval rejection, spent-attempt/new-stop rejection,
wrong hash/parent/writer/epoch rejection, source-window consistency and unchanged
process/recorder/predicate guards. Include the ten command-extent invariants within
this versioned regression; do not replay the already consumed16/17 attempts.
Exact groups, signatures and helper allowlist are required in the subsequent plan.

After exact GREEN only: ONE Prepare, exactly31new snapshots (186→217 total),
then ONE new contract host (actual consumed4→5), with ZERO child targets,
compilations or reviewers. This is one added host allowance, not reuse of the
spent old contract slot or diversion of the frozen three downstream hosts.
The new epoch has MaxHosts1/MaxChildren0/MaxCompilations0/MaxReviewers0. Record
consumed counters and new authorization separately from legacy ceilings; do not
reset cumulative totals or unfreeze unrelated allowances.

Push and verify the new frozen archive before launching the host. Preserve
120000ms host cap, owned-object-only cleanup,2000ms exit/drain bounds, bounded
evidence/journal and all five false capability flags. Do not alter environment,
runtime version/hash or privilege requirements to obtain a pass.

Acceptance requires original52ordered groups PASS, exact original stdout and
stderr, expected exit, actual parent EvidenceValid/Exact and post-source pin
integrity, zero children, no timeout/forced termination/overflow/rescue. A child
exit0 or partial evidence alone is insufficient. Full contract success still is
not product acceptance and does not authorize the next runtime stages.

## Stops, preservation and handoff

Every new computation/Prepare/host consumes its own CreateNew intent first.
First unexpected result, partial freeze, recording error or drift creates the new
stop and ends this scope. No repair/retry within the remaining allowance, hidden
fallback, historical reclassification or output replacement. Even full success
ends after the single host and preservation; core/lifecycle/continuation stay frozen.

Persist actual full tool input/results, first error and partial records in new
portable preparation/candidate/freeze/final archives; verify decoded originals
and retain all prior archives. Commit necessary source/tests/spec/plan/status/
recovery records to the existing branch and verify remote equality. No main merge.

Current consumption remains17auxiliary/4hosts/3compiles/186snapshots/0targets/
0reviews. Proposed additions are2auxiliary/1host/31snapshots only; none consumed
by drafting. Written design approval precedes a detailed implementation plan and
separate execution approval. Product remains NOT_ACCEPTED.
