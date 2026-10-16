# Leaf Stage Build-Only Diagnostics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a fixed vet-off compile/link diagnostic without changing required vet acceptance or the32-attempt retained sampling policy.

**Architecture:** A separate planner/controller owns the new run identity and explicit scope marker. It reuses existing stage/core/lifetime facts and native runner unchanged; a separate scope validator prevents altered plans/receipts from becoming scoped completion. Only the common ordinary guard gains a stricter metadata refusal.

**Tech Stack:** Windows PowerShell7/5.1,existing Add-Type C#/.NET helpers;no dependencies or installation.

**Spec:** [Approved written design](../specs/2026-10-16-leaf-stage-build-only-design.md).

**Status:** User approved the written spec after b8e7956dafd54fe884e0ad2ac1554d97944942fe.
User approved this plan and selected option1 Native/inline after8a25712f.
All implementation/review repairs are committed througha4bfe80a. Final21 suites
per host passed884 checks (11 static included),all42 commands exit0;66 source/test
pins matched. Final documentation/push preservation is in progress.
No real attempt is authorized by implementation approval.

## Global Constraints

- Reuse D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator,branch codex/c12-b01-task9-coordinator. Verify the linked worktree at execution;never reset user edits or work on main.
- New absent owned basename leaf-stage-build-only-NNN,three digits,beneath .superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator;no output/input overlap,reparse paths or multi-link inputs. Never spoof an old root name or replay001.
- Fixed argument sequence:test,-c,-x,-p=1,-vet=off,-count=1,-timeout=120s,-o=<RunRoot>/package.test.exe,talenro.local/platform/internal/nodecontrol/nodeleafbinding. One vet-off argument,no arbitrary overrides.
- Fixed20 environment keys:GOROOT=Toolchain;GOCACHE=RunRoot/go-cache;GOMODCACHE=ModuleCache;TEMP/TMP/GOTMPDIR=RunRoot/tmp;GOTOOLCHAIN=local;GOENV/GOPROXY/GOSUMDB/GOAUTH=off;GOVCS=all:off;GO111MODULE=on;GOWORK=off;GOFLAGS=-mod=readonly;GOOS=windows;GOARCH=amd64;TALENRO_FUZZ_DIAG_MODE/TALENRO_FUZZ_FIXTURE/GODEBUG empty. No sandbox/complete-cache-audit claim.
- Policy leaf-stage-build-only-v1,VetMode=off,VetEvidenceComplete=false,DiagnosticOnly=true,Coverage=sampled,NoRetry=true. Acceptance/PackagePassed/OriginalTestPassed/RuntimeAttempted stay false even for a complete diagnostic.
- Trigger schema1/policy leaf-test-compile-log-v1;fixed compiler/source directory/package;16777216 total bytes,65536 iteration/raw-line bytes,60000 polls;unchanged exact-root activation and clocks.
- 32 cumulative attempts,0x101000 access,inherit=false,no eviction/reopening/extra cohort.600s native deadline,500ms natural/shared5000ms cleanup. No C# changes or image filters.
- Preserve old full-stage/eager/default/matrix dispatch,arguments and source-list semantics. No new vet-only entry or automatic chaining;ordinary vet acceptance remains required outside this diagnostic.
- Implementation verification is fake-only:no Go/native Run,generated binary,PID query,full SDK/product inventory,Docker,download/install,security change or real vet. Fresh child shells run tests;Add-Type,file metadata/hash/JSON operations are allowed.
- Commit/push explicit necessary source/tests/docs under AGENTS.md,not raw receipts/logs/toolchains/caches/binaries/secrets. Ordinary dev-branch checkpoints do not grant acceptance/main merge. Confirm remote SHA and clean state at handoff.

## Review Focus

1. Plan and receipt altered together must fail fixed constants,not pass by agreeing with each other (Task2 JointTampering).
2. Existing evidence roots must validate without invoking the unused-root planner or inventing/deleting a root (Task2 ExistingRootReadOnly).
3. PS5.1 JSON arrays,empty environment values and typed false flags must survive without dropping pins (Task1 FlatPlanRoundtrip;Task2 TypedMarkerRoundtrip).
4. Diagnostic completion must remain vet/ordinary-ineligible even with trigger/flags removed from a marker-bearing copy (Task2 MarkerOnlyRefusal;Task3 ScopedCompletion).
5. Earlier failures must not suppress later independent audits,and provisional results/marker failures cannot imply durable completion (Task3 IndependentFinalAudits/PersistenceFailure).

