# Unified Saved-Data Checker Revision Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No agents/reviewers. Steps use checkbox syntax for tracking.

**Goal:** Address the complete checker audit in one new data-contract revision, test its valid/invalid data behavior before a separate assessment of the saved diagnostic, and preserve original failures.

**Architecture:** New pure saved-data validation functions, frozen literal data tests, and a separate fixed-path receipt helper. No process-launch controller or diagnostic replay. Original parent/recheck/probe and all archives remain byte-identical; a new assessment never changes their verdicts.

**Tech Stack:** Existing tool PowerShell context, strict UTF8, built-in JSON/AST APIs and file hashing. No diagnostic/capture host, child PowerShell, target, Add-Type, runtime module import, installs or production integration.

**Spec:** [Complete readonly audit A1–A6](../../roadmap/2026-11-01-offline-checker-readonly-audit.md), SHA256907E39483740AD39F34CFE68A573C52CE4C548E6E3B462E33C6A7AB21DFEB96E; [original diagnostic semantics](2026-11-01-offline-error-reference-diagnostic.md); [actual failed diagnostic](../../roadmap/2026-11-01-offline-error-reference-diagnostic.md) and [failed readonly recheck](../../roadmap/2026-11-01-offline-error-reference-readonly-recheck.md). The strict contract/type/framing changes below explicitly supersede the old predicate-only label correction; they are not retroactive original-run acceptance.

Status: PLAN_PENDING_EXECUTION_APPROVAL. Baseline fd0915b6906eec42188bcc98610385c7cb00aa19. User approved drafting only. No new source, data tests, assessment, runtime launch or budget release in this turn.

## Global Constraints

- Propose NEW0diagnostic/capture hosts/0targets/0compilations/0reviews. Separately request exactly1expected RED data-test pass,1GREEN data-test pass and1formal saved-data assessment in existing tool context. These computations are not a frozen integration grant or refund.
- Preserve all old sources/tests/oracles/attempts/stops/results/snapshots/raw/archives. No original parent command invocation or patch, old recheck retry, I1/I2 production fix, new runtime tree, main merge or release.
- Only NEW pure helper/test definitions may be dot-sourced for approved data tests. No original probe/capture/writer/runtime source/snapshot execution and no code invocation recovered from tool JSON. Parse old source as data only.
- Mutate only in-memory copies of JSON/byte data; never mutate original files or claim cloned data contains live ErrorRecord/Exception references. No production factory injection, real I/O fault experiment or C# fixture.
- Tests and literal expected models are fixed before RED; changing them after results is not authorized. Expected single RED below is deliberate; every other actual static/test/assessment/record/provenance failure stops without repair/retry/new suffix.
- Native finite PowerShell/data-TDD scope replaces skill-default setup/Bash/full-repository/task-done reruns, agents/reviews, automatic fix passes and deletion. Only the named data suite; not a repository green suite or dynamic persistence fault proof.
- Original diagnostic host allowance exhausted; v3 remainder6/7/1/conditional1 stays FROZEN, older limits unchanged. Five capability flagsfalse; I1/I2 and C1/C2 not closed, R2/Minor2 deferred.

## Review Focus

1. Reserved-variable and singleton mistakes can be syntax-valid: availability RED, positive singleton control and AST/name checks, tasks1–2.
2. Type/count/matrix relations must not depend on implicit coercion: explicit JSON kinds and single mutations, tasks1–2.
3. Framing/duplicate keys/source attribution cannot be normalized or substring-matched into validity: raw-byte/source-model negatives, tasks1–2.
4. Writing the new result must not replace its first failure or reopen/retry: isolated helper/static structure check, task2/3; compound I/O remains uninduced.
5. A successful new data assessment must not erase original failure, recover live references or release budgets: provenance/tri-state result and handoff, task3.

## Files and fixed ownership

Create only five new sources in `docs/diagnostics/isolated-fuzz/p2b-saved-data-contract-revision/`:
`.gitattributes` (LF), `saved-data-contract.ps1`, `saved-data-contract.tests.ps1`,
`expected-data-tests.json`, `receipt.ps1`. No executable parent/launch script.

