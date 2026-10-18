# P1 initially plain scalar ownership — bounded characterization proposal

Status=PROPOSED_NOT_APPROVED; NewTargetHosts=0/1UNAPPROVED_UNUSED.
Design only after bf1740cbe80bb975bf237301f2d34720ffd34e29.
No executable fixture/controller created, no target invoked, no candidate repair.
This is a bounded investigation, not a new product architecture or acceptance plan.

## Question and recommendation

Can mutable extended metadata attached AFTER two accepted evaluations of an
initially plain input become visible through another saved result or the input?
[Independent R1](2026-10-18-p1-token-readonly-review.md) establishes missing
evidence, not an observed runtime defect. Test this question before choosing a fix.

Recommend one candidate-only fresh ordinary PS7 host. Repeating already-passed
baseline/full suites does not answer this vector. Source-only reasoning alone
cannot establish the engine's actual extended-member ownership behavior.
No production changes, performance comparison, PS5 run or new reviewer proposed.

## Frozen subject and finite matrix

Subject: docs/diagnostics/isolated-fuzz/integer-token-local-candidate/offline-supervision-policy.ps1
SHA256 F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A.
Do not replace the original or either existing routing copy.

|Path|Initially plain input field|Saved decision field|Input type/value|
|---|---|---|---|
|Normal|progress.ExitCode|WouldExitCode|Int32 / 0|
|Normal|progress.ExitCode|WouldExitCode|Int64 / 0|
|Normal|progress.ExitObserved|WouldExitObserved|Boolean / true|
|Normal|cleanup.NaturalCompletion|WouldNaturalCompletion|Boolean / true|
|Normal|persist.Classification|WouldClassification|String / clean|
|Sticky supervisor-lost|persist.Classification|WouldClassification|String / clean|

Each row has three separate fresh scenarios: attach through the original input
field, saved result A field, or saved result B field. Total 6 x 3 = 18 cases.
Each case calls the candidate exactly twice BEFORE attachment, no post-attachment
reevaluation. Two separate plain-normal controls, before/after the matrix, each
call once: maximum 38 candidate calls, all in the single host. Stop early on any
failure; do not skip a case or add diagnostic candidate calls.

Normal uses arm/release/progress/cleanup/persist at sequence/time 0..4, exactly
the prior normal fixture. Sticky replaces observation 2 with actor L,
fault supervisor-lost; cleanup/persist remain. Evaluated=true in both; sticky is
an evaluated Aborted result, not a positive acceptance result or refused input.

## Independent oracle and fixture fidelity

Before execution, prepare a separate portable test and reviewed literal expected
objects from the existing normal/sticky contract, not candidate-generated output.
Use the complete 13-field result and 22-field Decision values in P1I1-Want
(normal/sticky) as the starting literal specification. Preserve exact names,
NoteProperty descriptors, nulls, reason/index, windows, flags and primitive values.
For the Int64 row explicitly expect Int64 WouldExitCode; the Int32 row expects
Int32. Do not rely on the old comparator that conflates these integer types.

For every case:

1. Make caller-owned fresh strings from nonempty char arrays and owned boxed
   numeric/Boolean values from typed storage; never decorate shared literals.
   Verify the actual stored input type/value and absence of extended members.
   Prepare a separate same-type/value field as a fixture-only attachment control.
2. Evaluate twice; compare both full results to independent literals. Require
   distinct result and Decision containers. Save both results without serialization,
   reconstruction, scalar casting or copying that might erase the ownership fact.
   Inspect all tested input/result leaves for initially absent extended members.
3. Attach a fresh NoteProperty containing a fresh mutable object[] directly through
   the selected EXISTING scalar field's PSObject.Properties. Do NOT replace the
   field, wrap/copy its scalar, or use the old P1I1-Decorate helper (it replaces it).
   Apply the same operation to the independent fixture-only control.
