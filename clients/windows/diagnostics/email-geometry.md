# Isolated email geometry diagnostics

Status: source implementation only. Portable tests do not compile the WinUI adapter.
No new native artifact, verified hashes or runnable fixed observation entry exists.
Native build and live execution require separate approval. C4 remains NOT_ACCEPTED.

The default-off `EnableEmailGeometryRecording` property permits Debug layout diagnostics
only, incompatible with close-stage recording/verification. App outputs use
`obj/email-geometry/` and `bin/email-geometry/`; Core paths are unchanged. Ordinary,
layout-only and Release versions exclude the probe. Existing C4/C5 runners and old
artifact hashes are unchanged. Do not substitute new hashes into old runners.

Activation requires exactly the fixed layout-test, en-US, rtl-mixed arguments plus
`--layout-email-geometry-run=<nonzero lowercase32 GUID>`. This is a schema description,
not an authorized launch command. Session output is under the new artifact's directory,
`email-geometry-records/<run>/geometry.jsonl`, with a persistent `<run>.claim` reservation.
Existing records/directories/claims reject; reparse ancestors reject. Trusted parent
directories are required; metadata checks are not protection against hostile parent
replacement races. No deletion or retry is automatic.

On naturally occurring layout events, capture at most2 attempts per Login/Account and
4 total, one physical queued callback. Navigation invalidates stale captures without
releasing the queue slot until that callback drains. No forced layout,
input, scrolling, focus changes, navigation, model actions or timer. Names identify the
real email controls; no text is read or recorded. Failed attempts are not refunded.

Snapshots contain allowlisted roles/properties/statuses only. Four corners are transformed
into RootDIP and normalized; parent-local layout slots remain separately labeled.
Unknown/error/nonfinite evidence cannot pass. Signed edge comparisons are box containment
only, not glyph visibility. Presenter bounds and supported explicit rectangular clips
are recorded for all measured roles, including root/presenter, as distinct `<role>:clip`
boundaries; implicit template/compositor clipping is not assumed. ClipChainComplete
remains false. Missing/ambiguous presenters stay missing, and absent second samples do
not imply stable layout. Render/text scale and scroll measurements carry field statuses.

Add buffers validated copies in memory. Closed stops/detaches handlers, then Finish
serializes UTF8/no BOM and synchronously writes, flushes and disposes. All cleanup is
attempted; failure produces exit3 (startup rejection exit2). No background writer.
Synchronous IO can block; the future observer's bounded-overrun policy remains required.
A footer is provisional. Accept transport evidence only with full schema/correlation
and independently verified original-process normal exit0. Zero samples are explicit;
crash/missing footer/partial write are incomplete. File presence alone proves nothing.

Portable checks cover evaluator, scheduling state, activation, receipts, injected failures,
MSBuild evaluation/validation (not Build/Compile/Restore), and unchanged normalized XAML.
They do not verify WinUI event order, template geometry, native compilation, actual output
metadata isolation, clipping reproduction or repair. Those remain later-stage checks.

API references: [TransformToVisual](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.uielement.transformtovisual)
and [LayoutInformation](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.controls.primitives.layoutinformation).
