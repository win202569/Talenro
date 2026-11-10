# C4 email geometry probe — short design

Status: user selected approach1 after42c0a97f; this written design awaits approval.
Design only: no implementation/build/test/native launch authority. Live allowance0.

## Intent and approaches

Locate which layout boundary puts the actual Account email outside the visible
area, without simultaneously fixing or simplifying the failing page. Prior native
run f6071741fab546ef8f856467def00375 shows clipping, not its exact layout cause.

- Selected: instrument the real shared pages in an independently built diagnostic
  variant. Preserves actual surrounding layout; costs a separately authorized build
  and bounded observation, and instrumentation can perturb timing.
- Alternative: minimal reproduction window. Easier variable isolation but can omit
  the real parent/template trigger; reserve it if real-page data stays inconclusive.
- Not a repair: do not add Center, offsets, widths, wrappers, change email direction,
  or make global RTL changes merely to produce a better screenshot.

This is a small diagnostic addition. The proposals below are executor design choices
for review, not prior user requirements. Keep existing C#/WinUI and PowerShell tools.

## Isolation and activation

Propose a dedicated opt-in Debug/layout-only compile gate, default off, incompatible
with close-stage recording/verification modes. Separate app obj/bin directories
under an email-geometry diagnostic namespace; no existing output directory reused.
Ordinary/Release artifacts must exclude probe code and recording behavior. Preserve
current Account/Login geometry, fixed text, event handlers and state transitions.
Nonvisual names/references may identify targets; they must not introduce layout
containers, focusable controls, bindings or user-visible probe UI.

Probe activation is fixed en-US/rtl-mixed plus a canonical nonzero lowercase run ID.
No arbitrary target/path/command selection. Normal layout options and C4/C5 entries
keep their existing contracts. App records and process observer share one run ID.
The plan must specify reject-before-launch activation and create-only ownership.

Existing C4/C5 entry files, shared process/native/identity modules and old fixed
artifact hashes remain unchanged. A later dedicated fixed entry for the new variant
must reuse shared observation logic, not copy the loop or relax source/hash/token
checks. Bind it to the new artifact only after a separately approved build yields
verified hashes; missing hashes mean NOT_READY, never an arbitrary path fallback.

## What to measure

Compare the actual Login email (control sample) and Account email, not the synthetic
heading's embedded address. Fixed target roles: root, PageHost, active page, scroll
viewport, content panel, actual email, and the relevant content presenter/ancestors
when discoverable. Missing or ambiguous template targets remain unavailable.

One snapshot on the UI thread contains:
- schema/run ID, sequence, monotonic elapsed time and page/target role;
- loaded/visible state; actual/desired sizes and local layout slot where available;
- local bounds transformed to one explicitly identified root coordinate space,
  in DIP, with normalized min/max edges (including mirrored-coordinate handling);
- effective flow direction, horizontal/text alignment, padding/margin and width
  constraints needed to interpret the relevant targets;
- scroll offsets/viewport dimensions, root/render/text scale;
- known viewport and explicit clipping bounds transformed into that same space,
  with individual acquisition status and completeness, not an invented full clip chain.

Do not mix screen pixels, parent-local layout slots and root DIP in containment
comparisons. Layout slots retain their own space label. Snapshot does not include
account text, passwords, credentials, process command inventories or exception messages.
Use roles, enum/numeric values and sanitized failure categories only.

## Acquisition and evidence limits

Propose at most two snapshots per target page, four total per process, after that
page is attached and layout has been observed. Bound event subscriptions and queued
callbacks; detach on completion/close. No timer, polling loop, forced UpdateLayout,
scrolling, resize, UI input, model dispatch or automatic page navigation by the probe.
If a second eligible sample never arrives, retain missing/unstable evidence; never
force layout or infer stability from a delay. Counts/scheduling detailed in the plan.

Use a bounded in-memory record buffer; keep disk IO out of layout callbacks. Flush
through an explicit owned create-only sink at a documented lifecycle point, without
using file presence as proof that all samples were captured. Fixed trusted output
root, unique run directory; reject collisions/reparse ancestors, no overwrite/delete.
Partial/failed flush remains incomplete evidence. Preserve primary errors and attempt
all acquired-resource cleanup; no recording failure may become a successful verdict.
The plan must resolve crash/close durability and finalization before implementation.

Independent geometry evaluator reports per-boundary Inside, Overflow (signed edge
distances), or Unknown. Invalid/nonfinite/missing/zero-area/mixed-space inputs cannot
pass. Preserve raw distances; any numeric tolerance is explicit and tested, not
silently used to conceal visible clipping. Unknown clip ancestry prevents claims
of complete visibility. Two equal samples establish only observed consistency.

Element-box containment is not glyph containment or screenshot readability. Geometry
may narrow cause but cannot prove a repair, compositor correctness, natural-language
shaping, or a general WinUI bug. Correlate measured page/profile with contemporaneous
supported UI evidence; do not reuse old screenshots as proof of new geometry.

## Verification and later runtime boundary

Implementation plan: test first for pure geometry/validation, activation, bounded
capture lifecycle and receipt ownership using memory doubles. Cover fully inside,
left/right/top/bottom overflow, mirrored transforms, missing targets, invalid values,
space mismatch, unstable samples, callback/flush/disposal failure and output collisions.
Static declaration checks alone cannot reproduce rendering. Include isolation checks
for invalid build-mode combinations and unchanged old observer/runtime files.

After implementation review, separately approve a build into new output paths and
verify old hashes unchanged. Do not build merely to complete this design/plan.
Only then propose a fixed-hash native route: one launch/no retry, startup30s, cap360s,
normal close by300s, exit assessment30s; Login then Account at scroll-top, no search,
connection, node selection or settings change. Operator observation and original
Process/SafeHandle evidence stay mandatory. No run is authorized by this proposal.

Success for diagnosis means usable correlated bounds identify the first measured
overflow boundary, or explicitly establish that available box geometry is insufficient.
If clipping disappears in the diagnostic variant, report non-reproduction; do not
call it fixed or overwrite the preserved failing artifact. A repair requires its own
scope and regression/native evidence. C4 stays NOT_ACCEPTED until separately resolved.

## Handoff

After written-design approval, prepare implementation tasks with exact file map,
interfaces, callback limits, serialization/cleanup contract and test commands.
Source/tests/design/plan/progress/recovery are committed to the existing development
branch; raw receipts/screenshots/binaries stay ignored. No main merge or broader
acceptance. Remaining uncertainty is measured layout behavior, not approval to guess.
