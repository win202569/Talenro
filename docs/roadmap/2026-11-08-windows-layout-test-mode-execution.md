# Layout diagnostics Stage A execution

Plan: docs/superpowers/plans/2026-11-08-windows-layout-test-mode.md.
User approved Native execution and Stage A budget after80307866. Base80307866.
Tasks1–4 targeted checks passed; Task5 in progress. No native launches authorized.

Budget used: targeted8/8; full0/1; cache-only restores0/2; builds0/3.
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
