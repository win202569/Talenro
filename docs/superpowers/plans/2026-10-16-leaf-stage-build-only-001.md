# Leaf Stage Build-Only001 Single-Run Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for the existing Native/inline method after this specific operational plan is approved. Steps use checkbox (`- [ ]`) syntax for tracking. This is execution of existing software,not implementation of a new feature;do not manufacture a live RED/GREEN cycle or change source to fit an observed outcome.

**Goal:** Obtain one real,explicitly scoped nodeleafbinding compile/link diagnostic while preserving ordinary vet/test acceptance and the32 cumulative admission limit.

**Architecture:** Use only the existing Invoke-LeafStageBuildOnly controller and its single native boundary. Reuse the existing owned copied SDK/reference and offline modules,but create a new output root and cold private cache/tmp. Preserve and read the actual outcome,including STOPPED;never retry or run the generated binary.

**Tech Stack:** Windows PowerShell5.1 in a fresh64-bit host,existing C#/.NET runner and PowerShell helpers;no new dependencies.

**Spec:** [Approved build-only design](../specs/2026-10-16-leaf-stage-build-only-design.md).

**Status:** PROPOSAL PREPARED;AWAITING SEPARATE EXECUTION APPROVAL. The user's approval following “prepare a specific execution proposal,do not run yet” authorizes this document/input inspection,not the real call. No Go/native Run,full SDK/product inventory,PID query or output-root creation occurred while preparing it.

**Verified implementation baseline:**4afc6baa304efaf6fad53ebc261b038a06462b57,
branch codex/c12-b01-task9-coordinator;controller/repair commit a4bfe80a.
Prior final fake verification:21 suites/884 checks per PS7/5.1 host (11 static
included each),all42 commands exit0;one fresh review/two Important repairs,no
deferred minors. This proposal does not rerun or reclassify that evidence.

## Global Constraints

- Windows PowerShell only. Existing worktree W=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator;scope S=W/.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator. Do not change worktree or branch.
- One fresh64-bit C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe host,NoProfile/NonInteractive. Record actual version/bitness inside it;do not assume readiness on a different computer.
- At most one Invoke-LeafStageBuildOnly call and one native compile;no retry,cache warm-up,Go version/env/list command,extra vet call,matrix or automatic chaining. A preflight failure stops without calling it.
- New absent S/leaf-stage-build-only-001 only;do not create it during preparation or reuse/delete/rename it after any attempted call. Never replay either historical eager/stage compile001.
- Fixed arguments:test,-c,-x,-p=1,-vet=off,-count=1,-timeout=120s,-o=<RunRoot>/package.test.exe,talenro.local/platform/internal/nodecontrol/nodeleafbinding. No free-form overrides. The120s Go argument does not replace the600s native deadline or authorize binary runtime.
- Fixed20 environment keys and values are exactly spec section3/Get-LeafStageBuildOnlyFixedFields:copied GOROOT,private GOCACHE/TEMP/TMP/GOTMPDIR,existing GOMODCACHE,GOFLAGS=-mod=readonly,GOTOOLCHAIN=local,GOENV/GOPROXY/GOSUMDB/GOAUTH=off,GOVCS=all:off,GO111MODULE=on,GOWORK=off,GOOS=windows,GOARCH=amd64,empty TALENRO_FUZZ_DIAG_MODE/TALENRO_FUZZ_FIXTURE/GODEBUG.
- Unchanged trigger schema1/leaf-test-compile-log-v1;16777216 total bytes,65536 iteration/raw-line bytes,60000 polls;exact creation-handle unsignaled check and ordered clocks;one registry,32 cumulative attempts,0x101000,inherit=false,no image filters/eviction/reopening.
-600s native deadline,500ms natural observation,shared5000ms cleanup. Full SDK/product audits and whole-log token validation are outside that native deadline;no global time/memory bound or OS-call preemption is claimed. Record phase timing so audit time is not labelled compiler time.
- No product/helper/test/C# modifications,generated-binary runtime,live post-hoc PID lookup,full module-cache inventory,Docker,Bash/WSL,download/install,security-setting change,PR/main merge,force push or branch deletion.
- Policy leaf-stage-build-only-v1,VetMode=off,DiagnosticOnly=true,Coverage=sampled,NoRetry=true. VetEvidenceComplete/Acceptance/PackagePassed/OriginalTestPassed/RuntimeAttempted remain false even if the diagnostic completes.
- Ordinary vet/test acceptance remains required. No absence-of-vet-process claim,authenticated-host claim,OS sandbox claim or exhaustive process-capture claim follows from this run.
- Preserve necessary report/status/recovery documents on the existing development branch under AGENTS.md;raw receipts/logs,SDK,module/cache files and binaries remain ignored locally.

