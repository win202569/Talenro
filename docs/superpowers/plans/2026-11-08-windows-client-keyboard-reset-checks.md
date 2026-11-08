# Windows offline demo keyboard/reset verification plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> after approval. Preserve inline execution; no delegated implementation.

**Goal:** Observe the remaining Login reset sequence and bounded keyboard/focus paths.
**Architecture:** One existing ordinary-user offline-demo process, fictional data
only. Record observed controls and outcomes separately from unverified coverage.
This is an observation plan, not authorization to implement or fix code.
**Tech Stack:** Existing WinUI3 executable, native UI observations, PowerShell.
**Spec:** ../specs/2026-11-07-windows-client-four-screen-design.md.

Status: user confirmed afterc4f64190; PID10372 observed and normally closed after
11.15minutes. This allowance1/1 consumed. Observed checks recorded in
../../roadmap/2026-11-08-windows-client-native-verification.md; overall PARTIAL
because intermediate content-entry focus remains unidentified. No new launch.

## Global Constraints

- At most ONE additional launch;20minutes total including startup and close.
  Startup30seconds, ordinary-close observation15seconds. Stop new scenarios at
  minute18. No restart/retry or extra launch to finish unchecked items.
- Use only the existing worktree executable:
  `clients/windows/Talenro.Demo.App/bin/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
  Verify ordinary token, no existing demo process/window, App/Core unchanged from
  0d8a436f, and executable SHA256
  `026D8642F47107CE84465BCF0E3133EC02242AD25B59D41068A58BC11CA67C95`.
  Stop on mismatch or unknown process identity; restricted invisibility is not
  absence. No claim that the executable hash attests all adjacent files.
- No source edits, automated test/build/restore, installation, real credentials,
  external links, network integration, or system/security/display/accessibility
  setting changes. No screen-reader activation within this allowance.
- Confirmed app defect/crash/hang/elevation prompt: stop scenarios; attempt one
  ordinary close if targetable, verify exact process exit; no kill or repair.
  User interruption stops UI input immediately. Tool errors are not app failures
  without evidence; refresh observations before any permitted input recovery.
- Retain prior results; do not repeat the permission/recovery/list-state matrix.
  Any repeated setup below exists only to reach a focus/reset checkpoint.
- Task11/12 and44/34/6/217 NOT_ACCEPTED unchanged; no main merge.

## Review Focus

1. A changed Login scenario must not leak through exit/reentry (Task1).
2. Keyboard focus must be visible and follow usable page order (Task2).
3. Disabled/read-only/decorative items must not act as misleading actions (Task2).
4. Closing a modal must return focus to its invoking control (Task2).
5. Long content must remain keyboard reachable without trapping focus (Task2).

## Task 1: Login reset sequence

Files: append evidence to docs/roadmap/2026-11-08-windows-client-native-verification.md;
update clients/windows/tests/manual-checklist.md only for fully observed subchecks.

- [x] Check preconditions; launch once; record exact PID/path/start and startup.
- [x] On Login, select invalid-input scenario with keyboard; observe disabled
  entry and explanation. Return selector to normal; observe enabled entry.
- [x] Enter demo, visit Account, exit demo. Inspect Login selector normal, blank
  notice and enabled entry. Record this specific sequence, not a stronger claim
  that a nonnormal selector survived entry (entry deliberately requires normal).

## Task 2: Bounded keyboard/focus observations

Same evidence files. Observe screenshot focus as well as accessibility metadata;
if they disagree and focus cannot be established, mark the subcheck UNVERIFIED.
Use one action then refresh. Pointer navigation may set up a page, but cannot
count as keyboard reachability. Record actual control sequence, not just PASS.

- [x] Login: Tab through actionable controls to the scenario selector; reverse
  with Shift+Tab. Observe visible focus and lower-control auto-scroll. Read-only
  fields may receive reading focus but must not become editable.
- [x] Keyboard-open registration placeholder, dismiss with Escape, verify focus
  returns to Registration. If inaccessible, record the observed limitation.
- [ ] Enter demo. Connection: traverse sidebar and content controls, observe one
  labelled primary action (decorative glyph not an extra action), lock/scenario
  reachability and reverse traversal. Do not repeat connection transition matrix.
- [ ] Nodes: keyboard-reach search, node list, Back/Apply and scenario selector;
  inspect arrows over country/disabled rows without applying an invalid selection.
  Confirm lower actions reachable and reverse traversal returns usable focus.
- [x] Set up jp-1 connected using explicit demo events, draft de-1 and open Apply
  confirmation. Inspect initial Cancel focus, Tab/Shift+Tab containment, Escape
  dismissal and return to Apply. Do not reconfirm a cross-country switch.
- [ ] Account: traverse available actions; keyboard-open Manage devices placeholder,
  dismiss with Enter and verify return focus. Inspect exit action reachability.
- [x] Close normally within the time limit and verify exact process exit/window
  disappearance. Mark any unvisited control/path UNVERIFIED; do not infer a
  whole-page or whole-app keyboard pass from a few successful controls.

Connection/Nodes/Account steps executed, but remain partially verified: primary
controls and recorded reverse paths were visible; each page's intermediate focus
stop after Settings was not identifiable. Account placeholder/Exit checks passed.
See the exact sequences in the evidence ledger; do not repeat whole matrices.

## Task 3: Preserve results

- [x] Reconcile the two tasks against observations; update manual subchecks without
  treating focus metadata or source tests as assistive-technology acceptance.
- [x] Update docs/roadmap/current-status.md,
  docs/roadmap/2026-11-08-windows-client-offline-demo-execution.md and
  docs/runbooks/repository-recovery.md.
- Preserve the explicit documentation checkpoint after diff checks; commit/push
  existing dev branch and compare remote/local SHA and clean worktree. The handoff
  reports that transport result; this file cannot attest its own future push.

## Deferred and self-review

Screen reader, contrast, enlarged text, exact1024x768/1440x1024 logical viewports,
RTL/long samples, English fallback, final design comparison, hidden atomicity and
network/storage/security acceptance remain outside scope. No setting changes to
obtain coverage. Scope checked against spec: this addresses remaining reset and
selected keyboard requirements only; other requirements retain their prior status.
No code interfaces or new test cycles are introduced. No further launch is implied
by an incomplete checklist; report the remaining evidence gap at the time limit.
