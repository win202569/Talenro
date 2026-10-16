# Fixed Ordinary Leaf Check Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task after plan approval and execution-method selection. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement and fake-test one fixed ordinary nodeleafbinding verification entry without launching real Go or a test binary.

**Architecture:** A dedicated planner/controller reuses the unchanged ordinary matrix recorded-process path and existing compile/query/shutdown predicates. A saved-only reader validates fixed scope, explicit artifacts and a hash-bound completion marker. Independent final audits preserve primary and secondary failures.

**Tech Stack:** Existing Windows PowerShell7 and Windows PowerShell5.1, existing C# result types and safe-file helpers; no new dependencies.

**Spec:** [Approved ordinary leaf design](../specs/2026-10-16-leaf-ordinary-check-design.md). User approved the written spec after `9fdda6ff` and selected Native/inline with1 after8fa1363c. Execution is fake-only;see the [actual implementation record](../../roadmap/2026-10-16-leaf-ordinary-check-implementation.md).

## Global Constraints

- Work only in `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`, branch `codex/c12-b01-task9-coordinator`. Verify existing linked worktree/clean state; do not create another checkout or touch main.
- This plan authorizes implementation and fake-only tests only after approval. No real Go/build/binary/fuzz/race/PID/Docker/SDK inventory or full repository suite. Importing existing C# result types in test hosts is allowed; no workload may enter the real native seam.
- Do not modify product Go/tests, C# collector, ordinary completion predicate, existing matrix/mapping/diagnostic entries or their tests. No installs/downloads/security changes, timeout increase, identity replay or diagnostic binary reuse.
- Exactly four public controller/planner inputs: `[string]RunRoot,Toolchain,ReferenceReceipt,ModuleCache`. Package fixed to `internal/nodecontrol/nodeleafbinding`; policy `leaf-ordinary-check-v1`, VetMode `default`, NoRetry typed true and Acceptance typed false.
- Output root must be absent,64-bit-host-owned `leaf-ordinary-check-NNN` under the established isolated-fuzz scope; toolchain must be an owned `uninstrumented-NNN/toolchain`. Reject overlap, reparse ancestors/files and unsafe/multiple-link inputs/artifacts using existing safe helpers.
- Fixed build args: `test,-c,-x,-p=1,-count=1,-timeout=120s,-o=<root>/package.test.exe,talenro.local/platform/internal/nodecontrol/nodeleafbinding`. Build cwd workspace; native600s. No vet override.
- Fixed runtime args: `-test.count=1,-test.timeout=120s,-test.v=true,-test.run=.`. Runtime cwd original package directory; native150s.
- Exact20 environment fields/values from spec section3, fresh private cache/tmp. No argument, environment, package, selector or native-mode passthrough.
- Ordinary `Run(...,false,false)` only;500ms natural observation/shared5000ms cleanup and schema3 predicates unchanged. Error87/unknown/forced/truncated/diagnostic evidence never passes.
- At most two fake seam calls in success, one after failed build, zero after failed preflight. Record attempted launch before seam entry; persistence/uncertain launch forbids retry.
- Actual product/SDK/artifact postchecks are independent; fake them during this plan. No historical file counts or old fake totals substituted for current evidence.
- Provisional result PackagePassed=false/FinalizationRequired=true. Only a verified real completed marker may establish scoped PackagePassed=true; overall Acceptance remains false. No broader gate closes.
- Preserve necessary work with explicit Git paths and ordinary pushes to existing origin/branch under AGENTS.md; verify remote SHA=local and clean. Ignore all fixtures/raw logs/caches/binaries. Real-run approval is a later separate task.

## Review Focus

