# C4 Email Geometry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task after plan/method confirmation. Steps use checkbox syntax. Sequential inline implementation with one final independent review is recommended.

**Goal:** Add an isolated, bounded real-page geometry probe that can narrow Account email clipping without changing the failing layout.

**Architecture:** Pure records and geometry evaluation are shared with the net10 test project. A UI-independent capture budget and owned receipt session feed a thin WinUI adapter, compiled only in a new diagnostic variant. New-artifact build/hash binding and native execution are later approval gates, not part of implementation completion.

**Tech Stack:** Existing C#/WinUI, System.Text.Json, xUnit and PowerShell7; no new packages.

**Spec:** docs/superpowers/specs/2026-11-10-windows-c4-email-geometry-design.md, user approved aftere9d85324.

Status: user approved inline implementation with final independent review on 2026-11-11.
Initial baseline aborted after VSTest's 90-second testhost connection timeout.
One approved ordinary-host comparison then passed286/286, exit0, administrator=False,
unchanged command/timeout. Task1 now implemented, full portable suite314/314 passed.
Tasks1–3 source implementation and final independent review/fix pass complete;
full portable suite372/372 passed, existing synthetic462 passed. Native adapter
compilation/output metadata/live observation remain unverified, separately gated.
See docs/roadmap/2026-11-10-windows-c4-email-geometry-execution.md. Live allowance0.

## Global Constraints

- Worktree D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator; existing codex/c12-b01-task9-coordinator branch; PowerShell only.
- No Center/offset/width/wrapper changes, text changes, input, focus changes or model dispatch. Preserve email LTR/root RTL and existing handlers.
- At most two snapshots per Login/Account page and four total per process. No timer, polling, forced UpdateLayout, forced navigation or layout stabilization.
- Use one labeled root-DIP coordinate space. Unknown/missing/mixed-space/invalid evidence never passes; box containment is not glyph readability.
- No edits to existing observe-c4/observe-c5 entry/tests or shared process/native/identity modules. No silent hash-policy changes.
- Preserve old layout-diagnostics/close-stage outputs and fixed EXE/DLL hashes. New diagnostic App obj/bin namespace only. No native App build or launch in this implementation stage.
- Proposed implementation authorization includes compiling/running the portable net10 test project with --no-restore, not building WinUI App. Missing SDK/packages/assets block tests; no installation or restore fallback without approval.
- No raw receipts/screenshots/private data/binaries committed. Ordinary dev checkpoints/push per AGENTS, no main merge/PR or acceptance promotion.

## Review Focus

1. Mirrored coordinates or unavailable clipping ancestry can falsely suggest visibility: preserve labeled spaces and raw signed distances; unknown evidence cannot produce a visibility verdict.
2. Navigation, repeated layout events or exceptions can exceed capture limits or retain handlers: reserve attempts before acquisition, bound subscriptions, detach on completion/close, never force layout.
3. Serialization/flush/disposal failure can resemble a completed run: retain partial evidence, attempt all owned cleanup, fail completion and never accept a footer alone.
4. Diagnostic activation/build routing can overwrite old artifacts or leak into ordinary builds: reject incompatible flags before build work and isolate output/intermediate paths; no unbound live launch.
5. Measurement itself can alter the reproduction: preserve layout declarations and page handlers, collect only read-only geometry; non-reproduction must not be called a repair.

## File map

All new pure types use namespace Talenro.Demo.App.Layout.EmailGeometry:
- Create clients/windows/Talenro.Demo.App/Diagnostics/EmailGeometry/GeometryEvidence.cs (records/enums/evaluator).
- Create .../EmailGeometry/GeometryCaptureBudget.cs (bounded state, no WinUI/IO).
- Create .../EmailGeometry/GeometryActivation.cs (fixed arguments, run identity).
- Create .../EmailGeometry/GeometryReceiptSession.cs (owned create-only session, serialization and cleanup).
- Create .../EmailGeometry/EmailGeometryObserver.cs (only WinUI adapter, snapshots/subscriptions).
- Create clients/windows/Talenro.Demo.Tests/EmailGeometryEvidenceTests.cs,
  EmailGeometryCaptureTests.cs, EmailGeometryReceiptTests.cs, EmailGeometryActivationTests.cs,
  EmailGeometryIsolationTests.cs, EmailGeometryAdapterContractTests.cs.
