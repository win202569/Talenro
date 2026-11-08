# Native verification — PARTIAL, NOT ACCEPTED

Five authorized ordinary-user sessions were observed and closed on2026-11-08,
after one interrupted request and explicit replacement approval. See
../../../docs/roadmap/2026-11-08-windows-client-native-verification.md for exact
PASS/UNVERIFIED subchecks. Unchecked compound rows below are not wholly verified.
The launch allowance is consumed; obtain a new finite allowance for remaining
checks. Compilation/XML tests do not substitute for native observations.

- [x] Initial Login; permanent offline notice on all four pages.
- [x] Read-only demo email/mask; normal entry plus all synthetic failure states.
- [x] Sidebar routes three authenticated pages; other destinations explain scope.
- [x] Explicit connect/cancel/success/failure/disconnect, including stale completion.
- [x] Search, country groups, Auto, disabled nodes, empty/loading/unavailable list.
- [x] Back discards draft; connected Apply names countries; Cancel preserves state.
- [x] Observed cross-country Confirm enters Connecting with DE lock; reopening
  Nodes shows de-1 selected (PID17108). This is visible-result evidence only.
- [ ] Runtime confirmation atomicity/no intermediate partial state: not established
  by screenshots. Reducer-level assertions are separate evidence; see coverage map.
- [x] Country lock unknown-country refusal; all three explicit recovery choices.
- [x] Expired/exhausted/unknown/stale/device-limit admission and account presentation.
- [x] Unknown/stale usage unavailable; usage aggregate, activated versus concurrent devices.
- [x] Excluded actions explain nonimplementation; no external links or false success.
- [x] Observed logout/reentry returns disconnected, lock off, active account,
  Auto selection and blank notice (historical native sessions).
- [x] After saved jp-1/JP lock and ordinary close, reopen starts Login;
  reentry shows disconnected/lock off and Auto/empty search (PID10208→14520).
- [ ] Exhaustive reset of every internal/UI field and persistence behavior:
  not established by the selected native sequences; model reset tests are separate.
- [x] Changed Nodes search/list reset to empty/normal/Auto on logout and reentry.
- [x] Changed Login scenario before entry, returned normal, then logout reset sequence.
- [x] Pointer scrolling visibly moves Connection/Nodes/Account content.
- [ ] Keyboard traversal, visible focus, Escape/Enter, dialog return focus.
- [x] Observed Login forward/reverse path and Registration Escape focus return.
- [x] Observed Connection/Nodes/Account primary controls and lower-control auto-scroll.
- [x] Node arrows skip disabled rows; confirmation Tab/Shift+Tab stays in dialog,
  Escape returns to Apply; Manage devices Enter dismissal returns to its button.
- [x] Correct extra Tab stop: Settings↔Connect/Search/MFA each observed in one
  forward/reverse key after PageHost fix; see focus-fix execution record.
  Runtime internal object identity remains unmeasured and metadata stale;
  this scoped visual check does not establish full keyboard/accessibility acceptance.
- [ ] Screen reader names, decorative glyph excluded, status text not color-only.
- [ ] Contrast and enlarged text; 1024x768 and 1440x1024 logical viewport.
- [ ] Long/RTL test-only samples, LTR email, no clipped actions; main content scrolls.

Known viewport constraint: reported1920x1200 at125% gives about1536x960 logical
display, insufficient for1440x1024 client area. Do not change display settings or
claim target coverage without separate agreement. Screenshot evidence, real
runtime/network/persistence behavior and final design comparison are unverified.

Evidence mapping and remaining scope:
[coverage map](../../../docs/roadmap/2026-11-08-windows-client-acceptance-coverage.md).
Checked historical observations are not fresh reruns on every subsequent build.