---

## File responsibilities and shared contracts

Create docs/diagnostics/isolated-fuzz/leaf-stage-build-only.ps1:planner,scope
validation and explicit single-call controller. Modify run.ps1 only to add the
BuildOnlyDiagnostic presence refusal before the existing trigger refusal.
No existing controller or C# body is refactored;imports and pure helpers are reused.

Create test-only leaf-stage-build-only-fixture.ps1,leaf-stage-build-only-plan.tests.ps1,
leaf-stage-build-only-scope.tests.ps1 and leaf-stage-build-only.tests.ps1 beside it.
The fixture exports helpers only;never dot-source a self-running legacy test to
obtain its factories. Follow the real DTO/log patterns in existing
leaf-stage-retained-compile.tests.ps1 without changing that file or adding test seams
to production classes. Fresh tiny fixtures go under a GUID-owned unit directory
inside the allowed diagnostic scope;no previous receipts,SDK or private cache needed.

The16 production SourceBinding paths are exactly:leaf-stage-build-only.ps1,
leaf-stage-retained-compile.ps1,leaf-retained-compile.ps1,retained-child-observation.ps1,
mapping-single-check.ps1,matrix-check.ps1,envelope-uninstrumented.ps1,envelope-sample.ps1,
envelope-observe.ps1,run.ps1,isolation.ps1,provenance.ps1,process-native.cs,
retained-lifetime.cs,compile-stage-trigger.cs,compile-stage-session.cs.

All paths in task file lists below are relative to docs/diagnostics/isolated-fuzz/
unless written with a docs/ prefix. Each verification command runs from the
worktree root. For two-host runs,use this command form with each task's exact
literal file list below (no live switches):

```powershell
$testHosts=@((Get-Command pwsh.exe).Source,'C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe')
foreach($testHost in $testHosts){foreach($testFile in $testFiles){
    & $testHost -NoProfile -NonInteractive -File (Join-Path 'docs/diagnostics/isolated-fuzz' $testFile)
    if($LASTEXITCODE -ne 0){throw ('Test failed: '+$testFile)}
}}
```

Each step defines $testFiles explicitly. Require exit0 and read every output;save
full logs in this plan's ignored workspace. No execution-policy changes.

### Task1: Fixed planner and source identity

**Files:** Create leaf-stage-build-only.ps1,leaf-stage-build-only-fixture.ps1,
leaf-stage-build-only-plan.tests.ps1.

**Interfaces:**
- Produces Get-LeafStageBuildOnlyPlan([string]RunRoot,[string]Toolchain,[string]ReferenceReceipt,[string]ModuleCache) -> PSCustomObject with existing stage-plan fields plus DiagnosticPolicy,VetMode,VetEvidenceComplete,PackagePassed=false,OriginalTestPassed=false;source names exactly the16 above.
- Produces Assert-LeafStageBuildOnlySources([object]Plan) -> no pipeline output;throws on missing/duplicate/wrong helper names,pins or actual go.exe drift. It performs only scoped file checks/hashes,not SDK inventories or process APIs.
- Test-only New-BuildOnlyTestFiles([string]FixtureRoot) -> object {Root,Toolchain,ReferenceReceipt,ModuleCache}. Root is an uncreated leaf-stage-build-only-901 directory;Toolchain lives at FixtureRoot/uninstrumented-901/toolchain;reference and tiny module/go files stay inside FixtureRoot.
- Test-only Check-BuildOnly([bool]Value,[string]Name) throws ASSERTION FAILED:<Name>,otherwise writes PASS:<Name>. Importing the fixture creates no run or process.

- [x] **Step1: Write planner RED tests and tiny file factory.** Import the new
  production file only if present,then assert a named missing-entry failure.
  Pin the complete argument array and20-key environment,not substring presence:

