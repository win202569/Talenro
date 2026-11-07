# Windows Four-Screen Offline Demo Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans for inline execution,
> if the user selects that method. This plan is not yet execution-approved.

**Goal:** Build the approved four-screen native offline demo without real credentials,
network, VPN services or payments.

**Architecture:** A deterministic in-memory state model supplies a native WinUI 3
shell and four views. Views dispatch explicit events; no timers or background
workers simulate success. Pure model tests and visual/native checks are separate.

**Tech Stack:** C#, WinUI 3, XAML, .resw, PowerShell; exact SDK/package/tool pins
must be qualified before this becomes a runnable plan.

**Spec:** [Approved interaction design](../specs/2026-11-07-windows-client-four-screen-design.md),
approved by user after86dfc037.

**Status:** Implementation breakdown complete; NOT READY TO EXECUTE. Toolchain
qualification and a finite execution allowance remain prerequisites. This document
does not authorize installation, package restoration, compilation or launching.

## Global constraints

- Native Windows only; no replacement HTML/browser prototype or new service.
- Fixed fictional credentials, no user-entered passwords; no networking, telemetry,
  persistence, account mutations, drivers, IPC, real tunnel or payment actions.
- State is memory-only; exit/reset returns to initial fixture. All active connection
  states say simulated. Secondary destinations explain their exclusion.
- Selected four PNGs are visual references, not production image backgrounds.
- Target1440x1024, minimum1024x768 logical layout; explicit keyboard/focus/DPI/RTL checks.
- zh-Hans first, English fallback resources; RTL/long strings are test-only samples,
  not approved Persian translations. Brand artwork remains provisional.
- Keep Go/C12/Task11/12 sources and acceptance untouched;44/34/6/217,NOT_ACCEPTED.
- Save reviewed source/tests/resources/pins/recovery docs; exclude SDKs, binaries,
  caches, private credentials and machine-specific build output. No main merge.

## Observed prerequisite gap and finite gate

At baseline86dfc037, repository search found no .csproj/.sln/.xaml/global.json.
Get-Command found no dotnet or MSBuild on PATH. Both standard Program Files dotnet
roots and both standard Visual Studio roots were absent. Windows Kits/10 exists,
but its Lib directory was not found in the initial check. This is a limited
filesystem/PATH inventory, not proof that no custom installation exists.

