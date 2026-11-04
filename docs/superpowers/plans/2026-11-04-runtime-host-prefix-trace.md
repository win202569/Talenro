# Real-host Bounded Prefix Trace Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Measure original host setup and the first real baseline wrapper/output path without reopening a stopped attempt.
**Architecture:** Generate three explicitly derived scripts beside their original sources so source-relative imports retain their directory. A separate bounded parent owns one diagnostic host; a silent tracing helper records stages in a new namespace.
**Tech Stack:** Native PowerShell7.6.5, .NET Process/Stopwatch, JSON, SHA256 and PowerShell AST; no installs.
**Spec:** ../specs/2026-11-04-runtime-host-prefix-trace-design.md

Status: PENDING_PLAN_APPROVAL / NO_EXECUTION.
User approved design6508337b with 批准. Design SHA256
5EEA657AFC9379268488DF10DD49B7B2FED55387834CF8541C40616D45A50B20.
Historical pending label in the design remains unchanged; this records approval.

## Global Constraints

- Fixed worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, branch codex/c12-b01-task9-coordinator; existing GitHub origin only.
- L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
- P=.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner.
- D=.superpowers/sdd/2026-11-04-runtime-host-prefix-trace-v1.
- Current27aux/6hosts/3compiles/217snapshots; proposed exactly computation28 and at most host7. Ceilings28/7/3/217.
- No tests beyond this single diagnostic, child targets, compilations, operational source snapshots, reviewers, installs, network from diagnostic code or security changes.
- All31operational sources and all historical evidence/stops remain byte-identical. Do not execute old driver/capture or reuse old grants/intents.
- Execution120000ms, owned-exit2000ms, shared stream-drain2000ms;8192-byte read buffers,65536bytes per stream.
- Journal4096events/1048576bytes, monotonic timing, UTF8, flush every event. No payloads/secrets.
- First unexpected preparation/execution result stops; preserve partial state, no repair/retry. Successful preparation consumes no runtime allowance.
- Explicit finite scope overrides generic skill smoke tests, RED/GREEN runs, full-suite checks, reviewers and cleanup. Static parsing/data verification only before the one run.
- NOT_ACCEPTED throughout. Design/plan approval does not imply product acceptance.

## Review Focus

1. Derived-file scope and source-directory behavior must not silently change imports, guards or wrapper variable lookup; Task1 compares unchanged statements/imports and scope boundaries.
2. A failure before the first wrapper returns must not reach later cases or masquerade as cutoff; Task1 checks bounded AST/control flow, Task2 checks actual outcomes.
3. Trace recording failure must preserve the original exception and cannot report success; Task1 checks latched error routing, Task2 inspects both errors.
4. Child output may stall or exceed bounds; parent must keep draining both streams and kill only its owned process; Task1 inspects loops/ownership and Task2 checks terminal records.
5. A successful write event without captured output is not completion; Task2 requires exact bytes, closed events and matching pre/post identities.

## File map and interfaces

Create in L:
- runtime-host-prefix-trace-common.ps1: definitions only; silent journal and identity checks.
- runtime-host-prefix-trace-parent.ps1: exclusive attempt/owned process/stream controller.
- runtime-host-prefix-trace-capture.ps1: generated capture-derived entry.
- runtime-host-prefix-trace-contract.tests.ps1: generated bounded suite and original wrapper.
- runtime-host-prefix-trace-test-helpers.ps1: generated helpers with result-write events.
- portable-runtime-host-prefix-trace-preparation.json and -final.json.

Create in D: prepare.ps1, preserve.ps1, progress.md, approval.json, input.json,
transformations.json, preparation-checks.json, preparation.tool-record.json;
then intent.json, child.claim.json, stages.jsonl, capture/journal.jsonl,
stdout.bin, stderr.bin, child-result.json, final.json, tool-record.json,
verification.json and stop.json when needed. Use exclusive creation for evidence.

Update only docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md,
and new docs/roadmap/2026-11-04-runtime-host-prefix-trace-result.md.

Common signatures:
- Write-HTOnce([string]Path, Value): CreateNew UTF8 JSON, no output.
- Assert-HTInputs(Binding): return observed SHA map; throw on mismatch/occupied absent pin.
- Write-HTEvent([string]Stage,[string]Kind,[string]Case): no output; Kind begin/end/error.
- Get-HTError(Record): structured actual exception identity/message/position/stack.
Parent: -InputSHA256 <64 uppercase hex>.
Generated capture retains original parameters and adds mandatory InputSHA256 and
IntentSHA256. Parent supplies fixed Grant contract-green, Epoch
history-contract-runtime and frozen hash186FE2C2B44A31A98EA5406109AC888A625FD1B0061C5338D0382268015D9BBA.
No caller-selected root, runtime, output path or scope.

