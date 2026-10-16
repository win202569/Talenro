# Ordinary Leaf Check001 Single-Run Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans with the existing Native/inline method only after this concrete operational plan receives execution approval. Track steps with checkboxes. This runs existing software; do not manufacture a live RED/GREEN cycle, modify code to fit an outcome, or dispatch implementation agents.

**Goal:** Obtain one bounded real ordinary nodeleafbinding default-vet build and, only if its evidence passes, one original test-binary execution.

**Architecture:** Use the existing fixed Invoke-LeafOrdinaryCheck controller and unchanged ordinary native collector. A new owned output root supplies private cold cache/tmp and durable attempt records; saved evidence alone determines the scoped outcome. No diagnostic binary or diagnostic policy is reused.

**Tech Stack:** Windows PowerShell5.1, fresh64-bit host, existing PowerShell/C# helpers, owned copied SDK and existing offline modules; no installation or download.

**Spec:** [Approved ordinary design](../specs/2026-10-16-leaf-ordinary-check-design.md); [verified fake-only implementation](../../roadmap/2026-10-16-leaf-ordinary-check-implementation.md).

**Status:** Specific execution approved by user “是” after2734cad9, then STOPPED before preflight execution: the executor's command had a PowerShell ParserError. No controller/native/build/test call or output/control root creation occurred. No retry was made;execution checkboxes remain open because the planned work was not performed. See the [stopped preparation record](../../roadmap/2026-10-16-leaf-ordinary-check-001.md). A new decision is required before corrected preflight;do not replay this approval.

**Baseline:** 53dd9c2da88203ece6f849d13095963d8b36b5bd on codex/c12-b01-task9-coordinator. Prior fake-only evidence:407 checks per PS7/5.1 host, including1 static each;16 commands exit0. This proposal supplies no new test result and no real package acceptance.

## Global Constraints

- W=D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; S=W/.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator. Keep this worktree/branch and verified origin https://github.com/win202569/Talenro.git.
- Use one fresh64-bit C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe host with -NoProfile -NonInteractive. Record actual version/bitness inside that host; do not inherit a PS7-specific PSModulePath through Start-Process. Direct invocation follows the verified host-launch pattern; no global environment/security changes.
- At most one Invoke-LeafOrdinaryCheck call, one native build and one conditional native runtime. No automatic retry, cache warm-up, extra Go/vet/version/env/list command, matrix replay, diagnostic mode or historical PID lookup. A failed/uncertain preflight also returns for a new decision rather than silently restarting.
- RunRoot=S/leaf-ordinary-check-001 and control=S/leaf-ordinary-check-001-control must both be absent, safe and non-overlapping. Do not create either while preparing this proposal. After any attempted call do not remove, rename or reuse them to obtain a fresh-looking attempt.
- Build from W, with exact arguments: test -c -x -p=1 -count=1 -timeout=120s -o=<RunRoot>/package.test.exe talenro.local/platform/internal/nodecontrol/nodeleafbinding. VetMode=default; no -vet=off or custom analyzer policy. Native build budget600s.
- Only after accepted build/receipt/compiler/binary bindings, execute that newly built binary from W/internal/nodecontrol/nodeleafbinding with -test.count=1 -test.timeout=120s -test.v=true -test.run=.; original Go timeout120s, native runtime budget150s. No fuzz/race/benchmark or narrower filter.
- Exact20-key environment: GOROOT=<Toolchain>;GOCACHE=<RunRoot>/go-cache;GOMODCACHE=C:/Users/Lenovo/go/pkg/mod;GOTOOLCHAIN=local;GOENV=off;GOPROXY=off;GOSUMDB=off;GOAUTH=off;GOVCS=all:off;GO111MODULE=on;GOWORK=off;GOFLAGS=-mod=readonly;GOOS=windows;GOARCH=amd64;TEMP/TMP/GOTMPDIR=<RunRoot>/tmp;TALENRO_FUZZ_DIAG_MODE/TALENRO_FUZZ_FIXTURE/GODEBUG empty. No caller overrides.
- Ordinary route remains IsolatedFuzzProcess.Run(...,false,false), schema3 completion/query guards,500ms natural observation and shared5000ms cleanup. No collector/error87 exemption or diagnostic sampling. Any timeout, forced termination, ambiguous identity/exit, truncation or failed evidence gate stops before another phase.
- Full copied-SDK/reference validation and actual product inventories before/after are mandatory controller work. Eligible final product/SDK/artifact audits run independently even after an earlier failure. Do not substitute historical file counts or skip audits. Native budgets are not global audit-time/memory/OS-call preemption guarantees; record phase times separately.
- No product/helper/test/C# edits, dependency acquisition, Docker, Bash/WSL, security changes, PR/main merge, branch deletion or force push. Historical build-only001 binary/root stays untouched and must never run.
- Provisional result retains typed PackagePassed=false,FinalizationRequired=true,Acceptance=false. Only valid completed.json and its explicit15-file manifest binding may establish scoped PackagePassed=true; overall Acceptance remains false. C11/C12/Task9/10/15/18/Batch01/security/production/main and historical failures stay open.
- Necessary plan/report/status/recovery documents are preserved on the existing development branch under AGENTS.md. Raw evidence, SDK/modules/caches/tmp/binary stay ignored locally; Git cloning is not environment readiness.

