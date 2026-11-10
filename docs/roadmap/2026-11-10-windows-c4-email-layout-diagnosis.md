# C4 account email clipping — source-only diagnosis

Date2026-11-10, Asia/Tbilisi. Baseline70d1ef33.
Scope: read source/history and prior native evidence; no implementation, build,
test execution, app launch or settings change. Product files remain unchanged.

## Confirmed facts

- Prior run f6071741fab546ef8f856467def00375 showed stable Account email clipping
  at the left window edge while accessibility exposed full demo@example.com.
  Login's separate real-email control was fully readable in that same process.
- MainWindow.xaml.cs sets Root.FlowDirection=RightToLeft for the rtl-mixed profile.
  PageHost uses horizontal/vertical content Stretch, in the main grid's second
  column beside the220-DIP sidebar. Account inherits this direction through its
  Page/ScrollViewer/StackPanel; its email locally overrides LeftToRight.
- AccountPage.xaml:2 uses StackPanel Padding40, Spacing20, MaxWidth820, with no
  explicit HorizontalAlignment. Its email at line4 has LTR and FontSize20, but no
  explicit element alignment, text alignment, width or wrapper.
- LoginPage.xaml:2 explicitly centers its MaxWidth460/Padding32 panel. Its email
  is also LTR (FontSize18). Connection likewise explicitly centers its limited-width
  panel. These differences are comparison clues, not a controlled one-variable test.
- AccountPage.xaml.cs refreshes entitlement/usage/device text, not the email or its
  geometry. LayoutStaticBindings updates named localized text/content, not this
  unnamed email. Theme.xaml contains brushes and an explicitly keyed button style,
  no custom implicit TextBlock geometry style.
- Account email and container declarations predate the diagnostic binding change
 99fc48f6; that change did not alter these two declarations. This does not prove the
  historical layout was correct or identify the exact SDK behavior in the old binary.
- Existing LayoutViewContractTests assert email LTR declarations only; they do not
  measure visible bounds. LayoutWindowObserver reports root/PageHost dimensions,
  not email/panel bounds, layout slots or clipping rectangles. Those tests/readouts
  cannot prove absence of this visual defect.

## Leading hypothesis, not confirmed root cause

The evidence points to element placement/clipping at the RTL-parent/LTR-email
boundary, with constrained/default-aligned Account container geometry a candidate
contributor. It does not support a truncated email value or a live account-data
submission issue. Text direction alone cannot guarantee containment in the viewport.

Microsoft documents direction inheritance and horizontally changed coordinates for
RTL layout, and notes that alignment depends on parent composition while MaxWidth
constrains Stretch. These general rules support investigating the boundary; they
do not prove a specific WinUI bug or that adding Center is a sufficient fix:
- https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.frameworkelement.flowdirection?view=windows-app-sdk-1.8
- https://learn.microsoft.com/en-us/windows/windows-app-sdk/api/winrt/microsoft.ui.xaml.frameworkelement.horizontalalignment?view=windows-app-sdk-1.8

Current source declares Microsoft.WindowsAppSDK2.6.5; the cited API explanations
are general semantic references, not version-matched proof of the preserved binary's
internal arrange behavior. No alternate renderer or numeric simulation was used to
claim reproduction. Login differs in several dimensions, so its success is not proof
that its Center setting alone explains the Account failure.

## Missing evidence and next proposed step

Exact cause remains unconfirmed: no runtime bounds were recorded for the email,
its StackPanel, ScrollViewer/content presenter or their shared-coordinate transforms.
Cannot yet distinguish parent arrangement, mixed-direction child positioning and
clip/content-presenter interaction conclusively.

Recommend a small, separately approved geometry-diagnostic plan before selecting a
fix: observe email/panel/viewport bounds in one root coordinate system and effective
directions/alignment; compare to the preserved failing evidence. Keep email LTR and
root RTL. Test the geometry verdict on in-memory inside/left-overflow/right-overflow
fixtures, but do not equate that with reproducing the WinUI rendering defect.
Any diagnostic build/live run needs explicit scope and budget; preserve old artifacts
and hashes, do not overwrite them or silently update the fixed C4 observer policy.
After cause confirmation, a minimal change needs regression evidence and native
revalidation. Do not remove LTR isolation, add guessed pixel offsets or claim a
centering-only repair now. No launch allowance remains; C4 stays NOT_ACCEPTED.

Only this report/status/recovery handoff is changed. No product/source/test fix,
main merge, system mutation or acceptance promotion is authorized by this diagnosis.
