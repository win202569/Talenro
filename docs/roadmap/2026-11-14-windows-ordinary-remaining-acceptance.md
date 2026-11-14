# Ordinary remaining acceptance and focus evidence — read-only consolidation

Date2026-11-14; checkpoint a0559154. No code changes, test/build/restore, live
process query or application launch. Read source and existing records only.

## Coverage ledger

| Item | Ordinary candidate status | Remaining evidence |
| --- | --- | --- |
| Exact launcher/command binding and normal exit | Verified on run f3c2f05a843b42abac087ae32517791f; original session ExitVerified=true, entry0 | No need to repeat merely for bookkeeping; each future run still needs preflight and exit validation |
| Login offline/read-only fields, disconnected/lock-off initial Connection | Observed in that run | No broader login/connection acceptance inferred |
| Default and qualitatively narrow Account | Email/use/device text readable; wrapping observed | Exact viewport/DPI/RTL coverage remains separate |
| Forward/reverse keyboard reachability and auto-scroll | UNVERIFIED on ordinary candidate | Establish visible button focus; observe traversal without pointer scrolling |
| Manage devices placeholder, keyboard close, focus return | NOT_RUN on ordinary candidate | Open exactly one harmless placeholder by keyboard; close by keyboard and observe returned outline |
| Exit-demo/reset and re-entry | NOT_RUN on ordinary candidate | Visible Login/sidebar reset and disconnected/lock-off re-entry only; no exhaustive state-reset claim |
| Exact620–660 root /1024x768 /1440x1024 and scales | UNVERIFIED | Ordinary screenshots are not root/client DIP measurements |
| Reserved-option rejection, other dialogs, screen readers, broad function/security/Release/C4 | Outside this minimal run / still unaccepted | Separate scope/evidence, not silently added |

## Focus evidence chronology and interpretation

The ordinary run first reported user input during attempted Tab. Subsequent
observations showed window/pane focus, no visible button outline; account heading
was clicked once and a later Tab still did not establish a button outline.
User resizing confirmation arrived during this sequence. No stable, visibly
identified button was established as the starting focus. No complete traversal
or key-delivery trace was captured. Thus the run is an inconclusive keyboard
observation, not a reproduced product defect or proof keyboard access works.

Earlier diagnostic run5b14bb63977d4ceb98cf0ea0a1cd2afb (2026-11-13) had visible
Settings/MFA/Manage/Language/Delete/Exit outlines and keyboard auto-scroll, plus
dialog close/focus-return evidence, while accessibility focus metadata STILL
reported pane2. That record explicitly treats rendered behavior as the evidence.
It used different hashes, en-US/rtl-mixed and measured650x620 diagnostic geometry;
it cannot be transferred as ordinary candidate acceptance. It demonstrates why
pane/window metadata alone is insufficient to diagnose broken keyboard focus.

## Source inspection facts (not runtime proof)

- AccountPage.xaml lines11–15 declare five ordinary Buttons in order MFA,
  Devices, Language, Delete, Exit. No explicit IsTabStop=false, TabIndex override,
  or keyboard-handler binding on them.
- MainWindow.xaml PageHost has IsTabStop=false, but this property alone is NOT
  treated here as evidence that all descendants are excluded. No framework
  behavior experiment was performed.
- AccountPage.Placeholder_Click shows ContentDialog then calls sender Control
  Focus(FocusState.Programmatic); same code is outside diagnostic conditionals.
  This expresses focus-return intent, not proof of a visible keyboard outline.
- MainWindow navigation dispatches state and Refresh replaces PageHost.Content;
  no explicit button-focus initialization for Account was found in those methods.
- Theme.xaml contains colors and primary-button visual setters, no focus/Tab
  setter. Inspected App.xaml merges that dictionary and XamlControlsResources.
- AccountPage.xaml/.cs and MainWindow.xaml/.cs have no Git diff against ordinary
  build source5ee8a339. No new code regression in these files was identified.

Possible explanations remain input/focus handoff after resize, automation key
delivery/observation limitations, or actual runtime focus behavior. Existing
evidence does not isolate which. Do not add a focus workaround based on this
inspection, or change shared guards/product code to make a report pass.

## Smallest proposed next step (not authorized/executed)

Prepare a bounded manual-keyboard follow-up plan, separate from implementation:
user completes resize, explicitly pauses mouse input, then manually presses Tab
one key at a time with visible observation. Establish a real button outline before
claiming the traversal begins. If keyboard focus cannot be established, preserve
UNVERIFIED, close normally and stop. Do not click Delete or activate security
settings. If sequence succeeds, check one Manage devices placeholder and reverse
navigation; reset/re-entry only if time remains before300s close deadline.

Keep one-launch/zero-retry/6min and exact preflight/retained-handle checks. New
launch requires separate approval. Do not repeat already observed default display
except necessary navigation, or silently expand to exact dimensions/all locales.
This is a proposal, not a live command or authority. Current allowance0.

References: [ordinary run](2026-11-14-windows-ordinary-exact-binding-live-observation.md),
[diagnostic keyboard run](2026-11-13-windows-keyboard-followup-observation.md).
Raw original evidence remains unchanged/ignored; sanitized coverage survives clone.
