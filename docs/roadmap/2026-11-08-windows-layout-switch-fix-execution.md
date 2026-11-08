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