- Modify clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj to link the four pure files explicitly, never the WinUI adapter.
- Modify clients/windows/Directory.Build.props and Talenro.Demo.App/Talenro.Demo.App.csproj for isolated compile/build gates.
- Modify Talenro.Demo.App/App.xaml.cs and MainWindow.xaml.cs only inside EMAIL_GEOMETRY_RECORDING guards.
- Modify Views/LoginPage.xaml and AccountPage.xaml only to add nonvisual x:Name references to ScrollViewer/content panel/Account email (Login email already named).
- Create clients/windows/diagnostics/email-geometry.md and docs/roadmap/2026-11-10-windows-c4-email-geometry-execution.md.
- Update current-status.md, repository-recovery.md and approved design/plan status at handoff.

No new live PowerShell entry in this stage: not-built artifacts have no verified
hashes. The later fixed-entry deliverable is explicit follow-up after a separately
approved build, not a missing implementation task or permission to use an ad-hoc runner.

## Shared data contract

Define these pure types in GeometryEvidence.cs (nullable data is accompanied by status):
- enum GeometryStatus { Ok, Missing, Unloaded, Invalid, Unsupported, Error };
  enum GeometryVerdict { Inside, Overflow, Unknown }.
- record GeometryPoint(double X, double Y);
  record GeometryRect(string Space, double Left, double Top, double Right, double Bottom).
- record GeometryField<T>(GeometryStatus Status, T? Value) for reference-type T;
  numeric properties within records are nullable where unspecified.
- record GeometryElement(string Role, GeometryStatus Status, bool Loaded, bool Visible,
  GeometryField<GeometryRect> RootBounds, GeometryField<GeometryRect> LayoutSlot,
  double? ActualWidth, double? ActualHeight, double? DesiredWidth, double? DesiredHeight,
  string? FlowDirection, string? HorizontalAlignment, string? TextAlignment,
  GeometryThickness? Margin, GeometryThickness? Padding, double? Width, double? MaxWidth,
  IReadOnlyDictionary<string, GeometryStatus> FieldStatuses);
  record GeometryThickness(double Left, double Top, double Right, double Bottom).
- record GeometryBoundary(string Role, GeometryField<GeometryRect> Bounds, bool Complete);
  record GeometryScroll(double? HorizontalOffset, double? VerticalOffset, double? ViewportWidth, double? ViewportHeight,
  IReadOnlyDictionary<string, GeometryStatus> FieldStatuses).
- record GeometrySnapshot(Guid RunId, int Sequence, string Page, double ElapsedMs,
  double? RenderingScale, double? TextScale, GeometryScroll Scroll,
  IReadOnlyList<GeometryElement> Elements, IReadOnlyList<GeometryBoundary> Boundaries,
  bool ClipChainComplete, IReadOnlyDictionary<string, GeometryStatus> FieldStatuses).
- record GeometryComparison(GeometryVerdict Verdict, double? LeftInset, double? TopInset,
  double? RightInset, double? BottomInset, string Reason).
- GeometryEvaluator.Normalize(string space, IReadOnlyList<GeometryPoint> corners)
  returns GeometryField<GeometryRect>; exactly four finite corners, normalize min/max.
- GeometryEvaluator.Compare(GeometryField<GeometryRect> element, GeometryBoundary boundary)
  returns GeometryComparison. Insets = element.Left-boundary.Left, element.Top-boundary.Top,
  boundary.Right-element.Right, boundary.Bottom-element.Bottom. Any negative => Overflow;
  zero tolerance, raw doubles retained. Missing/invalid/incomplete boundary or differing
  spaces => Unknown with null distances. Positive area required. Inside is per known
  box boundary only, never a global visual PASS; incomplete clip chain stays explicit.

