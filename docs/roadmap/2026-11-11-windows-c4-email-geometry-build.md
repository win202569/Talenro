# Email geometry isolated native build and artifact inspection

## Approved cache-only restore and successful build

After the initial missing-assets stop below, user approved cache-only locked restore,
fail on missing cache, then build/check without app launch. Source remains822fa1b8;
documentation baseline wasdeebb7b8. Administrator=False.

Existing layout assets listed16 package entries; all16 had metadata in the worktree's
clients/windows/.nuget/packages cache. Default user-global cache did not contain the
required package; the verified project cache was used instead.

Temporary ignored config:clients/windows/.nuget/email-geometry-offline.config. It clears
packageSources/auditSources and supplies only the absolute local project cache path;
globalPackagesFolder is that same cache. No remote source configured, no audit disabled
or warning suppressed. Restore executed with --locked-mode and explicit --configfile,
Debug/layout/email-geometry=true, close-stage=false. Restore exit0; subsequent same
--no-restore build exit0 in28.10s,0 errors,1 NU1905 warning. The local audit source has
no vulnerability data: vulnerability audit is NOT established by this offline build.
No app launch, network package source, version update, install or system change.

## Verified new artifacts

Directory:clients/windows/Talenro.Demo.App/bin/email-geometry/Debug/net10.0-windows10.0.19041.0/win-x64/

| File | Bytes | SHA256 |
| --- | ---: | --- |
| Talenro.Demo.App.exe | 312320 | CBDDE55CBFE2928E3F179250F9B0DFC560DB45D61FD05B8B014BADB0B0CFCDB3 |
| Talenro.Demo.App.dll | 329728 | 154E7E70F9FB2D2F980D16A920CA7A3CB69C5CA001CECF7D6A7DFFD1D1F043D9 |

PE metadata was read as bytes using PEReader, not loaded/executed as application code:
machine AMD64; EmailGeometryObserver/targets/evaluator/coordinator/receipt types and
LayoutOptions present; no CloseStage type names. Runtime config includes.NET10.0.14.
deps.json42809bytes and runtimeconfig.json359bytes present. This proves native adapter
compilation and inspected artifact composition, not runtime behavior or diagnosis.

Generated assets confirm outputPath=obj/email-geometry, packagesPath=project cache,
only the temporary local config/source, restoreLockedMode=true. Build file list:
586 entries, all inside App obj/email-geometry or bin/email-geometry. Core retains
its existing separate paths. Ordinary/close-stage modes were not rebuilt in this run.

Old layout EXE/DLL hashes below remain identical. Lock files also unchanged:
- App:53E1D67615BDF1ACE24D0B62EB1D08CA70BEB99E9ADB4CF45437FF9AF68F3766
- Core:A4364D81A6DE2707D66FBCED18FBBC926C882AB96E1D1AEBBB17E871285B9D75

Binaries/cache/generated assets and machine-local temporary config stay ignored.
For a new computer, obtain approved matching dependencies first; cloning does not
restore cache or binaries. To reproduce this offline path, create a local config
with both source sections cleared and pointing only to that machine's existing
project cache, keep globalPackagesFolder identical, and use --locked-mode with the
same diagnostic properties. Never copy old assets into a different output namespace.
Rebuilt hashes must be reverified; these recorded hashes are not assumed portable.

Next: prepare a separate fixed-hash observer entry using the shared observation
modules and these verified artifacts. Do not change old C4/C5 hashes or launch by
ad-hoc command. Live observation still NOT_RUN, allowance0; C4 remains NOT_ACCEPTED.

## Initial no-restore attempt — historical blocker

User approved independent native build and artifact inspection, explicitly no app
launch. Source baseline822fa1b8. Ordinary-host administrator check=False.

MSBuild property evaluation confirmed App paths:
- BaseOutputPath:bin/email-geometry/
- BaseIntermediateOutputPath:obj/email-geometry/
- MSBuildProjectExtensionsPath:App project/obj/email-geometry/
- TargetPath:bin/email-geometry/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.dll

Attempted once from the coordinator worktree:
```powershell
dotnet build clients/windows/Talenro.Demo.App/Talenro.Demo.App.csproj --no-restore -c Debug -p:EnableLayoutDiagnostics=true -p:EnableEmailGeometryRecording=true -p:EnableCloseStageRecording=false
```

Result:exit1,0 warnings/1 error,NETSDK1004: new isolated
obj/email-geometry/project.assets.json missing. Stopped before native compilation;
no new output files observed under bin/email-geometry. No restore, install, retry,
app launch or system-setting change performed. No source/lockfile edits.

Before/after old layout-diagnostics hashes identical:
- EXE:4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95
- DLL:7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB

Prior source-stage portable372/372 and synthetic462 results are historical, not
rerun here. Native compile/output metadata/new hashes remain unverified; C4 remains
NOT_ACCEPTED and no live allowance exists.

Next proposed scoped authority: generate the isolated dependency assets from existing
local package cache with locked package versions; fail if cache lacks packages rather
than download/install/change versions. Then repeat the approved build and artifact
checks, never launch. Do not copy old project.assets.json into new obj: it contains
path-sensitive generated metadata. Previous no-restore boundary remains until approved.