New recording packet R: `.superpowers/sdd/2026-11-01-offline-checker-unified-revision`.
It owns approval/progress/input identities, immutable test/source/plan snapshots,
static receipts, red/green/assessment consumed intents and complete tool records,
actual data-test results and unique new result.json. No original record replacement.

Original diagnostic packet P and its16files/9directories remain under
`.superpowers/sdd/2026-11-01-offline-error-reference-diagnostic`.
Original diagnostic portable SHA256 A12F16F6F42A8365B0D7D34BE8D70771BBDF9FAD96F5FC43EBFDC0B8B37E74D6;
failed recheck portable SHA256844D976FD867DDCD9899595E7720E07061E703328FD11358A333D516DEF447F1.
Check actual old bytes/bound8inputs, not reconstructed originals. No repeat of
the unrelated incomplete integration full-pin audit.

Create new actual report `docs/roadmap/2026-11-01-offline-checker-unified-revision.md`
and new leaf/portable-final.json; update current-status/repository-recovery.
All older reports/plans remain historical and unchanged.

## Exact validator and receipt contracts

`Assert-ODSavedReferenceData([byte[]]Bytes,[Collections.IDictionary]Context)`
returns exactly one ordered data report {Schema=1,CasesChecked=6,FaultCalls=7,
RecordedMeasurements=[six fixed ordered summaries]}; no success text or file I/O.
Mismatch throws RuntimeException with Message/FQID exactly
`ODR: saved data contract mismatch`. Arbitrary implementation exceptions are
unexpected failures, never valid rejection. Catch decoding/JSON parser errors
only at those narrow boundaries; do not translate every code bug into mismatch.

Context contains only the fixed original packet path, frozen probe path/hash,
frozen probe source text and independently derived source model. Caller creates
it from verified originals. Helper has no process/import/network/write authority
and must not mutate Bytes/Context or create operational ErrorRecords.

- Framing:1..1048576bytes, strict UTF8, no BOM/raw CR, exactly one terminal LF,
  no blank/extra row; remove only that required LF. Exactly7rows, original six
  case names/order followed by COMPLETE. Reject duplicate JSON keys recursively
  before a dictionary parser could discard them; depth≤32. Use built-in JSON
  token/document APIs as data parsing, no compilation/evaluation.
- Shape: require real JSON arrays/maps, exact ordered case/event/collector/error/
  final keys from original contract, case-sensitive names; no joined-key-list
  ambiguity or scalar .Count shorthand. Strings must be strings, flags boolean,
  integer fields JSON integer tokens representable as Int64, not strings/bools/
  fractions/exponents/coerced values. Validate array/index bounds before indexing.
- Facts: original fault counts1/1/1/1/1/2, event labels Outer or Inner/Outer or
  First/Second as explicit string arrays; indices0..n-1 and before/after counts
  continuous/bounded. Collector PrimaryIndex0, optional SecondaryIndices[1],
  Overflow=false; DirectOuter total1, independent total2 with records different.
- Matrices: exact n×n boolean arrays, true diagonal/symmetry, original collector
  count consistency. Same observed ErrorRecord implies same captured Exception;
  do not force rethrow off-diagonals to previously measured outcomes. Treat all
  matrices as recorded data, not restored references or new measurements.
- Errors: original Type/Id/Message exactly bound to fixed case absent.txt path.
  Parse Position first line as exact frozen path plus numeric line/column bound
  to Invoke-ODReferenceFault's ReadAllText AST extent. Parse the3Stack frames as
  frozen source/helper/Ref-Case/entry call locations. Derive permitted lines
  independently from original AST/case dispatch, not from observed values.
  Original source currently has fault16:5, per-case calls54/56/60/65/71/77 and
  entry140; require frozen hash and matching AST before adopting those locations.
  Position/stack line separators may be CRLF or LF inside JSON strings only;
  raw JSONL remains LF-only. No Contains-based source acceptance.
- COMPLETE: exact original keys/order, Schema1/Status COMPLETE/Cases6/FaultCalls7
  with strict types and all5flagsfalse. No original transport error suppression
  or reinterpretation of original parent Exact=false.

