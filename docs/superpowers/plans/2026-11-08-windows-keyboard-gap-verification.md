# Windows keyboard gap verification plan

Status: proposed, execution approval pending. Baseline b3c19c55; application
source last changed in d64de426. This is an observation plan, not implementation:
no TDD, code edits, delegated implementation or new build is required.

Goal: fill specific missing keyboard evidence without repeating the core-flow
acceptance sessions. Scope comes from clients/windows/tests/manual-checklist.md
and docs/roadmap/2026-11-08-windows-client-acceptance-coverage.md.
Execution: one operator in this chat using the existing Windows app-control tools.

## Bounds and preconditions

- Proposed allowance: one ordinary-user launch, maximum20minutes; stop starting
  new scenarios at18minutes. Startup30seconds, ordinary close observation15seconds.
  No retries/restarts, tests, builds, restores, installs, instrumentation or settings
  changes. An unresolved defect stops scenarios; close normally and report it.
- Read-only preflight: non-admin context, no existing demo process/window, unchanged
  App/Core against d64de426, expected Debug executable path and recorded hash
  83921BD364A6FEA64587E9684D7FEA2266299ACBE79D8EE8D2BF6439F60CC2B6.
  This hash alone does not attest all adjacent files. If the artifact/source
  mismatch or process ownership is unclear, stop rather than rebuild or terminate.
- Use latest returned window and a refreshed observation after each input. Inspect
  visible focus outline/caret; stale metadata alone is never a pass or failure.
  If observation cannot establish focus, mark UNVERIFIED, do not invent its target.
- Placeholder buttons are synthetic explanations only. Do not activate Delete
  account in this session. Never interact with real authentication/security UI.

## Missing-path matrix

All rows below are pending observation. Previously passed entry crossings,
ordinary forward paths, reset sequences and primary connection flows are not
being re-certified. Navigation/setup needed to reach a missing path is permitted.

| ID / priority | Path to observe | Pass criterion |
| --- | --- | --- |
| K1 / 1 | Connection, recovery panel hidden: focus Entitlement selector, then Shift+Tab through Disconnect completion → Locked-all-failed → Failure → Success → Country lock → Choose location → Connect | Each adjacent target has visible focus, with automatic scrolling as needed; no skipped enabled control or invisible intermediate stop |
| K2 / 1 | Nodes normal list: focus List scenario, reverse to Apply → Back → selected enabled list row → Search | Each target visible; entering list does not select a disabled/group row. Existing arrow-skip test need not be repeated |
| K3 / 1 | Account: focus Exit (do not activate), reverse through Delete placeholder → Language → Manage devices → MFA | Every enabled button reached in reverse with visible focus; scroll follows lower targets. Do not activate Delete |
| K4 / 2 | Sidebar reverse from Settings → Account → Devices → Traffic → Nodes → Connection, without activation | One visible focus target per key, no unexpected navigation |
| K5 / 2 | Keyboard Return opens Traffic, Devices, Language and MFA explanation, separately; Escape closes each | Explanation appears, no real action; focus returns visibly to the exact invoking control. Settings/Registration/Reset Escape and Manage-devices Enter already covered |
| K6 / 3 | Recovery panel visible: create jp-1 known-country simulation, enable JP lock, use explicit locked-country-failure scenario; traverse Country lock → Recover once → Retry country → Stay disconnected → Success and reverse | Recovery actions are keyboard reachable in both directions with visible focus; do not activate the three recovery actions (their outcomes already covered) |

K1–K5 use a fresh normal, disconnected session; K2 keeps Auto selected and the
list normal. K6 setup may use pointer input: select/apply jp-1 while disconnected,
request simulated connection, explicitly simulate success, enable JP lock,
request disconnect and explicitly complete it, request connection again, then
invoke locked-country failure while Connecting. Confirm each displayed state and
the visible recovery panel; do not assume setup succeeded. No real VPN is involved.
If20minutes cannot cover all rows, unexecuted
rows stay UNVERIFIED; do not extend the session or silently launch again.

## Execution and evidence handoff

- [ ] Complete read-only preflight, then one approved launch and startup observation.
- [ ] Perform K1–K6 in priority order; record actual target sequence and exceptions.
  A complete row needs every listed transition; partial rows list passed subpaths.
- [ ] Close normally before20minutes, confirm exact process exit and no target window.
  No forced termination is authorized; if close fails, report and request direction.
- [ ] Update native-verification record, manual checklist, current status, execution
  record and recovery document with evidence labels; commit/push explicit files to
  existing development branch under standing authorization. No main merge.

This does not prove full keyboard/accessibility acceptance: dropdown open-list
behavior, disabled-state permutations, screen reader, target viewports, contrast,
long/RTL/English layouts, runtime atomicity and exhaustive storage/reset remain
outside this session. Do not mark the compound keyboard checklist fully passed.

Self-review: sequences matched XAML control ordering and existing native records;
recovery controls were added because ordinary-page traversal does not cover the
conditionally visible panel. Delete is traversed only. Existing evidence is retained,
not rerun as a prerequisite. Each row has an observable result and bounded stop.
