# Saved Error Reference Data Readonly Recheck Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No agents/reviewers. Steps use checkbox syntax for tracking.

**Goal:** Check the already-saved six-case diagnostic data with the singleton-label correction, without rerunning a host or rewriting its failed parent verdict.

**Architecture:** A separately attested, single readonly data computation in the existing tool PowerShell context. Reproduce the original saved-data predicates, correct only array preservation for expected labels, and record a new limited data-contract conclusion. Original source, parent command, final.json and all earlier archives stay unchanged.

**Tech Stack:** PowerShell file reads, JSON parsing, SHA256 and AST/text comparison. No diagnostic source execution, module loading, Process/Start-Process, new PowerShell child, Add-Type or I/O fault induction.

**Spec:** [Actual diagnostic and parent defect](../../roadmap/2026-11-01-offline-error-reference-diagnostic.md), [original six-case diagnostic contract](2026-11-01-offline-error-reference-diagnostic.md), [readonly identity analysis](../../roadmap/2026-11-01-offline-error-reference-readonly-analysis.md). The explicit singleton-array correction below is the only proposed predicate change.

Status: PLAN_PENDING_EXECUTION_APPROVAL. Baseline f06674a0a4eb870107af6f6423364ed00889e52f. User approved plan drafting only. No saved-data recheck, source edit, runtime launch or budget release in this turn.

## Global Constraints

- NEW0diagnostic/capture hosts /0targets /0compilations /0reviews; no integration allowance released. Ordinary read-only calculations use the existing tool context, not an intentionally launched tested process. Integration6/7/1/conditional1 remains FROZEN; diagnostic1host exhausted.
- Preserve original probe/attributes,8input binding,16packet files/9directories, parent tool input/result, final.json Exact=false and earlier archives byte-for-byte. Do not fix the original command, replace failure records or claim an original successful run.
- Never execute/dot-source/import the saved parent command, probe, capture, writer, runtime module or snapshot. Treat stored code and measurements only as data; no Invoke-Expression, ScriptBlock invocation from records or arbitrary paths.
- No corrected runtime retry or full-C12/repository acceptance. New readonly contract-match result is distinct from original parent verdict, runtime acceptance and a new independently measured reference matrix.
- Single finite data-recheck attempt. Any actual static, input, parsing, computation or recording failure stops without adjustment/retry/new suffix. Preserve all available partial records and report UNKNOWN/FAILED.
- Approved Native/PowerShell readonly scope replaces skill-default setup/Bash/baseline/TDD/full-suite/task-done reruns, agents/review/fix/deletion. All5capability flagsfalse; no I1/I2/C1/C2 closure, main merge or release.

## Review Focus

1. Singleton expected-label pipeline must remain an array: explicit string-array assignment, task1/2.
2. Correction must not relax unrelated case/type/count/error/matrix contracts: source predicate comparison and complete saved-data checks, task1/2.
3. Original failure/transport versus new data assessment must stay distinct: readonly result schema below, task2/3.
4. JSON booleans report past live measurements, not restored object identity: no ReferenceEquals on decoded objects, task2/3.
5. Changed originals, unknown/missing rows or failed recheck cannot become success: archive-byte provenance and one-attempt/stop policy, all tasks.

## Exact ownership and inputs

Original packet P: `.superpowers/sdd/2026-11-01-offline-error-reference-diagnostic`.
Original archive: `docs/diagnostics/isolated-fuzz/p2b-error-reference-diagnostic/portable-final.json`, SHA256 A12F16F6F42A8365B0D7D34BE8D70771BBDF9FAD96F5FC43EBFDC0B8B37E74D6.
Original binding SHA256 E4484D836278B97B99743F42A30B98F01384D2AD3DDBA5B29D3BCA228A8EC2BA.
Original probe SHA256 494A5945B278A0F8DC299EECA3786839BEF722E65EDC884D4871D881AE0E1496.

New recording-only sidecar R:
`.superpowers/sdd/2026-11-01-offline-error-reference-readonly-recheck`.
Create approval.json, progress.md, preflight.tool-result.json, recheck.tool-input.json,
static.tool-result.json, attempt.intent.json, recheck.tool-result.json and
result.json only as actually produced, plus optional terminal handoff note.
No new executable source file/controller/runtime namespace.

Create ordinary execution report `docs/roadmap/2026-11-01-offline-error-reference-readonly-recheck.md`
and `docs/diagnostics/isolated-fuzz/p2b-error-reference-diagnostic/portable-readonly-recheck.json`;
update current-status/repository-recovery. The original diagnostic report is
historical and remains unchanged. New archive includes actual R records only;
references the existing original archive rather than replacing or relabelling it.

## Task1 — approval, original provenance and exact correction

**Consumes:** actual explicit approval, original source/data/archive, approved plan.
**Produces:** fixed readonly tool input/hash and static predicate-difference record.

