# Ordinary Exact Command-Line Binding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans in this session. Follow checkbox steps; no implementation until this checklist is confirmed. No live launch is included.

**Goal:** Replace the one incorrect expected serialization with one exact observed representation and bind the launcher component before acquisition.

**Architecture:** Amend only the ordinary wrapper and its focused synthetic tests. Separate read-only launcher evidence collection from validation, capture the validated command in backend state, and leave the shared identity/session modules unchanged.

**Tech Stack:** Windows PowerShell7.6.9, existing AST/memory fixture harness and Git; no dependencies/installations.

**Spec:** [Approved amendment](../specs/2026-11-14-windows-ordinary-exact-command-line-binding-design.md).

Status2026-11-14: design approved; this checklist awaits confirmation. No code/test changes or execution in this preparation. Preserve existing inline implementation method plus final independent scoped review.

## Global Constraints

- Worktree c12-b01-task9-coordinator, branch codex/c12-b01-task9-coordinator; PowerShell only.
- ExpectedCommandLine = U+0022 + absolute fixed EXE + U+0022 + U+0020. No Trim, case folding, regex tolerance, old-or-new acceptance, actual-CIM-derived expectation or retry.
- CommandLineProfile=`shell-quoted-exe-one-u0020-v1`; Arguments=''; ArgumentList absent. Shared StringComparison.Ordinal unchanged.
- Launcher command is `Microsoft.PowerShell.Management\Start-Process` (one literal backslash in PowerShell, not an escaped pair); Cmdlet; implementing type Microsoft.PowerShell.Commands.StartProcessCommand.
- Launcher PowerShell7.6.9/Core, assembly Microsoft.PowerShell.Commands.Management.dll version7.6.0.500, file7.6.9.500, SHA2562181516B3A55866A2263DDAB22A7547FC815240F474199088B8ABA030B92264A.
- All six artifact hashes, source5ee8a339b574171fc61f3cf0c0c636bd86421397, source/path/token/branch/clean-tree/inventory checks remain. Missing or updated runtime fails closed; no automatic download or alternative host.
- Default invocation returns NotStarted before imports/evidence collection. Public switch remains RunApproved only. Context remains exactly RunId,ExpectedPath,ExpectedCommandLine.
- No product/shared module/old entry changes, App restore/build/launch, native compilation or settings changes. Four existing Minors remain deferred; do not expand scope to fix them.
- Existing receipt fixtures may create tiny files in the ignored test workspace; no candidate/original receipt writes. Raw evidence/binaries/cache never committed.
- Future separately approved live limits remain one launch/zero retries/360000ms/max sleep2000ms/normal close by300s. This plan grants no live allowance.

## Review Focus

1. A whitespace near-match accidentally passes: direct real-guard positive/negative matrix (Task1 A).
2. Stale/ambiguous/shadowed launcher discovery appears valid: evidence faults, independent literal expectations and command capture (Task1 B/C).
3. Default execution or failed binding still acquires resources: effect counters and no-acquisition assertions (Task1 B/C).
4. Unverified or changed evidence enters the manifest: backend state/receipt ordering and evidence-copy tests (Task1 C/D).
5. Correct new path breaks retained-handle/cleanup/deadline semantics: existing receipt/identity suites preserved (Task1 E/Task2).

## Task1: Exact policy, verified launcher capture and synthetic proof

**Modify:** clients/windows/tests/observe-ordinary-account-center.ps1 and clients/windows/tests/observe-ordinary-account-center.tests.ps1 only.
**Read unchanged:** the five shared dependencies and observe-c4.ps1/.tests.ps1 under clients/windows/tests.
**Record:** docs/roadmap/2026-11-14-windows-ordinary-exact-binding-execution.md; update current-status.md, repository-recovery.md and this checklist at handoff.

### Interfaces and data flow

- Get-OrdinaryPolicy([string]$Root): add CommandLineProfile and LauncherExpected hashtable; exact constants above. Change only the terminal character expectation, not the executable path/arguments.
- Get-OrdinaryLauncherEvidence(): read-only collection returns a hashtable with Command (the single resolved CommandInfo), CommandName, CommandType, ImplementingType, PowerShellVersion, Edition, AssemblyPath, AssemblyName, AssemblyVersion, FileVersion, Sha256. Resolve with module-qualified Get-Command, require exactly one result; read the loaded implementing assembly, verify leaf/ancestor non-reparse paths before hashing. Missing/exceptional values throw, never return partial evidence. No command invocation here.
- Assert-OrdinaryLauncherBinding($Policy,$Evidence): validate every expected field, nonempty absolute assembly path, expected filename, and required Command; return a new binding object containing Command and a scalar-only Evidence copy (no CommandInfo serialization). No public injectable parameter.
- Assert-OrdinaryPreflight($Policy): retain all existing guards, collect/validate launcher evidence and return exactly that binding only after ALL guards pass. Initial entry caller assigns it to $null. Update synthetic callers that previously assumed no output.
- New-OrdinaryBackend($Policy): private mutable state starts Binding=$null. Its Preflight callback clears stale binding then assigns the full preflight result. AcquireWriter and AcquireProcess throw if binding absent. AcquireProcess invokes state.Binding.Command, not another Get-Command; preserve stopwatch/start ordering and original Process return. Capture function references as baseline does, but never resolve the launch command again after preflight.
- New-OrdinaryReceiptWriter($Policy,$Context,$Binding): validate required binding/evidence before filesystem actions; add CommandLineProfile, ExpectedCommandLine and Launcher to SchemaVersion1 manifest. Launcher contains only CommandName, CommandType, ImplementingType, PowerShellVersion, Edition, AssemblyName, AssemblyVersion, FileVersion, Sha256. Preserve existing fields, resource ownership and aggregation. No AssemblyPath or CommandInfo leaked into portable manifest. Direct test factory calls must now supply a valid synthetic binding.