Root space literal `RootDIP`; layout slots use `ParentDIP:<role>` and must not be
compared to root bounds. Width Auto/Infinity constraints encode null plus acquisition
status rather than nonstandard JSON numbers; invalid geometry itself stays Invalid.
FieldStatuses has one entry for every scalar/enum/thickness acquisition on its record,
using the exact property name; nested records carry their own statuses. No absent keys.
Auto Width and unbounded MaxWidth use Unsupported/null; unavailable reads use
Missing/Unloaded/Error as appropriate. Null with Ok, nonfinite serialized numbers,
unknown keys and missing status keys are rejected by receipt validation and tested.
Record only allowlisted enum/property values, roles, numbers and sanitized categories.

### Task1 — Pure geometry and bounded capture state

**Files:** GeometryEvidence.cs, GeometryCaptureBudget.cs, corresponding tests, test csproj.
**Consumes:** fixed schema above. **Produces:** evaluator plus
`GeometryCaptureBudget.TryReserve(string page, bool eligible, out int sequence) -> bool`,
`Stop() -> void`, `Stopped -> bool`, and per-page Attempts counts.
Login/Account only. Reserve before callback, consume failing attempts, never reset
on re-entry; stop permanently at4 or explicit Stop. Other pages/ineligible do not reserve.

Also in GeometryCaptureBudget.cs: pure `GeometryCaptureCoordinator(GeometryCaptureBudget budget)`,
`ChangePage(string? page) -> void`, `TryQueue(out long token) -> bool`,
`TryBegin(long token, bool eligible, out int sequence) -> bool`,
`Cancel(long token) -> void`, `Stop() -> void`. ChangePage invalidates capture eligibility
without releasing the physical pending slot or resetting budget. The stale callback
must drain before another enqueue. Tokens are unique/increasing; at most one pending token.
TryBegin consumes only its matching pending token, then calls TryReserve before reads;
stale tokens cannot clear newer work. Cancel handles dispatcher rejection; Stop
invalidates work and stops the budget permanently. UI-thread-only, no dispatcher/IO.

- [x] **1. Write failing evaluator tests with literal fixtures.**
  Inside: element(10,20,30,40), boundary(0,0,100,100), RootDIP => Inside,
  insets10/20/70/60. Left(-2,20,30,40) => Overflow,left=-2. Add right101,
  top=-1,bottom101, exact-touch0, negative coordinates inside a negative-origin box.
  Mirrored corners(100,0),(80,0),(80,20),(100,20) normalize80/0/100/20.
  Missing/error/NaN/infinity/zero/reversed area, three corners, mixed-space and
  incomplete boundary => Unknown, never Inside. Incomplete overall clip chain must
  remain distinguishable from known-boundary Inside in snapshot serialization.
- [x] **2. Run RED:** `dotnet test clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore --filter FullyQualifiedName~EmailGeometryEvidenceTests`.
  Expected failure from missing new production types, not missing dependencies. If
  dependency failure occurs, report blocked verification instead of treating it as RED.
- [x] **3. Implement pure evaluator and records; run same command.** Expected exit0,
  all named cases pass, no WinUI or process calls. No geometric rounding tolerance.
- [x] **4. Write capture tests then run RED** with filter EmailGeometryCaptureTests:
  Login2+Account2 total4; third Login rejects; unknown page/ineligible reserves0;
  exception after reservation consumes slot; Stop/re-entry/reentrant reservation
  cannot exceed limits. Expected missing budget behavior failure.
  Coordinator tests: duplicate queue rejects, navigation invalidates old token,
  stale cancel/begin preserves newer token, rejected enqueue releases pending slot,
  close invalidates all tokens, and ineligible acquisition consumes no attempt.