## Frozen inputs and preparation observations

Read-only preparation at baseline confirmed both proposed roots absent, four exact input hashes,13 helper bindings and the71-file aggregate below. Planning host was PS7.6.5/64-bit; this is not certification of the future PS5.1 execution host or full SDK/product readiness.

Controller inputs: RunRoot=S/leaf-ordinary-check-001; Toolchain=S/uninstrumented-001/toolchain; ReferenceReceipt=S/framework-001/copy-receipt.json; ModuleCache=C:/Users/Lenovo/go/pkg/mod.

| Input | Path | SHA256 |
| --- | --- | --- |
| Reference receipt | S/framework-001/copy-receipt.json | AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55 |
| Copy receipt | S/uninstrumented-001/copy-receipt.json | DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10 |
| Go executable | S/uninstrumented-001/toolchain/bin/go.exe | D6A71E6DC9806CB3F1FCB817D562175C5A7B1505FC3F6638299648E0B9F0F74F |
| Offline module marker | C:/Users/Lenovo/go/pkg/mod/github.com/google/uuid@v1.6.0/go.mod | 73DA47B6338B00A082FD451AA35A3273D3ADC09B8E9BBA98DAB01091E402AF6E |

Freeze all71 immediate .ps1/.cs files under W/docs/diagnostics/isolated-fuzz:
form Name|UPPERCASE_SHA256 per file, sort strings with StringComparer.Ordinal,
join LF with one final LF, hash UTF8 without BOM. Expected SHA256:
AF83F1057EA2B71E31BCB38BD59F6518C2C68424BE77F1B1FD62A04D17340869.
The controller independently freezes13 production helpers and four inputs.
Store all71 rows in the new control record only after execution approval;
direct JSON assignment must retain an array with71 unique names, not a
Value/Count wrapper. An aggregate/input/path/machine mismatch stops; do not
refresh expectations silently. A new computer needs a revised preparation and
approval, not replay001.

## Review Focus

1. Source/input drift between approval and boundary: Task1 recomputes exact pins and clean Git state; Task2 rechecks immediately before the sole controller call.
2. Existing/unsafe/overlapping root or uncertain call: Task1 requires absence and safe paths; Task2 records entry durably and never retries, including after wrapper/persistence failure.
3. Build exit0 with error87 or incomplete ordinary evidence: Task2 keeps the unmodified guard and forbids runtime; Task3 distinguishes compile exit from verified build success.
4. Final PASS with skips, missing tests or failed audits: Task3 reports top-level/subtest RUN/PASS/SKIP separately, rejects contradictions, and never equates output alone with PackagePassed.
5. Provisional/pending/contradictory/tampered evidence: Task3 uses the existing saved reader and verifies explicit manifest rows; never repairs records to manufacture completion.

## Files and deliverables

Preparation creates only this plan and status/recovery updates. No implementation changes or fake reruns are needed for a documentation-only proposal. After execution approval only, create absent ignored control for frozen rows, an execution wrapper, host/timing/call status and outer streams. Controller alone owns RunRoot. Preserve all primary/secondary errors.

After the attempt create docs/roadmap/2026-10-16-leaf-ordinary-check-001.md and update this plan's status, docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md and diagnostic README. Do not edit source to turn the same attempt green; a discovered defect requires separate diagnosis/repair approval and a later new run decision.

### Task1: Approved preflight and fresh-host binding

