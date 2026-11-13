# Windows remaining acceptance — reconciled 2026-11-12

2026-11-13 addendum: [new keyboard-only run](2026-11-13-windows-keyboard-followup-observation.md)
closes the isolated auto-scroll and keyboard-dialog-close visible return-focus
questions on this candidate at650x620. Together with the lower-controls run this
supports the scoped R1 path; not all dimensions or ordinary-build acceptance.
Mouse-close immediate outline remains unverified, not a confirmed defect. R2–R7
and broader limitations remain; no additional launch authority. Older text below
records its original evidence baseline, superseded only for these two questions.

Latest addendum: [lower-controls run](2026-11-12-windows-account-lower-controls-observation.md)
partially covers R1 (lower labels/access/traversal) and selected logout recovery in
R2 on the current candidate. Immediate visible dialog-return focus and isolated
keyboard auto-scroll remain unverified. Other gaps below are unchanged; no new
launch authority or overall acceptance. Original reconciliation baseline follows.

Documentation-only reconciliation at ec8bb093. No app launch, test rerun, build,
restore, settings change or raw-record rewrite. This is an evidence map, not a new
execution plan or acceptance approval. Live allowance0; C4 and overall native
checklist remain PARTIAL / NOT_ACCEPTED. Task11/12 and44/34/6/217 are not promoted.

## Current candidate: supported findings

Candidate EXE/DLL identities and independent preparation review are recorded in
[build/binding provenance](2026-11-12-windows-account-alignment-native-binding.md).
Do not generalize findings to rebuilt binaries or the ordinary build.

| Scope | Evidence | Supported conclusion and limit |
| --- | --- | --- |
| Default-width Account email, en-US/rtl-mixed, scales1 | [Old/new comparison](2026-11-12-windows-account-center-geometry-comparison.md), candidate run66db34b2 | root1136x620, page916x534; recorded right inset -82 to+258 DIP; no longer overflows known boxes |
| Narrow Account email, same profile/scales | [Narrow observation](2026-11-12-windows-account-center-narrow-observation.md), runf86a7c0d | root640x620, page420x534; two Account samples, visible complete email, known-box right inset70 DIP |
| Normal closure of these runs | Same run reports | Original observations ended normally; no residual targets then. Not proof the historical C3 shutdown defect is resolved |
| Targeted tests and independent review | Build/binding record | Historical scoped assertions and static review passed; no fresh full-suite or independent test rerun in this reconciliation |

Runef33d392 did not obtain the narrow condition and remains preserved, not a pass.
Login snapshots are not narrow-window evidence. Coordinates are RootDIP, not
desktop orientation. Raw Unreviewed/NotAccepted and provisional fields stay intact.

## Current-candidate gaps and what would close each

| ID | Remaining item | Required evidence / boundary |
| --- | --- | --- |
| R1 | Account lower controls and scrolling at narrow width | Observe access to lower actions and focus/scroll behavior at measured size. Panel height726 versus viewport534 alone is not a defect; the recent email-only run did not scroll. Do not infer whole-page fit |
| R2 | Current-build C4 route beyond email | Map inherited English/mirrored layout, technical-ID direction, fixture codepoints and logout/profile-retention checks to affected source before deciding the smallest retest. [2026-11-10 C4](2026-11-10-windows-c4-native-execution.md) passed these subchecks on an older artifact; not a current-build rerun |
| R3 | Complete clipping/longer stability evidence | Both runs have ClipChainComplete=false and near-adjacent snapshots. Either obtain an adequate measurement method under a new scope or retain the limitation; screenshots/known boxes cannot silently close it |
| R4 | LTR/other language and rendering/text-scale conditions | Define specific missing cases and appropriate fixed artifact/entry binding first. Existing account-center entry is en-US/rtl-mixed only; no parameter/hash override or OS-setting change is authorized |
| R5 | Requirement viewport coverage | Measure four-page/lower-action/dialog behavior at required1024x768 and1440x1024 logical content sizes. Neither1136x620 nor640x620 satisfies these sizes. Available desktop feasibility needs fresh evidence, not old display arithmetic |
| R6 | Independent dimension/DPI corroboration | Title and geometry agree but come from the same application. Original checklist asks for independent client/DPI evidence; do not label the current agreement independent |
| R7 | Regression on the intended deliverable | Select relevant full/ordinary-build checks after source/artifact impact mapping. Diagnostic candidate evidence does not automatically qualify ordinary default behavior or a newly rebuilt deliverable |

R1–R7 are missing evidence, not seven newly proven defects. They are not all
indispensable to the already-supported narrow email finding; they limit broader
claims. No source fix is justified solely by this list.

## Broader historical gaps, kept separate from the email fix

- Screen-reader speech and semantics, contrast and enlarged text remain open in
  the [methods review](2026-11-08-windows-accessibility-layout-methods.md). Source
  declarations/accessibility trees do not prove actual spoken output or contrast.
- Full keyboard coverage excludes expanded dropdown/disabled-state combinations;
  selected K1–K6 paths passed historically. Internal confirmation atomicity and
  exhaustive reset/storage behavior remain unverified beyond selected sequences:
  [coverage map](2026-11-08-windows-client-acceptance-coverage.md).
- C3 long-country UI observations passed selected checks, but its normal shutdown
  failed historically. Later normal exits do not identify or close that cause.
- C5 has scoped Unknown/Stale display and separate Unknown→Active restoration
  evidence; no uninterrupted Active→Unknown→Stale→Active route is implied.
  C6 default-after-profiles remains NOT_RUN in the reconciled records. See
  [manual checklist](../../clients/windows/tests/manual-checklist.md) and
  [Stage B record](2026-11-08-windows-layout-stage-b-continuation-execution.md).
- NU1905 offline vulnerability-data warning remains: build success is not dependency
  vulnerability/signature clearance. Runtime network/storage audit, production
  readiness and main merge are outside this narrow validation.

## Recommended order, not authority to execute

1. Next read-only task: map current source changes to R1/R2/R7 and the ordinary
   deliverable. Produce a minimal missing-evidence checklist, preserving old passes
   with their original baselines; do not repeat successful email scenes by default.
2. Resolve feasible evidence methods for R3–R6 and accessibility before proposing
   live runs. Do not add instrumentation, rebuild, change settings or request a
   broad launch budget merely to fill the table.
3. Only then seek a bounded authorization for the selected missing native checks.
   Include visibility confirmation before interaction, normal close, zero automatic
   retries and current fixed-binding/process checks. No allowance is granted here.

Preserve reports, both narrow attempts and local raw evidence. Fresh-computer
recovery starts from this development branch; Git restores source and sanitized
findings, not ignored artifacts, receipts, caches or environment acceptance.
