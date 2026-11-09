# Read-only layout observer / close-path review

Baseline5cbf7e29. Source inspection only: no application edits, build, tests,
launches, settings/service operations or additional runtime budget used.
Reviewed App/MainWindow, LayoutWindowObserver, LayoutMeasurement/LayoutTitle,
LayoutPresentation, NodesPage, DemoViewModel and test/project wiring.

## Verified source paths

- MainWindow.xaml.cs:25-37 creates all four pages, then conditionally constructs
  LayoutWindowObserver when DiagnosticPresentation exists in LAYOUT_DIAGNOSTICS.
  Removing the entire diagnostics build would also change resources/presentation
  and is NOT an observer-only comparison. Normal builds exclude Diagnostics/*.cs.
- Observer constructor:28-40 creates UISettings and subscribes TextScaleFactorChanged;
  subscribes root/pageHost Loaded, Unloaded, SizeChanged and window.Closed; calls
  Refresh. Only UISettings setup catches COMException/NotSupportedException.
- Refresh:43-70 checks disposed first, dispatches to UI queue when necessary,
  replaces XamlRoot.Changed subscription if root identity changes, reads sizes,
  raster scale and UISettings.TextScaleFactor, formats and updates window.Title.
  Equal titles skip the final title write, NOT the earlier property reads.
- TextScaleChanged:76 enqueues Refresh. A pending Refresh invoked after disposal
  returns at its first guard. Callback itself has no disposed check and ignores
  TryEnqueue's result; actual callback/thread ordering is not recorded.
- Closed:77 synchronously calls Dispose. Dispose:79-93 sets disposed=true first,
  removes six root/pageHost handlers, then window.Closed, XamlRoot.Changed and
  UISettings.TextScaleFactorChanged, and clears observedRoot. No explicit wait,
  sleep, process control, clipboard access or async task appears in this path.
  Removals have no per-step completion trace or exception handling. If one threw,
  later statements would not run and a subsequent Dispose would return immediately;
  this is a conditional control-flow observation, NOT an observed exception/defect.
- NodesPage.Search_Changed:93 calls only NodesPage.Refresh directly. It does NOT
  directly call MainWindow.Refresh or observer.Refresh. Selection dispatch during
  list rendering is guarded. Layout events may indirectly trigger the observer,
  but search-to-observer activity/count is not established from this source alone.
- MainWindow.Refresh:60 explicitly calls observer.Refresh on its own refresh path.
  The model PropertyChanged subscription is a captured local-model callback;
  its existence alone does not demonstrate an external root or exit blocker.
- LayoutTitle is pure formatting. LayoutMeasurementTests:46-55 lifecycle check
  reads copied source and searches markers; it never instantiates the live observer.
  Test project links formatting logic, not the live WinUI observer implementation.
  No claim that existing green tests prove event removal/close execution is valid.

## Evidence gap and next proposal

The two list-refresh modes exited normally but omitted this observer. That does
not incriminate the observer or exclude interactions. Earlier wait-chain evidence
names a service-host dependency; it does not locate a managed callback frame or
prove TextScaleFactorChanged subscription/removal caused that wait.

Prioritize bounded diagnostic stage records on a separately isolated diagnostic
build: close-handler entry, Dispose entry, before/after each event-unsubscription
boundary, Dispose return and close-handler return. Preserve original artifacts,
logic and normal close semantics. Record only fixed stage identifiers, sequence,
correlation and monotonic timing, not input text or clipboard/user content.
No exception suppression, forced exit or service changes. Logging transport,
failure handling and bounded overhead need design review before implementation;
synchronous logging could itself perturb the shutdown being measured.

If no entry is recorded, do not immediately conclude Closed never fired: first
establish recorder reliability. If the last reliable stage is before/after a
specific removal, narrow the next investigation there; if all stages complete but
native process remains alive, investigate after this handler rather than blaming
it. Absence of logs is not proof of absence of execution.

Alternative: observer-enabled/disabled isolated comparison offers coarse
localization but changes subscriptions, UISettings reads and title updates
together; a new small prototype can also miss original interactions. Neither is
a root-cause finding. Recommendation is a stage-recording design first, not an
implementation authorization. No new run approval; all broader NOT_ACCEPTED gates
and prior interrupted/timing records remain. No product fix justified yet.