[Microsoft WinUI guidance](https://learn.microsoft.com/windows/apps/winui) confirms
the native framework is delivered with Windows App SDK. The current
[getting-started guide](https://learn.microsoft.com/en-us/windows/apps/get-started/start-here)
is a reference for toolchain selection, not authorization to execute its install
or launch commands. No SDK versions are guessed or downloaded here.

### Task 1: Qualify a toolchain, then finalize execution inputs

**Output:** `docs/roadmap/2026-11-07-windows-client-toolchain-readiness.md` and an
approved amendment to this plan. No product files before qualification.

- [ ] Obtain a user-provided custom tool path, or explicit approval for a separate
  developer-tool installation scope. Do not repeatedly search the whole disk.
- [ ] Record exact .NET SDK, Windows SDK, Windows App SDK, build/test tooling,
  architecture and deployment approach compatible with official documentation.
  Identify required local packages versus downloads and any machine changes.
- [ ] Specify exact project TFMs, package versions, lockfiles, approved restore
  sources and per-command build/test/launch limits in this plan before execution.
  Choose a deterministic test runner and define the exact RED/GREEN commands.
  A missing tool/dependency is a setup failure, never an accepted test RED.
- [ ] Separately authorize SDK install/download if needed. Stop if unavailable;
  do not generate unbuildable client scaffolding as a substitute for readiness.

Tasks2–5 below are the fixed implementation scope, not executable permission.
Their command recipes cannot be finalized honestly until Task1 supplies the
toolchain. No automatic native allowance transfers from the old Go/C12 work.

## Planned file ownership

All new code under `clients/windows/` (new directory; no files created this turn):

|Files|Responsibility|
|---|---|
|global.json, Directory.Build.props, NuGet.Config, project lockfiles|Exact qualified build inputs; local scope only|
|Talenro.Demo.Core/Talenro.Demo.Core.csproj; DemoState.cs; DemoEvent.cs; DemoReducer.cs; DemoFixtures.cs|Pure immutable state, events, transition function and fictional fixtures|
|Talenro.Demo.Tests/Talenro.Demo.Tests.csproj; SessionTests.cs; ConnectionTests.cs; NodeSelectionTests.cs; CountryLockTests.cs; EntitlementTests.cs|Deterministic core coverage, no OS/tunnel calls|
|Talenro.Demo.App/Talenro.Demo.App.csproj; App.xaml/.cs; MainWindow.xaml/.cs; app.manifest|Native app lifetime and shell; no service registration|
|Talenro.Demo.App/ViewModels/DemoViewModel.cs|INotifyPropertyChanged adapter and dispatch only|
|Talenro.Demo.App/Views/{Login,Connection,Nodes,Account}Page.xaml/.cs|Four views and minimal UI event adapters|
|Talenro.Demo.App/Resources/Theme.xaml; Strings/{zh-Hans,en-US}/Resources.resw|Shared spacing/color styles and resource identifiers|
|README.md; tests/manual-checklist.md; tests/ui-copy-cases.json|Reproduction, scenario/visual evidence and test-only long/RTL strings|

## Review focus

1. Stale completion after cancel/logout must not revive a simulated connection.
2. Connected node/lock changes must be atomic and reversible on Cancel.
3. Unknown entitlement must not look like active entitlement or zero usage.
4. Placeholder/security actions must never imply a real account change.
5. Enlarged/RTL text and keyboard dialogs must preserve action meaning/focus.

### Task 2: Implement the deterministic demo model, test first

**Interfaces:** namespace `Talenro.Demo.Core`; `DemoFixtures.Initial(): DemoState`;
`DemoReducer.Reduce(DemoState state, DemoEvent action): DemoState`.
State records page, demo session, connection state, committed/draft preference,
saved country lock, pending confirmation, entitlement, scenario and notice.
Events contain no credentials/network handles. UI cannot write state fields.

- [ ] Define failing SessionTests: initial Login/no session; EnterDemo→Connection;
  Logout→exact initial state, clearing preferences, notices and pending operations.
- [ ] Define ConnectionTests: explicit request→Connecting; explicit scenario result
  →SimulatedConnected; Cancel→Disconnected; completion ignored outside Connecting;
  Disconnect→Disconnecting→Disconnected. Repeat events never start workers.
- [ ] Define NodeSelectionTests: Back discards draft; disabled node cannot apply;
  disconnected Apply updates only preference; connected Apply opens confirmation;
  Cancel preserves connection/preference; Confirm commits and enters Connecting;
  transitional states refuse node editing. Cross-country confirmation includes
  old/new country and updates saved lock only after confirmation.
- [ ] Define CountryLockTests: initial off; missing country refuses enable; locked
  failure offers exactly3 actions; one-time override does not alter saved lock,
  is consumed by the next explicit recovery transition, and cannot leak to later
  failures; retry remains in-country; stay disconnected makes no transition request.
- [ ] Define EntitlementTests: exhausted/expired/unknown/stale/device-limit each
  refuses Connect and confirmed reconnect; account remains accessible. Changing
  scenario to exhausted while connected returns Disconnected with hard-stop notice.
- [ ] Observe intended RED then implement minimal reducer/fixtures; GREEN the full
  core suite. Use exact commands/budgets from Task1, not implicit dotnet restore.
- [ ] Review defensive state ownership and commit only this unit plus pinned build
  inputs. No UI or OS runtime success is inferred from model tests.

### Task 3: Build the native shell, login and connection views

**Interface:** `DemoViewModel.State` read-only; `Dispatch(DemoEvent action)` applies
the reducer and emits property notifications. One shared model instance per app.

- [ ] Add shell/copy checks before views: permanent offline notice, fixed read-only
  demo email/password display, EnterDemo action, no editable password control;
  decorative power glyph not focusable; one primary connection action.
- [ ] Wire native navigation and excluded-destination explanation; only4 real pages.
  Scenario selector supplies explicit events and remains clearly separated from
  consumer actions. Login failure fixtures cannot submit credentials.
- [ ] Build against qualified pins, verify RED/GREEN checks and core regression;
  a UI compile is not a visual test. Commit shell/views/resources with recoverable
  project configuration. No native launch until the separate Task5 allowance.

### Task 4: Add nodes and account plus refusal dialogs

- [ ] Add checks for country grouping/search/no results/list unavailable, visible
  Protocol heading, disabled selection, draft discard and confirmation cancel.
- [ ] Bind Nodes to the reducer, not a second UI-local connection state machine.
  Bind Account to explicitly fictional usage and activated-device counts; unknown
  data shows unavailable rather than invented numbers. No purchase controls.
- [ ] Add explicit placeholder dialogs for reset/create/MFA/devices/language/delete;
  test that they leave state unchanged except notice. Exit demo dispatches Logout.
- [ ] Verify all core transitions and view bindings with exact approved commands;
  preserve original expected failures and stop for unexpected outcomes. Commit.

### Task 5: Bounded visual/accessibility verification and handoff

- [ ] After separate launch approval, open only the demo as an ordinary user; no
  elevated service, system route, account or network action. Use approved bounded
  session and graceful close; no arbitrary PID termination or retry loop.
- [ ] Compare four screens with saved targets, using the same viewport; check all
  alternate scenarios, node-change/country-lock dialogs and placeholders. Record
  design deviations rather than declaring image fidelity from compilation alone.
- [ ] Verify keyboard focus/order/escape/return, text+icon status, contrast,
  1024x768/1440x1024 and enlarged-text/DPI layouts, synthetic long/RTL strings,
  LTR-isolated demo email. Return dialog focus to invoking control. Capture only
  fictional data. Missing checks remain explicitly unverified.
- [ ] Close/reopen under the approved allowance to establish initial reset. Static
  absence of network/storage APIs is useful evidence, not proof of all OS behavior.
- [ ] Final whole-scope review per execution skill; preserve reviewed source,
  resources, pins and reproduction instructions. Push existing development branch,
  compare remote/local SHA, disclose remaining issues; no production acceptance.

## Self-review and current handoff

All approved screen states map to Tasks2–5; five review risks have concrete checks.
No schema/DTO from Task11/12 is consumed. Model and shell ownership are separated;
secondary flows stay excluded. No exact SDK/package/test-command claim is made
without evidence. This is deliberately a readiness-gated plan: Task1 is unresolved,
so no RED/GREEN invocation, scaffolding, restore, build or launch is ready now.

Recommended execution method once qualified: inline, because the four views share
one tightly coupled state model; one final independent review remains required.
Next user decision is whether to supply an existing custom toolchain location or
permit a separately bounded installation proposal. This plan itself installs nothing.