```powershell
Check-BuildOnly ($null -ne (Get-Command Get-LeafStageBuildOnlyPlan -ErrorAction SilentlyContinue)) 'BuildOnlyPlannerExists'
$plan=Get-LeafStageBuildOnlyPlan $f.Root $f.Toolchain $f.ReferenceReceipt $f.ModuleCache
$argsExpected=@('test','-c','-x','-p=1','-vet=off','-count=1','-timeout=120s',('-o='+(Join-Path $f.Root 'package.test.exe')),'talenro.local/platform/internal/nodecontrol/nodeleafbinding')
Check-BuildOnly (($plan.BuildArguments -join "`n") -ceq ($argsExpected -join "`n")) 'FixedArguments'
Check-BuildOnly ($plan.Environment.Count -eq 20 -and $plan.Environment.GOPROXY -ceq 'off' -and $plan.Environment.GOFLAGS -ceq '-mod=readonly') 'FixedEnvironment'
Check-BuildOnly ($plan.SourceBinding.Count -eq 16 -and -not(Test-Path $f.Root) -and $plan.BuildLimitSeconds -eq 600 -and $plan.VetEvidenceComplete -is [bool] -and -not $plan.VetEvidenceComplete) 'PinsAndNoLaunch'
```

  Define $f with New-BuildOnlyTestFiles in each case;assert every environment
  value and every helper name/hash too. Add ExistingRoot/WrongRoot,Overlap,
  MissingInput,PlanSourceDrift and FlatPlanRoundtrip cases. Reparse/link rejection
  uses existing safe-path primitives;do not change Windows settings/create privileged
  links merely for a test. Check the old eager/full-stage plans keep their exact
  vet-enabled argument arrays using the same tiny inputs and separate old root names.
- [x] **Step2: Observe RED.** Run
  `pwsh.exe -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/leaf-stage-build-only-plan.tests.ps1`.
  Expected:exit1,ASSERTION FAILED:
  BuildOnlyPlannerExists,not a syntax/import error or attempted executable launch.
- [x] **Step3: Implement the fixed planner/source verifier.** Import the existing
  stage helper,validate the actual new identity,produce the spec's constants and
  flat typed values. Freeze go.exe and the16 named paths;do not call the old planner
  with a synthetic root. Keep extra private helpers inside this new file only.
- [x] **Step4: Observe GREEN on both hosts.** Set
  $testFiles=@('leaf-stage-build-only-plan.tests.ps1') and use the two-host loop.
  Expected:exit0,all named assertions PASS,no native
  runner/SDK-product audit calls,no new Root/cache directories from planning.
- [x] **Step5: Inspect and commit only Task1's three files.** Run staged whitespace
  and explicit file-list checks;commit diagnostics: add fixed build-only diagnostic plan.
  Record actual RED/GREEN commands,counts,source pins and commit in the plan ledger.

### Task2: Scope facts and unconditional ordinary refusal

**Files:** Modify leaf-stage-build-only.ps1,leaf-stage-build-only-fixture.ps1,
run.ps1:Assert-ProcessCompletion;create leaf-stage-build-only-scope.tests.ps1.

**Interfaces:**
- Consumes Task1 Get-LeafStageBuildOnlyPlan(RunRoot,Toolchain,ReferenceReceipt,ModuleCache),Assert-LeafStageBuildOnlySources(Plan) and New-BuildOnlyTestFiles(FixtureRoot),with their exact Task1 signatures/return contracts.
- Produces Get-LeafStageBuildOnlyEvidence([object]Receipt,[object]Plan,[string]StderrPath) -> object {DiagnosticOnly,Coverage,Acceptance,VetEvidenceComplete,ScopeEvidenceComplete,ScopeError,TriggerEvidenceComplete,TriggerError,CoreNaturalComplete,CoreError,LifetimeEvidenceComplete,LifetimeError,ObservationPassed}. Scope false cannot be overridden by other true facts.
- Consumes unchanged Get-LeafStageRetainedCompileEvidence(Receipt,Plan,StderrPath),safe-path/hash helpers and existing native report DTOs;never strips metadata or calls a native Run.
- Test-only New-BuildOnlyNativeFixture([object]Plan) -> object {Native,Log}. Native is an IsolatedFuzzProcess+Result with root17,entry/PID23,one admitted/released entry,query100–110ms,open120–130ms,membership140–150ms,admission160–170ms,root-wait300–310ms,final500–510ms,frequency1000 and UTC epoch2026-10-16T00:00:00Z. Receipt spans0–1000ms;candidate10–20ms,root-check30–40ms with WaitResult258,activation50–60ms precede query1. Preserve existing valid core row/accounting/list fixture patterns. Log starts with the copied compiler,-p exact target,-pack and an absolute budget_test.go in Plan.Trigger.SourceDirectory,then copied context/context.go and internal/fuzz/fuzz.go compile-token lines. Quote Windows paths using existing doubled-backslash rules;hash only the first eligible raw LF record for trigger provenance. No live PID/handle or installed SDK required.
- Scope test file accepts [ValidateSet('all','scope','guard')][string]Focus='all' so both missing scope and missing guard can have separate observed REDs.

