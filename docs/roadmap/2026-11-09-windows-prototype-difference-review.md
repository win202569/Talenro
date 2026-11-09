# Demo / single-TextBox difference review

Read-only review after6ad301c7. No source change, build, launch, service operation
or settings change. The prototype did not reproduce the stall; manual-route timing
deviation remains. This comparison selects a candidate investigation, not a fix.

## Source-level comparison

| Area | Original diagnostic demo | Prototype |
| --- | --- | --- |
| Startup | Parses layout arguments, loads embedded fixtures and two resource catalogs; creates four pages and model | Creates one Window |
| Text input | Standard TextBox, TextChanged -> NodesPage.Refresh synchronously | Standard TextBox, no text handler |
| Refresh | Reapplies diagnostic Header/PlaceholderText, clears and rebuilds ListView items, updates enabled/notice/selection properties | No list, selection or refresh |
| Container/resources | Page/ScrollViewer/StackPanel, ContentControl navigation, theme brushes/button style, localized x:Uid content | Grid, static English header/placeholder, default control resources |
| Lifetime | Layout observer subscribes UISettings/TextScaleFactorChanged, XamlRoot and layout events; Window.Closed calls Dispose | No application-defined close/event observer |

Both retain the Window in App, call Activate, and use WinUI x64 unpackaged,
self-contained build settings and declared WindowsAppSDK2.6.5. Comparing every
prototype lockfile dependency target/package against demo lockfile found no
resolved-version/contentHash difference. This does not prove loaded native binaries
or complete runtime state are identical. Demo also references its Core project.

Inspected App.xaml.cs, MainWindow.xaml(.cs), Views/NodesPage.xaml.cs,
ViewModels/DemoViewModel.cs, Resources/Theme.xaml, Diagnostics/LayoutWindowObserver.cs,
LayoutResourceCatalog.cs, LayoutPresentation.cs, and prototype App/MainWindow sources.
Prior Nodes XAML/static-bindings review remains supporting context.

## What this does and does not explain

- Search_Changed synchronously clears/repopulates the list; rendering guard prevents
  model selection dispatch during rendering. No direct clipboard call was found.
- Diagnostic Copy uses the prepared dictionary; MainWindow.Copy falls back to a
  ResourceLoader only without DiagnosticPresentation. Do not claim resource loader
  creation on every search in the observed diagnostic configuration.
- Theme defines brushes and a Button style, not a custom TextBox template. This
  makes a custom-text-template explanation unsupported by inspected source.
- Layout observer has a disposed guard and explicit event unsubscription. This
  does not prove the close callback ran/completed or exclude native event effects.
- Refresh, layout invalidation, accessibility notifications, IME/focus and service
  state could interact; source alone cannot establish which interaction occurred.

## Recommended next single-factor boundary (not implemented/authorized)

Investigate whether synchronous list rebuilding during TextChanged changes the
prototype outcome. A useful diagnostic variant would keep the same TextBox and
ListView present in both modes, initialize the same fixture rows, and switch only
whether text changes trigger rebuilding. Keep theme, layout observer, resources,
close behavior and package versions unchanged between its modes; no demo references.
Do not compare a list-free window to a list window and call rebuilding the only
variable. No clipboard operations, force-exit workarounds or service changes.

Why first: it is directly on the observed input path and has a small switchable
boundary. This is an investigator's prioritization, not evidence of defect or
user approval to modify/build/run. The observer/resource hypotheses remain open.
Adding an entire demo page or disabling all layout diagnostics at once would change
multiple dimensions and provide a weaker discriminator. A variant would still not
be a fully equivalent copy of NodesPage.Refresh; record omissions explicitly.

Next obtain scope for creating/compiling this isolated diagnostic variant; runtime
budget remains zero until separately approved. Preserve the original one-TextBox
baseline/artifact and original application. All broader NOT_ACCEPTED gates remain.
