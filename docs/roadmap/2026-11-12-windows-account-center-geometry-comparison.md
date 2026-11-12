# Account-center geometry comparison and remaining acceptance evidence

## Limited conclusion

The candidate removes the previously measured Account email overflow in the
observed en-US/rtl-mixed condition. This is evidence for the Center candidate,
not full visual acceptance or a proof of the internal native arrange mechanism.
C4 remains NOT_ACCEPTED; no new live allowance exists.

Read-only comparison used old run55550244cd6449eab7d870733015eb23 and candidate
run66db34b22cc748bbb8fdf61f656b879a. Both geometry files retain their recorded
SHA256; all8 raw manifest/process/geometry/claim files match inventories. Both
manifests' EXE/DLL hashes match their corresponding preserved artifacts.
No app launch, raw-record rewrite, build or source change occurred during analysis.
This subsequent checkpoint records the already reported comparison in portable docs.

## Integrity and comparability

Each geometry file parses to header +4 snapshots +health; schema1, all wrapper
run IDs and normalized sample GUIDs agree. Sequences1–4 are ordered with increasing
elapsed time. Login2/Account2, Samples4, Failed=false, ErrorCodes empty and
Provisional=true. Process records correlate with contiguous sequences: old71/new24.
Their persisted outcomes remain Finalized=false; separately recorded terminal
results after cleanup are Finalized=true. No promotion of provisional raw receipts.

Same language/profile, rendering/text scale1, zero horizontal/vertical offsets,
root1136x620 DIP. Login viewport1136x534; Account916x534. Corresponding root,
PageHost and presenter RootBounds and FlowDirection match exactly across runs.
Login Elements/Boundaries/Scroll/FieldStatuses match for both ordinal samples,
excluding run identifiers and timestamps. This is a useful unchanged-page control,
not a randomized or repeated-run experiment.

## Measurements

Coordinates are recorded RootDIP, not desktop pixel coordinates. Relevant boxes
have Status=Ok, matching spaces, finite positive dimensions; recorded boundaries
are complete individually. Insets use element.Left-boundary.Left and
boundary.Right-element.Right (and equivalent top/bottom differences), matching
GeometryEvaluator.Compare's known-box rule. No raw mirroring or rounding applied.

| Account measurement (both samples) | Old | Candidate |
| --- | ---: | ---: |
| Panel alignment | Stretch | Center |
| Panel actual width | 820 | 480 |
| Panel horizontal bounds | 438–1258 | 438–918 |
| Email horizontal bounds | 1031.095703125–1218 | 691.095703125–878 |
| Presenter horizontal bounds | 220–1136 | 220–1136 |
| Email right inset against root/presenter/explicit presenter clip | -82 | 258 |
| Email left inset against presenter/explicit clip | 811.095703125 | 471.095703125 |
| Email top/bottom inset against presenter/explicit clip | 98 / 409.3984375 | 98 / 409.3984375 |

Old Account email overflows all3 known boundaries by82 DIP; candidate is inside
all3. Horizontal right-inset improvement340 DIP equals panel width reduction340.
Panel candidate horizontal margins within presenter are218 DIP on both sides;
its recorded MaxWidth820 and Padding40 remain, while actual width equals desired480.
Email actual width186.904296875 and height26.6015625 are unchanged; its LTR direction
and Left text alignment remain unchanged. Panel/root remain RTL. Do not describe
RootDIP right as desktop right without the prior coordinate-frame qualification.

New Account samples are approximately15.9538ms apart; old samples12.7476ms apart.
Within-run agreement does not establish long-duration stability. Login remains
inside all known boxes; its original0.5 DIP early layout shift repeats unchanged.

## Remaining evidence, not an execution authorization

- Both runs retain ClipChainComplete=false: known-box inclusion does not prove
  absence of implicit/compositor clipping. Original Unreviewed/NotAccepted values
  remain untouched; this document supplies a separate limited analysis.
- Narrow-window layout, LTR and other rendering/text scales are not covered by
  this pair. No system scaling/settings change is proposed implicitly.
- The observer captures at most2 samples per page. Resizing after those samples
  are consumed cannot supply a new geometry condition; do not reuse a run's old
  samples as evidence for a later resized screenshot.
- The current fixed entry is en-US/rtl-mixed only. It must not be given unreviewed
  profile/hash/path overrides to broaden coverage. Any new route needs appropriate
  preparation and any additional launch needs explicit bounded approval.
- Full-project regression, dependency security clearance, startup deadline and
  broader product/C4 acceptance remain separate from this narrow geometry result.

Next recommended preparation is to define the smallest missing narrow-window case
and its capture timing before requesting a new launch. Do not replay the already
successful default-width condition without a distinct question. Preserve branch,
old/new artifacts, all raw receipts and failed restore snapshots. Fresh clone recovers
these sanitized findings and source, not the ignored machine-local evidence.

Sources: [baseline review](2026-11-11-windows-email-geometry-readonly-review.md),
[candidate live inventory](2026-11-12-windows-account-center-live-observation.md),
[build/binding provenance](2026-11-12-windows-account-alignment-native-binding.md).