- [x] **Step1: Write independent scope/guard RED tests.** For the scope fixture,
  explicitly create Root/build,save the synthetic log,convert the fake native
  receipt with the production converter,add WorkingDirectory and only then add
  the exact BuildOnlyDiagnostic marker. Test fresh and JSON-roundtripped copies:

```powershell
Check-BuildOnly ($null -ne (Get-Command Get-LeafStageBuildOnlyEvidence -ErrorAction SilentlyContinue)) 'BuildOnlyScopeExists'
$e=Get-LeafStageBuildOnlyEvidence $receipt $plan $stderr
Check-BuildOnly ($e.ScopeEvidenceComplete -and $e.TriggerEvidenceComplete -and $e.CoreNaturalComplete -and $e.LifetimeEvidenceComplete -and $e.ObservationPassed -and -not $e.VetEvidenceComplete -and -not $e.Acceptance) 'TypedMarkerRoundtrip'
foreach($value in @($null,$false,'bad',[pscustomobject]@{Schema=1;Policy='leaf-stage-build-only-v1';VetMode='off';VetEvidenceComplete=$false})) {
    $ordinaryCopy|Add-Member BuildOnlyDiagnostic $value -Force
    $message=$null;try{Assert-ProcessCompletion $ordinaryCopy}catch{$message=$_.Exception.Message}
    Check-BuildOnly ($message -ceq 'Build-only diagnostics are not acceptance') 'MarkerOnlyRefusal'
}
```

  $ordinaryCopy is an otherwise valid ordinary converted empty/native DTO receipt
  with both mode flags false and no trigger metadata;prove it passes before adding
  the marker. Assert new marker is absent on ordinary conversions. Add MissingMarker,
  WrongSchema/Policy/Type,VetEvidenceTrue,ArgumentOrder/DuplicateVet/MissingVet,
  EnvironmentDrift,WrongRoot/Package/Executable/Streams,MissingOrDuplicatePin.
  JointTampering changes both plan/receipt arguments,environment or trigger limits
  together and still requires ScopeEvidenceComplete=false. ExistingRootReadOnly
  snapshots saved file hashes before/after validation and forbids planner invocation.
  MissingTrigger,cap32,zero entries,core87,source-query chronology and missing core
  rows must stay incomplete even with valid scope. Never repair a receipt to pass.
- [x] **Step2: Observe both REDs.** Run
  `pwsh.exe -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/leaf-stage-build-only-scope.tests.ps1 -Focus scope`,
  then the same command with -Focus guard. Expected:exit1 for BuildOnlyScopeExists and
  MarkerOnlyRefusal respectively. Legacy guards must otherwise still pass.
- [x] **Step3: Implement scope composition and the guard.** Validate plan against
  fixed constants/current16 source names/pins/go hash as well as receipt bindings;
  two altered records agreeing is insufficient. Existing-root saved validation
  must not rerun the unused-root planner. Keep errors per fact;ObservationPassed
  is scope AND existing helper's ObservationPassed. Add the BuildOnlyDiagnostic
  property-presence refusal first in run.ps1 with the exact message above;do not
  introduce a native DTO/default-null marker or alter other serializers/guards.
- [x] **Step4: Observe GREEN.** Set $testFiles=@('leaf-stage-build-only-plan.tests.ps1',
  'leaf-stage-build-only-scope.tests.ps1','compile-stage-receipt.tests.ps1',
  'retained-lifetime-receipt.tests.ps1') and use the two-host loop.
  Expected:all exit0,new scope/guard positives and negatives PASS;
  old native signatures/ordinary JSON/legacy refusal messages unchanged.
- [x] **Step5: Inspect four explicit task files and commit** diagnostics: bind build-only scope and refusal.
  Record actual scoped evidence and commit in the ledger;do not claim live vet
  absence or future capture completeness from synthetic metadata.

### Task3: Single-call controller,durable evidence and compatibility

**Files:** Modify leaf-stage-build-only.ps1,leaf-stage-build-only-fixture.ps1;
create leaf-stage-build-only.tests.ps1 and docs/roadmap/2026-10-16-leaf-stage-build-only-implementation.md.
Update docs/roadmap/current-status.md,docs/runbooks/repository-recovery.md,
docs/diagnostics/isolated-fuzz/README.md and this plan/spec execution status only.