4. Retrieve the selected field again through its owner. Verify the exact property
   descriptor, payload ReferenceEquals, unchanged CLR type/base value and mutation
   from initial to changed through that descriptor. The separate control must
   retain its own distinct payload. Lost/ephemeral attachment or changed base
   value/type is FIXTURE_INVALID, never evidence of successful isolation.
5. Inspect the other saved result and input fields directly for the property and
   payload, then mutate the selected payload once and inspect again. No metadata
   from the selected owner may be visible through any other owner. Full base-value
   literals must remain unchanged. BaseObject ReferenceEquals alone is diagnostic,
   NOT the isolation oracle. Record direction, field and first observed breach.

Use fixed SDK member inspection, including PSMemberSet.Members reflection, so
plain scalars do not repeat the historical strict-mode .psextended.Members error.
Enumerate descriptors without invoking ScriptProperty/CodeProperty getters;
only the newly installed known NoteProperty's Value is read. No executable
metadata or global type-data/runspace changes. Do not claim callback coverage
merely from a counter when this fixture installs no executable callbacks.

## Proposed preparation, host budget and fail-stop outcomes

Approval would cover portable fixture/oracle/pins/controller preparation AND
one execution attempt together. Source/data parsing and integrity checks may
precede the host; no preparatory probe, warmup, test import, helper invocation or
second target host. Consume the one slot immediately before launch. Freeze the
new files and exact ordered output inventory before execution; no oracle edits
after observing candidate output.

Use the established ordinary fresh PS7 capture pattern: 30s wall budget, bounded
2s owned-host exit confirmation and 2s stream drain. Preserve raw stdout/stderr,
hashes, exact command/engine/source identities, elapsed/exit/timeout and capture
facts. Terminate only the newly owned host if needed; never name-wide/PID-reuse
cleanup or weaken the existing ownership/exit/capture checks. No timeout increase,
retry, baseline run, full regression, source fix or reviewer follows automatically.

- ISOLATED_WITHIN_MATRIX: all 18 cases plus controls pass, complete expected
  output, exit0, no timeout, empty stderr, complete owned-exit/capture evidence.
  This supports only the stated PS7 vectors; it does not globally close I1/R1.
- OWNERSHIP_GAP_OBSERVED: valid attachment/control checks establish cross-owner
  descriptor/payload visibility. Stop at first breach; report exact observation,
  not an exploit claim. Any correction needs separately agreed scope.
- FIXTURE_INVALID: attachment/identity/type/control fidelity fails. Stop; do not
  reinterpret a failed mutation as isolation or silently change the fixture.
- EXECUTION_INCOMPLETE: launch, budget, exit, stream, integrity, literal-oracle
  or other unexpected failure. Stop and retain the unfinished result.

Any non-pass freezes remaining cases; there is no repeat-until-pass allowance.
Recheck existing 51 scoped and 476 inherited pins before/after preparation/run,
plus the new frozen identities. Preserve all historical failed/raw records.
Missing old packets on a fresh computer cannot reset budgets or prove integrity;
resolve that evidence limitation explicitly before executing this proposal.

## Approval boundary and recovery

Current turn only records this proposal and status/recovery. NEW max1 host is
UNAPPROVED/UNUSED. All previous direct2/2, full4/4, review1/1 spent and older
failed/frozen/ineligible allowances remain unchanged. No new agent authorized.
FormalAcceptance=false;PromotionApproved=false;I1CandidateDisposition remains
InsufficientEvidence. No main, security, runtime, C12 or release approval.
PS5, other scalar values/paths, concurrent mutation, hostile runspaces, performance
and the prior review's 18+5 declined judgments remain outside this finite matrix.

Restore the latest codex/c12-b01-task9-coordinator checkpoint, not main. Git
preserves this design and existing source/reports, not local raw packets, engines,
credentials or execution authority. Next: user approval of preparation plus the
single bounded attempt; not another proposal-only or prep-only approval cycle.