1. A plan and receipt altered identically must not establish scope; independently regenerate constants and require exact types/keys/arrays (Task1 tampered-plan controls, Task3 saved mutation).
2. A binary/helper changes after successful build but before runtime; recheck bindings and launch no runtime (Task2 boundary mutation tests).
3. Primary launch failure plus failed error persistence or multiple final audits must not suppress audits or become success (Task2/3 simultaneous-failure tests).
4. Provisional result, pending marker or contradictory STOPPED/completed must never become package success; marker publication failure must fail closed (Task3 saved/finalization tests).
5. PS5.1 serialization or misleading PASS/SKIP/subtest output must not lose typed flags/evidence or inflate test coverage (Task2 output tests, Task3 JSON roundtrip).

## Files and test commands

Create under `docs/diagnostics/isolated-fuzz/`:

- `leaf-ordinary-check.ps1`: planner, fixed-scope validation, controller and saved reader; dot-source `mapping-single-check.ps1` for unchanged ordinary helpers.
- `leaf-ordinary-check-fixture.ps1`: import-only fake factories/scenario driver; never execute fixture binaries or native processes.
- `leaf-ordinary-check-plan.tests.ps1`: planner/constants/path/type/binding tests.
- `leaf-ordinary-check.tests.ps1`: flow/output/final-audit tests; parameter `[ValidateSet('flow','all')][string]$ReviewFocus='all'`.
- `leaf-ordinary-check-saved.tests.ps1`: final marker/artifact/saved-only and JSON tests.

Documentation: update diagnostic README/current-status/recovery and this plan; create
`docs/roadmap/2026-10-16-leaf-ordinary-check-implementation.md` with actual evidence.
Implementation subsequently began after approval;the actual record,not this
original file checklist alone,identifies current commits/results and open repairs.

For each named test script run both fresh hosts with `-NoProfile -NonInteractive
-File <script>`, preserving output/exit in this plan's ignored workflow directory.
Use installed `(Get-Command pwsh).Source` and
`C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe`; verify64-bit and host
versions in each host. Missing host stops verification, not installation. Every
verification below means both hosts, exit0 with no failed assertion. Do not treat
a pipeline's last command as proof of the child exit code.

### Task1: Fixed planner and process-free fixtures

**Files:** Create helper, fixture and planner test files listed above.

**Interfaces:**
- Consume safe file/path helpers from the mapping import chain; no native calls or inventories during planning.
- Produce `Get-LeafOrdinaryCheckPlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache) -> PSCustomObject` with Schema1, Policy, fixed paths/args/env/budgets, NoRetry/Acceptance, executable/input/helper bindings.
- Produce `Assert-LeafOrdinaryCheckPlan([object]$Plan,[bool]$RequireNew) -> void`; independently reconstruct constants and safe paths, never trust two matching altered records. RequireNew=false permits owned existing-root saved reads without invoking the planner.
- Fixture `New-LeafOrdinaryTestFiles([string]$FixtureRoot) -> {Root,Toolchain,ReferenceReceipt,ModuleCache}` uses fake data files only. `Check-LeafOrdinary([bool]$Value,[string]$Name)` throws on false and emits one PASS assertion on true.

- [x] **Step1: Write named planner assertions.** Include `PlannerExists`, `PlanningLaunchesZero`, `FixedScope`, `FixedArguments`, `FixedEnvironment20`, `TypedFlags`, `BindingsPresent`, `WrongOrExistingRoot`, `OverlapRejected`, `UnsafeInputRejected`, `TamperedPlanRejected`, `InputOrSourceDriftRejected`. Set SDK/product/native seams to throw if planning calls them. Representative exact assertions:

```powershell
Check-LeafOrdinary ($p.BuildLimitSeconds -eq 600 -and $p.ExecuteLimitSeconds -eq 150) 'FixedBudgets'
Check-LeafOrdinary ($p.Policy -ceq 'leaf-ordinary-check-v1' -and $p.VetMode -ceq 'default') 'FixedScope'
Check-LeafOrdinary ($p.Acceptance -is [bool] -and -not $p.Acceptance -and $p.NoRetry -is [bool] -and $p.NoRetry) 'TypedFlags'
Check-LeafOrdinary ($p.Environment.Count -eq 20 -and $p.BuildArguments -cnotcontains '-vet=off') 'DefaultVetNoOverride'
Check-LeafOrdinary (-not (Test-Path -LiteralPath $f.Root)) 'PlannerDoesNotCreateRoot'
```