- [ ] Record approval of this complete plan, clean existing branch/origin, absent R/new archive. Bind actual plan SHA and approval text in R/approval.json. No runtime source/file changes.
- [ ] Read original archive; require exact16file/9directory inventory and safe relative paths. Compare decoded lengths/SHA256/bytes with actual original P files before any assessment; originals stay in place. Check8bound original inputs against binding hashes, original binding identity and original probe identity. This is a new scoped diagnostic-data provenance check, NOT a retry or completion of the earlier failed integration full-pin audit.
- [ ] Read parent-invocation.tool-input.json as text/data only. Extract the original readonly validation predicates for comparison, never evaluate its launch/intent/recording code. Author a new inline readonly computation with those ordered row/key/type/event/error/matrix/collector/final predicates. Preserve numeric/reference conditions exactly; actual transport fields are read from original final.json, not re-observed through a process.
- [ ] Sole predicate correction: preserve expected-label array via `[string[]]$labels=@(if($i -eq 0){'Outer'}elseif($i -eq 5){'First';'Second'}else{'Inner';'Outer'})`. Do not change original event names/indices/counts or use substring/Contains/wildcard acceptance. Static AST/text comparison must account for only this label expression within predicates, plus explicitly necessary readonly input/result plumbing. No replacement of unknown reference outcomes with expected booleans.
- [ ] Save exact new tool input as JSON data in R/recheck.tool-input.json; AST-parse only for syntax and forbidden process/import/invocation/write commands in the computation. File mutation is limited to new R receipts/result, not any original file. Preserve full actual static output after completion. No actual static failure correction pass or experiment.

## Task2 — one saved-data computation, zero tested hosts

**Consumes:** task1 fixed input/hash, original immutable data and lifecycle result.
**Produces:** new data-only result below, not original-run regrading.

- [ ] Write R/attempt.intent.json via CreateNew with actual tool-input SHA, plan SHA, Consumed=true and0/0/0/0. Flush/close before the sole computation. Existing marker, lost/unknown completion or recording failure forbids repeating the assessment.
- [ ] Compute original transport preconditions from final.json: Started/OwnedExit/StreamsComplete true, PID10448, ExitCode0, no timeout/termination, empty original streams and recorded PostInputsIntact=true. Assert original Exact=false and recorded parent error remain unchanged; do not require the original PrimaryError to be null or overwrite it. Validate original tool-result exit1 and its correspondence to the saved failed parent. These are historic recorded facts, not new process observations.
- [ ] Apply all original data checks to exact LF/UTF8≤1MiB observations.jsonl:7ordered rows,6case names, faults1/1/1/1/1/2, exact ordered keys/types/events and fixed missing-file message/Type/Id/actual probe source position+stack. Check current source text only as data. Preserve the complete matrix dimensions/boolean diagonal/symmetry, collector0primary/optional1secondary/reference-count continuity, independent-fault control, COMPLETE6cases/7faults/5falseflags.
- [ ] Off-diagonal matrices remain measurements; do not preselect true/false for rethrow cases merely because the report already summarizes them. Never call ReferenceEquals on deserialized errors/matrices to pretend original live references are restored. Summarize stored matrix entries as recorded facts only. A mismatch or arbitrary exception is failure/unknown, not a different acceptable “rejection”.
- [ ] After computation, verify original16bytes/hashes unchanged again and preserve full actual tool object/chunks after terminal completion. Do not run the old command with a corrected label. Save a NEW result.json via CreateNew with exact root keys: Schema=1,Scope="SavedDiagnosticDataOnly",Outcome (READONLY_CONTRACT_MATCH/FAILED/UNKNOWN),OriginalParentExact=false,OriginalParentExitCode=1,OriginalProbeExitCode=0,PredicateChange="SingletonExpectedLabelsArrayOnly",CasesChecked (actual0..6),RecordedMeasurements,OriginalArchiveSHA256,ToolInputSHA256,OriginalsUnchanged,NewDiagnosticHosts=0,Targets=0,Compilations=0,Reviews=0,IntegrationRemainder="FROZEN6/7/1/conditional1". Only full data checks and byte-preservation success permit READONLY_CONTRACT_MATCH. Result persistence failure itself stops; never rewrite it or substitute a success file.
- [ ] Stop even on READONLY_CONTRACT_MATCH. It means the saved data matches the explicitly corrected readonly predicate, not original Exact=true, accepted runtime proof, I1/I2 repair or permission to launch anything.

## Task3 — provenance-preserving handoff

- [ ] New report separates original failed parent, recorded probe measurements and new data-only assessment. State whether any original predicates were not completed, exact error/missing data and zero new runtime consumption. Earlier integration full-pin audit remains uncompleted; no retrospective claim.
- [ ] Archive every actual new R file/directory/partial/unknown/tool object as portable-readonly-recheck.json with lengths/hashes/base64; byte-compare decoded originals. Preserve previous archive/reference hashes. No fabricated receipts/results or inclusion of caches/credentials. Update recovery/status, explicit necessary docs/new archive only; staged-byte verification, commit/push existing development branch and remote SHA=HEAD. No merge/deletion/source changes.

## Self-review and execution gate

Five risks are covered by array correction/exact predicate comparison, full
data checks, explicit result scopes, matrix-as-recorded rules and pre/post byte
provenance. The scope is one readonly computation and0new tested processes,
targets, compilations or reviews. Original parent failure and frozen budgets
survive even a matching new data assessment.

Please explicitly approve Native execution of this zero-host readonly recheck,
including the singleton-array predicate correction and separate result/archive.
No original source/result modification, runtime revalidation, budget release,
production repair or automatic further step is authorized by that approval.
