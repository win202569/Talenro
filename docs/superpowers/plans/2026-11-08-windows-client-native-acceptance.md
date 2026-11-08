# Windows Offline Demo Native Verification Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> after explicit user approval. Preserve the existing inline execution method.

**Goal:** Observe the compiled offline demo's four native pages and reset behavior
without claiming production readiness or expanding system permissions.

**Architecture:** Inspect the existing self-contained x64 executable in at most
two ordinary-user sessions. Exercise real controls through the native UI, record
observed results separately from static/model tests, then close gracefully.

**Tech Stack:** Existing WinUI3/.NET10 build, PowerShell, native UI observation.

**Spec:** ../specs/2026-11-07-windows-client-four-screen-design.md

Status: PROPOSED, NOT APPROVED. Planning does not consume or grant a launch.
Baseline source0d8a436f; evidence checkpointfe9c9f14:159 tests pass and repaired
App compile passed. This is verification-only, not a feature implementation plan.

## Global Constraints

- Only the existing Talenro offline demo, fixed demo@example.com and masked dummy
  password; never enter real account information or open external links.
- Permanent notice: “离线演示，不会连接 VPN 或提交账号信息”. All connections simulated.
- No installs, restore, builds, automated test reruns, code fixes, MSIX registration,
  services, routes, VPN, account/payment actions or security/UAC/firewall changes.
- No display scaling, resolution, OS language, accessibility settings or global
  keyboard changes. Resize only this app window within the available display.
- No native screenshots containing unrelated applications, credentials or personal
  data. Crop to this app; commit only reviewed fictional UI evidence.
- No main merge; Task11/12 and44/34/6/217 NOT_ACCEPTED unchanged.

## Proposed finite allowance

Exactly at most2 launches, not2 retries: session1 up to20minutes for functional/UI
checks; session2 up to5minutes only to verify close/reopen reset. Each startup
observation capped30seconds; each graceful-close observation15seconds. Check
progress at intervals<=60seconds. Time limits are observation budgets, not a
guarantee of OS process termination. Do not fill unused time with new scope.

If session1 fails startup, crashes, hangs, requests elevation/setup, or exposes a
functional defect: stop further scenarios and do not use session2 as a retry.
Capture the specific failure, attempt one ordinary window-close, and report any
remaining process. No arbitrary PID/name kills. Repairs need their own scoped
proposal and test/build allowance; never patch while the inspected app runs.

Executable relative to worktree:
`clients/windows/Talenro.Demo.App/bin/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
Resolve exact absolute path before launch. If absent or source differs from tested
baseline, stop; do not rebuild implicitly. Binaries remain ignored, not portable
source artifacts; a new computer needs separately prepared tools/dependencies.

## Review Focus

1. Startup/resource failures must not be disguised by successful compilation.
2. Confirmation Cancel and Escape preserve saved node/lock and return usable focus.
3. Recovered current country must survive completion without changing saved lock.
4. Unknown/stale usage stays unavailable; placeholders never imply real success.
5. Logout and close/reopen clear model and page-local selectors/drafts/notices.

### Task 1: Verify preconditions and launch once

**Files:** read existing executable and clients/windows/tests/manual-checklist.md;
create docs/roadmap/2026-11-08-windows-client-native-verification.md for results.

- [ ] Recheck clean source state, executable path/hash, and command token.
  AdministratorToken must beFalse; otherwise stop without launch.
- [ ] Observe available desktop size/scaling without changing it. If app control
  is unavailable, stop and request user-assisted observation, not a tool install.
- [ ] Start that exact executable once, visibly as an interactive app; record
  process identity/path/start time. User approval of this plan permits this visible
  window, not other apps. If a matching instance already exists, stop to clarify.
- [ ] Within30seconds observe initial Login, readable fixed fields, localized copy
  and permanent notice. A process existing alone is not successful startup.

### Task 2: Inspect real page interactions within session1

**Evidence:** update the native verification record and manual checklist. Each
row must be PASS/FAIL/UNVERIFIED with observation, not inherited test results.

- [ ] Login: normal entry and invalid/submitting/rejected/unavailable/MFA demo
  states; no editable password or real identity claim. Return to normal explicitly.
- [ ] Connection: connect/cancel/late success ignored; connect/explicit success;
  disconnect/explicit completion. One labelled primary action; power glyph not
  keyboard-focusable. No real IP/speed/protection claim.
- [ ] Nodes: search, country groups, Auto, disabled rows, no results/loading/list
  unavailable. Back discards draft. Connected Apply opens explanatory confirmation;
  Cancel/Escape preserve preference/lock and restore focus; Confirm restarts once.
- [ ] Country recovery: chooseJP, lockJP, explicit locked failure, one-time DE
  recovery and completion; saved lock remainsJP. Toggle off/on now usesDE. Also
  inspect retry-in-country and stay-disconnected choices, without automatic retry.
- [ ] Account: active/exhausted/expired/unknown/stale/device-limit scenarios;
  unavailable values never become zero. Count means activated, not concurrent.
  Placeholder destinations/actions explain nonimplementation; no purchase flow.
- [ ] Keyboard Tab/Shift+Tab and dialog Enter/Escape: visible focus, labelled
  controls, usable focus after dismissal. Inspect all four pages for clipping and
  scrollability at current available size; compare saved design targets without
  claiming pixel equivalence. Save up to6 app-only screenshots if safely possible.
- [ ] Logout: returns Login, resets selectors, node draft, lock and notices. Leave
  a changed fictional preference in a new demo session before ordinary close so
  session2 can test process-lifetime reset. Verify this exact process exits.

### Task 3: Reopen reset and bounded handoff

- [ ] Only if session1 completed and its process exited, launch same executable
  once more. Observe initial Login, enter demo and confirm initial disconnected
  state/no saved lock or node selection. Do not infer reset merely from Login.
- [ ] Close normally and verify this exact process exits. Record both sessions'
  duration, launch count and all failures/unverified items.
- [ ] Preserve result summary, checklist and safe evidence in existing development
  branch; explicit staging, ordinary commit/push, remote/local SHA comparison.

## Explicitly unverified / separate future scope

No claim of1440x1024 client-area coverage where it cannot fit: last reported
1920x1200 at125% offers roughly1536x960 logical display. Reobserve actual current
conditions, and mark unavailable targets; do not change display settings.
Exact1024x768 client dimensions require measured evidence, not window outer size.
Enlarged-text settings, synthetic RTL injection, English OS-locale switching and
screen-reader interaction are deferred if not already available without changing
the user's settings. Merely having English resources does not verify runtime
fallback. Record contrast as unverified unless actually measured.
Process launch and source inspection cannot prove absence of all incidental SDK
network/storage behavior; no security/production/runtime audit claim is authorized.

## Self-review and handoff

All four views, confirmation/recovery, unknown entitlement and both reset types
are assigned observable checks. Native-only gaps remain explicit above instead
of weakening acceptance. No source change/TDD cycle is proposed in this plan.
Await user approval of the2-launch/25-minute maximum scope before execution.