- [x] **Step2: Run `leaf-ordinary-check-plan.tests.ps1`.** Expected RED at missing planner or an exact missing behavior; fake scaffolding/path errors are not valid behavioral RED. Record actual assertion/exit.
- [x] **Step3: Implement planner/fixed validator and fixture factories.** Bind exactly13 production helpers: `leaf-ordinary-check.ps1`, `mapping-single-check.ps1`, `matrix-check.ps1`, `envelope-uninstrumented.ps1`, `envelope-sample.ps1`, `envelope-observe.ps1`, `run.ps1`, `isolation.ps1`, `provenance.ps1`, `process-native.cs`, `retained-lifetime.cs`, `compile-stage-trigger.cs`, `compile-stage-session.cs`. Bind four inputs: reference receipt, `copy-receipt.json` beside the toolchain directory, `toolchain/bin/go.exe`, and `GOMODCACHE/github.com/google/uuid@v1.6.0/go.mod`, with safe file/hash checks. This is not a full SDK audit. Keep current bindings reproducible without old ignored data. Hash rechecks belong to controller boundaries and saved validation as well.
- [x] **Step4: Rerun planner suite.** Expected GREEN in both hosts, every required assertion, launch0, no output root. Record actual counts and static-only checks separately.
- [x] **Step5: Commit Task1.** Explicitly stage only these three files; message `feat: add fixed ordinary leaf planner and fake controls`. Expected commit with frozen old helper/source files unchanged.

### Task2: Ordinary build/runtime flow and output evidence

**Files:** Modify helper/fixture; create controller tests.

**Interfaces:**
- Consume Task1 planner/validator and bindings; existing `Invoke-MatrixRecordedProcess`, `Assert-ProcessCompletion`, `Assert-MatrixCompileEvidence`, `Assert-CopiedCompileEvidence`, `Assert-MappingQueryEvidence` and safe-file helpers.
- Produce `Invoke-LeafOrdinaryCheck` with Task1's four string inputs. In this task return/persist only provisional result; Task3 adds authoritative completion.
- Produce `Get-LeafOrdinaryTestOutput([string]$Stdout,[string]$Stderr) -> {TopLevelRun,TopLevelPass,TopLevelSkip,SubtestRun,SubtestPass,SubtestSkip}` arrays of observed names, not a source-inventory coverage claim; reject invalid PASS/FAIL/instrumentation conditions.
- Fixture `Invoke-LeafOrdinaryTestCase([string]$Scenario) -> {Root,NativeCalls,Calls,Result,Error,Saved,AuditCalls}` captures errors and actual saved provisional state. Replace `Invoke-MatrixNative` in fixture scope, never `[IsolatedFuzzProcess]::Run` itself; calls record executable/args/env/cwd/limit/order and write fake receipt streams.

- [x] **Step1: Write flow/output RED cases.** Scenarios: `success`, `preflight`, `build-exit`, `build-timeout`, `build-error87`, `build-diagnostic`, `build-schema`, `build-query`, `missing-binary`, `missing-compile-evidence`, `binary-before-run`, `source-before-run`, `runtime-exit`, `runtime-timeout`, `runtime-forced`, `runtime-query`, `no-tests`, `no-final-pass`, `fail-line`, `diagnostic-stdout`, `diagnostic-stderr`, `skip-and-subtest`, `native-throw`, `native-throw-persist`. Use both build/runtime sides for diagnostic/truncated/unknown-member controls. Exact assertions include:

```powershell
Check-LeafOrdinary ((Invoke-LeafOrdinaryTestCase 'preflight').NativeCalls -eq 0) 'PreflightLaunchesZero'
Check-LeafOrdinary ((Invoke-LeafOrdinaryTestCase 'build-error87').NativeCalls -eq 1) 'Error87StopsBeforeRuntime'
Check-LeafOrdinary ((Invoke-LeafOrdinaryTestCase 'binary-before-run').NativeCalls -eq 1) 'BoundaryBindingStopsRuntime'
```

