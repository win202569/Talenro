# Content-entry focus diagnosis — read-only, runtime identity unconfirmed

Baseline:8bfc37d8. User requested the recommended read-only investigation.
No application launch, build, test, restore, instrumentation, source edit or
system setting change was performed. This record is not a repair or acceptance.

## Evidence and comparison

Observed in PID10372: Settings → one stop with no identifiable focus outline →
the first content control (Connect, Search or MFA). Focus metadata returned a
root pane or stale Search value, so it did not identify the intermediate object.

- `clients/windows/Talenro.Demo.App/MainWindow.xaml:16`: shared `PageHost` is a
  ContentControl immediately after Sidebar. It sets content alignment only;
  no IsTabStop or focus-visual override. This declaration already existed in
  391e2944; the later country/notice fix did not introduce it.
- MainWindow.xaml.cs: Refresh assigns an existing Page instance to PageHost.Content.
  Navigation dispatches state changes; no explicit navigation-focus redirect.
  The only explicit shell Focus call is after a placeholder dialog closes.
- Each of the four Views/*.xaml has Page → ScrollViewer → StackPanel. No project
  IsTabStop, TabIndex, TabFocusNavigation or UseSystemFocusVisuals override was
  found. Theme.xaml defines brushes and a primary-button style, not container
  focus styles. App.xaml merges XamlControlsResources and that theme.
- Actual restored WinUI package is Microsoft.WindowsAppSDK.WinUI/2.6.5, under
  the worktree's clients/windows/.nuget/packages, as recorded by project.assets.json.
  Read-only inspection of its net6.0-windows10.0.17763.0/Microsoft.WinUI/Themes/
  generic.xaml found the default ContentControl style at8726: alignment setters
  and a plain ContentPresenter template, no IsTabStop override, focus states or
  system-focus-visual setter. The unkeyed ScrollViewer style at9062 inherits
  DefaultScrollViewerStyle at33549, which explicitly sets IsTabStop=False and
  enables the system-focus-visual resource. Thus disabling all ScrollViewers is
  not supported by this evidence and would target a control already excluded
  by the inspected default style.
- The same package's Microsoft.WinUI.xml at2860 documents IsTabStop's default
  as true; at3168 it documents UseSystemFocusVisuals' default as false. Microsoft
  Learn independently describes [IsTabStop](https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.uielement.istabstop?view=windows-app-sdk-2.0-experimental)
  and [focus visuals](https://learn.microsoft.com/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.uielement.usesystemfocusvisuals).
  Web API version is not proof of local2.6.5 runtime values; local package evidence
  is the version-specific reference here. No dependency/cache files are committed.
- Login/Account placeholders explicitly restore sender focus; Nodes restores Apply
  after a canceled confirmation. This agrees with the observed working dialogs,
  and is distinct from ordinary sidebar-to-content Tab traversal.
- ShellContractTests checks offline notice placement, fixed credential text,
  entry wording, decorative glyph and separate scenarios. It does not assert a
  noninteractive PageHost tab-stop policy. Existing tests were inspected, not run.

## Diagnosis

Strongest hypothesis: shared PageHost itself is the extra noninteractive tab
stop, without a visible focus treatment. Its position, inherited tab-stop default
and plain template explain all three page observations with one common cause.
The inspected ScrollViewer default is contrary evidence for blaming scrolling.

Confidence: high as a source-level explanation, not confirmed runtime object
identity. Page/container defaults, applied runtime styles and capture behavior
were not measured live. The stale tool metadata remains independently unresolved;
it must not be treated as proof that focus stayed in Search. Login's successful
button sequence started after initial focus and does not exclude a host stop.

## Smallest proposed next change — requires implementation approval

Explicitly set `IsTabStop="False"` only on PageHost. Keep it enabled and preserve
child tab navigation; do not disable the host, hide it from accessibility, rewrite
page navigation or alter all ScrollViewers. Add a static regression asserting
that PageHost is a non-tab-stop container while existing content controls remain
present. Static checks establish declaration intent, not runtime focus behavior.

Suggested bounded validation if approved: one focused regression RED before the
change and GREEN after (2 calls), one existing full suite, one app build, then
one ordinary-user launch up to10minutes (stop new scenarios at8minutes;
startup30seconds, close observation15seconds; no retries). Observe Settings →
Connect/Search/MFA without the extra stop, reverse traversal back to Settings,
lower controls still reachable/auto-scrolled, and representative modal focus
return. Stop on unexpected failure; do not consume more calls or weaken gates.
If the stop remains, retain the failed hypothesis and request scoped live focus
identity diagnostics; do not stack speculative fixes.

This is a proposed scope, not newly authorized execution. All previous launch
allowances remain consumed. Full keyboard/accessibility acceptance and
Task11/12 plus44/34/6/217 NOT_ACCEPTED remain unchanged; no main merge.