`Write-ODSavedDataReceipt([byte[]]Bytes)` in receipt.ps1 writes ONLY fixed
R/result.json via CreateNew; it has no path/callback/retry parameter. Retain
first actual operation ErrorRecord locally, dispose held stream exactly once,
keep bounded cleanup detail without replacement and throw the original cause
after cleanup. No throw from finally, reopen or fallback success file. This
guarantees code structure only in this scope; no dynamic Write+Dispose fault is
approved. Normal receipt write at handoff is the only real helper I/O exercise.
Do not promise raw ErrorRecord identity survives a calling catch; local first
record retention and original cause are distinct from wrapper propagation.
Initialize private ReceiptAttempted=false/ReceiptFailed=false and typed
ReceiptSecondary List[ErrorRecord] once when loading this NEW helper. Consume
ReceiptAttempted before Open; any later call rejects before file I/O. Latch
ReceiptFailed on first actual failure; retain at most8cleanup records separately
for caller diagnosis, never serialize them as new live authority. Do not reset
or reload the helper to bypass a failed attempt; durable assessment intent and
absent-result checks independently prevent a cross-context retry.

## Task1 — immutable tests and deliberate new-API RED

**Files:** new LF attributes, tests and literal expected JSON; R approval/ledger/test snapshots/intents.
**Consumes:** explicit approval and immutable original data/source.
**Produces:** fixed42group suite and one expected availability RED, no old algorithm replay.

- [ ] Obtain explicit approval of this complete revision, new helper/test files and finite data-test budget. Verify clean linked development branch/origin, absent new leaf/R/archive, original archive byte equality/8inputs. Record consent and actual plan/input hashes. No runtime budget release.
- [ ] Write the tests and expected42names/order/signatures FIRST. Freeze those files and hash them before RED. Get-Command availability check must be data-context-only, not import/execute old code. RED is exactly FunctionAvailable FAIL plus41NotRun, exit1 because the NEW API is not yet implemented; no unrelated exception is accepted as this RED.
- [ ] The42GREEN groups are1availability,2positive,36negative and3static checks. Positive1 uses original bytes unchanged. Positive2 uses an in-memory valid alternate CatchBare record/exception all-true matrix with collector1/empty secondaries and second event AfterCount1; other facts unchanged. This proves the validator is not overfitted to the previously observed different-record outcome. Never present this artificial positive as a measured runtime event.
- [ ] Each negative must get ONLY the exact mismatch exception; otherwise fail the test. Build named independent in-memory variants, never mutate originals:

| IDs | Fixed negative variants |
| --- | --- |
|01–06|missing case row; duplicate case row; swap first2rows; unknown case name; extra root key; swapped root key order|
|07–11|DirectOuter Events scalar; ErrorProjections scalar; RecordSame scalar; ExceptionSame row scalar; SecondaryIndices scalar|
|12–18|string ObservationIndex; string PrimaryIndex; string secondary index; string COMPLETE.Schema; string COMPLETE.Cases; string COMPLETE.FaultCalls; string case.FaultCalls|
|19–22|string matrix cell; false diagonal; asymmetric matrix; same-record/different-Exception variant based on positive2 (all other invariants retained)|
|23–27|nonnumeric Position line; valid-but-wrong statement line1; different source path; wrong ErrorId; wrong Message|
|28–32|extra terminal LF; missing terminal LF; CRLF raw framing; UTF8 BOM; duplicate physical JSON property|
|33–36|invalid UTF8byte0xFF; size1048577bytes; interior blank row; wrong helper Stack line17|

- [ ] Static groups: NoReservedVariableWrites (case-insensitive assignments/parameters/loop variables include Error/Host/PID/HOME/Args/Input/PSVersionTable/PSCommandPath/PSScriptRoot; explicitly permit writable preferences); PureValidatorNoIOOrProcess (AST); ReceiptFirstCauseStructure (no finally throw/reopen/retry and first-cause-before-cleanup propagation). These inspect NEW source only; do not dynamically fault the receipt helper.
- [ ] Run the single availability RED in existing tool context, save full actual tool result and explicit consumed red intent. Unexpected result stops; do not redefine the frozen expected suite.

## Task2 — unified implementation and one GREEN suite

**Files:** create saved-data-contract.ps1 and receipt.ps1 only; tests/expected data frozen.
**Consumes:** Task1 fixed suite/context source model.
**Produces:** strict pure validator and separately reviewed receipt structure, all42groups exact.

