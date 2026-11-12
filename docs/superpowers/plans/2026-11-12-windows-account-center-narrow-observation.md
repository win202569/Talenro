# Account-center Narrow Observation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Obtain one bounded narrow-window Account observation without rebuilding or changing the application.

**Architecture:** Reuse the existing fixed-hash entry. Resize on Connection before Account consumes its two capture attempts; analyze immutable receipts after normal closure.

**Tech Stack:** Windows, ordinary PowerShell7+, existing WinUI diagnostic build and approved UI interaction tools.

**Spec:** [Approved design](../specs/2026-11-12-windows-account-center-narrow-observation-design.md).

## Global Constraints

- Current live allowance0. Plan approval is not launch authorization.
- One separately approved launch, zero relaunches/retries; unchanged six-minute outer budget.
- Root width640 +/-20 DIP, height620 +/-10 DIP; actual Account page width below480 DIP.
- en-US / rtl-mixed; rendering and text scale both1; zero scroll.
- No code, entry, hash override, build, install, system/security-setting change or forced termination.
- Preserve all old/new records; Login is not narrow evidence; C4 remains NOT_ACCEPTED.
- Work in existing codex/c12-b01-task9-coordinator worktree; PowerShell only.

## Review Focus

1. Hash/source drift or wrong token: preflight must reject, not be weakened (Task1).
2. Outer pixels mistaken for content DIP: verify diagnostic title then raw root/page measurements (Tasks2–3).
3. Account visited before resize: invalidate condition, never try to refill capture budget (Task2).
4. Missing/mismatched samples or scale change: inconclusive, never select only favorable evidence (Task3).
5. Close/cleanup disagreement: preserve provisional/final distinction; no kill/retry (Tasks2–3).

No product code is implemented, so no RED/GREEN cycle or full test rerun is planned.
Checks below validate observations, not synthetic replacements for live evidence.

### Task1: Readiness and authorization

**Files read:** approved spec; `clients/windows/tests/observe-email-geometry-account-center.ps1` and its six policy Sources; existing build/live inventories.
**Interface:** produces readiness result and explicit one-launch approval, or blocker.

- [ ] Confirm plan review, then separately ask for one launch/zero retries and availability for up to six minutes. Do not infer this from design approval.
- [ ] Read current computer-use instructions before any UI interaction; use normal window controls, not a new UI automation script.
- [ ] Verify ordinary Windows token, current branch, clean worktree, policy source blobs matching HEAD, exact EXE/DLL hashes from spec, and successful zero-match target/prototype process query. Query errors are blockers, not zero counts. Reuse existing entry safety rules; no check overrides.
- [ ] Confirm approved UI tools expose a safe resize operation or user is available to resize manually. If neither is available, stop before launch.
- [ ] Record readiness and approval scope. Entry repeats its complete preflight at execution time; no stale check substitutes for it.

### Task2: One live condition, then normal closure

**Files consumed:** unchanged fixed entry and bound candidate.
**Produces:** new RunId, terminal result, screenshot/title evidence and ignored receipt files, or documented invalid/failed condition.

- [ ] Only with Task1 complete, run once from the worktree in ordinary PowerShell7+. Resolve the available executable; do not install or guess a versioned path:

```powershell
$observationPwsh = (Get-Command pwsh -ErrorAction Stop).Source
& $observationPwsh -NoProfile -File '.\clients\windows\tests\observe-email-geometry-account-center.ps1' -RunApproved
"Observation exit code: $LASTEXITCODE"
```

If pwsh is unavailable, stop before invocation and locate an existing approved runtime. Never rerun merely because terminal output is delayed; retain the original session.

- [ ] Select the single observed candidate window by current identity/path. Enter demo, remain on Connection; do not initiate connection or visit Account.
- [ ] Resize using normal controls. Release resize and inspect title twice at least one second apart: same dimensions, width620–660, height610–630 DIP, rendering/text scale1, expected profile. If unavailable/incorrect, close normally and mark invalid, no relaunch.
- [ ] Navigate to Account once. No further resizing, scrolling or navigation. Save observed title/screenshot and whether email is fully visible. If Account was visited early, mark invalid rather than revisiting.
- [ ] Click normal title-bar X within the original deadline. Poll the same observation session in bounded intervals; record terminal Outcome, ExitVerified, Finalized, OverrunMs and exit code. Success requires exited/agreement, true/true,0 and exit0; otherwise preserve actual failure, no forced termination.
- [ ] Read-only check for residual target/prototype processes. Residuals/query errors are blockers; do not start another instance. Keep all raw files, even for invalid conditions.

### Task3: Read-only evidence and portable handoff

**Read:** new run manifest/process/geometry/claim files below candidate output; old run66db34b22cc748bbb8fdf61f656b879a inventory for comparison.
**Create:** `docs/roadmap/2026-11-12-windows-account-center-narrow-observation.md` (sanitized evidence summary).
**Modify:** this plan checkboxes/results, `docs/roadmap/current-status.md`, `docs/runbooks/repository-recovery.md`.
**Interface:** limited supported/failed/inconclusive condition, not full product acceptance.

- [ ] Hash/inventory files; validate run IDs, artifact bindings, sequence ordering, capture errors and process/geometry correlation. Record raw provisional fields unchanged, separately from final terminal cleanup results.
- [ ] Require both Account snapshots: usable finite positive geometry, matching root/page dimensions within approved root ranges, page width below480, scales1/profile unchanged, offsets0. Missing/inconsistent data => inconclusive. Login samples cannot satisfy this check.
- [ ] Compute all known-boundary insets with existing GeometryEvaluator rules; report both samples, panel/page widths and screenshot result separately. Any known-box overflow or visible email clipping fails this narrow condition. Preserve ClipChainComplete=false limitations; never equate RootDIP right with desktop right.
- [ ] Write limited conclusion and remaining acceptance gaps. Do not claim Login-narrow/LTR/other-scale/full-regression/security/C4 acceptance. No automatic follow-up run.
- [ ] Review explicit docs-only diff and links; run `git diff --check`. Stage only listed necessary docs, commit/push existing development branch, verify remote SHA equals HEAD and worktree status. Never stage ignored binaries/receipts or unrelated files.

## Plan review record

Latest execution: a NEW explicit bounded approval allowed run
f86a7c0d8cad4a02a6fd307c1652c790 after the earlier invalid run. User confirmed window
visibility and resize completion. Condition640x620/page420x534 obtained before first
Account navigation; both Account samples and visible email support known-box
containment. Normal exit0 at193.91s/no residual processes. Read-only inventories
and limited result recorded in the linked run report. Live0; no further launch.
Earlier unchecked steps describe the first attempt, not missing authority to retry.

Execution checkpoint: checklist and one-launch approval received. One run
ef33d39285174f1095ac69feae7e6da5 ended normally at306.26s/exit0, residual processes0.
Manual resize completion was not observed; root stayed1136x620 and Account was
never visited. Narrow condition NOT_OBTAINED, not a layout pass/failure. No retry.
Tasks1/2 reached the authorized invalid-condition close path; Task3 recorded the
missing-evidence outcome, not a successful geometry comparison. Unchecked capture
steps remain unmet and MUST NOT be interpreted as authority for a new launch.
See [run result](../../roadmap/2026-11-12-windows-account-center-narrow-observation.md).

Self-review: approved scope, timing, artifact identity, dimension/scale conditions,
invalid-run handling, close/receipt distinction and recovery gates mapped above.
No implementation or observation executed while preparing this plan. Execution
was subsequently authorized once as recorded above; that live allowance is now0.
