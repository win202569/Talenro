# Account Lower Controls Verification Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this checklist in the current session after review and separate live approval.

**Goal:** Fill the narrowly affected Account scrolling/focus and visible logout-route gaps without replaying the email investigation.

**Architecture:** Reuse the existing fixed account-center diagnostic entry and artifact. Observe lower controls at the already feasible narrow size, one harmless explanation dialog, then logout/re-entry and normal close. No implementation changes.

**Tech Stack:** Existing WinUI diagnostic build, ordinary PowerShell7+, supported native UI tools.

**Spec:** [Remaining acceptance scope](../../roadmap/2026-11-12-windows-remaining-acceptance.md), R1/R2; size and binding constraints from [approved narrow design](../specs/2026-11-12-windows-account-center-narrow-observation-design.md).

Status: checklist confirmed and one separately approved run completed; live0.
See [scoped result](../../roadmap/2026-11-12-windows-account-lower-controls-observation.md).
Lower controls/traversal and selected recovery observed; immediate visible return
focus and isolated keyboard auto-scroll remain UNVERIFIED. No automatic retry.

## Source-impact basis

Compared with historical C4 baseline818f1604, Account adds geometry names and
panel HorizontalAlignment=Center. The latter affects every panel child and is
shared with ordinary builds. Account action handlers are unchanged except the
conditional geometry target accessor. Core, view-model, Connection/Nodes views,
resources, shell XAML and layout presentation/resource catalog had no diff in
the inspected paths. Diagnostic App/MainWindow adds startup/capture/close hooks;
absence of business-code changes does not prove absence of runtime interaction.
App/Core/Directory.Build.props have no further diff from build sourcebed155cb.

Thus do not repeat all connection/node/country-lock/entitlement scenarios solely
for this alignment change. Do not qualify the ordinary build with this diagnostic
session: ordinary-build layout/packaging remains a separate R7 item.

## Global Constraints

- No launch in preparation. Proposed future scope: one separately approved launch,
  zero relaunches/retries, existing360s outer budget; begin normal closure by300s.
- Unchanged `clients/windows/tests/observe-email-geometry-account-center.ps1`,
  its fixed EXE/DLL hashes, source/ordinary-token/process/path guards and time cap.
- en-US/rtl-mixed; root640 +/-20 by620 +/-10 DIP, scales1; no OS-setting change.
- Confirm user sees the window before interactions. Resize on Connection before
  entering Account. No real credentials, connection, account mutation or external links.
- Never activate Delete account, even though its current handler is a placeholder.
  Focus traversal may pass over it without pressing Enter/Space/clicking it.
- No builds, tests, instrumentation, entry override, forced termination or new helper.
- Capture attempts remain2/page. Screenshots after scrolling/logout are separate
  UI evidence, not new geometry samples; never reuse old samples for a later state.

## Review Focus

1. Lower controls clipped/unreachable: inspect every label and scroll access (Task2).
2. Focus disappears during scrolling: observe actual visible focus and movement (Task2).
3. Dialog return focus: use only Manage devices explanation, check invoking control (Task2).
4. Logout/profile route changes: observe selected visible reset state, not hidden reset (Task2).
5. Deadline/close failure: stop scenarios, preserve partial result, no relaunch/kill (Tasks2–3).

### Task1: Ready the bounded observation

**Read:** this checklist, fixed entry, artifact provenance and current UI tool instructions.
**Produces:** verified readiness plus separate explicit live authorization, or blocker.

- [ ] Review checklist; obtain separate one-launch/zero-retry approval and user availability.
- [ ] Recheck clean correct worktree, ordinary token, fixed artifacts/source identities,
  no matching target/prototype processes. Existing entry repeats all guards.
- [ ] Only then invoke the unchanged fixed entry once with RunApproved. Keep its
  original observation session. Activate target and wait for user visibility confirmation.
- [ ] On Connection, obtain narrow size as specified; confirm stable title twice at
  least one second apart before first Account navigation. If not feasible, close
  normally and mark the condition invalid, rather than entering Account anyway.

### Task2: Minimal visible checks

**Consumes:** Task1's single live session; **produces:** per-check PASS/FAIL/UNVERIFIED notes.

- [ ] Enter Account once. Using its existing scroll pane, inspect readable labels
  and access to MFA, Manage devices, Language, Delete account and Exit demo/reset.
  Record wrapping/clipping and that scrolling moves content; no email remeasurement.
- [ ] Establish visibly confirmed focus on MFA (Tab traversal from an observed
  current focus). Tab through the five buttons to Exit demo/reset, then Shift+Tab
  back to Manage devices. Record each actual focus position and automatic scrolling;
  do not infer focus from stale accessibility metadata or activate intermediate buttons.
- [ ] With Manage devices visibly focused, press Enter once. Inspect explanatory
  dialog and its close action; dismiss using its labelled Close button. Confirm
  visible focus returns to Manage devices and one subsequent Tab reaches Language.
  No repeated dialogs or destructive-button activation.
- [ ] Scroll to and activate only Exit demo/reset. Observe Login, hidden sidebar,
  English/RTL presentation and fixed demo fields. Enter demo once; observe
  Connection disconnected, country lock Off and retained language/RTL presentation.
  Do not connect or select a node. This is selected visible recovery, not exhaustive reset.
- [ ] Close normally by300s (earlier on defect/unusable UI); preserve remaining
  budget for observing exit. Stop further scenarios on failure, mark unvisited
  checks UNVERIFIED. No retry, force-close or deadline extension.

### Task3: Read-only result and recovery record

**Create:** `docs/roadmap/2026-11-12-windows-account-lower-controls-observation.md`.
**Update:** this checklist, current-status, repository-recovery and scoped manual-checklist addendum.

- [ ] Record run/artifact identities, measured condition and each Task2 outcome.
  Correlate final process result/exit code with preserved raw inventories; distinguish
  provisional saved outcome from final post-cleanup console result. Check residuals.
- [ ] Keep raw records local/ignored. No inference of ordinary-build acceptance,
  complete C4, screen-reader/contrast/other-scale coverage or historical C3 repair.
- [ ] Review explicit documentation diff and links, run git diff --check, commit/push
  necessary files on the development branch and verify local/remote SHA equality.

## Self-review

Scope fits R1 plus the affected visible logout route of R2. Manage devices shares
the unchanged placeholder handler with the other account buttons, so one dialog
is a representative check, not proof of all dialogs. Exact focus transitions and
normal close are observed, not assumed. No source change or test cycle is proposed.
Execution occurred in the current session. The single allowance is consumed;
unchecked steps do not authorize another run or imply every subcheck passed.
