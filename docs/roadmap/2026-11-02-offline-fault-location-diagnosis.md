# Fault-location diagnosis — actual single attempt

Plan02b53d98 approved by user “是”; preparation committed/pushed cf6cd11a before
the only invocation. No product source, existing test, oracle or historical
artifact changed. No host, target, compile, runtime snapshot or reviewer ran.

## Observed result

Computation15 returned LocatedMismatch/exit0, with655visited checks:
654matched and one first mismatch. This is diagnostic success, NOT contract PASS.
The mismatch is zero-based row8 (ninth observation), at
`$[8]['Value']['View']['PrimaryError']`.

The saved error is System.ArgumentException from ConvertFrom-Json in
Read-LSLedger, scripts/private/offline-diagnostics/sinks.ps1, line36column50.
Expected type/message/error ID match; the unchanged OC-SourceError predicate
rejects at its line33 because the model has no exact function/operation/line/
column candidate. Its sole Read-LSLedger/ConvertFrom-Json candidate instead
starts at line36column9 and spans the whole assignment.

Raw AST data records CommandAst36:50 → PipelineAst36:16 →
AssignmentStatementAst36:9. Get-ODFailureExtent promotes the command to that
assignment. This establishes the first saved-data rejection and supports a
command-specific extent correction as the next repair topic. It does not prove
there are no subsequent defects: traversal stopped at this failure and no fixed
source or full contract was run.

## Integrity and provenance

69inputs have identical pre/post hashes, including31operational sources and
nine historical stops. DiagnosticError/RecordingError are null. Actual exception
stack and complete tool input/result retained; extracted helper definitions
preserved their original line numbers and bodies. The recorded diagnostic stack
is a new offline stack, not a substituted original host stack.

- DiagnosticSHA E781A9149422669512281DD297DEC390D3EE53FC8DDB66B1554D89F5DF880057.
- InputSHA C7B18E568EDA29F5DDAEE6FABB795378E4206FADE64B742F3E15AA0364FC6B61.
- FinalSHA F911770E1FC860017BD7EF1114F88BBD7700173F465F7B7E9571D714177137CE.
- Preparation archive525files/135directories/31zero-byte, decoded equal,
  SHA66A5AE59317B4D65431430B05262AE9504B9DB610E880C9EE76800FCC7C6F899.

The old contract-green final still says EvidenceValid=false/EvidenceSHA256=null;
its1PASS/1FAIL/50NotRun is not regraded. There is no new diagnostic stop because
the located-mismatch outcome ends normally, but all original stops still apply.
Core/lifecycle/continuation intents remain absent.

## Budget, rulings and next boundary

Final portable-fault-location.json contains528actual files/135directories/
31zero-byte entries (all P/Q plus the two required H originals), decoded bytes,
lengths, hashes and directories verified. SHA
EB5B1E3743889DB1DD695D23CA2131E5E49380799423665CA9D5942FA2825BB3.
Input and diagnostic hashes rechecked after preservation and remain unchanged.
The embedded ledger describes the state just before final archive creation;
this report records the subsequent completed archive verification.

Consumed15auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews. Auxiliary
allowance exhausted. Remaining3hosts/7targets/conditional1review stay frozen;
no compile/snapshot allowance. Product/I1/I2 remains NOT_ACCEPTED.

Rulings: followed the approved PowerShell-only single-attempt workflow instead
of generic extra tests/reviews/cleanup; this diagnostic was not separately tested
before its single attempt and cannot establish product acceptance. The old final
provided stream hashes but no paths, so filenames were obtained from the frozen
driver's actual naming statements and verified against those hashes; any mismatch
would have blocked preparation. No speculative path or evidence replacement.

Next work needs a separately approved minimal repair/test/freeze scope. Preserve
exact location matching; do not accept arbitrary columns or weaken negative tests.
Do not rerun this diagnostic or the failed original contract attempt.
