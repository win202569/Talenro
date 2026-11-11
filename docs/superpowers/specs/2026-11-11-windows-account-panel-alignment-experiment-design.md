# Account panel alignment experiment — short design

Status: approach 1, written design and inline implementation plan approved by user; execution blocked at baseline verification, before candidate edits.
Scope: small, single-variable candidate experiment, not an accepted fix.

## Intent and evidence

Test whether Account panel's current Stretch arrangement contributes to horizontal
overflow in rtl-mixed. Existing run 55550244cd6449eab7d870733015eb23 recorded panel
overflow 122 DIP and email overflow 82 DIP. Email fits within the panel. RTL root
coordinates explain why recorded right overflow corresponds to screen-left clipping.
The exact arrange cause is still a hypothesis, not proven. See
[read-only analysis](../../roadmap/2026-11-11-windows-email-geometry-readonly-review.md).

## Approaches and selection

1. Selected: explicitly center AccountGeometryPanel. One declared layout property
   changes; retains MaxWidth=820, Padding=40, Spacing=20 and LTR email direction.
   Tradeoff: removing Stretch may change actual width and wrapping as well as position.
   These are effects to measure, not additional independently edited variables.
2. Not selected: remove MaxWidth while retaining Stretch. This tests a different
   constraint and changes wide-window presentation; do not combine it with option 1.

## Candidate and boundaries

Only application change: add HorizontalAlignment="Center" to the existing named
StackPanel in clients/windows/Talenro.Demo.App/Views/AccountPage.xaml.
No Login, shell, text/font/direction, width, padding, scroll setting, model, event
handler, diagnostic adapter, evaluator, observer entry or process-safety changes.
This shared XAML affects ordinary builds too; it is an experimental development
checkpoint, not a diagnostics-only switch or release-ready change.

No installs, security changes, application launch, native rebuild, artifact rebinding,
main merge or live acceptance are authorized by this design. Keep existing fixed
binaries and raw receipts intact. Do not run the old observation entry against a
future changed binary or update its bindings merely to bypass a hash failure.

## Offline verification design

After written-design approval and implementation planning, add a failing source
contract test first, then the single attribute. Require exactly one named Account
panel with Center alignment. Preserve all other XAML via the existing canonical
baseline check: assert the new attribute specifically, remove only that attribute
on that node plus the already approved diagnostic names from an in-memory copy,
and compare to the original Account baseline hash. Do not replace the baseline hash
or broadly strip alignment attributes. Login baseline remains unchanged.

Regression coverage must reject missing/wrong alignment, alignment on the wrong
node, and any unrelated layout edit. Keep geometry evaluator, clip-unknown and
isolation checks. Test source fixtures must be current, not stale copied output.
The implementation plan will specify bounded test commands after inspecting test
project dependencies; no native application build/launch is part of offline checks.
Record actual test results, not expected results. Passing source/synthetic tests
proves the candidate's scope and contracts, NOT that WinUI clipping is repaired.

## Later native evidence and stopping rules

A native build and subsequent live observation need separately scoped authorization,
new isolated artifact paths/hashes and a matching inert observation entry. Preserve
the old artifact and run. No live budget is granted here (remaining allowance 0).

For any later approved comparable run, retain en-US/rtl-mixed and matching viewport,
scale, navigation and initial scroll position. Compare panel/email bounds, actual
and desired widths and visual text visibility, and retain normal-close evidence.
Do not call email containment alone success if panel or other content is clipped.
Unknown clip-chain coverage stays unknown; source tests cannot certify it. LTR,
narrow-window and broader C4 checks remain outstanding even if this one case improves.

If offline checks fail, stop and explain rather than loosen gates. If later native
evidence still overflows, worsens layout, or is not comparable, mark the hypothesis
unsupported/inconclusive; do not add width/direction changes or retry automatically.
If needed, revert only this candidate through a reviewed edit/commit, preserving
other work and all evidence. Even improvement supports this candidate, not a proven
framework-internal root cause or full C4 acceptance.

## Handoff

Builder may choose test organization and clear test names within these constraints.
Any extra application property, new diagnostic mode or wider verification budget
requires revisiting scope. Next after design approval: implementation plan, then
approved offline candidate work. Current disposition: not implemented, C4 NOT_ACCEPTED.