- [x] **5. Implement budget, run filter then full portable suite.**
  `dotnet test clients/windows/Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore`.
  Expected exit0; record actual counts, do not assume old counts remain current.
- [x] **6. Commit explicit Task1 files:** `feat: add bounded email geometry evidence model`.

### Task2 — Activation and create-only recording lifecycle

**Files:** GeometryActivation.cs, GeometryReceiptSession.cs, activation/receipt tests,
explicit pure links in test csproj.
**Consumes:** Task1 immutable snapshots and budgets.
**Produces:** `GeometryActivation.Parse(IReadOnlyList<string> args) -> GeometryActivationResult`
record(Guid? RunId, IReadOnlyList<string> LayoutArguments, string? ErrorCode).
Require exactly layout-test/language=en-US/profile=rtl-mixed and
`--layout-email-geometry-run=<lowercase nonzero32 GUID>` once; any extra/malformed,
wrong language/profile, duplicate or close-run option rejects. Remaining layout args
still go through existing LayoutOptions parser unchanged.

`GeometryReceiptSession(Guid runId, Stream ownedStream)`, `Add(GeometrySnapshot) -> void`,
`Finish() -> GeometryFinishResult`, `Dispose() -> void`; FinishResult(bool Completed,
int Samples, IReadOnlyList<string> ErrorCodes). Add memory only, max4; thread affinity
captured at construction; wrong-thread/invalid-run/sequence violation disables capture.
Session owns stream immediately; failures always attempt disposal. Finish idempotent,
no second writes; Dispose without Finish only releases resources, never fabricates success.
Private factory `OpenNew(string trustedAbsoluteRoot, Guid runId) -> Stream` creates
`<root>/<runId>/geometry.jsonl` without overwriting existing directory/files, rejects
existing reparse ancestors. No automatic deletion/retry. Root supplied only internally.

Lifecycle choice: create sink before MainWindow; collect in memory in layout callbacks;
on normal Closed, detach all probe handlers, then synchronously Finish outside layout
callbacks. At most4 snapshots,16 elements and16 boundaries per snapshot; enforce all
caps. Serialize UTF8/no BOM, schema1 header, snapshots, then provisional health footer,
flush/dispose. Completed true only after cleanup succeeds; close handler sets process
ExitCode=3 on recording/cleanup failure (startup rejection remains2), never force-exits
the process. File footer alone is not proof of completion: future reader requires
valid full schema/correlation plus independently verified normal process exit0.
Crash/forced termination/missing footer => incomplete, no reconstruction/retry.
Synchronous IO may block:360s observer overrun remains unsuccessful; no hard deadline
claim, no hidden background worker or polling added. Preserve primary and cleanup
errors internally, expose only sanitized categories, no exception messages in receipts.

- [x] **1. Write activation tests; run RED** filter EmailGeometryActivationTests.
  Valid literal run11111111111141118111111111111111 returns same ID and3 layout args;
  empty/uppercase/all-zero/duplicate/extra/close-run/wrong-profile cases reject.
  Expected missing-type/behavior failure. Implement parser; same command GREEN exit0.
- [x] **2. Write recording tests; run RED** filter EmailGeometryReceiptTests.
  Memory tracking streams assert Add performs0 writes, Finish writes header+0..4
  snapshots+footer, exact IDs/sequence/schema, no user text, disposal before Completed.
  Inject serialize/write/flush/writer-construction/dispose and combined failures;
  assert Completed=false, all owned cleanup attempted, primary retained, no retry.
  Dispose without Finish, repeated Finish/Dispose, wrong thread, caps and invalid
  snapshot values fail safely; zero samples cannot imply diagnosis success.
- [x] **3. Implement session; run receipt filter GREEN.** Factory tests use a test-owned
  temporary directory with cleanup limited to that exact directory (no real reparse
  creation needed; inject path metadata below factory for reparse cases). Existing
  file/directory content unchanged; invalid IDs cause no filesystem effects; stream
  ownership on construction failure explicitly tested. No new public CLI seams.