## Frozen inputs and current read-only observations

All relative paths below resolve against W,except the absolute module path.
Planner-only inspection confirmed output absent,64-bit current planning host,
fixed9 arguments and16 helper pins. It did not certify the full SDK or product.

| Input | Exact path | SHA256 |
| --- | --- | --- |
| Reference receipt | S/framework-001/copy-receipt.json | AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55 |
| Copy receipt | S/uninstrumented-001/copy-receipt.json | DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10 |
| Executable | S/uninstrumented-001/toolchain/bin/go.exe | D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F |
| Existing module marker | C:/Users/Lenovo/go/pkg/mod/github.com/google/uuid@v1.6.0/go.mod | 73DA47B6338B00A082FD451AA35A3273D3ADC09B8E9BBA98DAB01091E402AF6E |

Controller inputs:

- RunRoot=S/leaf-stage-build-only-001.
- Toolchain=S/uninstrumented-001/toolchain.
- ReferenceReceipt=S/framework-001/copy-receipt.json.
- ModuleCache=C:/Users/Lenovo/go/pkg/mod.
- Project/working directory=W;no alternate package or directory.

Freeze all66 immediate .ps1/.cs files in W/docs/diagnostics/isolated-fuzz:
for each file form `Name|UPPERCASE_SHA256`;sort those strings with
StringComparer.Ordinal;join with LF and append one final LF;hash the UTF8 bytes
without BOM. Expected count66,aggregate SHA256
`FAFE2BBCF5DE1AA67E96D492761CCB657F69A4D8F33557C279C1A315964BB86D`.
This definition does not require an ignored prior session file. Also preserve
the actual66 rows in the new ignored control record when execution is approved.
The controller independently binds/rechecks its16 production helpers and go.exe.
Different hashes,paths,machine or source changes require stopping and revising
the proposal;do not silently refresh these frozen expectations.

## Review Focus

1. A changed source/input between approval and launch must stop before the controller call;Task1 recomputes all frozen fingerprints and checks clean Git state.
2. An existing output identity must not be replayed or removed;Task1 requires absence and Task2 never retries after any call/uncertain launch.
3. PS5.1 JSON array wrappers must not drop pins or cause a mistaken second launch;Task1 uses direct assignment,explicit count66/unique-name validation,and a pre-call status record.
4. Exit0/capped/zero-entry samples or metadata alone must not become ordinary acceptance;Task3 checks independent scope/trigger/core/lifetime/compile facts and all false flags.
5. Provisional result or completed.pending.json must not be interpreted as final completion;Task3 checks the actual completed.json file,manifest binding and STOPPED contradictions without repairing saved artifacts.

---

## Files and deliverables

This proposal changes no implementation or test file. After execution approval,
create only a new ignored,absent control directory
S/leaf-stage-build-only-001-control for source/input snapshots,an execution
wrapper,outer stdout/stderr and status/timestamps. Check it is safe/non-overlapping
and absent too;do not reuse old wrappers or clear directories to obtain absence.
The controller alone owns RunRoot and its cache/tmp/build/evidence files.

After the attempt,create docs/roadmap/2026-10-16-leaf-stage-build-only-001.md;
update current-status.md,repository-recovery.md,diagnostic README and this plan's
execution status. Keep implementation/source frozen. If a defect is found,stop
and propose a separate tested repair,not an opportunistic code edit/live retry.

### Task1: Approved preflight and fresh-host binding

**Interfaces:** Consume the four frozen inputs,source aggregate,existing
Get-LeafStageBuildOnlyPlan(RunRoot,Toolchain,ReferenceReceipt,ModuleCache) and
Assert-LeafStageBuildOnlySources(Plan). Produce safe ignored source/input
snapshots and wrapper status;do not return or invent native evidence.

- [ ] **Step1: Confirm specific execution approval and clean branch.** Read this plan/spec;verify branch/origin,W,clean index/worktree and the tested baseline ancestry. Only reviewed documentation may differ from4afc6baa;the66 source aggregate must match exactly. Expected:no implementation drift,no existing RunRoot/control identity.
- [ ] **Step2: Verify frozen inputs with safe file helpers.** Check each of the four hashes above,non-reparse/single-link inputs,owned/non-overlapping paths and existing module marker. Do not execute Go or inventory the SDK here. Expected:four exact matches;any failure stops before controller invocation.
- [ ] **Step3: Prepare the new ignored control record and fresh host.** Record source rows/input rows and host version/bitness. If JSON is read,use direct assignment,then assert count66 and66 unique names;never use a Value/Count wrapper as the row list. Set ControllerCalled=false in the durable status before any import/call;record preparation timing. Expected:all rows present,64-bit Windows PS5.1,no RunRoot yet.
- [ ] **Step4: Validate the fixed planner inside that host.** Import only the existing helper;call its planner with these four inputs;assert16 helper pins,nine exact arguments,20 fixed environment entries,600s limit,policy/vet-off/typed false flags and output absence. Recheck frozen inputs/source fingerprints immediately before the call boundary. Expected:planner/source validation succeeds without native execution;otherwise stop and preserve the failure.

