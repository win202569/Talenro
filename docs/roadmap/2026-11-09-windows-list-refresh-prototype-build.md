# Independent list-refresh prototype: build-only checkpoint

User approved creating/compiling the isolated diagnostic spike proposed after
fe40ab4c. This is experimental, not production code or a defect fix. No runtime
launches, settings/service operations, or baseline/demo edits are in this scope.

## Implemented boundary

New clients/windows/diagnostics/TextBoxListRefreshProbe uses the existing pinned
WinUI dependency and self-contained x64 build configuration without project
references. Both modes share the same XAML, default resources, TextChanged handler,
three initial synthetic rows, nonselectable ListView and ordinary window lifetime.
Default/--mode=static does not rebuild; --mode=rebuild clears/recreates the same
three TextBlocks synchronously after input changes. Initialization is guarded.
Invalid/multiple arguments throw before window construction. Mode must be recorded
by a future observation harness; no UI toggle or mode-specific visual was added.

No filtering, selection updates, dynamic header/placeholder, demo resources,
navigation, layout observer, clipboard calls or close interception. Consequently
this isolates a reduced rebuild operation, not all of NodesPage.Refresh. Build
does not establish runtime equivalence, root cause, or whether the stall reproduces.

## Verification

- Offline cache restore succeeded (353 ms); locked-mode restore succeeded (190 ms).
- Debug/x64 --no-restore build exited 0 in 15.07 s: 0 warnings, 0 errors.
- NuGetAudit=false was used only for offline restore; no vulnerability audit claim.
- Original demo and TextBoxShutdownProbe tracked files unchanged; existing EXE/DLL
  hashes match earlier recorded values (below).
- No runtime tests/full suite or launches performed; approved throwaway build-only
  scope. Prior baseline 2/2 budget remains consumed; new runtime budget is zero.
- Necessary new source, dependency lock and recovery notes only; bin/obj/caches
  excluded. All prior NOT_ACCEPTED conditions and manual-route timing deviation
  remain, with no acceptance promotion.

SHA256, local Debug artifacts (not portable committed binaries):

| Artifact | SHA256 |
| --- | --- |
| New variant EXE | 32E1E7392B46D89DC297DA03C097A1661894B61A919DE697BB087738A76C16B5 |
| New variant DLL | 9A390F3F19ED94F4E33A0E1905A29A705655A1E324E85F9F933D601525E86BD1 |
| Baseline EXE | D19AFA9B50F68A8340CF25F3F8FB967717C68BA426061BEB1A7286C864793612 |
| Baseline DLL | D31308E544BC72C3671195C78A6A7D072B776E27768C9D1CF7DD8304D5C5D6A7 |
| Original demo EXE | 4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 |
| Original demo DLL | 7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB |

Next obtain separate bounded runtime scope for the two modes, including input
method, identity/argument checks, observation deadlines and normal-close handling.
Do not infer permission to kill processes or change Windows settings. Recover
source and commands from this branch and the new README on another computer.
