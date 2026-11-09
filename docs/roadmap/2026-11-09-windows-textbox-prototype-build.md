# Single-TextBox diagnostic prototype — build-only checkpoint

User approved creating and compiling an isolated WinUI single-TextBox sample after
9330db92, explicitly without running it or modifying the original application or
system settings. This is an experimental diagnostic spike, not a product feature,
fix, accepted UI or demonstrated minimal reproduction. Source retention follows
the standing fresh-computer progress instruction; it is not product promotion.

## Scope

Added `clients/windows/diagnostics/TextBoxShutdownProbe/` with project, App/Window
XAML and minimal constructors/startup, README, local ignore file and NuGet lock.
Same declared WindowsAppSDK2.6.5, net10.0 Windows target, x64/self-contained/unpackaged
settings as the demo. One empty TextBox with name ProbeInput; jp-1 is placeholder,
not text. No handlers for text/close, filtering, clipboard logic, forced exit,
observer, account/network code or Talenro project reference. Not added to solution/CI.
Shared SDK pin/build properties remain unchanged. Prototype output is isolated.

## Verification performed (ordinary token, PowerShell)

1. First restore attempt stopped before restore: denied access to user .dotnet
   directory. Used the project's existing ignored .dotnet-home with per-command
   DOTNET_CLI_HOME; no system-wide environment or permissions change.
2. Local-cache restore succeeded using --source ./.nuget/packages --packages
   ./.nuget/packages -p:NuGetAudit=false --disable-parallel, no package downloads
   or installs. Dependency audit was not performed by this offline restore.
3. `dotnet build ./diagnostics/TextBoxShutdownProbe/TextBoxShutdownProbe.csproj
   --no-restore -c Debug -p:Platform=x64` succeeded in29.46s:0warnings/0errors.
4. Repeated restore with --locked-mode and same cache flags: success.
5. Enumerated untracked source: only the eight intended source/config/readme/lock
   files; bin/obj excluded. Original demo EXE/DLL hashes unchanged.

No runtime tests or app launch, no original-project suite rerun. This prototype
contains only framework startup/UI declarations; no mocked or source-text tests
are presented as runtime proof. Compilation is the only behavioral gate authorized
in this turn. Shutdown cause and all NOT_ACCEPTED gates remain unresolved.

## Artifact identity

Output under prototype `bin/x64/Debug/net10.0-windows10.0.19041.0/win-x64/`:

- EXE SHA256 D19AFA9B50F68A8340CF25F3F8FB967717C68BA426061BEB1A7286C864793612
- DLL SHA256 D31308E544BC72C3671195C78A6A7D072B776E27768C9D1CF7DD8304D5C5D6A7

Original demo retained SHA2564180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95
and7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB.
Binaries/cache are ignored, not pushed. Fresh computers rebuild using README/lock;
rebuilt artifacts need their own identity before runtime observation.

## Next boundary

Prepare a separately approved runtime comparison for this prototype with updated
executable/name/hash binding and no demo-specific arguments or Nodes navigation.
Keep one normal close, bounded waits, identity checks, no retry/kill/service changes.
No launch allowance granted by creation/compilation approval. A successful or failing
prototype run alone would not establish a tool/service defect or repair Talenro.