## Task1: Prepare and statically prove the bounded host

- [ ] Confirm branch/origin/clean scope and all proposed new paths absent. Read spec,
  original capture.ps1, observation-contract.tests.ps1, test-helpers.ps1,
  recording.tests.ps1, publisher.psm1, original source-model/validator and prior
  diagnostic parent/common as data. Expected: no consumed D or path collision.

- [ ] Create mechanical preparation/check machinery first, without evaluating
  any operational functions. Bind the original31source inventory, all maps from
  diagnostic27 input, that input itself, its final archive
  72F4638D83566BEF22B4DE90A1B388D0B4B7F0CC5BA49B1CCEF126C6B2DBFDEF,
  original runtime completion archive, design/plan and every historical stop.
  Bind fixed engine/SMA hashes from diagnostic27 input, exact historical evidence
  order and frozen manifest. No self/future generated receipt in a hash map.

- [ ] Author the common helper. Events carry Schema1, Seq, EventId, ParentId,
  Stage, Kind, Case, ElapsedTicks, Frequency. Begin allocates ID; end/error uses
  active ID; nesting is exact; cases only empty/ContractAPIAvailable/BaselineCore.
  Latch journal failures. No success pipeline output; no mutation of automatic
  variables or operational collector state. Preserve first recording error.

- [ ] Generate scripts using AST-located edits, not global replacement. Every edit
  records OriginalPath/SHA, start offset, removed text, inserted text and generated
  offsets/hash. Require unique anchors, nonoverlapping edits, zero parser errors.
  Reverse edits in explicit integer descending-offset order and require exact
  original text/hash. Unedited spans must be byte-identical; encode without BOM.

  Capture-derived changes:
  1. Extend param block with bootstrap hashes; validate fixed invocation and exclusive
     child claim before business setup. Start trace clock/journal before setup.
  2. Keep original repo/packet variables for input lookup. Change only CaptureRoot
     to D/capture; original capture journal stays distinct from stages.jsonl.
  3. Retain original runtime/identity checks, collector initializations and original
     Capture-Failure/Capture-Value/Journal-ODFailure/Write-ODJournal definitions.
     Keep six guard-observed function definitions unmodified.
  4. Preserve write-boundary, record-owner, oracle, recording tests, guard start,
     publisher import and entry.tests imports in original order. Redirect only
     test-helpers import to derived helper and contract-tests import to derived suite.
     Restrict dispatch to the contract-green arm; remove other executable arms.
  5. Replace post-suite recorder-completion/tests-complete/full-result block with
     diagnostic cutoff checks. Remove Capture-Evidence publication invocation and
     publication catch path. Retain original outer catch collection/journal logic,
     terminal Console.Out.Flush and conditional Capture-ErrorSummary; close the
     original capture journal, then save diagnostic results. No synthetic throw
     for a normal cutoff; genuine errors keep exit failure.
  6. Insert silent stage calls around imports/setup, caught-error entry, collector,
     journal projection, terminal flush/error summary and diagnostic completion.
     Preserve the original caught ErrorRecord in a uniquely named variable BEFORE
     trace recording; recording failure must not replace it or suppress cleanup.

  Suite-derived changes:
  1. Keep original function scopes, original wrapper body/pipeline, and original
     validator/source-model imports. Do not wrap the suite in a new module/function.
  2. In Invoke-ODContractTests change only first loop limit23 to1; keep first
     iteration body and call syntax unchanged. Remove executable statements after
     that loop from this function. Mutation branch remains unreachable at index0.
     Add a diagnostic flag immediately after first wrapper return; no second call.
  3. Trace availability, source import, manifest hash/read, each evidence load and
     each context, baseline copy, wrapper entry/body return, suite update and
     wrapper return. Keep original body capture via ForEach-Object, rejection
     matching, LS-Result and LS-Need ordering.
  4. Do not instrument source-model/validator in this experiment: their original
     files are imported intact. Coarser stages avoid adding another comparison variable.

  Helper-derived changes: insert result-write begin/end around the ORIGINAL
  Console.Out.WriteLine in LS-Result; no added Flush there. Other statements,
  parameter lists and definitions unchanged.

- [ ] Verify generated files remain in L, so PSScriptRoot imports and root-relative
  computation retain original directory behavior. Record changed ScriptBlock.File/
  line provenance explicitly; do not claim transformed positions equal originals.
  Publisher is original module loaded by original path, retains private state and
  zero exports. Guard references originate from original six functions before
  observation. No post-run recorder test or publication probe is added.

