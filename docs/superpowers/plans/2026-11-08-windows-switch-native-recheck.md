# Windows Switch Native Recheck Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this verification plan task-by-task.

**Goal:** Verify the1ac1c19e switch-language correction in fresh ordinary and diagnostic artifacts, without claiming full Stage B acceptance.
**Architecture:** Retain sequential execution by the current implementer. Build each isolated variant once, inspect metadata/hash identity, then observe four bounded sessions using existing simulated UI actions.
**Tech Stack:** Existing .NET10/WinUI2.6.5, PowerShell, installed native UI observation tools; no new dependencies.
**Spec:** docs/superpowers/specs/2026-11-08-windows-layout-test-mode-design.md. Correction/evidence: docs/roadmap/2026-11-08-windows-layout-switch-fix-execution.md.

Status: proposed after1ac1c19e, awaiting approval. Preparation authorizes no runtime action. Existing226/226 suite is historical and will not be repeated in this scope.

## Global constraints

- Maximum2 build invocations and4 app process invocations. No retries, restores, downloads, test-suite calls or Release build. Failed starts consume a slot.
- Ordinary Windows token required before each build/launch; stop on administrator token or unclear ownership. Broader tool execution permission does not permit OS elevation.
- No system/display/text-scale/locale/security changes, no resize, no real network/account action, no source changes or main merge. Existing allowances are not reused.
- Build ceiling10minutes each; startup30seconds, interactive observation10minutes each, normal close30seconds. Poll at most30seconds per wait. Unexpected error, mixed language, inaccessible required control or timeout stops dependent work. No force-kill or automatic relaunch.
- Offline notice, account identity, state handlers and enablement remain unchanged. Core behavior is not repaired during verification.

## Review focus

1. Old binaries mistaken for fixed outputs: Task1 records fresh DLL/apphost hashes and paths, not old Stage A values.
2. Diagnostic payload leaks into ordinary mode: Task1 metadata script and Task2 ordinary/no-flag sessions.
3. Off is translated but On is not: Task2 checks Off→On→Off in each session.
4. Refresh/reset overwrites selected language: Task2 navigation and logout/re-entry checks both diagnostic languages.
5. Auto node cannot establish a country lock: Task2 explicitly selects jp-1 while disconnected before enabling lock, and checks JP identity.

## Task1 — Build and identify artifacts

Working directory: D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator/clients/windows.
Read AGENTS.md, correction ledger and existing UI tool instructions. Confirm clean expected development branch, source1ac1c19e plus documentation-only commits, ordinary token, no demo processes/windows. Preserve all old receipts.

- [ ] Confirm cached assets exist for both App configurations. Set only process-local DOTNET_CLI_HOME to .dotnet-home, NUGET_PACKAGES to .nuget/packages and existing telemetry/first-time opt-outs. Missing prerequisites block execution, not a restore allowance.
- [ ] Build ordinary once: `dotnet build Talenro.Demo.App/Talenro.Demo.App.csproj --no-restore -c Debug -p:EnableLayoutDiagnostics=false`. Use C:/Program Files/dotnet/dotnet.exe. Expect exit0,0warnings/errors.
- [ ] Build diagnostic once: same command with `-p:EnableLayoutDiagnostics=true`. Expect exit0,0warnings/errors. Do not proceed on unexpected warnings/errors.
- [ ] Under an existing PowerShell7 runtime, run tests/verify-layout-artifacts.ps1 with -OrdinaryAssembly and -DiagnosticAssembly pointing to the respective managed DLLs below. It reads PE metadata without executing App code; expect its static isolation PASS. Missing PowerShell7 blocks, no install.
- [ ] Record SHA256 for App.dll and App.exe in both directories; bind each subsequent row to exact path/hash. Ordinary directory: Talenro.Demo.App/bin/Debug/net10.0-windows10.0.19041.0/win-x64. Diagnostic: Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64. Both executable names Talenro.Demo.App.exe. Build success plus isolation/hash evidence, not apphost hash alone, identifies fresh outputs.

## Task2 — Four sessions, in this order

| ID | Variant | Arguments | Expected language/title |
| --- | --- | --- | --- |
| S1 | Ordinary | none | Existing system-selected language; no LAYOUT TEST marker |
| S2 | Diagnostic | --layout-test --layout-language=en-US --layout-profile=baseline | English; LAYOUT TEST en-US/baseline |
| S3 | Diagnostic | --layout-test --layout-language=zh-Hans --layout-profile=baseline | Chinese; LAYOUT TEST zh-Hans/baseline |
| S4 | Diagnostic | none | Ordinary language/state after explicit-language sessions; no marker |

- [ ] Start each resolved executable directly and capture process handle/PID/start time. One process at a time, visible interactive window, never dotnet run. Observe Login/offline notice then Enter demo. Do not submit real credentials.
- [ ] On Connection observe initial Off/关. Navigate Nodes, select existing jp-1, Apply while disconnected. Confirm Connection remains disconnected. Toggle CountryLock on: require On/开 and locked-country JP; then off: require Off/关 and localized lock-off status. Verify by visible content as well as available UI text, not title alone. For S1/S4 derive expected On/Off or 开/关 from their observed ordinary language; if another language appears, record it without changing OS settings or inventing translations.
- [ ] S2/S3 additionally keep lock on, navigate Account then back Connection: state and selected-language On/开 retained. Toggle off, navigate Account, scroll to Exit demo/reset, logout and re-enter: require disconnected/off state and selected language retained. Verify relevant Tab focus remains visible and offline notice persists; no broader keyboard/accessibility claim.
- [ ] Close normally and record actual exit0 before next row. On failure stop all remaining rows; preserve evidence, do not patch or consume an extra launch. No exact-size claims from self-reported title; all earlier measurement gaps remain open.

## Task3 — Save evidence and hand off

Modify docs/roadmap/2026-11-08-windows-layout-switch-fix-execution.md (append fresh results/hashes), docs/roadmap/current-status.md, docs/runbooks/repository-recovery.md and clients/windows/tests/manual-checklist.md. No production file changes planned.

- [ ] Record build results and each S row as PASS/FAIL/UNVERIFIED/NOT_RUN with identity, arguments, language, state transitions, exit and limits. Check no owned process remains or report the unresolved PID.
- [ ] Close the specific V3 switch-language finding only if S2/S3 On/Off/refresh/reset observations pass; ordinary regressions require S1/S4. Do not mark broader V3 baseline or Stage B complete: its interrupted English dialog route, V4–V7 profiles and other unobserved requirements still need separate validation. S4 provides narrow no-activation evidence, not all old V8 checks.
- [ ] Commit only necessary portable records, push existing codex/c12-b01-task9-coordinator branch, verify remote SHA equals HEAD and clean status. Exclude binaries/caches/raw logs/private screenshots. Unused slots do not carry forward after handoff.

## Self-review and approval

Scope matches the minimal correction: shared resources require ordinary and diagnostic builds; two explicit languages cover four state strings; no-flag sessions check default-path behavior. Task1 identities feed Task2, whose observations feed Task3. No implementation/TDD/reviewer cycle added to this verification-only plan. Exact sizes, text enlargement, reader speech, contrast, resource fallback and broader NOT_ACCEPTED gates remain unresolved. Please confirm2 builds plus4 launches before execution.
