# Account Panel Alignment Experiment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a single-property Account alignment candidate with offline scope regression evidence, without claiming visual repair.

**Architecture:** Preserve existing XAML and geometry machinery. A test-only validator permits exactly the selected Center attribute on AccountGeometryPanel while retaining the original canonical baseline hash. No runtime helper or diagnostic switch is added.

**Tech Stack:** C#/net10.0, xUnit, System.Xml.Linq, Windows PowerShell; existing dependencies only.

**Spec:** [Approved short design](../specs/2026-11-11-windows-account-panel-alignment-experiment-design.md).

## Global Constraints

- Only application change: AccountGeometryPanel HorizontalAlignment="Center".
- Retain MaxWidth=820, Padding=40, Spacing=20 and LTR email direction.
- No application launch, native rebuild, artifact rebinding, installs, security changes or main merge.
- Preserve fixed binaries and raw receipts; live allowance 0; C4 NOT_ACCEPTED.
- Ordinary shared XAML also changes: this is an experimental checkpoint, not release acceptance.
- Offline scope covers managed test/Core compilation and existing MSBuild evaluation checks only; no restore or native app compilation.
- Do not substitute new expected hashes for original XAML baselines, or weaken clip/process checks.

## Review Focus

1. Missing, wrong or duplicate target alignment must fail, not normalize away (Task 1 mutation cases).
2. Center on another node must not satisfy the contract (Task 1 wrong-node case).
3. Width, padding, spacing or email direction drift must fail the old baseline (Task 1 mutation cases).
4. Stale copied source must not produce a false pass (Task 1 source/output byte hashes).
5. Login or diagnostic safeguards must remain unchanged (Task 1 Login baseline, geometry regression and diff checks).

Native narrow-window/LTR rendering and incomplete clip-chain uncertainty cannot be certified by these offline tests; retain them as explicit future verification needs.

## Task 1: Single-property candidate and strict source contract

**Files (relative to worktree root):**
- Modify: clients/windows/Talenro.Demo.App/Views/AccountPage.xaml
- Modify/test: clients/windows/Talenro.Demo.Tests/EmailGeometryAdapterContractTests.cs
- Update: docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md
- Create: docs/roadmap/2026-11-11-windows-account-panel-alignment-experiment.md (actual results)

**Interfaces:** Test-only `private static void AssertApprovedPage(XDocument source, string page, string baseline, int expectedNames)` consumes a source document and the unchanged existing baseline constants. It clones the document, checks the exact candidate, removes only allowed attributes and asserts the original SHA256. No application API changes.

- [ ] **1. Confirm starting state.** Read spec, instructions and git status; preserve unrelated changes. Record pre-change fixed EXE/DLL and receipt hashes from the existing inventory. Inspect test assets and available dotnet without installing anything. Test project references Core, not App; App files are copied content or linked pure C#.

- [ ] **2. Add the RED contract.** Refactor `OnlyApprovedNonvisualNamesAreAdded` to call `AssertApprovedPage` and rename it to `OnlyApprovedPageChangesArePresent`. Preserve original Login/Account InlineData hashes. For Account, require exactly one `StackPanel` named `AccountGeometryPanel`, and assert:

```csharp
Assert.Equal("Center", (string?)panel.Attribute("HorizontalAlignment"));
```

Only after that assertion remove that one attribute from the cloned document. Retain the exact existing diagnostic-name count/removal and canonical hash assertion. For Login, do not remove or relax any alignment check. No new production attribute yet.

- [ ] **3. Add mutation tests against the same validator.** Construct an in-memory candidate from the current Account fixture (set only the target attribute to Center), assert it is accepted, then mutate independent clones. `AccountCandidateRejectsUnapprovedMutations` must reject: target alignment removed; Stretch; Left; Center moved to ScrollViewer; duplicate target name; MaxWidth=821; Padding=41; Spacing=21; email FlowDirection=RightToLeft; extra HorizontalAlignment on email. Use `Record.Exception` plus `Assert.NotNull` for rejection; do not swallow exceptions inside the validator. `AccountCandidateValidationDoesNotMutateInput` asserts unchanged `ToString(SaveOptions.DisableFormatting)` before/after acceptance. These checks exercise the validator, not native layout.

- [ ] **4. Run RED once**, from `clients/windows` in ordinary PowerShell:

```powershell
dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometryAdapterContractTests' --logger 'console;verbosity=normal'
```

Expected: Account source contract fails because Center is absent; mutation/clone and Login checks pass. Record exit code and failing assertion. A dependency, compile, timeout or fixture failure is NOT valid RED; stop to diagnose without restore/retry or gate weakening.

- [ ] **5. Apply candidate.** Add exactly `HorizontalAlignment="Center"` to AccountGeometryPanel. No other application changes.

- [ ] **6. Run GREEN once** with the identical command. Require zero failures and exit 0. Compare SHA256 of source `Talenro.Demo.App/Views/AccountPage.xaml` and copied `Talenro.Demo.Tests/bin/Debug/net10.0/AppSources/Views/AccountPage.xaml` (also Login), using Get-FileHash. A mismatch blocks acceptance of the test result; do not use `--no-build` or silently retry.

- [ ] **7. Run geometry regression once:**

```powershell
dotnet test .\Talenro.Demo.Tests\Talenro.Demo.Tests.csproj --no-restore --filter 'FullyQualifiedName~EmailGeometry' --logger 'console;verbosity=normal'
```

Require exit 0 with no failures/skips; record actual count. Existing isolation tests evaluate MSBuild properties/guard targets without native build/restore. Each command has a 120-second observation budget, polling at most 30 seconds per tool call. If unfinished, stop further commands, report its identity/state and do not force-kill or launch a replacement. No automatic retry. Missing assets/tools block execution rather than authorize installation. This is targeted regression, not a claim that all project tests passed.

- [ ] **8. Review and preserve.** Inspect full diff and `git diff --check`; verify only one application attribute changed, old hashes/safety sources/raw evidence unchanged. Obtain the review required by the chosen execution workflow before reporting the candidate ready. Record actual RED/GREEN/regression outcomes, review findings, unresolved native acceptance and exact candidate commit in the execution record/status/recovery. Stage only the explicit changed files, commit to current development branch and push; verify remote SHA equals HEAD. Do not include generated tests output/raw evidence.

## Handoff and self-review

Plan self-review: every design constraint maps to Task 1 or the later-native exclusion;
the sole validator signature is shared by source/mutation tests. Original baseline
hashes remain authoritative after narrowly asserted normalization. Native rendering
is deliberately untested here and remains a blocker to claiming a fix.

Recommended execution: implement inline in this session, followed by a fresh final
review, since this is one tightly coupled test/code task. Alternative: delegated
implementation plus independent review. User confirmation of plan/execution method
is pending. No steps have run; approval of this plan does not authorize native work.
