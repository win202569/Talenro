# Account-center narrow-window observation design

Status: approach 1 selected by user; written design awaiting approval. No live
authorization. Documentation only; no source, entry, build or raw-record changes.

## Intent and approaches

Determine whether the existing Center candidate's Account email stays within the
recorded boundaries when available page width is below its observed 480 DIP panel
width. Previous comparison covered only root1136x620, Account916x534 DIP.

1. Selected: reuse the fixed candidate and resize on Connection BEFORE first
   Account navigation. No rebuild; only Account receives the narrow condition.
2. Considered: a new fixed-startup-size artifact could capture Login and Account
   narrow, but requires source/build/hash binding work. Not selected.

This is a small observation design, not a product layout change or full C4 test.

## Fixed scope and proposed condition

Reuse `clients/windows/tests/observe-email-geometry-account-center.ps1` without
overrides and its existing preflight, process safety and time limits. Candidate:

- EXE SHA256 CF7C577C2FF3F4A690B2EFEE054411B59006210440333C5E8EA1838A7062A48F
- DLL SHA256 6011FA7BB63A45D5FAEB4A44529904FEC0857741A4E810B20B38555655C838C1
- en-US / rtl-mixed; rendering and text scale both1, as in the prior comparison.

Proposed by the assistant, pending design approval: root width640 +/-20 DIP,
height620 +/-10 DIP, measured by the diagnostic title and ultimately the records,
not outer-window desktop pixels. With the existing220 DIP sidebar this should
yield page width400–440 DIP, below480. Actual recorded page width must confirm it.
Do not change Windows scaling, text settings, language or security settings.
No explicit application sizing constraint was found in MainWindow/App startup;
this does not prove the platform will permit the proposed dimensions.

## Capture order for a separately authorized run

1. Recheck ordinary token, zero matching target/prototype processes, fixed hashes
   and existing entry preflight. If blocked, stop; do not loosen checks.
2. Only after separate bounded approval: one launch, zero relaunches/retries,
   existing six-minute outer budget unchanged. Start with Login, enter the demo
   and remain on Connection. Do not visit Account yet or initiate a connection.
3. Resize through normal window controls while on Connection. Observe the
   diagnostic title twice at least one second apart after releasing resize:
   matching dimensions within the proposed range, same scale and profile.
   This is a readiness check, not proof of settled Account geometry.
4. Navigate to Account once. Do not resize, scroll or navigate again. Capture
   title/screenshot evidence and inspect whether the email is fully visible.
   Every page has at most two ATTEMPTS; returning cannot replenish its budget.
5. Close normally with the title-bar X within the existing deadline. Record final
   console outcome/exit code separately from provisional persisted receipts.
   Never force-kill, dump, restart or extend the observation to obtain a pass.

If Account was visited too early, sizing cannot be achieved, the profile/scales
change, or UI interaction cannot be completed safely, mark the condition invalid
and close normally. Preserve all records; a new run needs new approval. A process
remaining after the deadline is a blocker for read-only investigation, not kill
authority. Do not represent completion of this design as an allowance to launch.

## Evidence and decision

After normal closure, read only: inventory hashes and correlate run ID, manifest,
process records and geometry. Require two Account snapshots with complete usable
fields, no capture errors, root dimensions within the proposed ranges, page width
below480 DIP and matching dimensions between snapshots. Check both snapshots;
do not select only a favorable one. Confirm actual scales/profile and zero scroll.
Login samples are startup-condition controls, NOT narrow-window evidence.

Compare email RootDIP bounds against every recorded usable boundary with the
existing evaluator rules. Report raw insets, panel/page widths and screenshot
observations separately. Any known-boundary overflow or visibly clipped email
fails this narrow condition; missing/inconsistent data makes it inconclusive.
Known-box inclusion plus full visible email supports only this observed condition.
ClipChainComplete=false, provisional fields and original NotAccepted values are
not rewritten. No full clipping-chain, long-duration, LTR, other-scale, regression,
security or C4 acceptance follows. Do not equate RootDIP right with desktop right.

## Handoff and next gate

Obtain approval of this design, then prepare the execution checklist. Request
separate one-launch/zero-retry authority when ready; current live allowance is0.
This document alone is not an executable instruction. Fresh clones recover source
and this design, not ignored binaries/receipts; missing artifacts require their own
preparation, never replacement of fixed hashes just to pass preflight.

References: [comparison](../../roadmap/2026-11-12-windows-account-center-geometry-comparison.md),
[prior live inventory](../../roadmap/2026-11-12-windows-account-center-live-observation.md).
