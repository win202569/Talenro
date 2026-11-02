# Command extent — local repair design

Status: DRAFT / PENDING_APPROVAL. User selected the local repair approach and
authorized this design, not implementation or additional computations yet.
Baseline f383286043b5ae655daba0ef2eb6f6db65799298; existing branch
codex/c12-b01-task9-coordinator. Product remains NOT_ACCEPTED.

## Evidence and intent

[Single diagnosis](../../roadmap/2026-11-02-offline-fault-location-diagnosis.md)
located the first rejection at saved row8 Value.View.PrimaryError. The actual
Read-LSLedger/ConvertFrom-Json error points to sinks.ps1:36:50. The model instead
uses the assignment extent starting36:9; unchanged OC-SourceError rejects at33.
Diagnosis final SHA F911770E1FC860017BD7EF1114F88BBD7700173F465F7B7E9571D714177137CE.
This establishes a specific model defect, not the absence of later defects.

## Approaches

Selected: correct this command's modeled extent and validate the local boundary
with one RED and one GREEN. Smallest behavior change; full runtime acceptance
remains deferred. Alternative: combine repair with a new runtime freeze and full
contract chain; deferred because it adds controller work and execution allowances.
Accepting any column within a statement is rejected: it weakens the predicate.

## Exact change boundary

Only operational source allowed to change:
docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/source-model.ps1,
inside Get-ODFailureExtent. Add a narrow early return of the node's own Extent
only when Node is CommandAst and GetCommandName() is exactly ConvertFrom-Json.
Retain the existing ancestor-walk body unchanged for every other node type/name.
Do not hardcode line36, column50, function names or input paths in production.

No changes to Read-ODSourceModel's inventory or output schema, named/dynamic
calls, pin helper, historical/runtime context constructors, observation-contract,
driver, recorder, old tests/oracles or any other operational source. No synthetic
extra candidate at the old assignment column. The exact function/operation/
source/line/column and stack checks remain byte-identical.

The source candidate will no longer match its OLD runtime manifest pin. This is
intentional development state, not a new runtime freeze. Do not update that pin,
run the old driver, bypass its mismatch checks or claim runtime readiness.

## Proposed local regression

Create a separate frozen regression and immutable phase inputs/receipts. Extract
only explicitly allowlisted actual source-model and pure-contract functions;
the implementation plan must enumerate dependencies and exact input hashes.
No full-script imports, module/profile/control/writer body, child process or
compilation. Use the original saved row8 error unchanged with the historical
context, source bytes and manifest already preserved. Bind the candidate helper
source independently; never rewrite or fabricate the old runtime manifest.

Ten ordered groups, fixed before RED:

1. SavedErrorAccepted: the unchanged OC-SourceError accepts the actual saved
   error under the source-derived historical context, including its entire stack.
2. CommandExtent: modeled ConvertFrom-Json extent equals the independently
   parsed command AST extent (start/end/text), not its assignment ancestor.
3. WrongColumnRejected: cloned error at old assignment column9 is rejected.
4. WrongLineRejected: a cloned wrong line is rejected even if first stack-frame
   line is changed consistently; remaining error data stays fixed.
5. WrongOperationRejected: cloned expectation with a different operation rejects.
6. WrongFunctionRejected: cloned expectation with a different function rejects.
7. WrongSourceRejected: cloned expectation resolving to a different pinned
   source rejects. No file/manifest mutation is used to manufacture negatives.
8. MethodsUnchanged: all modeled method-call faults in the fixed historical
   source inventory equal those from the preserved original helper.
9. ThrowsUnchanged: all modeled throw faults in that inventory remain identical.
10. ModelBoundaryUnchanged: source identities, Functions, Calls and DynamicCalls
    remain identical; all non-ConvertFrom-Json faults remain identical and each
    changed command fault equals its own independently parsed command extent.

Negative groups must reject with the actual expected RuntimeException and exact
Message/FQID 'OD: observation contract mismatch', not merely any exception.
Original-helper comparisons run in separate local scopes within the same phase;
no function from a previous attempt may leak into candidate evaluation.

RED must be0PASS/1FAIL/9NotRun with the first group's actual unchanged
OC-SourceError location rejection. A setup error, unexpected pass or different
failure stops before any repair. Only after that expected RED add the early
return and commit its independently hashed candidate before GREEN.
GREEN must be10PASS/0FAIL/0NotRun with null setup/recording errors and unchanged
pre/post candidate/input hashes. Later mismatches in the full22rows are not
silently folded into this repair or claimed fixed; this scope tests the identified
error and model invariants, not the full contract.

## Allowance and stopping

Propose exactly TWO new auxiliary computations: RED16 and GREEN17, ceiling17.
Consumed15auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews unchanged until
execution. Added hosts/targets/compiles/runtime snapshots/reviewers all ZERO.
The original remaining3hosts/7targets/conditional1review stay frozen.
Native work only. No separate full-suite run, extra probe, rerun of diagnosis15,
replay of contract-green, or automatic repair after an unexpected result.

CreateNew phase intent before each attempted invocation; setup/recording failures
consume the attempt and stop the remainder. New uniquely named stop/results
preserve actual error and full tool output. All nine existing stops, old failures,
diagnosis and archives remain untouched; no refund or reclassification.
Even successful GREEN ends this scope. A new runtime freeze/controller recovery
and full acceptance require a separate proposal and explicit allowances.

## Preservation and next step

Keep tests, preparation inputs, source candidate and final results in separate
committed checkpoints on the existing development branch. New portable archives
must preserve raw phase records and required historical inputs, verify decoded
bytes/hashes, and never overwrite earlier archives. Update status and recovery
instructions; verify remoteSHA==HEAD. No merge to main or product acceptance.

This is a small local design, not a broader architecture project. After written
design approval, prepare the implementation plan with exact helper/input lists,
test interfaces, receipt schemas and commands, then obtain execution approval.
No source edit or test execution occurred while drafting this design.