### Task2: One real compile-only controller call

**Interfaces:** Invoke exactly once
Invoke-LeafStageBuildOnly([string]RunRoot,[string]Toolchain,
[string]ReferenceReceipt,[string]ModuleCache),using the frozen values above.
It owns SDK/product inventories,one native call,receipt conversion/evidence,
independent final audits and persistence. Do not call its native seam directly.

- [ ] **Step1: Cross the single call boundary.** Durably record ControllerCalled=true immediately before calling the controller once;capture outer output/error and exit status without a retry loop. That boolean denotes entering the controller,not proof that native execution began. Expected:zero native attempts if preflight inside the controller fails,otherwise at most one native attempt. Preserve actual result/receipt rather than guessing from the wrapper boolean.
- [ ] **Step2: Let existing pre/post audits and finalization run.** Full copied-SDK checks require the established11536-file manifests;product before/after comparison uses actual inventories,not an assumed historical701 count. Do not skip them to save time. Use the owned tool session for bounded progress updates;no extra process/PID investigation or second Go command. Expected:either completed diagnostic or durable STOPPED/provisional evidence,never automatic retry. Full audits may take several minutes independently of compile time.
- [ ] **Step3: Preserve any launch/persistence failure.** Keep primary and secondary errors,outer logs and all files actually written. If ControllerCalled is true or launch is uncertain,do not start another controller;even proven pre-call failure is reported for a new decision,not bypassed automatically. Expected:no fabricated receipt,identity deletion,receipt rewrite or completed-marker repair.

### Task3: Saved-only evaluation and recovery checkpoint

**Interfaces:** Consume the actual saved plan/receipt/streams,result,manifest and
markers using existing Get-LeafStageBuildOnlyEvidence/Assert-ProcessCompletion
and safe file/hash helpers. Produce a factual report and ordinary dev checkpoint;
no native Run or live PID query is permitted in this task.

- [ ] **Step1: Establish the observed outcome.** Separately record NativeAttempts,compile exit/time,CompileEvidenceComplete,ScopeEvidenceComplete,TriggerEvidenceComplete,CoreNaturalComplete,LifetimeEvidenceComplete and final audit facts/errors. Validate saved scope with the existing-root validator,not the unused-root planner. At least one admitted entry and fewer than32 attempts are necessary;cap/zero/missing evidence means incomplete regardless of exit0 or cleared cap flag. Expected:actual facts,not a predicted success target or post-hoc PID/command join.
- [ ] **Step2: Verify durable evidence.** Check every explicit manifest row's safe path/hash/length,artifact binding,source/go/four-input fingerprints and saved SDK identity. Never walk cache/tmp. Only completed.json as a real file with correct policy/false flags and manifest SHA may establish finalization;result.json is provisional,completed.pending.json is never a marker. If STOPPED or persistence contradictions exist,report them and do not repair records. Independently confirm ordinary Assert-ProcessCompletion refuses the marker-bearing receipt. Expected:scoped completion or explicit STOPPED/unverified persistence,never acceptance.
- [ ] **Step3: Record limits and phase timing.** Preserve admission/query/release counts,cap/errors,trigger timing and optional saved image metadata without claiming exhaustive capture,vet absence or error87 resolution. Report pre-audit/native/post-audit intervals separately. Even successful diagnostic keeps vet/package/test/runtime/production/main gates false/open. Saved-only review must not trigger a rerun;no new implementation changes are authorized by this plan.
- [ ] **Step4: Preserve progress.** Commit only the factual result/report and necessary status/recovery/README/plan updates;ordinary-push the verified existing origin/branch and confirm remote SHA=local and clean state. Raw control/output/SDK/cache/binary remain ignored. On a new computer,clone latest dev tip and read the report;missing machine-bound inputs require a new proposal,not replay001. Expected:recoverable Git progress with actual failures and remaining gates clearly labelled.

## Approval handoff

Please approve this specific single-run proposal before Task1 execution.
Preparation approval alone does not consume the one native attempt or authorize
full live audits. On approval,use the existing Native/inline method and these
fixed boundaries;do not ask for routine dev-branch preservation permission again.
No success is promised:the useful outcome may still be a stopped sample.