- [x] **4. Run full portable suite** using Task1 command; expected exit0. Record actual
  test totals and lifecycle limitations in execution ledger.
- [x] **5. Commit explicit Task2 files:** `feat: add fixed email geometry capture receipts`.

### Task3 — Isolated WinUI adapter, build gates and recoverable handoff

**Files:** EmailGeometryObserver.cs, App/MainWindow hooks, nonvisual XAML names,
Directory.Build.props, App csproj, isolation/adapter-contract tests, guide/ledger/status.
**Consumes:** Tasks1/2 types and owned session.
**Produces:** `EmailGeometryObserver(FrameworkElement root, FrameworkElement pageHost,
Func<string?> currentPage, Func<EmailGeometryTargets?> currentTargets,
GeometryReceiptSession session) : IDisposable`; `RefreshPage() -> void`;
`EmailGeometryTargets(Page Page, ScrollViewer Scroll, StackPanel Panel, TextBlock Email)`
is adapter-local WinUI data. Named target getters compiled only under new gate.

Build property EnableEmailGeometryRecording defaults false; define
EMAIL_GEOMETRY_RECORDING only true+Debug+layout diagnostics. Add initial validation
EMAILGEOMETRY001 for nonboolean flag or enabled without Debug/layout, enabled with
close-stage recording or any CloseStageVerificationMode. Disabled flag leaves existing
configuration behavior unchanged. App paths: obj/email-geometry/ and bin/email-geometry/;
MSBuildProjectExtensionsPath matches new obj, no Core path overrides. Exclude all
EmailGeometry/*.cs when flag not true; never include probe in ordinary/Release output.

Names: LoginGeometryScroll/LoginGeometryPanel (existing DemoEmail),
AccountGeometryScroll/AccountGeometryPanel/AccountGeometryEmail. Add only x:Name;
no geometry properties or extra containers. App parses activation and opens session
before constructing window. MainWindow wires adapter after shared pages exist and
calls RefreshPage after selecting PageHost.Content; currentPage maps only Login/Account.
On Closed try detaching observer then finishing session even if detach fails; any
failure => ExitCode3, no exception payload publication. All non-probe behavior unchanged.

Adapter subscribes only active page LayoutUpdated while it has budget, plus bounded
Loaded/Unloaded handling for that page and window close. RefreshPage detaches old
page first; capture only loaded visible page actually in PageHost with same XamlRoot.
Use two distinct naturally occurring LayoutUpdated events, at most one pending
dispatcher capture; generation token discards navigation-stale queued work before
reservation. Reserve once before reading. Detach after two page attempts or total4;
leave completion pending/Unknown if events do not arrive. No event sender assumption.

Read four transformed corners of each element into RootDIP; retain local layout slot
as ParentDIP:<role>. Acquire dimensions/enum properties individually with sanitized
statuses. Traverse only email-to-root ancestor chain, maximum16; no app-wide visual
tree walk. Scroll content presenter must be unique on that chain; do not guess via
first descendant match. Capture viewport/explicit clip bounds only when coordinate
mapping is supported; missing/template/nonrectangular clip information => Unknown
and ClipChainComplete=false. Never infer full clip chain from parent ActualWidth.
No filesystem IO, UI mutation, UpdateLayout, model dispatch or forced navigation in
capture. Time from Stopwatch, same run ID, sequential reserved sequence. A read failure
records target Error; it cannot reset budgets or cause automatic recapture.

- [x] **1. Add failing isolation tests** verifying evaluated MSBuild properties/Compile
  items without build: ordinary, layout-only, geometry-valid, geometry+Release,
  geometry without layout, geometry+close-mode, invalid string flag. Plan uses
  `dotnet msbuild ... -getProperty:BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,DefineConstants -getItem:Compile`
  for evaluation only, and `-target:ValidateEmailGeometryConfiguration` for gates;
  no Build/Restore/Compile targets. Test assertions: valid paths isolated, gate
  errors named EMAILGEOMETRY001, ordinary excludes all probe files. Expected RED
  missing target/gate, not missing SDK failure. Implement gates then GREEN.
- [x] **2. Write adapter contract tests before hooks, run RED** filter
  EmailGeometryAdapterContractTests. Compare XAML before/after removing only approved
  x:Name additions: same element tree/attributes. Verify UI-independent scheduling
  using recording event/dispatcher doubles: navigation before queued capture,
  duplicate events, acquisition throw, no second layout event, close-before-callback,
  detachment failure; assert caps, disposal attempts and no stale-page records.
  Drive Task1 GeometryCaptureCoordinator through those doubles; adapter uses the same
  coordinator, cancelling tokens when enqueue fails and stopping before detachment.
  No new scheduling interface or real UI execution in tests.
- [x] **3. Implement thin adapter and guarded hooks.** Validate API signatures against
  installed references/official docs; do not substitute WPF APIs. Suggested APIs:
  UIElement.TransformToVisual, GeneralTransform.TransformPoint and
  Controls.Primitives.LayoutInformation.GetLayoutSlot. Missing property/template
  support remains Unsupported, not invented data. Run Task3 tests then full portable
  suite. Expected exit0; WinUI adapter compilation remains UNVERIFIED until later build.
- [x] **4. Add guide and execution ledger.** Explain pure/static versus live evidence,
  synchronous close-IO limits, incomplete-on-crash, full file+normal exit requirement,
  no recorded text/private payload, old-artifact preservation, no new runnable entry
  until verified hashes exist. Mark C4 NOT_ACCEPTED/no launch allowance.
- [x] **5. Final scoped verification.** Full portable dotnet suite --no-restore; separate
  hosts for observe-c4, observe-c5, c5-observation-session, observation-session-guard,
  process-observation, native-process-observation `.tests.ps1` via pwsh -NoProfile -File.
  Expected existing counts152/32/120/48/32/78, all exit0. Git diff --check; existing
  observer/module diff empty versus implementation baseline; old EXE/DLL hashes
 4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 /
 7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB unchanged.
  No native App build/launch; declare adapter compile and actual output metadata
  isolation unverified, not passed by source/evaluation tests.
- [x] **6. Commit:** `feat: isolate real-page email geometry diagnostics` with explicit
  necessary files. Obtain one final independent read-only whole-plan review for inline
  execution, covering all five focus items. Fix Important/Critical with RED→GREEN;
  record deferred Minor items and every boundary the reviewer declines to judge.
- [x] **7. Preserve:** update current status/recovery/plan, commit review record, ordinary
  push to verified origin existing branch, verify remote SHA=local and clean worktree.
  Final message states not built/not run. Request separate native build authority only
  after implementation review, never auto-advance to build/hash binding/native route.

## References and self-review

Microsoft LayoutUpdated documentation states that events can reflect other tree
changes and sender is null; it does not promise repeated stable samples:
https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.frameworkelement.layoutupdated
This is why capture uses selected-page/generation checks and accepts missing second samples.

Spec map: geometry/unknown => Task1; limits/activation/ownership => Tasks1/2;
real-page reads/isolation => Task3; old artifacts and later fixed runner => explicit
build follow-up boundary. Five review-focus cases have owner tests above. Shared
interfaces are records/budget then session then adapter; avoid parallel edits to them.
The test project links pure sources and does not reference the WinUI App project.
Tests compile portable code only; implementation approval must explicitly include
that scope. Design approval alone has not run any command listed in this plan.

Recommended execution: sequential inline, final fresh review. Alternative: fresh
implementer/reviewer per task, then whole-plan review; stronger interim checks at
higher coordination cost. User selected inline; ordinary-host baseline now verified.
