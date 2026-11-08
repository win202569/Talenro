# Windows Layout Test Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the approved default-off diagnostic presentation mode without changing ordinary demo semantics.
**Architecture:** Guard the existing executable with an explicit build property and startup flag. A pure App-side presentation adapter supplies test text; the same four views and handlers render it. An event-driven title readout measures actual content without adding page UI.
**Tech Stack:** Existing C#/net10.0 Core/tests, WinUI net10.0-windows10.0.19041.0 App, Microsoft.WindowsAppSDK2.6.5, xUnit and PowerShell; no added package.
**Spec:** docs/superpowers/specs/2026-11-08-windows-layout-test-mode-design.md (approved after f376dcbd).

Status: Native execution and Stage A allowance approved after80307866; Stage A
implemented/verified; the minor test gap is closed by the test-only follow-up
(targeted16/16 and full222/222). See
docs/roadmap/2026-11-08-windows-layout-test-mode-execution.md for all receipts/rulings.
Stage B is not authorized. Historical proposal wording below is superseded here.
Existing worktree: D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator.
All paths below are relative to it; commands run from clients/windows unless stated.

## Global Constraints

- EnableLayoutDiagnostics defaults false in every configuration, including Debug. Release with it enabled fails the build.
- Startup requires --layout-test; language zh-Hans|en-US; profiles baseline|long-status|long-country|rtl-mixed|unknown-usage.
- No live switching. No OS locale/display/text-scale changes, installs, real network actions, persistence or external fixture paths.
- No Core reducer/events, account identity, offline-warning meaning, control enablement or selection changes.
- Compile diagnostic payload only into explicitly enabled builds; preserve ordinary Copy/resource behavior.
- Actual root/client measurements are not screenshot or outer-window dimensions. Unavailable values remain unknown.
- No new product language UI, viewport resizing helper, general localization framework, extra window or content toolbar.
- All tests/builds use an ordinary-user token; stop on administrator token or unclear process ownership. Never change UAC/firewall/timeouts.
- Previous160/160 is historical. No main merge; Task11/12 plus44/34/6/217 remain NOT_ACCEPTED.

## Review Focus

1. Reserved arguments without activation, duplicated and unknown values must not silently run ordinary mode (Task1 matrix).
2. Initialization/Refresh/logout can overwrite test strings or cause mixed-language views (Task3 refresh wiring and resource coverage checks).
3. Invalid/NaN/stale dimensions or scale changes can produce false target claims (Task4 formatter and lifecycle checks).
4. Alternate builds can reuse stale generated XAML/assets or leak diagnostics to ordinary outputs (Task1 path contracts and Task5 artifact checks).
5. RTL/long labels can alter node identity/search/confirmation or hide state meaning (Task2 invariance and Task3 existing-handler checks).

## Execution scope and proposed finite budget

Stage A implements and verifies code only: four targeted RED/GREEN pairs (8 test
calls), one full-suite call, at most two cache-only locked restores (App diagnostic
graph and test graph if required), three build calls (ordinary Debug, diagnostic
Debug, expected rejected Release+diagnostics). No native app or Narrator launches.
MSBuild evaluation/property reads are allowed separately; they must not invoke
compile/restore targets. Each dotnet operation has a10minute ceiling; stop and
report an overrun without launching another or killing unrelated processes.
Unexpected failure stops dependent work. Expected RED and the deliberate Release
rejection are not defects. No unbudgeted reruns or new dependencies/downloads.
Read-only investigation and clearly unfinished documentation checkpoints remain
allowed after failure. Stage B native observations require a later finite allowance.

Before Stage A, locate the previously used dotnet/runtime/cache paths from local
assets and execution ledger; verify access, token and no demo window/process.
Do not assume C:/Users/Lenovo/.nuget/packages exists (not found during planning).
Keep host/cache paths out of portable build files. Missing prerequisites block
execution; they do not authorize install or a change in machine settings.

## File responsibilities

