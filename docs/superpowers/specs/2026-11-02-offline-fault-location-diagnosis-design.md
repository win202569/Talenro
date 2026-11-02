# Fault-location diagnosis — recovery addendum

Status: DRAFT / PENDING_APPROVAL. The user selected the diagnosis-first direction
and authorized writing this addendum; this is not execution approval.
Baseline: 6dc5677ee154b2ee83172defc37c04184a91bd0c, existing development branch
codex/c12-b01-task9-coordinator. Product remains NOT_ACCEPTED.

## Intent and evidence

Identify the first saved BaselineCore error-location mismatch before proposing
any source repair. The actual contract host ended 1PASS/1FAIL/50NotRun; its inner
stack reaches OC-SourceError line33 (exact function/operation/line/column match).
The saved failure does not itself identify the observation row. Static inspection
suggests ConvertFrom-Json command column50 versus assignment column9 at sinks.ps1
line36, but this is a hypothesis, not a confirmed complete root cause.

Authority remains the original failed final, including EvidenceValid=false and
EvidenceSHA256=null. Reading the evidence file does not change that verdict.
See [actual execution](../../roadmap/2026-11-02-offline-pin-binding-recovery.md).

## Approaches and selected scope

1. Selected: one bounded, offline diagnosis over preserved data and pinned source
   definitions. This can identify a specific mismatch without replaying a host,
   but it cannot establish a repaired contract or product acceptance.
2. Directly change the modeled extent and rerun the contract: deferred because
   the failing row is not established and this requires new repair/runtime budget.
3. Loosen position matching to a range: rejected; it changes acceptance semantics
   instead of establishing why the exact predicate failed.

This is a small diagnostic addendum, not an architecture redesign. Native work
only; no agents or reviewers. Keep the existing PowerShell environment.

## Proposed allowance — not yet granted

Request exactly ONE additional auxiliary computation (number15, ceiling15).
Existing consumption stays14auxiliary/4hosts/3compiles/186snapshots/0targets/
0reviews until the new attempt begins. No refunds or reassigned spent attempts.
Added hosts, targets, compilations, runtime snapshots and reviewers: all ZERO.
The original remaining3hosts/7targets/conditional1review remain frozen.

The single computation may extract only explicitly enumerated, inspected pure
source-model/contract helper definitions from pinned source ASTs, reconstruct the
historical BaselineCore context from saved data, and trace saved error objects in
the original traversal order. The later implementation plan must enumerate that
helper allowlist and exact input paths/hashes before execution. No whole-script
dot-sourcing, module/profile/control/writer execution, child process, installation,
network, full-suite invocation or synthetic claim of an original runtime result.
If the task cannot be expressed within that allowlist, stop before execution.

## Inputs and outputs

Use the immutable repair manifest and its source snapshots, the original contract
final/evidence/streams/tool record, and the exact historical baseline rows/oracle
referenced by the frozen test. Trace their provenance before using them. Do not
substitute current machine paths silently: any relocation must be explicit and
preserve logical source identity, hashes and historical coordinates.

Required anchors:

- repair-green.frozen.json: A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556
- pin-binding.stop.json: D5D3CE0845438682B770E2FA747607B5052A38B39A1A861A0021CEAB707066AD
- contract-green.evidence.json: 5993F5D9A5C93BDE307C4EDD14E365D04774C1EF718895134C0636B95E7C4F40
- portable-pin-binding.json: 8FF8BA63EFA96FFA476CFB2AA165312459DF09979610EB2829F64FC75B638029

Freeze a separate diagnostic script and input identities before the computation.
Create a unique new intent before invoking it, recording attempt15 as consumed.
Write new result/tool records only; never modify old artifacts. A result should
include the traversed row index and object path, actual error and expected
source/function/operation, parsed line/column, actual model candidates and their
AST extents, and the precise failed predicate. Preserve preceding traversal
results so a later bad-JSON row is not falsely called the first mismatch.

The diagnostic wrapper may add trace records outside the frozen predicate; it
must not edit that predicate or manufacture a successful acceptance result.
Record pre/post input and source hashes, actual exceptions and full tool output.
If the first mismatch is elsewhere, report it without forcing the hypothesis.

## Stop conditions and interpretation

Stop on missing/changed inputs, ambiguous extraction, unsupported context,
recording failure, unexpected execution needs or any source drift. An attempted
run consumes the allowance even if setup fails. No retries, second probe, fixes,
new source freeze or runtime continuation. Preserve partial records and a new
diagnostic stop record without deleting or bypassing the nine historical stops.

Even a successful diagnosis ends this allowance. Classify the result as a
located mismatch, hypothesis contradicted, or inconclusive; never PASS for the
original contract. A subsequent fix requires a separate reviewed scope and
explicit test/source-freeze/runtime allowances. Exact source, operation, line,
column, stack and process-safety predicates remain unchanged.

## Preservation and handoff

Implementation planning follows approval of this written design. Before any
diagnostic execution, the implementation plan must freeze the input/helper list,
receipt schema, unique filenames and command; execution approval is separate.
After execution or stop, preserve actual new records in a new portable archive,
verify decoded bytes/hashes, and update status/recovery documentation. Retain all
existing archives. Commit/push necessary files to the existing branch, verify
remote equality, and never merge main or claim whole-project acceptance.

This drafting turn changed documentation only and consumed no execution budget.
