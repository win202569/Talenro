# Windows test-only presentation and measurement mode

Status: APPROVED by user after f376dcbd; implementation plan awaits execution approval.
Baseline85f8663c; no implementation, test/build or launch is authorized here.
Independent read-only builder check: nothing important left before planning.
This checks design clarity, not implementation correctness or native acceptance.

## Intent and scope

Make the missing native layout evidence obtainable in the existing offline demo:
render the committed synthetic samples, select Chinese/English presentation and
report actual logical dimensions without changing Windows settings. Reuse the
existing Login, Connection, Nodes, Account and dialog implementations, not copies.
Default application behavior, Core reducer, simulated events, offline warning,
read-only credentials and existing security/acceptance boundaries stay unchanged.

This is a test facility, not a product language setting, translation delivery,
screen-reader verdict, contrast certification or runtime/persistence acceptance.
It cannot make a too-small monitor fit the1440x1024 client target. No automatic
connection, login, state seeding, real network, external input files or persistence.

## Approaches and decision

1. Selected: guarded mode in the existing application, sharing production views.
   Best correspondence to the actual UI; requires explicit gates and default-path
   regression checks to avoid leaking test presentation into ordinary runs.
2. Considered: independent test executable sharing views. Stronger executable
   separation, but adds startup/resource maintenance and possible host divergence.
   Not selected. A copied gallery would be weaker evidence and is excluded.

The detailed choices below are assistant proposals, not prior user requirements.

## Two gates and artifact separation

- A dedicated build property, proposed name EnableLayoutDiagnostics, defaults
  false in every configuration, including Debug. Release with it enabled fails
  the build. Debug alone must not enable test mode.
- The diagnostic build additionally requires an explicit startup switch
  --layout-test. Without it, it starts the ordinary demo with no diagnostics.
- Supported test options are --layout-language=zh-Hans|en-US and
  --layout-profile=baseline|long-status|long-country|rtl-mixed|unknown-usage.
  Defaults after explicit activation: zh-Hans and baseline. No live switching.
- Reject unknown, duplicate, conflicting or malformed test options before making
  a MainWindow. A normal build given any reserved layout option rejects it; never
  silently pretend a requested diagnostic run was ordinary success.
- Parsing errors are reported without echoing arbitrary argument values; return
  a distinguishable failure exit code. No partial window, retries or file writes.
- Keep diagnostic output/intermediate directories separate from ordinary outputs,
  so a test build cannot overwrite the existing accepted-for-observation artifact.
  Compile test-only loader/overrides/measurement support and embedded samples only
  into the diagnostic build. The small reserved-option rejection guard may remain
  in ordinary builds; no arbitrary reflection/plugin surface is introduced.

## Presentation contract

One immutable configuration is chosen before creating any view. A diagnostic run
always starts at Login with fresh ordinary DemoState. Navigation and simulated
events continue through existing handlers. Profiles never dispatch model events,
change node IDs/tags, lock country, entitlement, control enablement or selection.
Their effect survives ordinary Refresh calls; logout retains the chosen test
profile for this process but resets model state exactly as before. Closing and
reopening without activation returns to ordinary behavior; nothing is saved.

Use a small presentation adapter with a production identity/default path and a
diagnostic override implementation. Keep MainWindow.Copy behavior unchanged when
diagnostics are inactive. Dynamic labels and static x:Uid-derived labels must use
the same chosen language in test mode, including sidebar, scenarios and dialogs.
Diagnostic language text is read from build-embedded copies of the existing resw
sources, not duplicate translations, OS locale overrides or user-selected files.

Apply static localized properties through an explicit reviewed mapping of known
view/control targets and string properties, not arbitrary resource-driven property
assignment/reflection. The builder may choose generated or handwritten mapping;
coverage checks must account for every current x:Uid/property resource used by the
four pages and shell. Dialog buttons/titles and dynamic state labels use that same
adapter. New unmapped localization targets fail a contract check rather than being
silently mixed-language. Reuse the same visible control instances and handlers.

Required fixture IDs come from tests/ui-copy-cases.json embedded at build time.
Validate IDs, uniqueness, required fields, direction and bounded text length;
invalid/missing required fixtures stop diagnostic startup before a demo window.
Builder sets a documented conservative length cap above current samples.
Do not edit existing fixture text or call it a translation.

| Profile | Exact presentation target / precondition | Preserved semantics |
| --- | --- | --- |
| baseline | All four pages in selected language, no synthetic replacement, existing LTR layout | All labels and state follow ordinary selected-language presentation |
| long-status | ConnectionStatus while model is Connecting only | Operator explicitly requests connection; other phases retain normal status, action label/enablement unchanged |
| long-country | Display label of JP group and jp-1 list row; append original jp-1 identity to row | Other rows unchanged; IDs, selection, confirmation country semantics and original search matching remain unchanged. Does not certify searching injected long labels |
| rtl-mixed | Root layout RightToLeft; Login heading displays rtl-mixed text; other pages inherit RTL | Actual demo email blocks remain explicitly LTR on Login/Account; technical identifiers keep LTR isolation. Synthetic bidi text itself is a stress sample, not natural-language shaping proof |
| unknown-usage | Account UsageValue only when entitlement is Unknown or Stale | Operator chooses an existing scenario; no fabricated numeric usage, entitlement mutation or override of known states |

