# Email geometry isolated native build — blocked before compilation

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