**Interfaces:**
- Consumes Get-LeafStageBuildOnlyPlan(RunRoot,Toolchain,ReferenceReceipt,ModuleCache),Assert-LeafStageBuildOnlySources(Plan),Get-LeafStageBuildOnlyEvidence(Receipt,Plan,StderrPath) with the exact Tasks1/2 signatures/return contracts,and existing JSON,SDK/product audit,compile-token,binding and safe-file helpers;no old-controller invocation or C# edits.
- Produces Invoke-LeafStageBuildOnlyNative([object]Plan,[string]Stdout,[string]Stderr) -> IsolatedFuzzProcess+Result through the existing RunWithCompileStageRetainedDiagnostics call with600s. This is the sole external execution seam.
- Produces Invoke-LeafStageBuildOnly([string]RunRoot,[string]Toolchain,[string]ReferenceReceipt,[string]ModuleCache) -> one finalized diagnostic state on success;throws after durable stopped/provisional evidence on failure. Plan and arguments cannot be supplied as overrides to this entry.
- State preserves existing stage-controller fields and adds DiagnosticPolicy,VetMode,VetEvidenceComplete,ScopeEvidenceComplete;Evidence is Task2's result. Both marker schema1/policy and vet-off/false evidence remain in receipt/result/completion records. NativeAttempts is0 before launch,then1;all acceptance/runtime flags stay false.
- Test-only Invoke-BuildOnlyTestCase([string]Scenario) -> object {Root,Plan,NativeCalls,ToolAuditCalls,SourceAuditCalls,Failure,Saved,Output};local fake native/tool/product boundaries use New-BuildOnlyNativeFixture,real source/hash/log validators and persistence code. Restore no global state;each test file runs in a fresh shell.

- [x] **Step1: Write controller RED and scenario tests.** Replace only full SDK
  audit/product inventory/native boundaries in test scope. Spy checks exact exe,
  arguments,environment and600s budget before writing synthetic streams/binary:

```powershell
Check-BuildOnly ($null -ne (Get-Command Invoke-LeafStageBuildOnly -ErrorAction SilentlyContinue)) 'BuildOnlyControllerExists'
$good=Invoke-BuildOnlyTestCase 'success'
Check-BuildOnly ($good.NativeCalls -eq 1 -and $good.ToolAuditCalls -eq 2 -and $good.SourceAuditCalls -eq 4 -and $good.Output.Count -eq 1 -and $good.Saved.ObservationPassed -and -not $good.Saved.VetEvidenceComplete -and -not $good.Saved.PackagePassed -and -not $good.Saved.RuntimeAttempted) 'ScopedCompletion'
Check-BuildOnly ((Test-Path (Join-Path $good.Root 'completed.json')) -and -not(Test-Path (Join-Path $good.Root 'stopped.json'))) 'BoundCompletionMarker'
$bad=Invoke-BuildOnlyTestCase 'multi-audit'
Check-BuildOnly ($bad.NativeCalls -eq 1 -and $bad.ToolAuditCalls -eq 2 -and $bad.SourceAuditCalls -eq 4 -and $bad.Saved.AuditErrors.Count -eq 3 -and -not $bad.Saved.ObservationPassed) 'IndependentFinalAudits'
```

  Execution correction:four source checks include the three lifecycle audits
  plus fixed-current-pin validation inside the saved scope validator.
  Add Existing/WrongRoot/PreTool cases(no native call),NativeThrow(no fabricated
  receipt;launch-error cleanup unverified),Exit/Timeout/Forced/Core87/MissingCore,
  MissingTrigger/BadPrefix/BadClock,Cap32/ZeroEntry/ReleaseFailure,ScopeDrift,
  MissingBinary/CompileTokens,PostProduct/PostTool/SourceDrift/ArtifactDrift and
  PersistenceFailure(result,manifest,completion marker). Failed calls emit no
  success object,no completed marker,no second native/test-binary/vet call.
  In multi-audit,start with nonzero compile,then product/tool/source final failures;
  preserve the original compile failure and all three audit errors. SourceDrift
  mutates only a tiny fake executable or saved pin,never repository/SDK sources.
  ScopeDrift changes Plan.BuildArguments[4] to -vet=all inside the fake native
  seam after recording the invocation;fixed scope must refuse the returned state.
  Verify manifest hashes/lengths and completion binding,scope fields/false flags,
  excluded cache/tmp,safe paths and continued audits after persistence faults.
  For file-as-directory persistence faults,check markers with -PathType Leaf;
  an obstructing completed.json directory is not a completion marker and must not
  be removed to force success. Missing result.json means Saved may be null.
  STATIC routing checks mark code inspection separately;never call the real native
  seam. Inspect routing before running tests to prevent an accidental live boundary.
