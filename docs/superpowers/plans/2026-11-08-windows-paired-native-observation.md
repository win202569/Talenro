# Windows Paired Native Observation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans sequentially. Verification only; no application or recorder implementation changes.

**Goal:** Compare native and managed observations of one retained process handle during one short search-and-close attempt.

**Architecture:** Retain the launched Process and its SafeHandle once, while alive. Record native then managed snapshots with a shared correlation ID and sequence number. Never replace conflicting evidence with another object's properties.

**Tech Stack:** Existing PowerShell7, reviewed native/managed modules, supported Windows UI controls; no installs.

**Spec:** docs/superpowers/specs/2026-11-08-windows-native-process-observation-design.md; prior route and limits in 2026-11-08-windows-same-object-live-observation.md.

Status: approved after0fbb76e0; one attempt consumed.73 pairs recorded, shutdown
timed out without signal; see paired-native-observation-execution ledger. No
remaining runtime allowance. Approval request below is historical.

## Global constraints

- Exactly one launch attempt including failed start; no replacement/retry. No builds/restores/source edits/settings changes/debugger/dump/kill/network account actions.
- Ordinary-user PowerShell token; current isolated development branch. Keep all NOT_ACCEPTED gates.
- Startup30s, route2minutes from launch, normal-close criterion30s, absolute observation cap360s from launch. No wait longer than1000ms; UI only while observer remains active.
- Native read scope is only the retained demo process handle: zero-time wait, exit code and process times through the reviewed helper. No arbitrary PID/current-process probe.
- Caller owns original Process/handle. No Refresh/reopen/duplicate/manual CloseHandle or early disposal. Dispose original Process once in finally after observation, never separately dispose its borrowed SafeHandle.
- No stdout/stderr redirection. App disappearance or launcher exit0 is not shutdown acceptance. Keep raw receipts/session screenshots out of Git.

## Review focus

1. Wrong artifacts/identity: bind files and creation time before UI actions.
2. Cached managed properties: preserve native and managed receipts separately.
3. Non-atomic transitions: retain intervals and fixed native-before-managed order.
4. Failed APIs/invalid handle: stop route, never replace handle or manufacture0.
5. Signal/property disagreement: preserve it; do not broaden acceptance.

## Task1 — Preflight and session preparation

- [ ] Read AGENTS, this plan, both module specs and execution ledgers, supported UI instructions. Record explicit approval and budget0/1 in a new paired-observation execution ledger.
- [ ] Require clean codex/c12-b01-task9-coordinator worktree, ordinary token, usable UI controls and no existing demo process/window; failed or conflicting queries block without launch.
- [ ] Bind original recorder to68ec4a7c and native module/helper to3b5e38dc by Git blob equality; tests at20e45326 document78/20/32/11 historical passes, not live acceptance.
- [ ] Require existing layout-diagnostics EXE under clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64; EXE SHA2564180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 and DLL SHA2567F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB. Missing/mismatch blocks; no rebuild.
- [ ] Use a fresh PowerShell7 observation host, import both modules once. Check expected exports. Compile/load native helper explicitly before launch using its private loader (no Capture); verify type identity/signatures using existing compile-check knowledge. Failure blocks without consuming launch budget. Do not run compile tests in this same host, force-reimport or adopt a preloaded type.
- [ ] Record HEAD/file hashes, runtime/OS versions, host identity and exact one-session command. Use session-local command only, no persistent replacement launcher implementation.

## Task2 — Single retained-handle paired observation

- [ ] Launch exact EXE using UseShellExecute=false, own working directory and --layout-test --layout-language=zh-Hans --layout-profile=long-country. Count attempt1/1 immediately; retain original Process, SafeHandle, PID, StartTime UTC, executable path, launcher identity and monotonic clock. Acquisition failure stops route without reopening or retry.
- [ ] Capture a native snapshot with Get-TalenroNativeProcessObservation -ProcessHandle $retainedHandle -CorrelationId $launchId, then managed snapshot with Get-TalenroProcessObservation -ProcessObject $ownedProcess. Envelope both with identical launch ID and sequence number, recording pair start/end times. Serialize depth at least12; do not format away statuses/nulls/UInt32 halves.
- [ ] Before UI input require valid managed identity and native creation time. Reconstruct creation FILETIME losslessly as UInt64(high)*4294967296+low and compare to retained StartTime.ToUniversalTime().ToFileTimeUtc(). Never convert raw exit time to an accepted date. Identity mismatch/read failure stops route; no alternate identity lookup substitutes for this check.
- [ ] Repeat pairs approximately every5s while observer active; fixed native-first order is deliberate and non-atomic. Preserve original helper timestamps, snapshot errors and individual statuses. Recheck creation identity for every successful native time read.
- [ ] Use bounded original Process.WaitForExit(min(1000,remainingMilliseconds)) between pairs. True outer wait triggers one final pair plus separately recorded trigger, then ends. At360s capture final pair and TIMEOUT. Check remaining time before UI actions; no further input once observer has ended. Observation overhead can affect timing; do not infer causation from one run.
- [ ] Only after initial identity pair passes: Enter demo -> Nodes -> screenshot/text -> explicit SearchBox click -> confirm caret -> type jp-1 once -> stable value/filter JP group and jp-1 row. No selection/apply/connect/dialog or alternate input injection. Route error stops input; request normal close once if the identity-bound window remains available.
- [ ] Record UTC and monotonic close-request times before one Alt+F4. Do not close again. Missing shutdown evidence at30s is failure/unresolved; continue only read-only observation to cap, never retry/kill.
- [ ] Invalid/closed handle, identity mismatch, snapshot/helper error or invalid receipt stops paired reads and UI route. Request one normal close only if not yet requested and safely identity-bound; record ABORT, release caller resources, report any remaining app without forced termination. Expected managed exit-property errors while running do not abort; native API failures are preserved and abort the route rather than escalating rights.
- [ ] In finally dispose original Process only after no reads remain. Independently inspect remaining process/window state afterward; absence after disposal does not prove disposal caused exit.

## Task3 — Classification and portable handoff

- [ ] Report search and shutdown separately. Scoped shutdown success requires verified identity, actual signal, native exit-code API success/output0 and managed successful numeric0 within30s, no contradictory final signal evidence. Nonzero fails; missing/error/implausible time or disagreement remains unresolved. One success cannot clear historical intermittent failure or broader gates.
- [ ] If native wait times out while native code0/raw exit time0 appears, record the disagreement at the OS-return boundary, not a proven OS bug. If native/managed differ, retain both intervals and consider sequential transition/caching; no immediate app/runtime fix.
- [ ] Save portable field transitions, API return/error distinctions, identity comparison, close/timeout timing, budget1/1 or blocked0/1 and limitations. Do not include raw process lists, private exception payloads or screenshots in Git.
- [ ] Update execution/current-status/recovery; diff-check and explicit-doc commit/push existing development branch; verify remote SHA=HEAD. No launch allowance carries forward.

## Self-review and handoff

Preflight feeds identity into the paired loop; the loop feeds conservative classification.
Native reads and managed properties remain separate. No new module behavior is
required; compilation happens before app launch. Identity checks and stop conditions
are explicit. This plan authorizes nothing until approved: request one launch,
same-handle native observation and the short offline UI route only.