For `success`, assert two calls with exact global constants, launch progress already saved at each seam entry, no diagnostics and provisional PackagePassed=false. For `skip-and-subtest`, assert names/counts in separate arrays, one skip is disclosed and not counted as an executed test body. For persistence failure, require original native exception plus secondary persistence error and no retry.

- [x] **Step2: Run `leaf-ordinary-check.tests.ps1 -ReviewFocus flow`.** Expected behavioral RED on missing controller/flow/output gate, no real native entry. Capture failure before implementing.
- [x] **Step3: Implement fixed controller flow and output reader.** Record each attempt durably, check constants/source/input/executable bindings immediately before the seam, require both ordinary guards and initial-query evidence for each phase. Bind build receipts/streams/binary before considering runtime; no downstream launch after any failed check. Use result fields from spec; missing/unknown evidence remains false/error, not fabricated.
- [x] **Step4: Run planner and flow suites.** Expected GREEN in both hosts; success2, failed-build1, preflight0; source/binary drift prevents runtime. Audit/finalization may remain explicitly provisional until Task3, never passed.
- [x] **Step5: Commit Task2.** Explicit helper/fixture/controller-test paths only; message `feat: add ordinary leaf flow with fail-closed fake tests`. Expected no old guard/collector/product changes.

### Task3: Independent audits, durable saved evidence and final verification

**Files:** Modify helper/fixture/controller tests; create saved tests and implementation report; update README/current-status/recovery/plan.

**Interfaces:**
- Consume Task2 flow/result and Task1 fixed validator in existing-root mode.
- Produce `Get-LeafOrdinarySavedEvidence([string]$RunRoot) -> PSCustomObject` with typed PackagePassed/Acceptance, scoped facts and errors. It reads only explicit saved evidence plus bounded current file/source/input hashes; no native call, planner, PID query, full SDK/product/cache inventory or mutation. Invalid/missing evidence throws or yields PackagePassed=false with error; never repairs records.
- Completed marker: Schema1, Policy=`leaf-ordinary-check-v1`, Package=`internal/nodecontrol/nodeleafbinding`, VetMode=`default`, NoRetry=true, Acceptance=false, PackagePassed=true, FinalizationRequired=false, ManifestSHA256. Provisional `result.json` keeps PackagePassed=false/FinalizationRequired=true.
- `evidence-manifest.json`: Schema1, same policy/package, explicit unique root-relative Files rows `{Path,Length,SHA256}`. Exclude manifest/pending/completed markers and cache/tmp; require saved scope, source/input bindings, attempts, both receipts/streams, binary, inventories, artifact binding and result for successful completion.

Exact successful files: `plan.json`, `source-binding.json`, `input-binding.json`,
`progress.json` (durable attempts), `before-product.json`, `after-product.json`,
`artifact-binding.json`, `result.json`, `package.test.exe`, and `receipt.json`,
`stdout.txt`, `stderr.txt` under each of `build/` and `execute/`. The manifest binds
these15 files; it does not bind itself or `completed.pending.json`/`completed.json`.
`stopped.json` and phase-local `launch-error.json` are failure records, not required
successful artifacts. Saved scope validation checks the explicit required set,
not just whatever subset a mutated manifest happens to list.