**Interfaces:** Consume the four frozen inputs and71-source aggregate; use Get-LeafOrdinaryCheckPlan(RunRoot,Toolchain,ReferenceReceipt,ModuleCache) and Assert-LeafOrdinaryCheckPlan(Plan,true). Produce safe ignored control snapshots/status, not native evidence.

- [ ] Read spec/implementation/this plan and confirm explicit approval of this exact real run. Check W/branch/origin, clean index/worktree and baseline ancestry. Only reviewed documentation may differ;71-source aggregate must match.
- [ ] Check both roots absent, path safety/non-overlap, all four hashes and existing module marker using existing safe single-link/non-reparse helpers. No Go call or full module-cache inventory. Stop on any failure.
- [ ] Create only the new control directory after these checks. Record source/input rows and ControllerCalled=false durably; record preparation timestamps. Start the fresh64-bit Windows PS5.1 host directly and confirm host/module resolution without changing global environment or execution policy.
- [ ] In that host import the existing entry and call only its planner. Assert policy leaf-ordinary-check-v1, default vet,13 helpers/four inputs, eight build arguments/four runtime arguments,20 environment keys,600s/150s budgets,NoRetry=true,Acceptance=false and RunRoot absence. Recompute frozen pins. No root/native workload yet.

### Task2: One ordinary controller invocation

**Interfaces:** Call Invoke-LeafOrdinaryCheck([string]RunRoot,[string]Toolchain,[string]ReferenceReceipt,[string]ModuleCache) once with the frozen values. It owns full SDK/product checks, phase guards, attempt records, native execution and finalization; do not call its native seam directly.

- [ ] Recheck source/input pins and durably set ControllerCalled=true immediately before the sole call. Capture outer output/error/exit and timings. This flag means entering the controller, not proof of a native launch; use actual progress/receipt evidence to count phases.
- [ ] Let controller perform mandatory pre-audits, at most one build and conditional one runtime. Each attempt is recorded before its native boundary. First unverified outcome stops forward work while eligible independent final audits continue. No second call even if launch/persistence status is ambiguous.
- [ ] Preserve actual failure/result, outer logs, receipts/streams and final audit errors. Report bounded progress from the owned session; no extra Go/PID/workload commands. Do not interrupt slow full audits merely because native600s elapsed; report pre/build/runtime/post times separately. Missing STOPPED persistence is unverified failure, never success.

### Task3: Saved-only evaluation and recovery checkpoint

**Interfaces:** Use Get-LeafOrdinarySavedEvidence([string]RunRoot) for successful completion; on failure inspect only actual saved records using safe file/hash helpers. Never invoke the new-root planner/controller for saved evaluation or run a generated binary manually.

- [ ] Record actual attempted phases and evidence, build/runtime exit and duration, CompilePassed,DefaultVetBuildPassed,ExecutionAttempted,OriginalTestPassed,ProductUnchanged,ToolchainUnchanged,ArtifactsUnchanged and primary/secondary errors. Do not claim native start from a pre-launch attempt row alone.
- [ ] Require schema3 ordinary completion/query evidence for both phases; unchanged original-package/copied-compiler tokens and executable bindings; final nonempty stdout line exactly PASS, at least one original top-level Test PASS, no FAIL/instrumentation marker. Disclose observed top-level/subtest RUN/PASS/SKIP names and counts without asserting exhaustive body/analyzer/process coverage.
- [ ] For success require actual completed.json, correct policy/typed flags, no STOPPED/pending/launch-error contradictions, manifest hash and exactly15 safe relative artifact rows with matching lengths/hashes. Saved reader must accept; provisional result alone remains PackagePassed=false. On refusal preserve the reason; never rewrite evidence or remove markers to pass. Recheck frozen71/four-input pins without walking cache/tmp.
- [ ] Write the factual result and phase times with all remaining gates. Commit only explicit necessary documentation, verify origin/branch, push normally and confirm remote SHA=local and clean state. Keep raw control/output/SDK/modules/caches/binary ignored. A stopped attempt consumes this identity/approval; another attempt requires a new separately scoped decision.

## Approval handoff

Execution was subsequently approved, but the first preflight command failed during PowerShell parsing. The explicit stop rule was followed without retry. Native/inline remains the method;any corrected preflight/controller call needs a new decision, including explicit permission if retaining the still-absent001 identity. No package/main/production acceptance or historical error87/timeouts/fuzz explanation follows from this stopped preparation.
