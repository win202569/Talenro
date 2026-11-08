# Windows offline demo execution — stopped at dependency restore

User approved the execution amendment after5c61627c. Work occurs in existing
codex/c12-b01-task9-coordinator, not main. Task A is incomplete; Task B unstarted.

## Actual invocation

R1 ran once from clients/windows using C:\Program Files\dotnet\dotnet.exe:
`restore Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --configfile NuGet.Config --use-lock-file`.
Process-local telemetry opt-out1, skip-first-time1, local DOTNET_CLI_HOME and
NUGET_PACKAGES were set exactly as originally planned.
UTC start2026-11-08T07:11:27.5201731Z; end2026-11-08T07:11:34.6104226Z; exit1.
SDK10.0.403 reported. Core (no direct dependencies) restored in61ms; Tests failed
NU1301 after5.64s: api.nuget.org:443 host unknown, service index unavailable.
This identifies failure at name resolution in this execution environment, not
proof of a machine-wide outage or incompatible package pins. No network retry,
source substitution, DNS/security setting change or workload update was made.

First-use output also reported a failed global-tools PATH addition, workload
validation warning, and an installed ASP.NET Core HTTPS development certificate.
Read-only CurrentUser/My inspection for localhost/ASP.NET returned no matching
certificate. Actual certificate persistence is therefore unconfirmed; the message
alone is not proof. No trust/delete/export operation was performed.

[Microsoft CLI environment documentation](https://learn.microsoft.com/zh-cn/dotnet/core/tools/dotnet-environment-variables)
states skip-first-time is unsupported since.NET Core3.0; certificate generation
and global-tools PATH addition default on. Corrected future recipe explicitly
sets DOTNET_GENERATE_ASPNET_CERTIFICATE=false and DOTNET_ADD_GLOBAL_TOOLS_TO_PATH=false,
with DOTNET_NOLOGO=1. These settings are process-local and have not been tested.

## Preserved checkpoint and allowance

Only configuration/projects exist; no C# implementation, test cases or app project.
Core empty dependency lock is valid for its dependency-free project. Failed Tests
restore produced an empty dependency lock: NOT a valid resolved test graph. Its
original is retained only in this plan's ignored workspace; no incomplete test lock
is committed. Portable evidence is this record, not a dependency on private cache.

Consumed: R1=1/4 restores; T=0/14; G=0/3; B=0/2; native launches0.
R1 cannot silently be repeated as L1 (locked validation), R2 (app restore), or L2.
Next proposed action: approve one additional bounded R1 attempt with explicit
first-use opt-outs and authorized network access. If approved, inspect resulting
locks before L1 and only then write the first behavioral tests. No approval is
inferred from remaining quotas. No tests discovered/executed; no behavioral RED.

## Executor rulings and remaining gates

- PowerShell/manual ledger replaces Bash-only skill helpers to honor the shell
  constraint; cost is manual evidence bookkeeping rather than automated helpers.
- Baseline is clean Git plus absence of a Windows project, not old Go runtime
  tests; explicit allowance excludes those; no Go regression claim is made.
- First-use recipe corrected from official evidence; no rerun, tool install or
  certificate cleanup is implied. Raw failure remains failure.

No final implementation review was dispatched because implementation has not
begun. Task11/12 and44/34/6/217,NOT_ACCEPTED unchanged. No main merge.