- [ ] Implement the exact contracts above, using task-specific names (recordProjection,caseRow,etc.) and explicit arrays. Parser syntax success alone is insufficient; source lint/type/framing/source-location cases are mandatory.
- [ ] Snapshot all5new sources/approved plan/approval with actual hashes before GREEN; input identity manifest separates originals from mutable/new sources. Green intent consumed via CreateNew before invoking tests. Dot-source only the new pure helpers/tests; forbid process creation, old source imports, Add-Type or external suite calls.
- [ ] Run GREEN once; require42PASS/0FAIL/0NotRun, all36wrong inputs rejected by exact mismatch, both positive controls accepted and all3static checks pass. Preserve bytes and complete tool chunks. Any actual failure freezes assessment and handoff as unaccepted; no patch or second GREEN within this approval.
- [ ] Verify original bytes and frozen test hashes unchanged. Coverage claims: A1/A2→lint/availability/positives; A3→07–22 and strict kinds; A4→23–27/36; A5→01–06/28–35; A6→static receipt structure+later normal write only. Compound record I/O remains dynamically unverified and must be reported as such.

## Task3 — one new formal saved-data assessment, then stop

**Files:** new R assessment intent/result; actual report/portable/status/recovery.
**Consumes:** exact GREEN/new source binding and immutable original data.
**Produces:** scoped data assessment, no original-run success or production repair.

- [ ] Only exact GREEN permits one assessment. Create consumed assessment intent with actual source/test/plan/approval hashes. Read original transport/failure/tool correspondence as immutable facts: PID10448, owned exit/drain, probe exit0, parent exit1/Exact=false, original stderr/stdout empty. Do not require original parent PrimaryError=null or infer a fresh process result.
- [ ] Call the SAME pure validator on original15228byte data, not the positive2 fixture. Require exact six summaries without preselecting reference outcomes, pre/post original16file/9directory/archive/8input checks. Source paths must remain original pinned values; mismatch stops rather than rewriting context on a new computer.
- [ ] New result schema2 keys: Schema,Scope=SavedDiagnosticDataOnly,Outcome(DATA_CONTRACT_MATCH/FAILED/UNKNOWN),OriginalParentExact=false,OriginalProbeExitCode=0,OriginalParentExitCode=1,CasesChecked,RecordedMeasurements,OriginalsState(NOT_CHECKED/MATCH/MISMATCH),PredicateRevision=UnifiedAuditA1A6,SourceBindingSHA256,NewDiagnosticHosts=0,Targets=0,Compilations=0,Reviews=0,IntegrationRemainder=FROZEN6/7/1/conditional1,DynamicCompoundReceiptIOVerified=false. Start NOT_CHECKED; set MISMATCH only on detected actual original drift, MATCH only after complete before/after equality. Failure before post-check stays NOT_CHECKED.
- [ ] Use fixed receipt helper once for new result; any persistence failure preserves actual tool error/partial bytes, no fallback result or further writes through a failed sink. Stop after assessment even if matching. No production collector repair, integration retry or more data-test pass is implied.
- [ ] On success OR earlier stop, archive all actual R records/directories/source snapshots/test variants as labelled fixture data/results/tool chunks/partial/zero-byte files in new leaf/portable-final.json, decode and byte-compare. Link original archives, never replace them. Report finite consumption, original versus new verdicts, expectedRED versus unexpected failures and unverified compound I/O. Explicit necessary new sources/docs/archive, staged-byte checks, commit/push existing branch/remote SHA equality. No main merge/deletion/raw/cache/secret commit.

## Self-review and approval gate

All audit A1–A6 plus reporting ambiguity have named implementation and test
coverage/limitations. Exact suite:1availability+2positive+36negative+3static=42;
RED1FAIL/41NotRun, GREEN42PASS, then at most1formal original-data assessment.
0new tested processes/targets/compiles/reviews. Old failures stay immutable;
source-level receipt restructuring is not dynamic I/O fault proof.

Please explicitly approve Native implementation of this complete unified
revision and its finite data-TDD/assessment computations, with0new diagnostic
hosts. Approval of drafting alone does not authorize code/tests/assessment.
Any unexpected actual failure stops without another patch/retry or release
of the still-frozen integration remainder.