- [x] **Step1: Add final-audit and saved RED assertions.** Controller scenarios: `post-product`, `post-sdk`, `post-artifact`, `all-audits-fail`, `result-write`, `manifest-write`, `pending-write`, `marker-publish`, `stopped-write`, `native-throw-all-audits-fail`. Assert every eligible final audit attempted, primary/secondary errors preserved, PackagePassed false/no valid completed marker on failure. Saved tests: `CompletedBindingPasses`, `ProvisionalRefused`, `PendingRefused`, `ContradictionRefused`, `ManifestHashMutation`, `ArtifactMutation`, `UnsafeOrDuplicateRow`, `MissingRequiredRow`, `AlteredPlanAndReceipt`, `SourceInputDrift`, `TypedBooleanMutation`, `PS51Roundtrip`, `SavedReaderLaunchesZero`. Use copied fake output, not old diagnostic identities.
- [x] **Step2: Run full controller and saved suites.** Expected RED on the missing independent-audit/final-marker/saved behavior. Record each failure motivating a fix; do not accept a syntax/fixture defect as RED proof.
- [x] **Step3: Implement independent audit and finalization behavior.** Attempt product/SDK/artifact checks in separate guarded blocks, collect all audit errors, save actual timing and inventories. Serialize nested receipts/bindings at sufficient depth and test roundtrip. Publish owned pending marker without overwrite only after safe explicit manifest validation. Preserve outer failure if error persistence also fails. Saved reader independently verifies constants, required files, marker/manifest binding and contradictions before scoped PackagePassed=true; no circular hash or inference from provisional booleans.
- [x] **Step4: Run three new suites in both hosts.** Expected GREEN; successful fake completed marker only, no real calls, all negative cases fail closed, current exact counts recorded.
- [x] **Step5: Run final allowlist in both fresh hosts after freezing all source/test hashes.** Exactly eight scripts: `leaf-ordinary-check-plan.tests.ps1`, `leaf-ordinary-check.tests.ps1`, `leaf-ordinary-check-saved.tests.ps1`, `mapping-single-check.tests.ps1`, `matrix-check.tests.ps1`, `leaf-stage-build-only-plan.tests.ps1`, `leaf-stage-build-only-scope.tests.ps1`, `leaf-stage-build-only.tests.ps1`. All paths under `docs/diagnostics/isolated-fuzz`. Expected16 child commands exit0, separately reported actual assertion/static counts and matching post-run pins. Do not run other `*.tests.ps1` by wildcard. Stop and diagnose a failure; do not expand authorization or hide it with an unchanged retry.
- [x] **Step6: Complete documentation and one fresh whole-increment review.** Record actual RED/GREEN sequence,16 final command outcomes, host versions/counts/pins, current limitations and any unresolved finding. Review scope is this new entry's correctness and compatibility, not an OS explanation of87 or real-environment acceptance. Follow the chosen execution skill's review/fix rules; no second reviewer as a substitute for covering tests. Expected no unresolved Critical/Important findings before declaring implementation complete; deferred minors explicitly listed.
- [ ] **Step7: Preserve the implementation checkpoint.** Review explicit new files and four progress/recovery documents, whitespace/local links, frozen old-helper/product diffs; commit `feat: finalize ordinary leaf evidence and fake verification` plus necessary follow-up docs. Ordinary-push existing verified origin/branch and verify remote=local/clean. Expected portable implementation/evidence summary, no raw files/secrets/binaries. A real run is NOT the next automatic command; propose it separately only after this plan is complete and reviewed.

## Plan self-review and handoff

Coverage: spec inputs/constants→Task1; ordinary policy/launch/output→Task2;
audits/manifest/marker/saved reader→Task3; process-free compatibility and recovery
→Task3 final allowlist/docs. Five Review Focus risks have named tests above.
User selected Native/inline with1 after8fa1363c. All three tasks and one review's
two Important repairs have verified fake results:final407 checks per host,
16 commands exit0,71 source/test pins unchanged. Resume from the actual
implementation record,not by repeating
the original approval or running an old identity. No real execution is approved.

Review repair clarification:Task2's shared recorded wrapper can lose the primary
native exception during its own error persistence. To meet the approved spec
without changing shared helpers,use a leaf-owned
`Invoke-LeafOrdinaryRecordedProcess(Plan,Build,State)` adapter over the unchanged
`Invoke-MatrixNative` ordinary call. Keep its typed native fields/record paths,
safe persistence and one-call boundary covered by the new inner-persistence
regression. Add leaf-local native-model shape validation before unchanged shared
predicates;receipt truthiness or scalar collections must not establish success.