### Test-first steps

- [ ] A. Write exact-string tests before production edits. Independently construct the full expected fixture literal for D:\fixture with spaces plus the fixed candidate path, closing quote and `[char]32`. Assert profile and expected equality. Feed the real unchanged identity guard valid native/CIM timestamps and equal PID/path; only that exact string succeeds. Table rejects old quoted-only, two spaces, tab, CR, LF, NBSP, appended diagnostic/other text and changed path/case. Assert rejection cause is identity mismatch, not missing fixture data. Change the old trailing-space rejection test to reject an EXTRA space; preserve old form as a new negative case.
- [ ] B. Extend inert-default sentinels to reject launcher discovery, hash/file probes, import and acquisition. Add independent LauncherExpected literals. Model one valid complete evidence fixture and table faults for zero/multiple command results, function/alias type, wrong implementing type, PS version/edition, assembly filename/version/file version/hash, missing field, unreadable file and every reparse ancestor. Collector side effects are mocked at Get-Command/file metadata/hash boundaries; validator itself stays real. Each invalid fixture must fail at its named guard with writer/process counters0.
- [ ] C. Add complete backend wiring tests: valid preflight result captured, launcher callback called once with fixed path/working directory/PassThru and no ArgumentList. Place an unrelated unqualified Start-Process trap and prove it is not called. Rebind that unqualified name after preflight and prove the already captured synthetic command is used. A failed second preflight clears previous binding; writer/process callbacks reject absent binding. No OS launch or native helper allowed. Test the real collector separately with mocked discovery evidence; do not confuse a mocked valid descriptor with proof that real cmdlet-type checking works.
- [ ] D. Update direct factory fixtures for binding. Assert exact manifest fields/count, one-space expectation, profile and every launcher literal; no CommandInfo/local assembly path. A missing/mismatched binding must fail before destination/open. Mutate source evidence after binding creation and prove recorded scalar copy does not change. Keep real CreateNew no-overwrite and all four combined-failure tests intact; update signatures, never remove assertions to pass.
- [ ] E. Run focused test once as RED against old production entry. Expected explicit new policy/binding assertion failure, exit1, NOT parse/module/setup errors. Preserve full output in this plan's ignored workspace. If failure is unrelated, stop and report rather than continue a blind fix loop.
- [ ] Implement only the wrapper changes at the Interfaces above. Preserve all shared comparisons and checks; ensure evidence includes no ambient real calls when tests replace boundary readers. Keep manifest binding separate from strict Context. Do not launch a probe to learn formatting.
- [ ] Run the focused test once as GREEN; expected exit0 with PASS count, no real process/CIM/native/sleep. Record actual count rather than expecting598 after new cases. Unexpected failure: preserve evidence, stop and request bounded continuation; no silent retry.

Future commands from worktree root, existing PowerShell7 host only (not to execute during planning):

```powershell
$PSVersionTable.PSVersion
& (Join-Path $PSHOME 'pwsh.exe') -NoProfile -File '.\clients\windows\tests\observe-ordinary-account-center.tests.ps1'
"Synthetic exit code: $LASTEXITCODE"
```

- [ ] Review explicit changed files and git diff --check. Record RED/GREEN evidence and any deviations. Save an ordinary checkpoint of only scoped source/test/docs; label unfinished if review or regression remains. Do not stage ignored fixtures.

## Task2: Regressions, independent review and recovery handoff

**Consumes:** Task1 verified policy/binding/manifest plus focused GREEN results. **Produces:** regression evidence, review disposition and clean pushed checkpoint; no live execution.

- [ ] Inspect then run each existing suite once, fresh PowerShell host, same command pattern: observe-c4.tests.ps1, c5-observation-session.tests.ps1, observation-session-guard.tests.ps1, process-observation.tests.ps1, native-process-observation.tests.ps1. Expect exit0/PASS; stop at first unexpected failure. No loader/compile/dotnet/App suites. Do not rerun focused GREEN merely for bookkeeping.
- [ ] Obtain one read-only independent review of the scoped diff/spec, especially the five Review Focus cases. Include actual counts, any skipped coverage, unchanged modules and four deferred Minors. No review-time test/live launches. Important findings remain blockers; remediation/retesting beyond the bounded runs needs an explicit continuation, not automatic acceptance.
- [ ] Confirm no product/shared/old-entry diff, no tolerant comparison and no fallback. Update execution/status/recovery documents; explicitly stage only scoped files, commit/push existing branch and verify remote SHA equals local plus clean worktree. Do not merge main.
- [ ] Handoff: no active launch allowance. A new live request must still name one launch/zero retries/max6min, ordinary host token with authorized CIM access, same six artifact hashes and current preflight. Changed runtime/candidate blocks; no substitution. Preserve prior identity-failed outcome and all runtime/viewport/security acceptance limits.

## Self-review and approval boundary

Design coverage mapped to A–D (serialization/host/evidence), E (test-first), Task2
(regression/review/recovery). Returned preflight binding and direct writer argument
changes are explicit to prevent accidentally discarding validation or breaking
existing fixtures. Shared Context is unchanged. The module separator is one real
backslash, not Markdown/JSON escaping. No alternative serialization is accepted.
All boxes remain unchecked: user has approved design, not implementation of this
new checklist. Next requested authority is scoped implementation and synthetic
tests only, never a live run.
