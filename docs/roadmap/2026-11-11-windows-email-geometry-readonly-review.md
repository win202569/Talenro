# Email geometry: read-only evidence review (2026-11-11)

Run `55550244cd6449eab7d870733015eb23`; follows the
[observation record](2026-11-11-windows-email-geometry-first-observation.md).
No launch, retry, build, application/source change, raw-record rewrite or gate relaxation.

## Record checks

- Geometry JSONL parses as six records: header, four snapshots, health footer.
  Schema 1; wrapper run and normalized snapshot GUID agree. Snapshot sequences
  1–4 are ordered, with increasing nonnegative elapsed times; Login 2, Account 2.
- Health reports Samples=4, Failed=false, ErrorCodes=[], Provisional=true.
  This is receipt health, not visual acceptance or proof of a complete clip chain.
- Process JSONL parses as 71 records: start 1, native-identity 1, pair 68,
  outcome 1. All run IDs agree and sequences are contiguous 1–71.
  Persisted outcome is ExitVerified=true, Finalized=false; user-supplied final
  terminal output separately reports Finalized=true and entry exit code 0.
  Do not rewrite the persisted provisional record to match terminal output.
- Manifest matches the previously recorded EXE/DLL hashes and en-US/rtl-mixed
  arguments. All four raw-record hashes match the observation inventory after
  review. No raw evidence is committed or restored by cloning this repository.

## Measurements and limits

Coordinates below are RootDIP, not screenshot pixels or parent layout slots.
GeometryStatus 0 means Ok; 4 means Unsupported. Relevant bounds are Ok.
RenderingScale and TextScale are 1; horizontal/vertical scroll offsets are 0.

| Snapshot | Page | Email horizontal bounds | Presenter horizontal bounds | Known-boundary result |
| --- | --- | --- | --- | --- |
| 1 | Login | 521.7861–690 | 0–1136 | Inside |
| 2 | Login | 521.2861–689.5 | 0–1136 | Inside |
| 3 | Account | 1031.0957–1218 | 220–1136 | Right overflow 82 DIP |
| 4 | Account | 1031.0957–1218 | 220–1136 | Right overflow 82 DIP |

Account email top/bottom are 141–167.6016, within presenter 43–577.
Root horizontal bounds are 0–1136; explicit presenter clip equals presenter
bounds. Thus Account has the same 82 DIP right overflow against all three
recorded boundaries. Its panel spans 438–1258 (width 820), extending 122 DIP
beyond the presenter right edge. Email width is 186.9043, desired width 187.
Both Account samples match, but are only about 12.75 ms apart: this is not proof
of long-term stability. Login panel/email shift 0.5 DIP between its samples.

All four snapshots have ClipChainComplete=false. Email Width/MaxWidth are
Unsupported, not zero; they cannot establish an explicit width constraint.
Inside known Login boxes does not prove all clipping absent. The visual record
separately showed a fully visible Login email and left-side Account clipping.

## What is and is not established

The recorded Account email box exceeds known RootDIP boundaries. This supports
a layout/coordinate issue, but does not yet explain the screen-side clipping:
the screenshot showed clipping on the LEFT while recorded overflow is RIGHT.
The adapter calls TransformToVisual(root), transforms four corners, and labels
the normalized result RootDIP. Page/panel inherit RightToLeft while email is
LeftToRight. RootDIP must not be silently equated to screen orientation under
that mixed-direction tree. No manual mirroring correction has been applied.

Account XAML has panel Padding=40, MaxWidth=820 and an LTR email TextBlock.
These are investigation inputs, not proof that MaxWidth or text direction is
the defect. Missing compositor/implicit clipping also remains unknown.

Next narrow step: read-only trace of root direction, ancestor transforms and
coordinate adapter contract; determine how RootDIP relates to screen orientation
before proposing a minimal repair. Do not change app layout or diagnostic
coordinates based solely on this discrepancy. No new live allowance exists.

Review disposition: limited structural and known-boundary analysis completed;
visual root cause unresolved. Existing Unreviewed/NotAccepted receipts remain
unchanged; C4 remains NOT_ACCEPTED. Startup-within-30s remains unproven.
