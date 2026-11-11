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

## Follow-up: RTL coordinate trace

Read-only source and existing-receipt review, no application execution:

- MainWindow sets Root.FlowDirection=RightToLeft for rtl-mixed, then passes that
  exact Root to EmailGeometryObserver. The snapshot confirms root, host, page,
  scroll, presenter, panel and recorded extra ancestors are RightToLeft; email
  alone is LeftToRight. Root width is 1136 DIP.
- TransformToVisual(root) produces coordinates relative to that specified
  element, not desktop coordinates. Microsoft documents that an RTL element's
  local coordinate origin is at its upper right. Thus the prior right/left
  discrepancy is consistent with root-local versus screen-oriented axes, not
  evidence that the receipt is corrupt.
- Derived root-relative left-origin interval is [1136-right,1136-left]:
  Account email [-82,104.904296875], panel [-122,698], presenter [0,916].
  Login snapshot 2 email becomes [446.5,614.7138671875]. These are arithmetic
  interpretations of existing bounds, NOT newly measured desktop rectangles.
  Account left overflow of 82 DIP agrees directionally with the visual evidence;
  the screenshot has not been independently pixel-calibrated to verify magnitude.
- Account email is inside the panel's own horizontal bounds, ending 40 DIP before
  the panel's RootDIP right edge (padding 40). The panel itself exceeds presenter
  by 122 DIP, so the evidence points upstream of the email text's box placement.
  Do not change the email to RTL or alter raw coordinate values as a repair.
- Working Login panel explicitly uses HorizontalAlignment=Center, width 307;
  Account panel uses default Stretch, MaxWidth=820, actual width 820 and desired
  width 480 within a 916-wide viewport. These differ in several variables; Login
  is not a controlled experiment proving alignment is the cause.

The precise arrange/centering mechanism remains unproven: no per-ancestor affine
matrices or native arrange inputs were captured. Existing adapter contract tests
cover synthetic bounds/clips and nonvisual XAML naming, not native RTL arrangement.
Next useful work is a bounded design for an isolated alignment/width experiment
with offline regression checks first; native execution still needs new explicit
approval. Do not treat repeated document review as further runtime validation.

Official references inspected for this interpretation:
- [FlowDirection: RTL local coordinate frame](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.frameworkelement.flowdirection?view=windows-app-sdk-1.8)
- [TransformToVisual: specified target coordinate frame](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.uielement.transformtovisual?view=windows-app-sdk-1.8)

Geometry receipt SHA256 remains AE5ABE46357E8E5AB2DF100A09843CB6C62BB4E9E8B865D226FE5619AB97DB1A.
No source/tests/raw records changed. C4 remains NOT_ACCEPTED.
