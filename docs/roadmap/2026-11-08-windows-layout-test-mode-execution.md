# Layout diagnostics Stage A execution

## Next-stage planning checkpoint

After7e25a4fc, prepared docs/superpowers/plans/2026-11-08-windows-layout-native-stage-b.md:
14 proposed invocations (6 rejected startups,8 visible sessions), sequential and
bounded with no retries/builds/settings changes. Await user approval. This planning
turn ran no tests/builds/restores/native launches. Prior222/222 evidence unchanged;
all native outcomes remain unverified until observed. No acceptance or main merge.

## Follow-up after 1e0a5c4f — minor coverage finding closed

User approved continuing with the recommended bounded test-only correction.
Added five in-memory negative cases (Shell/Login/Connection/Nodes/Account): a new
localized control without resource properties must be rejected by the same coverage
validator used for real declarations. The validator now rejects zero matching keys;
removed the vacuous invented-key absence assertion. App/Core sources unchanged.

Fresh ordinary-user verification, Debug and --no-restore:
- RED: LayoutViewContractTests, 16 total, 5 expected failures and 11 passes, exit1.
  All five failed because the old validator did not throw InvalidDataException.
- GREEN: same filter, 16/16 passed, 0 skipped, exit0, no warnings.
- Full suite: 222/222 passed, 0 failed/skipped, exit0, no warnings.

Ignored local receipts under clients/windows/Talenro.Demo.Tests/TestResults:
layout-unmapped-red.trx, layout-unmapped-green.trx, layout-unmapped-full.trx.
The follow-up allowance is consumed: targeted2/2 and full1/1. No application build,
dependency restore, native launch or system-setting change. Earlier build hashes
below are historical, not freshly checked here. The deferred minor finding below
is superseded by this closure. Native Stage B remains unapproved/unverified;
next prepare its finite observation matrix. Wider NOT_ACCEPTED and no-main-merge
boundaries remain unchanged. Raw receipts/caches are not needed to recover source.

## Original Stage A record (historical)

Plan: docs/superpowers/plans/2026-11-08-windows-layout-test-mode.md.
User approved Native execution and Stage A budget after80307866. Base80307866.
Tasks1–5 Stage A checks complete, with one deferred minor test-coverage finding.
Native behavior remains UNVERIFIED; no native launches authorized/performed.

Budget used: targeted8/8; full1/1; cache-only restores1/2; builds3/3.
Remaining restore slot is not a new test/build/launch allowance and is not needed.
Preflight: ordinary token(False), no Talenro.Demo.App process, clean linked worktree
on codex/c12-b01-task9-coordinator; dotnet at C:/Program Files/dotnet/dotnet.exe.
Existing assets identify clients/windows/.nuget/packages and .dotnet-home locally.

Ruling: use PowerShell commands and this portable ledger instead of Bash-only
workspace/task helpers; preserve the task brief in the committed plan. No extra
task-done or baseline test calls: the approved GREEN calls are the completion
checks and the final full suite is reserved for Task5. Cost: less helper automation,
not reduced approved verification. Native launch budget remains zero.

Pre-flight interfaces: Task1 options consumed by2/4 match; Task2 presentation and
fixture catalog consumed by3 match; Task3 window/state integration consumed by4
matches; Task5 verifies all earlier build/source artifacts. Task1 startup resource
validation depends on Task2, so wiring is completed with Task2/3 before any app
build. Intermediate commits are partial checkpoints, not runnable acceptance.

Task1: RED22/22 expected failures (unimplemented parser/guard, missing build gates),
GREEN22/22 passed,0 skipped,exit0. Filter FullyQualifiedName~LayoutOptionsTests,
dotnet test --no-restore -c Debug, ordinary token checked before both calls.
Local GREEN receipt: clients/windows/Talenro.Demo.Tests/TestResults/layout-options-green.trx.
Build declarations are static checks only; evaluated isolation and actual builds
remain Task5. Diagnostic startup resource/presentation wiring remains Tasks2/3.

Task2: RED15/15 expected missing implementation/exception-contract failures,
GREEN15/15 passed,0 skipped,exit0 with CS9124 warning. Constructor parameter was
captured by methods as well as copied into Options. Removed duplicate capture by
using Options consistently; behavioral equivalence is to be rechecked by the
remaining planned full suite, not an extra targeted rerun. RED stub CS9113 warnings
were removed with implementation. GREEN receipt layout-presentation-green.trx.
Task2 commit is a tested-behavior checkpoint with warning cleanup pending full
verification, not a warning-free build claim. Startup/view integration not yet built.

Task3: RED11/11 missing bindings/hooks, GREEN11/11 passed,0 skipped,exit0 and no
compiler warnings. Explicit typed bindings cover shell/four pages and both resource
inventories. Source/declaration checks do not establish runtime wiring. Added
diagnostic initialization guard to Nodes so XAML selection callbacks cannot bind
not-yet-created controls; final full suite/build still pending. Receipt layout-views-green.trx.