Create under clients/windows/Talenro.Demo.App:
- Layout/ReservedLayoutArguments.cs: small ordinary-build reserved-option guard.
- Diagnostics/LayoutOptions.cs: immutable options and strict diagnostic parser.
- Diagnostics/LayoutFixtureCatalog.cs: bounded embedded JSON validation.
- Diagnostics/LayoutPresentation.cs: pure language/profile mapping, no UI/model mutation.
- Diagnostics/LayoutStaticBindings.cs: explicit typed control/property bindings.
- Diagnostics/LayoutMeasurement.cs: pure measurement record/title formatter.
- Diagnostics/LayoutWindowObserver.cs: WinUI event subscriptions/readout lifecycle.
Modify App.xaml.cs (gated startup), MainWindow.xaml.cs (presentation/observer
integration), MainWindow.xaml and Views/{Login,Connection,Nodes,Account}Page.xaml
(names only where necessary), and their .xaml.cs files (guarded presentation hooks).
Modify Directory.Build.props and App csproj for early isolated paths/compile/embed
gates. Core files, existing resource text and fixture text remain unchanged.
Create four test classes: LayoutOptionsTests, LayoutPresentationTests,
LayoutViewContractTests, LayoutMeasurementTests in Talenro.Demo.Tests; extend its
csproj to link only pure new sources and copy relevant declarations/fixture files.
Create clients/windows/tests/verify-layout-artifacts.ps1 for non-executing PE metadata
checks; it must never load/run the target App or its initialization code.

## Interfaces used across tasks

All new types use namespace Talenro.Demo.App.Layout; diagnostic files compiled only
with LAYOUT_DIAGNOSTICS, except pure copies linked into the test assembly.

- ReservedLayoutArguments.ContainsReserved(IReadOnlyList<string> args): bool;
  reserve case-insensitive --layout prefix; ordinary unrelated args keep existing behavior.
- LayoutOptions(string Language, string Profile); LayoutParseResult(bool Enabled,
  LayoutOptions? Options, string? ErrorCode). LayoutOptions.Parse(IReadOnlyList<string>
  args): LayoutParseResult; ordinal exact supported names/values, defaults zh-Hans/baseline.
- LayoutFixtureCatalog.Parse(string json): LayoutFixtureCatalog;
  Text(string id): string; required four IDs, text1..4096 UTF-16 units, JSON <=64KiB,
  no duplicates/unknown IDs, direction RightToLeft only on rtl-mixed; reject malformed input.
- LayoutPresentation(LayoutOptions options, LayoutFixtureCatalog fixtures,
  IReadOnlyDictionary<string,string> strings); Copy(string key): string;
  Text(string target, DemoState state, string ordinaryText): string;
  IsProfileActive(DemoState state): bool; IsRtl: bool.
- LayoutStaticBindings.Apply(string viewId, FrameworkElement root,
  LayoutPresentation presentation): void. viewId is Shell/Login/Connection/Nodes/Account.
  Bindings have explicit resource-key and typed assignment; no property reflection.
- LayoutMeasurement(double? RootWidth, double? RootHeight, double? PageWidth,
  double? PageHeight, double? RenderingScale, double? TextScale);
  LayoutTitle.Format(LayoutOptions options, bool active, LayoutMeasurement values): string.
- LayoutWindowObserver(MainWindow window, FrameworkElement root,
  FrameworkElement pageHost, Func<bool> active, LayoutOptions options);
  Refresh(): void; Dispose(): void. Produces title only, never model events.

## Task1 — Fail-closed activation and build isolation

- [x] Write LayoutOptionsTests plus compilable unimplemented signatures above.
  Assert empty args => Enabled=false; --layout-test => zh-Hans/baseline; every valid
  language/profile combination accepted; selectors without activation, duplicate
  flags/selectors, --layout-test=true, unknown --layout-* and invalid casing/values
  produce a fixed error code and no Options. Diagnostic parser rejects unrelated
  arguments too; ordinary guard preserves unrelated args. Errors never echo input.
  Assert project gates exclude diagnostic Compile/EmbeddedResource when false,
  explicit opt-in creates LAYOUT_DIAGNOSTICS, and Release opt-in has an error target.
- [x] Run targeted RED with filter FullyQualifiedName~LayoutOptionsTests, confirm
  assertion/NotImplemented failure for intended contract, not environment failure.
- [x] Implement parser, fixed error code2 startup rejection and early build guards.
  App reads process arguments excluding executable; ordinary build uses only the
  reserved guard. Diagnostic build validates options/resources before MainWindow.
  Report fixed error text to stderr when available plus exit code; no error window.
  Disabled diagnostic build branch creates exactly the original ordinary window.
- [x] In Directory.Build.props, conditional on MSBuildProjectName=Talenro.Demo.App
  and opt-in, set early BaseIntermediateOutputPath=obj/layout-diagnostics/ and
  BaseOutputPath=bin/layout-diagnostics/. Ensure MSBuildProjectExtensionsPath follows
  isolated intermediate location. Exclude ALL obj/** and bin/** from source globs,
  not just the active variant. Core/tests retain their own paths; no outputs deleted.
  Embed the existing JSON/resw only with opt-in; exclude Diagnostics/**/*.cs normally.
