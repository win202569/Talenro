# CountryLock language fix — blocked before valid RED

Base8c5fe2a6. User requested continuing the recommended diagnosis/correction.
Announced scope: one targeted RED, one targeted GREEN, one full suite; zero App
builds/restores/native launches. Prior Stage B unused launch slots are not reused.

Investigation: Stage B V3 directly observed English Connection labels with Chinese
off-state '关'. CountryLock XAML has x:Uid but no OnContent/OffContent; diagnostic
LayoutStaticBindings applies CountryLock.Header only. Neither language resource
file nor RequiredKeys includes the two state-content keys. Existing coverage only
checks declared resources, so framework-provided text escaped it. Installed WinUI
2.6.5 Microsoft.WinUI.xml documents OnContent/OffContent as the displayed state
content. Proposed minimal fix: explicit On/Off and 开/关 resources, required-key
validation and typed state-property bindings. No production change made yet.

Added four theory cases in LayoutPresentationTests:
SwitchStateTextUsesSelectedLanguageAndCannotBeMissing, en-US/zh-Hans × OnContent/
OffContent. Exercises real resource parsing, presentation.Copy and missing-key
validation; removing state content or its required validation must fail. Existing
view contract tests will check resource-to-typed-property mapping after correction.
These tests do not execute WinUI controls; eventual native recheck is still needed.

Attempt1 (ordinary token, Debug, --no-restore):
dotnet test Talenro.Demo.Tests/Talenro.Demo.Tests.csproj --no-restore -c Debug
--filter 'FullyQualifiedName~LayoutPresentationTests|FullyQualifiedName~LayoutViewContractTests'
--logger 'trx;LogFileName=layout-switch-red.trx'

Core/tests compiled, then vstest could not connect to testhost within90seconds.
Command exit1, run aborted. Ignored TRX result summary: total0/executed0, outcome
Failed, connection-timeout error; reported process13868 still running at error time.
Subsequent read-only Get-Process -Id13868 returned no process in this execution
context. No process killed. This is infrastructure failure, NOT an assertion RED;
do not implement the proposed production fix on this evidence.

Budget: targeted1/2 consumed with no valid RED; full0/1. No automatic retry,
VSTEST_CONNECTION_TIMEOUT increase, security changes or fallback launch. Remaining
calls cannot supply both a replacement RED and GREEN within the announced ceiling.
Next: separately diagnose testhost connectivity read-only, then obtain a bounded
replacement verification allowance. Preserve the original receipt; never overwrite
it with a rerun. All four new cases remain unexecuted; prior222/222 is historical.

Checkpoint is intentionally unfinished test-only progress. English native defect
remains open, no new application binary, no native acceptance, no main merge.

## Read-only environment investigation after0b14d3b6

No tests, builds, restores or application launches performed in this follow-up.
Compared original receipts: layout-unmapped-full.trx completed222/222 at19:31:17
host time; layout-switch-red.trx spans19:50:54.133 to19:52:24.352+04:00, executed0.
The failure is before test execution/host connection, not a failing new assertion.

Observed configuration/files:
- global.json pins SDK10.0.403 with rollForward disabled; installed SDK directory
  10.0.403 exists. Runtime directories10.0.14 and8.0.33 exist. Test runtimeconfig
  requests Microsoft.NETCore.App10.0.0/net10.0, not an absent major version.
- No tracked change from7e25a4fc to0b14d3b6 in test csproj, global.json,
  Directory.Build.props or test packages.lock.json.
- testhost.exe/.dll, TestPlatform communication/engine assemblies and xUnit
  adapter/execution assemblies exist in the test output. Presence is not integrity
  or successful-load proof; versions were not changed or reinstalled.
- VSTEST_CONNECTION_TIMEOUT, VSTEST_HOST_DEBUG, VSTEST_RUNNER_DEBUG, DOTNET_ROOT,
  DOTNET_ROOT_X64 and DOTNET_ROLL_FORWARD are unset in the current process/user/
  machine scopes. This is current state, not a capture of the failed child's env.
- Current-context dotnet/testhost process count0; no process terminated.
- Application event-log query for IDs1000/1001/1026 at19:48–19:53 host time was
  denied in restricted execution. Read-only approved-access query succeeded but
  found no matching events. This does not prove absence of an unlogged crash or
  of events in other logs. No security/event-log settings changed.

Conclusion: connectivity/startup failure is localized, but its root cause is NOT
established. Existing evidence cannot distinguish scheduling/resource delay,
host startup stall, transport failure or execution-context restrictions. Do not
attribute it to firewall/UAC or enlarge the timeout on speculation. The original
run had no --diag log; its TRX alone cannot identify the failed connection step.

Recommended next scope, awaiting confirmation: one replacement targeted RED call
using the same ordinary-user context, filter, Debug and --no-restore settings,
adding --diag with a unique ignored local path and a new TRX filename. Keep the
90second default unchanged. No automatic second retry, no App launch/build/restore.
Capture actual outcome and runner/host diagnostics, redact private environment
details from portable notes. If intended assertions fail, obtain/confirm subsequent
fix/GREEN/full scope; if infrastructure fails again, stop and inspect saved logs.
This proposal does not consume or repurpose the old remaining test/native slots.
