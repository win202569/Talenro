# Ordinary candidate minimal native observation design

Date2026-11-13. Proposed design only, awaiting approval. No entry implementation,
synthetic tests, build, restore or launch performed/authorized by this document.

## Purpose and approaches

User requested a minimal ordinary-version run plan bound to the inspected binary.
Recommended small adaptation: a dedicated fixed entry based on observe-c4.ps1,
reusing unchanged c5-observation-session.psm1 and its process/identity modules.
Existing C4/geometry entries pass diagnostic arguments and cannot be used directly.
Manual direct launch lacks the existing retained-handle receipt guarantees;
adding in-app geometry would cease to test this ordinary artifact. Neither is
selected. Exact new entry names/budget/route below are proposals for approval.

No product behavior changes, new instrumentation or shared observer refactor.
Future files: clients/windows/tests/observe-ordinary-account-center.ps1 and its
focused .tests.ps1. Original entry/tests/modules remain unchanged.

## Fixed artifact binding

Build source5ee8a339b574171fc61f3cf0c0c636bd86421397; static verification at1680985e,
report checkpointbe67d633. Base relative to repository:
clients/windows/Talenro.Demo.App/bin/ordinary-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/.

| Bound file | SHA256 |
| --- | --- |
| Talenro.Demo.App.exe |FEDA8E835E8390ED7F166FA764413FBEC9BCE591C7C7CF7E8E43E028B03C0375|
| Talenro.Demo.App.dll |773AE00EC4E5552D65ECDFA527FEEE6688AE380D8ED4DB4D3282DD10C1B76A0B|
| Talenro.Demo.Core.dll |67FDB3F42E69350400B8A9FC11BCFB54670233B4571F2E27BA91F30E8D19DE18|
| Views/AccountPage.xbf |A50DDC403AD9E41EA977AE868551AE25B6079674094EBE0BE38CD12C2B7EFE6F|
| Talenro.Demo.App.deps.json |99DD4D70751D66997DF9DA858C09DDA2BB68A8FDFAC7F671E4FAB5B35D910ED6|
| Talenro.Demo.App.runtimeconfig.json |3FA241FAED5CC59579FA3B630E1A873E5539428B9CED9A4CA48F5828488E9666|

No path/hash/argument/timeout override parameters. Hash mismatch stops before
launch; never rebuild or substitute an older ordinary/diagnostic artifact.
Bind source evidence to the recorded build commit (App/Core/Directory.Build.props)
and require no tracked differences in those paths; observer sources separately
match HEAD blobs. This allows documentation/new-entry commits, not unnoticed
product edits. Fixed hashes are not an adversarial full-install signature audit.

## Entry and safety contract

- Only public switch RunApproved; without it return NotStarted before imports,
  environment probes, receipt writes or process acquisition.
- Arguments empty; omit Start-Process ArgumentList entirely, not an empty array
  or diagnostic flags. Fixed expected command line is quoted absolute EXE only,
  with no trailing argument separator. Preserve strict command-line identity;
  if host serialization differs, abort and record it, never relax/retry live.
- Same ordinary-token, correct branch, clean-worktree, observer-source, native/
  managed/CIM original-handle identity guards; successful no-target/prototype
  process query. Validate fixed artifact/receipt paths and existing ancestors
  against reparse points. No fallback on query/identity/hash failure.
- One separately approved visible launch; zero retries;360000ms outer cap,
  max sleep2000ms. Begin normal titlebar close by300s, earlier if unusable or a
  check fails. Preserve original session; no force-kill, service change, dump,
  detached replacement or automatic second run. Timeout is not verified exit.
- Receipt root fixed under candidate directory: ordinary-account-observations;
  fresh nonzero RunId directory, CreateNew manifest/samples, no overwrite.
  Manifest Observation=OrdinaryAccountCenter, all six hashes, empty arguments,
  source commit, cap and default-environment language/profile (not forced English/
  RTL). Do not invent geometry receipts; geometry recorder is absent.
- Retain failure/cleanup semantics of chosen baseline; inspect and test any
  necessary wrapper differences. All writes local ignored; original receipts
  untouched. Final process success is distinct from manual UI acceptance.

## Minimal proposed operator route — one session, partial results allowed

1. Confirm the actual window is visible within30s. Record default language,
   direction and title; no LAYOUT TEST marker. Login fixed demo fields/offline
   notice visible. No real credentials or external actions.
2. Enter demo, observe disconnected/lock-off Connection; navigate Account without
   connecting or opening Nodes. At default size inspect complete email, usage,
   device count and account labels. This is ordinary resource behavior, not RTL
   stress coverage or locale fallback acceptance.
3. If timely, user narrows only horizontal edges; retain height, no system scale
   changes. Record observed window/screenshot dimensions only as outer bounds,
   never label them root/client DIP. Ordinary title has no size readout. Mark
   exact620–660/1024x768/1440x1024 conditions UNVERIFIED; do not borrow old titles.
   If resize cooperation is unavailable, retain default-only evidence and skip
   narrow check; no second launch or prolonged resizing attempts.
4. On Account, no pointer scrolling first. Tab from visibly established focus
   through MFA, Manage devices, Language, Delete (focus only), Exit (focus only);
   observe automatic scroll and outlines. Shift+Tab back to Manage devices;
   Enter opens one placeholder dialog; keyboard Close, confirm returned outline.
   Use actual localized labels, not hardcoded English or stale focus metadata.
5. If time remains, activate only Exit demo/reset, verify Login/sidebar hidden;
   enter once and observe disconnected/lock-off state. No exhaustive reset claim.
   Never activate Delete. No connection/node/country-lock/entitlement scenarios.
6. Close normally by300s regardless of incomplete checks; observe retained-handle
   exit and report per-check PASS/FAIL/UNVERIFIED. Preserve original persisted
   pre-cleanup versus console-finalized distinctions. Then read-only residual
   query and evidence/hash inventory, no launch retry.

Ordinary reserved-option rejection needs a separate process invocation; excluded
from this single-launch route. Other dialogs, expanded dropdowns, screen readers,
contrast, enlarged text, exact viewport/DPI and historical shutdown diagnosis are
not silently included. NU1905 and earlier valid-guard test Minor remain unchanged.

## Before any live authorization

After design approval, prepare detailed implementation/test checklist. Entry must
first pass synthetic-only tests: inert default, all six exact bindings, genuinely
absent ArgumentList, strict quoted no-arg context, manifest consistency, changed
source/hash/path rejection before process acquisition, existing/reparse receipt
failure, cleanup/error propagation, clock deadline and no retry/termination.
Reuse existing test helpers and shared-module suites; no real target launched by
tests. Independent scoped review then verifies differences from baseline and
that all safety gates are retained. Entry implementation approval is not live
approval. Only after committed clean preparation may user grant one run and a
continuous cooperation window. No run command is provided at this design stage.

## Completion boundary

This proposal closes neither runtime nor overall acceptance. A future successful
run can support only its observed default/narrow qualitative account route and
normal exit on this exact ordinary Debug candidate. No Release/production or
C4/security promotion. Preserve necessary source/tests/sanitized records on the
development branch; no main merge or raw binary/receipt commit.