Task4: RED9/9 expected unimplemented formatter/missing observer, GREEN9/9 passed,
0 skipped,exit0,no compiler warnings. Receipt layout-measurement-green.trx. Native
observer itself is not executed by pure/declaration tests. Local WinUI2.6.5 XML
documents ActualWidth/ActualHeight, Loaded/SizeChanged, XamlRoot.Changed and
RasterizationScale (physical pixels per view pixel); SDK10.0.19041.57 XML documents
UISettings.TextScaleFactor/TextScaleFactorChanged. Runtime correspondence still
requires Stage B. Readout never asserts target PASS; unavailable values are unknown.

## Task5 results

- Full suite217/217 passed,0 failed/skipped,exit0,no warnings; layout-full.trx.
  This is the Windows demo suite, not acceptance of the unrelated217 runtime gate.
- Ordinary Debug build: exit0,0warnings/errors,10.24seconds.
- Diagnostic Debug build: exit0,0warnings/errors,10.98seconds.
- Release+diagnostics: expected exit1,LAYOUT001,0warnings; rejected before compilation
  in0.36seconds, not a missing-assets error. No app execution.
- Cache-only restore1/2: --locked-mode with the sole empty local feed and existing
  package cache, NuGetAudit=false,exit0. No package/version/lockfile changes.
- MSBuild evaluation before/after restore confirms ordinary obj/bin and diagnostic
  obj/layout-diagnostics/bin/layout-diagnostics paths, distinct assets, diagnostic
  symbol only in opt-in. No bin/obj generated sources leaked into evaluated Compile.
- verify-layout-artifacts.ps1 used PE metadata only under PowerShell7.6.9: ordinary
  has only reserved-argument guard, no diagnostic resource/type payload; diagnostic
  has9 required types and3 resources. Script created after reviewer range and
  inspected/executed by main agent, not independently reviewed. No Assembly.Load.
- Final process query found no Talenro.Demo.App. Combined read-only command's exit1
  came from expected Get-Process absence; metadata script explicitly reported PASS.

Artifacts below are local ignored build outputs under clients/windows/Talenro.Demo.App.
They are not portable downloads or complete dependency-attestation records.

| Variant | Relative output directory | App.dll SHA256 |
| --- | --- | --- |
| Ordinary | bin/Debug/net10.0-windows10.0.19041.0/win-x64 | C39E19BC4700DA25CFA6EC90A696E8627C922E7FD8B026BD08EB2A29D5F36C8E |
| Diagnostic | bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64 | C16A8479C6BD6826BDD649651645C28A49A772AE6B7BD9CC8DCBDB38C0D96C63 |

Both App.exe apphosts hash A32B55694C5C5BC380474C6E2F66F750A1DD9F7C9780D4E141CA05325975FF71.
Therefore executable hash alone cannot distinguish these variants: record the
managed App.dll hash and exact directory too. Prior native-session hashes/results
are historical, not fresh observations on these outputs.

## Independent review and rulings

Fresh reviewer examined80307866..72572cc4 read-only: no Critical/Important source
finding. Final: minor (deferred): LayoutViewContractTests skips an x:Uid with zero
matching resource keys; its invented-key absence assertion is not a genuine
negative coverage test. Current controls are covered, but future new untranslated
controls could escape this check. Needs a real failing fixture/validator test and
fresh test allowance; not silently fixed after targeted budget was consumed.

Reviewer exclusions were explicitly reconciled:
- Final: Ruling: compilation/evaluated isolation/Release/binary leakage are decided
  by the main agent's Task5 outputs above, not reviewer inference; cost if wrong is
  accidental diagnostic leakage. All listed checks obtained, no native equivalence claim.
- Final: Ruling: full suite/warnings use the fresh217/217 parent run and two clean
  builds, not historical targeted counts; cost if wrong is a regression being missed.
- Final: Ruling: startup exit/window behavior and diagnostic-no-switch behavior stay
  UNVERIFIED until separately budgeted launches; cost is an undiscovered startup bug.
- Final: Ruling: live refresh/logout/namescopes/focus/dialogs/clipping/RTL/title
  visibility stay UNVERIFIED; source coverage does not substitute. Cost is a UI defect.
- Final: Ruling: root/client equivalence, exact dimensions and event timing/text
  scale availability stay UNVERIFIED; cost is false layout evidence if overclaimed.
- Final: Ruling: reader speech/contrast/enlarged text/Windows fallback are outside
  Stage A and stay UNVERIFIED; cost is an accessibility claim unsupported by evidence.
- Final: Ruling: artifact script and handoff docs, outside reviewer commit range,
  were checked by main agent; cost is less independent coverage of this small harness.

Ruling: preserve existing development branch/worktree and push necessary files
under AGENTS standing instruction; no PR, merge or cleanup. Use already approved
full-suite evidence on unchanged tested source rather than an extra finishing-skill
rerun. Cost: none of this constitutes integration or native acceptance.

Next recommended work: close the deferred negative-localization test gap with
separately approved RED/GREEN and full-suite calls, then prepare Stage B launch
matrix. Neither next step has run. No system settings/install/network app action.
