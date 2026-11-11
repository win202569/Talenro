# Isolated email geometry diagnostics

Status: approved isolated native build and artifact inspection completed2026-11-11;
see docs/roadmap/2026-11-11-windows-c4-email-geometry-build.md for verified hashes.
The offline build has NU1905 (no local vulnerability audit data), not an audit pass.
Fixed entry source now exists at tests/observe-email-geometry.ps1; its synthetic
verification does not authorize live execution, which requires separate approval.
Independent review found unresolved end-to-end failure-category loss through the
shared session and entry catches. Entry readiness is BLOCKED, not ready for a live
approval: first design and test entry-owned sanitized failure bookkeeping without
changing frozen shared modules or weakening original Process/SafeHandle ownership.
C4 remains NOT_ACCEPTED. Portable tests alone do not compile the WinUI adapter.

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

## Fixed entry (no live run authorized)

No-argument invocation returns NotStarted before imports, probes or writes. The only
public switch is RunApproved; it is not evidence of approval and must not be used
without a separate bounded-run decision. No path/hash/profile/run-ID overrides exist.
The entry binds the two verified new hashes, generates one run ID and checks ordinary
token, branch/clean/source identity, hashes, conservative process inventory and both
receipt destinations. It reuses the existing shared observation engine, not a copied loop.

Process records:email-geometry-observations/<run>/manifest.json and samples.jsonl.
App record:email-geometry-records/<run>/geometry.jsonl. Both under the fixed artifact
directory, correlated by run ID. A missing/partial/unhealthy geometry record or absent
samples cannot be accepted merely because process exit0 was verified; result stays
GeometryEvidenceStatus=Unreviewed and Acceptance=NotAccepted.

After future approval: one launch/no retry, startup observed within30s, Login then
Account at scroll-top, no search/connection/node/settings change; normal titlebar close
by300s, exit assessed within30s, total observation cap360s. Operator/UI evidence and
retained original Process/SafeHandle evidence are both required. No forced termination.
These are operator obligations, not implemented automatic navigation/close timers.