- [x] Run targeted GREEN (same filter), expect all selected tests pass/exit0.
- [x] Review explicit file diff and commit scoped checkpoint; incomplete runtime
  integration remains clearly labelled, no native acceptance claim.

## Task2 — Pure profile and language presentation

- [x] Write LayoutPresentationTests. Assert exact fixture text at ConnectionStatus
  only for Connecting; JP group text exactly long-country and jp-1 row exactly
  long-country + " · jp-1"; UsageValue override only Unknown/Stale; rtl heading exact
  fixture string; all other target/state combinations return ordinaryText.
  Target IDs: ConnectionStatus, CountryJP, NodeJp1, LoginHeading, UsageValue.
  Assert input DemoState equality before/after calls, default text unchanged,
  exact zh/en resource values, and missing required key fails diagnostic preparation.
  Assert invalid fixture size/IDs/duplicates/direction rejected; no original text edits.
- [x] Run RED filter FullyQualifiedName~LayoutPresentationTests.
- [x] Implement the pure catalog/presentation interfaces. Load selected embedded
  resw using bounded XML parsing with DTD disabled, one dictionary keyed by original
  dotted resource keys. No OS language override or custom missing-key fallback.
  Fail startup if either language's required view/dynamic/dialog key is absent.
  IsProfileActive: baseline always; long-status on Connection+Connecting; long-country
  on Nodes; rtl-mixed always (root mirrored); unknown-usage on Account+Unknown/Stale.
  Report condition activity, not a claim that an offscreen target has been inspected.
- [x] Run GREEN same filter; review and commit only this task's files.

## Task3 — Shared-view integration and language coverage

- [x] Write LayoutViewContractTests against copied source declarations and binding
  manifest. Enumerate each view's x:Uid plus applicable resw property; require one
  explicit binding per view/target/property, both languages covered. Verify dynamic
  and dialog keys from current source/resource inventory, including composed state
  keys. Keep existing ShellContractTests assertions (offline banner, noneditable
  credentials, PageHost no Tab stop). Add tests for init/Refresh/logout hooks and
  preserved handler/Tag/search/confirmation expressions. Label these static checks,
  not proof of live event wiring or focus. Add failure case for a new unmapped UID.
- [x] Run RED filter FullyQualifiedName~LayoutViewContractTests.
- [x] Add stable x:Name only where bindings need it; implement explicit typed
  assignments in LayoutStaticBindings, with an inspectable key/target manifest
  matching actual setters. No generic property setter or new visual controls.
  Apply selected-language static text after InitializeComponent and before profile
  replacements. Keep adapter initialized before constructors can trigger callbacks.
  Route diagnostic MainWindow.Copy through adapter; ordinary path unchanged.
- [x] Apply profile text after relevant Refresh output; reapply Login heading after
  Reset. Nodes changes display label only, preserving search on original selected-
  language labels and node identity/selection/confirmation. For RTL set root direction
  only in mode; isolate email and technical node/lock identifiers LTR without adding
  focusable wrappers. Keep actual state/action labels, notice meaning and enablement.
  Title observer must see state/page refreshes without subscribing to new actions.
- [x] Run GREEN same filter. Review default and active paths separately; commit.

## Task4 — Actual dimensions and readout lifecycle

- [x] Write LayoutMeasurementTests. Assert prefix LAYOUT TEST, language/profile,
  active/inactive, distinct root/page dimensions, render/text scales; unknown for
  null/zero/negative/NaN/infinity, invariant formatting with3 decimal places.
  Assert no equality decision based on rounded title. Add declaration checks for
  event-based observer, unsubscription on close, no timer/forced layout/Dispatch.
- [x] Run RED filter FullyQualifiedName~LayoutMeasurementTests.
- [x] Implement pure record/formatter; inspect installed framework reference docs
  before selecting native read APIs. Use actual laid-out Root and PageHost bounds
  in logical units plus root rendering scale. Read text scale only through supported
  read-only API; otherwise unknown. Use Loaded/SizeChanged/root-scale and state/page
  notifications; invalidate unavailable/old measurements on page/root replacement.
  Handle UI-thread delivery safely, suppress identical titles, dispose subscriptions
  on close. No live-region declaration, polling, model mutation or automatic resize.
- [x] Run GREEN same filter. Record API evidence/unknown fields and commit.

## Task5 — Whole-change checks and portable handoff (no native launch)

