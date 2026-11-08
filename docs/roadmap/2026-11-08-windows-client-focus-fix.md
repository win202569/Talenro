# PageHost focus correction — execution record

Baseline: e19a7e38. User approved the diagnosis proposal: one targeted RED,
one targeted GREEN, one full test suite, one app build and one ordinary-user
native session up to 10 minutes (stop new scenarios at 8 minutes; startup
30 seconds; close observation 15 seconds; no retries).

Scope: only PageHost IsTabStop=False and a static declaration regression.
Native observation remains necessary; declaration tests do not prove focus.
No package restore, security changes or broader acceptance. NOT_ACCEPTED remains.

Ruling: use this portable execution record as the ledger and PowerShell commands
instead of the Bash-only plan helpers, preserving the user's Windows constraint.
The approved static regression is a declaration contract, not a behavioral proof;
runtime verification must independently check navigation. No new architecture.

RED: targeted test ran once with non-admin token, --no-restore, exit 1;
1 failed as expected: PageHost IsTabStop expected False, actual null.
Production change adds only IsTabStop=False to PageHost.

## Verification results

- Targeted GREEN: 1/1 passed, 0 skipped, exit 0 (second and final targeted call).
- Full Demo.Tests suite: 160/160 passed, 0 skipped, exit 0 (one call).
- App build: --no-restore -c Debug, exit 0, 0 warnings/errors, 22.50 seconds.
- Each dotnet command checked the token and would abort if administrator;
  all ran using the previously working ordinary-user execution context and
  existing local CLI/package directories. No restore or settings change.
- Read-only independent code review: no critical/important/minor findings.

Commands from clients/windows: dotnet test Talenro.Demo.Tests/Talenro.Demo.Tests.csproj
--no-restore -c Debug, with filter
FullyQualifiedName~PageHostDeclaresNoTabStopWithoutDisablingPageControls for RED/GREEN;
unfiltered for the full suite. Build target: Talenro.Demo.App/Talenro.Demo.App.csproj.
Executable SHA256: 83921BD364A6FEA64587E9684D7FEA2266299ACBE79D8EE8D2BF6439F60CC2B6.

## Native session — scoped checks passed

Single launch, PID 9008, exact worktree Debug/win-x64 executable; process start
2026-11-08T17:22:05.0638602+04:00. Window appeared within the initial call.
Alt+F4 issued about 215 seconds after launch; exact PID absent at
2026-11-08T17:25:44.8862008+04:00, and refreshed window enumeration empty.
Immediate post-close enumeration still showed the window; refreshed observation
confirmed it was gone. No forced termination or second launch.
Default restricted shell initially could not see the live process; read-only
inspection in the working execution context found and tracked PID 9008.

Observed one action and refreshed screenshot at each step:
- Connection: Settings focus after closing its explanation with Escape;
  one Tab directly to Simulate connect, one Shift+Tab directly back to Settings.
- Nodes: Settings explanation Escape returns visible focus to Settings;
  one Tab directly to Search with caret/underline, one Shift+Tab back to Settings.
- Account: sidebar Account then Tab to Settings, Tab directly to MFA,
  Shift+Tab directly back. No intermediate invisible stop on these three paths.
- Account: Tab through MFA, Manage devices, Language, Delete placeholder and
  Exit demo. Lower controls automatically scroll fully into view and show focus.
  No delete action, account operation or settings change was invoked.

The observed extra-stop symptom is resolved for the three tested paths on this
build. Runtime internal focus-object identity was not instrumented. Accessibility
metadata still returned stale root-pane data, so conclusions use visible focus
outlines/caret, not that metadata. This is not full screen-reader acceptance.

Reviewer exclusions resolved: runtime traversal/modal/scrolling are assessed by
the observations above; test/build results by parent command output. Applied
style internals and runtime object identity remain unmeasured. Broader acceptance
and main merge remain outside scope. No review findings were deferred.

Budget consumed: targeted 2/2, full suite 1/1, build 1/1, launch 1/1.
Task complete for this narrow correction. Full keyboard/accessibility and
Task11/12 plus44/34/6/217 NOT_ACCEPTED unchanged. Next: review remaining checklist
coverage without launching; any additional runtime work needs its own finite scope.
