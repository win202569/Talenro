# Native verification — PARTIAL, NOT ACCEPTED

2026-11-13 scoped update: keyboard-only Tab auto-scroll and visible focus return
to Manage devices after Return closes its dialog observed at650x620/page430x534.
See [keyboard result](../../../docs/roadmap/2026-11-13-windows-keyboard-followup-observation.md).
Normal exit0, residual query0. Only these questions supersede the older unverified
keyboard entries below; no compound checkbox/C4 acceptance promotion. Live0.

Latest lower-controls scope (2026-11-12): at629x620 root/409x534 page, five Account
buttons visible after scrolling; forward MFA→Exit and reverse Exit→Manage devices
focus outlines observed. Manage devices dialog closed; next Tab reached Language.
Immediate return outline and isolated keyboard auto-scroll remain UNVERIFIED.
Logout/re-entry visible state restored; normal exit0, no residuals. See
[scoped record](../../../docs/roadmap/2026-11-12-windows-account-lower-controls-observation.md).
No compound checkbox or C4 acceptance closed; live0.

Latest reconciliation (2026-11-12): current Account Center candidate has scoped
email evidence at root1136x620 and640x620 DIP, en-US/rtl-mixed, scales1. Known-box
containment and visible email do not close the compound layout/accessibility rows.
See [remaining acceptance map](../../../docs/roadmap/2026-11-12-windows-remaining-acceptance.md).
Historical checks below retain their original artifact baselines; C4/overall remain
NOT_ACCEPTED, live allowance0. No app/test/build executed for this reconciliation.

Latest continuation: C1 English confirmation/Cancel, original selection retention
and logout/re-entry observed; C2 exact long-status Connecting-only text and Cancel
restoration observed; both normal close exit0. C3 exact long-country labels/jp-1
search/selection and Chinese Cancel route observed, but normal close left a
windowless process: shutdown FAIL, row not accepted. C4 started prematurely before
C3 exit confirmation, was closed/exit0 without profile validation; C5/C6 NOT_RUN.
See docs/roadmap/2026-11-08-windows-layout-stage-b-continuation-execution.md.
No compound acceptance box or broader gate is closed by this partial attempt.

Latest scoped recheck (2026-11-08): fresh ordinary/diagnostic builds and static
isolation passed; four sessions exited0. See switch-fix execution ledger in
docs/roadmap for exact artifact hashes, arguments and observation limits.

- [x] Scoped CountryLock state text: explicit en-US On/Off and zh-Hans 开/关;
  select jp-1 while disconnected, observe JP on and localized off status.
- [x] Scoped selected-language retention: Account→Connection while on; then off,
  Exit demo/reset and re-enter; disconnected/off and selected language retained.
- [x] Scoped ordinary Chinese paths: ordinary build and diagnostic/no-flag build
  show no diagnostic marker; initial off and JP on/off transition observed.
- [x] Local Tab observation in both diagnostic languages: next scenario button
  has visible outline, not activated; offline notice persists.

These close only the V3 switch-language finding. Broader V3 dialogs, V4–V7 profiles,
remaining V8 checks and all compound acceptance rows remain pending. No exact
size, complete keyboard traversal, accessibility metadata or screen-reader claim.
No test suite rerun in this scope. Earlier pending-native statements below are
historical and superseded only for the explicit checks above.

Latest source-only follow-up: CountryLock state-language resources/bindings fixed;
targeted35/35 and full226/226 passed. No App rebuild or native recheck yet. V3
closure remains pending; all old artifact hashes refer to pre-fix outputs. Observe
both On and Off in English/Chinese after newly authorized builds and sessions.

Latest Stage B attempt:9/14 invocations, stopped at English baseline V3 because
CountryLock off text remained Chinese '关'. Ordinary four-page route and Chinese
baseline route/dialog/reset observed; six rejection exit/message checks passed.
V4–V8 NOT_RUN; exact-size/resize and transient-window absence UNVERIFIED. All
processes exited. See docs/roadmap/2026-11-08-windows-layout-native-stage-b-execution.md
from repository root. Compound rows below remain unchecked; older no-native-check
wording is historical, not the latest status.

Latest Stage A diagnostic implementation has passed code/build checks only; none
of the native rows below were rerun on its new ordinary/diagnostic outputs.
See2026-11-08-windows-layout-test-mode-execution.md in docs/roadmap for hashes and
the closed minor negative-localization test gap (latest full222/222). This test-only
follow-up does not complete any native row. Additional pending checks:

- [ ] Ordinary executable retains behavior; reserved options reject without window.
- [ ] Diagnostic executable without activation behaves as ordinary; invalid options reject.
- [ ] Activated zh-Hans/en-US baseline and all four samples render at intended conditions.
- [ ] Measured title root/page dimensions agree with independent client/DPI evidence.
- [ ] Resize/scale updates and normal shutdown clean up observer subscriptions.

Six authorized ordinary-user sessions were observed and closed on2026-11-08,
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
- [x] K1–K4: remaining three-page reverse paths and sidebar reverse sequence,
  with visible focus/automatic scrolling in the tested normal state (PID8700).
- [x] K5: Return/Escape on MFA, Language, Devices and Traffic explanations;
  visible focus returns to each invoking control (PID8700).
- [x] K6: JP-failure recovery panel Country lock↔Recover once↔Retry country↔
  Stay disconnected↔Success, each adjacent step in both directions (PID8700).
  Recovery choices not activated. Expanded dropdown/disabled-state permutations
  and full accessibility remain outside these scoped checks.
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

## C5 follow-up — 2026-11-09

One approved unknown-usage launch stopped at observer identity validation before
functional navigation (Process/CIM creation-time string precision mismatch).
Normal close requested; later target process/window inventories empty. No original
handle exit verdict. C5 remains NOT_RUN/NOT_ACCEPTED, not passed or an application
display failure. See docs/roadmap/2026-11-09-windows-unknown-usage-native-execution.md
from repository root. No further launch allowance.

## C5 fixed-entry follow-up — 2026-11-09

Separate one-launch approval consumed. Initial Active numeric values, Unknown and
Stale unavailable values, active Account/inactive Connection title conditions were
observed. Final Active restoration NOT_RUN to preserve close-by300s. C5 PARTIAL /
NOT_ACCEPTED, not fully passed. Original native/managed exit0 agreed; observer
completed290.697s, no residual targets; unchanged old artifacts. No retry allowance.
See docs/roadmap/2026-11-09-windows-c5-fixed-entry-native-execution.md. Historical
shutdown cause, C4/RTL, C6 and broader acceptance remain unchanged.

## C5 Unknown→Active restoration — 2026-11-09

Separately approved single launch: PASS for Unknown unavailable values followed by
Active numeric restoration12.4/100GB and2/5, title inactive.110 complete pairs,
native/managed exit0, no residual target. No Stale rerun; original uninterrupted
Active→Unknown→Stale→Active route is not implied. No broader acceptance promotion.
See docs/roadmap/2026-11-09-windows-c5-active-restoration.md. No runtime allowance remains.
