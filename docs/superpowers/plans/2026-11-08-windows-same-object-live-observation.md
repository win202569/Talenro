# Windows Same-Object Live Observation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans sequentially. This is verification-only, with no implementation or new review cycle.

**Goal:** Obtain wait/exit/property-error observations from the original launched Process object during one short search-and-close route.

**Architecture:** Reuse the unchanged diagnostic app and the reviewed observation module. The launcher retains its Process object and handle, periodically calls the module on that same object, and emits structured session receipts. External process lists remain supplementary evidence.

**Tech Stack:** Existing PowerShell7, System.Diagnostics.Process, supported Windows UI controls; no new dependencies.

**Spec:** docs/superpowers/specs/2026-11-08-windows-process-observation-design.md. Incident and short route: docs/roadmap/2026-11-08-windows-c3-exit-observation-execution.md.

Status: approved after4c5a323d, one launch consumed. Input passed; same-object false wait/true HasExited/exit0/1601 exit time persisted through360s timeout. See execution ledger; no runtime allowance remains. Historical approval request below is superseded.

## Global constraints

- One application launch attempt total, including failed starts; no retry or replacement. Zero builds/restores/app-suite runs/source edits/installs/settings changes. No debugger, dump, force-kill or main merge.
- Ordinary-user token; PowerShell only. No real accounts/network actions. Use only the existing offline demo and fixed arguments.
- Startup30s; input route2minutes from launch; normal-close deadline30s. Launcher absolute observation cap360s from launch, unchanged from the previous short route. No individual wait longer than1000ms; model-facing polls at most5s. Stop UI work on errors; no repeated close or input retries.
- Windows snapshots must include screenshot and text before each action. Explicitly click the SearchBox, verify focus/caret and only then type. Do not trust accessibility focus alone.
- Expected ExitCode/ExitTime getter errors while running are observations, not permission to repair code. Failed identity/invalid schema/module invocation errors stop the UI route and trigger safe normal close, not a replacement launch.
- All Task11/12 and44/34/6/217 NOT_ACCEPTED remain. Do not count successful receipt collection as an application fix.

## Review focus

1. Wrong binary/module/object: Task1 binds artifacts; Task2 passes only the original object.
2. Missing/failed exit values becoming0: Task2 retains explicit field Status and numeric type.
3. Disagreement hidden by different contexts: Task2 collects all fields on the same object; no external value substitution.
4. Snapshot delay or expected getter errors: Task2 preserves intervals and errors without claiming atomicity.
5. False acceptance on timeout/window disappearance: Task3 requires actual signal and successful exit-code read.

## Task1 — Preflight

Consumes approved plan; produces identity/readiness record in
docs/roadmap/2026-11-08-windows-same-object-live-observation-execution.md.

- [ ] Read AGENTS, plan/spec, incident and observation-module ledger, and supported UI instructions. Confirm explicit new approval, current branch codex/c12-b01-task9-coordinator and clean worktree.
- [ ] Verify ordinary token and no demo process using permitted broader CIM/name checks and UI inventory; query failure/disagreement blocks. UI support must work before launch.
- [ ] Use clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64. Require DLL SHA2567F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB and EXE SHA2564180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95. Missing/mismatch blocks without rebuild.
- [ ] Confirm clients/windows/tests/process-observation.psm1 matches committed68ec4a7c content (compare Git blob/hash, no edits), import it and verify only the two expected exports. Record current HEAD, PowerShell/runtime versions and module identity. No synthetic rerun is needed;32/32 remains historical.
- [ ] Initialize ledger budget0/1 and record the planned launch command. Do not create persistent replacement launcher code; use a single session command invoking the unchanged module.

## Task2 — Single session

Consumes Task1 identities; produces session-local structured receipts and portable summaries.

- [ ] Launch exact exe with UseShellExecute=False, its own directory and `--layout-test --layout-language=zh-Hans --layout-profile=long-country`. Immediately retain original Handle, PID/start/path/launcher identity and monotonic elapsed clock. Handle acquisition failure consumes the attempt and stops the route.
- [ ] Call `Get-TalenroProcessObservation -ProcessObject $ownedProcess` on the original object immediately, then approximately every5s. Emit each record as JSON with sufficient depth for Fields. Record actual timestamps; never claim exact scheduled timing. Keep same object/reference throughout; no Get-Process substitution, Refresh or PID reopening in the launcher.
- [ ] Between snapshot intervals use original-object WaitForExit(1000), bounded by remaining absolute time. On a true wait, immediately capture a final module snapshot and record that triggering wait result separately, then end observation. If outer wait is false but snapshot reports HasExited=true or an exit code, retain disagreement and continue only to the approved cap. At cap capture a final snapshot and explicit TIMEOUT; Dispose only the caller-owned observation object, never kill the app. Outer script exit0 is not APP_EXIT0.
- [ ] Enter demo -> Nodes -> fresh screenshot/text -> click inside SearchBox -> refresh to confirm editable focus/caret -> type jp-1 once -> refresh until value and list are stable within route deadline. Require visible/text value jp-1 and only JP group/jp-1 row. Do not select/apply, connect or open dialogs. If input fails, stop and close normally; no alternate injection.
- [ ] Record UTC timestamp immediately before one Alt+F4 on the identity-bound window. Read receipts and supplementary process/window snapshots. By30s after close, classify lack of signal/actual exit code as unresolved and stop UI work. Continue only bounded read-only observation to the absolute cap; no close retry, attach or termination.
- [ ] After observer ends, independently inspect remaining process/window identities. Preserve presence, absence or query errors as observed; disappearance after Dispose does not establish causality.

## Task3 — Interpret and preserve

- [ ] Report input result separately from shutdown. Scoped normal shutdown requires actual wait signal and same-object ExitCode Status=ok numeric0 within30s of close, with verified identity. Nonzero is failure; missing/error/invalid values are unknown; disagreement is retained. If outer wait and final snapshot disagree, mark unresolved rather than select the favorable value.
- [ ] Summarize each field's transitions/exceptions and the triggering wait, with timestamps. Do not conflate a method-wrapper exception type with a proven underlying Windows error. A single success does not close the historical intermittent symptom.
- [ ] Update execution/current-status/recovery with launch1/1 or blocked0/1, limitations and next diagnostic choice. Keep raw receipts/screenshots local to the session, not Git. No secrets/private error messages or unrelated process data.
- [ ] Review explicit documentation diff, run git diff --check, commit/push existing dev branch and verify remote SHA=HEAD/clean status. No new runtime allowance carries forward.

## Self-review / handoff

The only substantive observation change from the successful-input/failed-exit short route is use of the reviewed same-object recorder. It may still affect timing, so this is evidence gathering, not a controlled proof of causation. Task1 binds Task2 identity and Task2 supplies Task3 outcomes. Please approve the single launch and live-module use; no wider runtime/debugger authority is requested.
