# Windows C3 Exit Observation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans sequentially. Verification only; no product implementation or subagent cycle.

**Goal:** Reproduce the C3 route once and distinguish application exit evidence from launcher timeout and process-list disappearance.

**Architecture:** Reuse the existing diagnostic artifact and UI route. The launching PowerShell retains the original Process object and explicitly acquires its handle immediately after start; short bounded waits produce exit receipts. Separate read-only process snapshots supplement, but never replace, that receipt. This changes observation, not application code, and is not a controlled root-cause comparison with the old launcher.

**Tech Stack:** PowerShell, System.Diagnostics.Process, CIM and supported native UI tools; installed tools only.

**Spec:** docs/superpowers/specs/2026-11-08-windows-layout-test-mode-design.md; incident constraints/evidence in docs/roadmap/2026-11-08-windows-layout-stage-b-continuation-execution.md (read-only diagnosis appendix).

Status: approved afterfeb90856; one attempt consumed. First UI click failed with unavailable input geometry, so the route stopped at Login and closed with actual exit0. Original C3 route remains unverified. See docs/roadmap/2026-11-08-windows-c3-exit-observation-execution.md; no launch allowance remains. Historical approval requests below are superseded.

## Global constraints

- One launch attempt total, including failed startup; no replacement or retry. Zero builds, restores, suite runs, installs, product-source changes or settings changes.
- Ordinary Windows token; PowerShell only. No real account/network operations, force-kill, debugger attach, dumps, environment/security changes or main merge.
- Startup30s; route maximum5minutes from launch. Request normal close at route completion or deadline. Observe up to30s from close request. Launcher absolute lifetime maximum360s from launch, without killing the application.
- WaitForExit calls at most1000ms each; no630-second single wait. Model-facing polls at most5s while working, communicate at least every60s. Timeout is not a successful app exit.
- Close only the identity-bound demo window. If targeting/ownership becomes uncertain, stop inputs and report; do not close another window. No second launch under any result.
- Task11/12 and44/34/6/217 stay NOT_ACCEPTED. Historical226/226 is not rerun. C4-C6 remain outside scope.

## Files and evidence interface

- Read: existing App.exe/App.dll and incident ledger; no edits to app or build outputs.
- Create at execution: docs/roadmap/2026-11-08-windows-c3-exit-observation-execution.md.
- Update at handoff: docs/roadmap/current-status.md and docs/runbooks/repository-recovery.md.
- Session-local receipt only: PID, start timestamp, resolved executable path, launch arguments, handle-acquisition success, launcher PID, close-request timestamp, timestamped wait results, actual app ExitCode if available, observation errors and timeout reason. Do not commit raw traces/screenshots or handle values.
- Separate snapshots record query timestamp/context, PID/start/path/parent/thread count where available, or explicit query error/absence. Do not silently treat access errors as absence or reuse an old PID identity.

## Review focus

1. Wrong artifact or existing process: Task1 blocks before launch; no rebuild allowance.
2. Launcher exit0 mistaken for app exit0: Task2 uses explicit application receipt only.
3. Process API disagreement/reused PID: Task2 preserves both observations and creation time.
4. Native window gone but app not exited: Task2 starts a separate30s close window.
5. Missing UI facility/expired budget: Task1 checks capability; Task2 never replaces a failed attempt.

## Task1 — Preflight, no launch until all checks pass

Consumes: approved plan and original incident identity. Produces: verified artifact/UI readiness or blocked ledger.

- [ ] Read AGENTS.md, spec, incident ledger and UI-control instructions. Confirm explicit approval of this one-launch scope.
- [ ] Confirm worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator, branch codex/c12-b01-task9-coordinator, clean state and AdministratorToken=False. Confirm supported UI inventory works before starting.
- [ ] Check no demo process/window via both permitted broader CIM/name queries and UI inventory. Any access error/disagreement blocks launch; record it, do not bypass.
- [ ] Resolve diagnostic directory clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64. Require DLL SHA2567F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB and EXE SHA2564180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95. Missing/mismatch blocks, without rebuild.
- [ ] Write execution ledger with preflight results, current HEAD and budget0/1. Prepare launcher command for explicit one-shot execution; no persistent helper or product code changes required.

## Task2 — One observation session

Consumes: Task1 identities. Produces: actual app exit receipt or explicit unresolved result, not inferred success.

- [ ] Start exact exe with UseShellExecute=False, its own directory as WorkingDirectory and arguments `--layout-test --layout-language=zh-Hans --layout-profile=long-country`. Keep the original Process object alive; immediately acquire its Handle, record identity and start a monotonic Stopwatch. Handle failure consumes the attempt: stop route, request normal close if safely targetable, mark observation failure; no replacement.
- [ ] Launcher repeatedly calls WaitForExit(1000) up to360s total. On true, read ExitCode from that same retained object and emit an explicit APP_EXIT receipt with timestamp. On false, continue bounded observation. On exceptions, emit OBSERVATION_ERROR; never emit exit0 by default. At deadline emit OBSERVATION_TIMEOUT and dispose the observation object, not the app. Shell success has no application-success meaning.
- [ ] Observe the original route within5minutes: Enter demo; Nodes; search jp-1; select/apply disconnected; simulate Connect/success; Nodes; clear search; select de-1/Apply; Cancel; Back; reopen Nodes. Record final jp-1 selection and Nodes page. Do not add JP search, reset/logout or other profile checks. If a step differs or errors, stop the route and request normal close rather than repairing it.
- [ ] Immediately before Alt+F4, record close-request timestamp and confirm target identity. Close once normally. Poll launcher and independently inspect CIM/name-based state immediately, around5s and at30s if no APP_EXIT receipt. Record actual elapsed times rather than pretending delayed queries occurred on schedule. Window disappearance alone is not exit evidence.
- [ ] If APP_EXIT arrives, record actual code. If no receipt by30s after close, mark shutdown unverified/timeout and stop all UI work. Keep only bounded read-only observation until launcher deadline; no retries, additional close commands, attachment or forced termination. Report any process still present when the launcher ends. If deadline coincides with process disappearance, preserve the timing without inferring causality.

## Task3 — Classify and preserve

Consumes: Task2 receipts and snapshots. Produces: portable result, no broadened acceptance.

- [ ] Classify: APP_EXIT0 within30s after close = scoped normal-exit observation; nonzero = failure; no actual code/API errors/late exit = unresolved or out-of-bound result. One successful repetition does not prove the original cause fixed or close broader acceptance.
- [ ] Preserve differences from C3, including changed observer handle acquisition/wait strategy. If reproduced, recommend a separately approved diagnostic capture only; if not, record intermittent/unreproduced and propose the next comparison, not an automatic rerun.
- [ ] Update execution/current-status/recovery with budget1/1 or blocked0/1, actual outcomes and remaining gates. Run git diff --check and inspect explicit documentation diff; commit/push current development branch, confirm remote SHA=HEAD and clean state. Do not commit raw receipts or binary artifacts.

## Self-review / handoff

No application fix is proposed without a root cause. The independently queried process list supplements the original handle; it does not guarantee root-cause identification. Bounds are shorter and close-specific, while the absolute launcher cap prevents indefinite observation. The one-attempt ceiling prevents accidental C4 continuation. This is the existing sequential verification method, not new parallel implementation. Please approve this one-launch observation scope before execution.