- [x] Independent whole-change code review after Tasks1–4; resolve findings only
  within scope. New verification calls beyond budget require further allowance.
- [x] Read-only evaluate build properties for ordinary and diagnostic App, confirm
  output/intermediate/project-assets isolation and effective source/embed item lists.
  A declared path without evaluated evidence is insufficient. Use dotnet msbuild
  -getProperty / -getItem only, not an implicit Build target.
- [x] If isolated assets need restore, use at most two locked cache-only restores.
  Use an existing verified package cache and an explicit empty local feed directory
  under .cache as the sole --source, with --locked-mode and NuGetAudit=false; no
  network source or lockfile update. Diagnostic App restore carries
  -p:EnableLayoutDiagnostics=true. Missing package stops work; no download fallback.
- [x] Run full tests once with --no-restore -c Debug, expect exit0 and all tests
  pass/none skipped. Record actual total, never assume it remains160.
- [x] Build ordinary Debug once (--no-restore), then diagnostic Debug once with
  -p:EnableLayoutDiagnostics=true. Both must exit0,0 warnings/errors. Run Release
  opt-in once with --no-restore, expect fixed diagnostic-build gate error before
  compilation rather than missing-assets failure; gate scheduling must ensure this.
- [x] Implement/run verify-layout-artifacts.ps1 against the two managed App outputs
  using System.Reflection.Metadata/PEReader, not Assembly.Load or app execution.
  Ordinary must have no diagnostic types, fixture JSON or embedded diagnostic resw;
  diagnostic must have them. Check effective MSBuild item list and output paths too.
  Metadata presence alone does not prove ordinary startup behavior. Hash executable
  and managed App files separately, recording exact paths/configuration.
- [x] Update manual-checklist.md with pending new checks (do not check them off),
  README.md, current-status.md, execution ledger and repository-recovery.md. Record
  changes, calls consumed, errors, artifacts and unresolved native checks. Explicit
  file staging/commit/push to codex/c12-b01-task9-coordinator; verify remote SHA and
  clean status. Exclude all artifacts/caches/raw session receipts from commits.

## Commands and outcomes

Use PowerShell in clients/windows, the verified existing dotnet executable, process-
local cache/CLI variables and ordinary-user precheck. For each of Tasks1–4, two calls:
`dotnet test Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore -c Debug --filter FullyQualifiedName~CLASS`
where CLASS is the named class in that task. RED must fail for its intended assertion;
GREEN exit0. Full suite omits --filter. Builds target
`Talenro.Demo.App/Talenro.Demo.App.csproj --no-restore -c Debug`, with/without opt-in.
Release rejection uses -c Release -p:EnableLayoutDiagnostics=true. Capture exit code
immediately; commands shown are templates, not permission to run in this planning turn.

## Stage B — required future evidence, NOT authorized by Stage A

Prepare a separate session matrix after artifacts pass: ordinary executable no
arguments; diagnostic executable no activation; explicitly activated baseline in
zh-Hans and en-US; four profile conditions. Invalid-option rejection also requires
separately counted process invocations (not free because no window should appear).
Do not squeeze these independent launches into an old one-launch allowance.

For dimensions, record reported root/page values, rendering/text scales, independent
native client physical bounds and effective DPI, page/state/profile, visible full
client availability and timestamp. Use logical comparison tolerance ±0.5 DIP per
axis; additionally require independent converted client bounds to agree within
±0.5 DIP. Include the title's ±0.0005 DIP rounding interval: the entire possible
interval must fall inside tolerance, not merely its rounded midpoint. Title alone
cannot establish target attainment without independent client/DPI observation. If
that observer is unavailable, the readout may be smoke-tested but exact-size result
stays UNVERIFIED. Root/client matching must be checked for this window chrome before
equating them. Never force offscreen sizes to claim1440x1024. Text scale unknown,
reader speech, contrast and fallback stay explicitly unresolved.

## Self-review and execution handoff

Spec coverage maps gates→Task1, fixtures/languages→Task2, shared controls/default
behavior→Task3, title/lifecycle→Task4, artifact isolation/regression/recovery→Task5.
All five Review Focus risks have tests in their owning task. Interfaces use the
same names throughout. Native proofs are deliberately separated from source tests.
No code or verification command has been executed for this plan.

Recommend Native execution (one implementer in this chat, independent review at end):
the four implementation tasks share startup/presentation interfaces and are easier
to integrate sequentially. Alternative is fresh implementer/reviewer per task for
more intermediate independence at higher coordination cost. Await user choice and
Stage A budget approval before implementation; Stage B remains separately gated.