- [ ] Write static assertions named ReversibleEdits, SameDirectoryImports,
  GuardDefinitionsUnchanged, WrapperPipelineUnchanged, SingleCaseReachable,
  PublicationUnreachable, OriginalFlushPlacement, OriginalErrorPreserved,
  TraceBounds, OwnedProcessOnly, NoExtraLaunch and FixedInputs.
  Each must inspect actual source/AST/maps; fail if a condition is unknown.
  Reachability checks require loop initial0/limit1/increment1, no index reassignment,
  exactly one wrapper site, no executable suffix, and capture dispatch only fixed
  contract-green. Availability failure cannot set cutoff/completion.
  Expected: all checks pass without importing/executing generated host or suite.

- [ ] Implement parent using prior bounded-parent discipline as a separate file.
  Refuse existing intent/claim/partial outputs/stops; CreateNew intent consumes28
  before validation. Hash-check helpers/input/runtime, start one hidden process
  with ArgumentList, fixed working directory and actual capture arguments.
  Drain both ReadAsync streams concurrently, preserve capped output, terminate
  only invocation-local Process on timeout/error/overflow; no lookup/tree kill.
  Record preflight/execution/cleanup separately. Before/after identity checks
  required even on failed process outcome where possible. Save actual child
  result only when produced; never synthesize a missing child completion.

- [ ] Preparation run (mechanical data transformation only):
  `& ./.superpowers/sdd/2026-11-04-runtime-host-prefix-trace-v1/prepare.ps1`
  Expected: static checks pass,31sources unchanged, all bound historical bytes
  match, transformations reversible, DiagnosticInvoked=false; fixed input hash.
  Save actual tool input/results. Any unexpected result creates stop and ends
  preparation; do not patch and retry under this allowance.

- [ ] Archive preparation originals and all five new helpers using preserve.ps1
  -Stage preparation. Require decode equality, length/hash/path containment,
  normalized outer JSON only. Report limitations and new input hash, commit explicit
  files, push existing branch and verify remote==HEAD before Task2.

## Task2: Exactly one run, independent interpretation and preservation

- [ ] Invoke once:
  `& ./docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/runtime-host-prefix-trace-parent.ps1 -InputSHA256 <Task1 actual hash>`
  Keep initial tool response and every poll; poll existing session, never replay.
  Expected outcome is evidence, NOT a guaranteed pass. Intent consumes28 even
  before start; host counter increases to7 only if process actually started.

- [ ] Independently read terminal artifacts. Validate stage schema/sequence/
  monotonic ticks/stack, byte/event bounds and EOF; retain torn final line as
  failure evidence. Compare input maps against actual files again.
  Require exactly two actual successful Groups named ContractAPIAvailable and
  BaselineCore, original SchemaSuite two entries, first wrapper return flag,
  no aborted/primary/secondary/recording failure and no unclosed intervals.
  Require stdout exactly UTF8(no BOM):
  PASS: ContractAPIAvailable followed by CRLF,
  PASS: BaselineCore followed by CRLF; stderr exactly empty.
  Require actual exit0, owned exit/streams complete, no timeout/termination/overflow.
  Only then classify HOST_PREFIX_COMPLETED, never FULL_SUITE_PASS.

- [ ] Failure statuses distinguish ASSERTION_OR_WRAPPER_FAILED, RECORDING_FAILED,
  INPUT_DRIFT, TIMEOUT and INCOMPLETE; preserve overlapping actual errors rather
  than hiding them behind one label. Save stop on any noncomplete outcome.
  No repair/retry/new tests. Successful cutoff also ends this execution scope.

- [ ] Write measured stage durations and last completed/unmatched intervals to
  result report. Do not add nested inclusive durations or infer exact blocking
  instruction from a missing end. Explicitly retain parent transport, derived
  source locations, tracing overhead and excluded full-publication limitations.

- [ ] Archive all actual D files including empty streams plus helper bytes using
  preserve.ps1 -Stage final; decode-verify. Update status/recovery, counters and
  remaining blockers. Stage explicit files, check staged bytes/diff, commit/push
  and verify remote==HEAD plus clean worktree. No main merge.

## Self-review and handoff

Spec coverage: preservation/scope/transformation/cutoff/error routing are Task1;
bounded one-shot execution, actual byte checks, interpretation and portability
are Task2. Review Focus conditions each have explicit static checks and actual
result gates; static checks are not represented as behavioral tests.
Shared interface: Task1 frozen input hash and helper identities consumed by Task2.
No extra run/reviewer is hidden in verification. Preparation and final evidence
are distinct and never overwrite prior archives.
Await plan approval, including proposed28/7 allowance, before implementation.
