# Windows offline demo — native implementation checkpoint

Current verified result: final full suite159/159 passed,0 skipped,exit0 after
0d8a436f. Both review fixes now pass regression; repaired App compile also passed.
Historical pending-result notes below are retained for traceability. Native
startup/visual/accessibility acceptance is still unrun and separately gated.

Latest fix checkpoint: country lifecycle and localized notices have targeted
GREEN8/8 and19/19; app compilation passed. Full159 run had158 pass and one stale
CurrentCountry expectation. That expectation is corrected but not rerun; do not
read the historical132-pass result below as verification of this latest edit.

Contains the in-memory Core reducer, five behavioral test classes, a native WinUI
four-page App, and two static XAML/resource contract test classes. Latest full
model/declaration suite:132 passed,0 failed/skipped,exit0. Static checks do not
prove native click wiring, visual fidelity, focus, accessibility or runtime safety.
First shell compile passed; final four-page compile/review status is recorded in
the execution record. Do not infer acceptance from this implementation checkpoint.

Read [execution record](../../docs/roadmap/2026-11-08-windows-client-offline-demo-execution.md)
and [approved plan](../../docs/superpowers/plans/2026-11-08-windows-client-offline-demo-execution-amendment.md)
before running commands. App dependency restore and locked restore passed with
Microsoft.WindowsAppSDK2.6.5. Both test and app package graphs are locked.
Use the approved non-admin execution context: default agent context blocks local
test-host connections with10013. Check the token before each test; never change
UAC/firewall/timeout to force a pass. Original test/build allowance is now consumed;
any further verification requires a scoped allowance. Native launch remains
unauthorized. Cloning does not restore tool binaries or caches.

Four pages: Login uses fixed read-only demo credentials; Connection dispatches
explicit simulation events; Nodes edits a draft and confirms connected changes;
Account shows fictional usage and activated-device counts. Unknown/stale usage
shows unavailable. Excluded actions show explanations, not real account changes.
No VPN, account submission, billing or persistence is implemented.

Pending native checks are in tests/manual-checklist.md; long/RTL samples in
tests/ui-copy-cases.json are synthetic test fixtures, not approved translations.

SDK/package caches and outputs are intentionally ignored. No secrets or live data
are required. This is not a production VPN client or an accepted build.