- [x] **Step2: Observe controller RED.** Run
  `pwsh.exe -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/leaf-stage-build-only.tests.ps1`.
  Expected:exit1,ASSERTION FAILED:BuildOnlyControllerExists,not native
  launch or a fixture parser error. No placeholder native implementation to fake GREEN.
- [x] **Step3: Implement the separate lifecycle.** Follow spec6 using established
  primitives:plan/source check/full pre-audits,one attempt,normal conversion plus
  marker,artifact binding,independent scope/trigger/core/lifetime/compile facts,
  all applicable final audits and stopped/provisional/manifest/completion protocol.
  Call Invoke-LeafStageBuildOnlyNative($plan,$stdout,$stderr) exactly once after
  preflight;do not bypass that replaceable seam with an inline native call.
  Freeze scope in all final records;do not infer vet absence or reuse ordinary
  acceptance guards. Keep primary errors alongside final audit errors. No collector,
  old-controller dispatch or arbitrary override refactor. Add no new profile to matrix.
- [x] **Step4: Run GREEN and final fake-only batch on both fresh hosts.** Run the
  three new test files and these18 existing files,one fresh child host per file:
  compile-stage-trigger,compile-stage-session,compile-stage-receipt,
  leaf-stage-retained-compile,matrix-check,pair-sample,envelope-uninstrumented,
  envelope-sample,envelope-compile,envelope-preparation,envelope-observe,
  mapping-single-check,retained-lifetime,retained-lifetime-integration,
  retained-lifetime-receipt,retained-child-observation,leaf-retained-compile,
  short-child-observation (each suffixed .tests.ps1).
  Define $testFiles as these18 literal filenames plus
  leaf-stage-build-only-plan.tests.ps1,leaf-stage-build-only-scope.tests.ps1,
  leaf-stage-build-only.tests.ps1,and use the two-host loop without Focus switches.
  Expected:21 suites per host,42 exit0 commands,no real execution. Record exact
  observed assertion/static counts;698 is historical,not this batch's expected total.
  Inspect default modes before running;never enable live/runtime switches or install
  missing dependencies. New fixtures must be portable;disclose any legacy planning
  checks that depend on local SDK/module paths. Verify changed PS syntax,staged
  whitespace and local doc links separately;freeze source hashes for final log binding.
- [x] **Step5: Request one fresh whole-increment review after inline execution**
  against this spec/plan,Review Focus,interfaces,old-mode isolation and exclusions;
  use the chosen execution skill's review flow if subagent-driven instead.
  Resolve Important/Critical findings with named RED/GREEN and rerun the complete
  approved batch on fixed sources. Record deferred minors/rulings/test/live limits;
  no repeated review dispatch solely to declare a tested repair addressed.
- [ ] **Step6: Record evidence and preserve only necessary files.** Write actual
  implementation report/status/recovery with suite counts,commits,remaining gates;
  archive raw logs/bindings in a new ignored evidence identity,not Git. Review the
  explicit staged file list and commit diagnostics: add isolated build-only controller.
  Ordinary-push the existing branch,confirm remote SHA=local and clean state. This
  checkpoint is not main acceptance or permission for another real compile.

## Self-review and handoff

Spec1–4 map to Task1's fixed planner and Task3's unchanged native route;spec5 to
Task2's scope/guard and Task3's marker/result protocol;spec6 to Task3's audits and
persistence;spec7 to all three RED/GREEN steps and the final21-suite batch;spec8
to approved-scope documentation/preservation. Each of the five Review Focus
conditions has a named owning test above. All dependencies have exact interfaces;
fixture additions are test-only and do not count among16 production helpers.

Recommended method:Native/inline,because three tasks share one fixed plan/scope/
controller contract and reuse an already-tested collector. It keeps implementation
in this session with one fresh final reviewer;subagent-driven offers independent
per-task reviews at greater context cost. The user must review this plan and choose
the method before either begins. Later real execution still requires separate
approval,frozen inputs,new identity,full pre/post audits and no binary runtime.