The sample's exact text must appear at its target (long-country row may add the
stable node suffix). Never replace the permanent offline warning, account identity,
primary action meaning or error/refusal messages with unrelated samples. Ordinary
profile/language does not add extra Tab stops or remove accessible control names.
If injected labels are clipped, record the defect; do not silently modify layout
or reduce font sizes as part of this facility.

## Measurements without changing the page under test

Use the native window title for a compact test-only marker/readout; do not add an
in-content toolbar, overlay, extra window or keyboard shortcut. This avoids taking
space from the measured layout or changing page focus traversal. Title includes:
LAYOUT TEST, selected language/profile, root logical width/height, PageHost content
width/height, rendering scale and actual text-scale factor if reliably available.
Label root/client and page-content measurements separately; neither outer-window
size nor screenshot pixels may substitute. Unavailable values say unknown, never1
or a requested target. Explicitly distinguish requested/selected profile from its
active target condition (for example long-status inactive outside Connecting).

Measure after layout and update on actual size/scale/page/state changes; avoid a
timer, polling loop, forced layout or live-region announcement of the title.
Implementation must check the installed framework's supported measurement APIs;
this design specifies meanings rather than assuming screenshot dimensions are DIPs.
Values must be finite/positive after layout, otherwise show pending/unknown.
Builder chooses rounding (retain sufficient precision to avoid false exact-size
claims) and harmless duplicate-update suppression. Do not dispatch model events.

No automatic resize, maximization, monitor move, display/text-scale change, locale
change, clipboard write or diagnostic file export. In separately authorized native
sessions, operator can resize normally and record measured values. A target counts
only when the measured client/root size matches the agreed tolerance and the full
client is available; offscreen windows and approximate outer bounds do not qualify.
The implementation plan must define the comparison tolerance and evidence fields
before runtime checks. This facility does not waive unavailable monitor targets.

## Validation requirements for the subsequent plan

These are required categories, not permission to run commands or a test budget:

- Pure tests: activation/option matrix, invalid input, fixture validation, exact
  profile mapping, conditional overrides, no alteration of input DemoState, and
  identity/default output when inactive.
- Declaration/resource checks: both language mappings complete; dialog/dynamic
  strings covered; offline warning and email LTR preserved; fixture placement uses
  existing views; no diagnostics controls in ordinary keyboard traversal.
- Artifact checks: ordinary build excludes diagnostic payload; Release+opt-in
  rejected; separate output paths; diagnostic build without switch is ordinary.
- Regression: existing suite plus new checks, ordinary and diagnostic build
  results recorded separately. Previous160/160 is not a fresh result.
- Bounded native checks after explicit allowance: ordinary-mode comparison;
  baseline language presentation; each fixture at its exact condition; title
  dimensions versus independently verified client measurement; resize update;
  default model behavior/ordinary close. No automatic reruns.
- English diagnostic rendering verifies explicit English presentation only.
  Missing-resource fallback is still a separate requirement; do not emulate a
  custom fallback and claim Windows resource fallback passed.
- Synthetic RTL/enlarged layouts, reader speech, contrast, accurate dimensions
  and screenshot evidence remain unverified until independently observed. Text
  scale readout alone does not test enlarged text. No font-size override here.

## Error handling and security boundaries

All inputs are fixed enums or committed synthetic resources. No path/URL option,
secret, account submission, shell execution, installation or network dependency.
No change to Core reducer or authentication/security gates. Keep diagnostic
markers explicit so screenshots cannot be mistaken for product-language approval.
Unexpected startup/fixture errors fail closed without opening the demo; errors
after window creation stop evidence collection and are reported, not repaired by
changing settings. Normal close remains available. This is not a security boundary
against someone rebuilding source; it prevents accidental test-mode shipping/use.

## Scope, open implementation choices and handoff

Expected touchpoints: App startup/build configuration, a small App-side presentation
adapter, explicit page/shell integration, test-only measurement title, existing
fixture/resource embedding, tests and recovery documentation. No new executable,
Core behavior, general localization framework or product language settings page.
Builder chooses class/file names, supported framework APIs and test organization
within these contracts. A native API limitation that changes isolation, measurements
or language behavior must be reported, not silently replaced with weaker evidence.

After written-design approval, create the implementation plan with explicit files,
regressions, build/launch budgets and stop conditions. Approval of this design does
not itself authorize runtime checks. Retain Task11/12 plus44/34/6/217 NOT_ACCEPTED,
no main merge, and save necessary checkpoints to the existing development branch.
