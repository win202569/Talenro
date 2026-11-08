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

## Approved single diagnostic attempt aftere877100a

One diagnostic-enabled replacement RED authorized and consumed1/1. Same ordinary
token, default execution context, Debug, --no-restore and two-class filter as above;
no debug/timeout override present. Added --diag at
Talenro.Demo.Tests/TestResults/layout-switch-diag-red.log and new receipt
layout-switch-diag-red.trx. Original receipt retained. No retry or production edit.

Result: exit1, aborted after90second host-connection timeout; TRX total0/executed0.
Host log layout-switch-diag-red.host.26-11-08_19-57-24_58464_5.log supplies the new
evidence (host timestamps+04:00): runner20432 listened on127.0.0.1:51504 at19:57:24.580;
testhost8172 started and attempted that endpoint at19:57:24.707; at19:57:24.728 its
connection callback reported SocketException10013, socket access forbidden by
access permissions, through TcpClient.CompleteConnectAsync. Earlier generic
'connected' trace lines are superseded by the failed connection task/exception;
they do not establish a completed protocol handshake. Runner timed out19:58:54.605.
Final current-context testhost count0; no process killed.

Diagnosis narrowed: immediate access-denied local socket connection, not slow
assertions or a missing testhost executable/runtime. Exact policy owner remains
unproven: execution restrictions versus host security/network policy. Runner also
logged optional assembly-discovery warnings (Fakes/AsyncInterfaces); those are not
established as causal given the direct host connection exception. Do not install
packages, disable protection, modify firewall/UAC or increase timeout from this.

Next recommended controlled comparison, requires new approval: ONE identical
diagnostic targeted invocation through the tool's approved broader execution
context, still enforce ordinary Windows token and stop if administrator. This
changes the execution restriction boundary, not OS security settings or timeout.
Use new receipt/log names; no automatic retry. A success would support a context-
specific restriction, not prove which policy caused it. A repeated10013 would
require further host-side diagnosis. This turn does not perform that comparison.
Application defect remains unfixed; no valid assertion RED yet, no App build/launch.

## Approved context comparison after3b33c9ad — valid RED obtained

One approved broader-execution invocation consumed1/1. AdministratorToken=False;
same Debug/--no-restore/filter, no timeout/debug override, new --diag
layout-switch-context-red.log and TRX layout-switch-context-red.trx under the same
ignored TestResults directory. No source or system-setting change between attempts.

Command completed in about4.5seconds, exit1:35 executed,31 passed,4 failed,0 skipped.
All four failures are the new SwitchStateTextUsesSelectedLanguageAndCannotBeMissing
cases (en-US/zh-Hans × OnContent/OffContent), throwing InvalidDataException:
'Missing layout resource.' from LayoutPresentation.Copy. This is the intended
assertion-level RED, unlike the two earlier0-executed connection failures. Final
current-context testhost count0. Original receipts/logs preserved, not committed.

The controlled comparison supports an execution-context-specific connection
restriction; it does not identify the exact enforcing policy or prove every future
run will succeed. No reason established to alter firewall/UAC or timeout. Optional
discovery warnings in previous logs are not evidence requiring package installation.

Next recommended scope: implement the minimal two-language state resources,
required-key entries and typed OnContent/OffContent bindings, then ONE targeted
GREEN (same35 cases) and ONE full suite using approved broader execution with the
same nonadministrator guard. No replacement RED needed on unchanged source/tests.
No App builds/native launches included; those require later artifact verification.
Await confirmation of this implementation/verification scope. This comparison turn
only obtained evidence and saved documentation; the production fix is still pending.
