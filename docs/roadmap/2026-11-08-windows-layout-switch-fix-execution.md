# CountryLock language fix — scoped native recheck passed

## Latest — approved native recheck, 2026-11-08

Executed docs/superpowers/plans/2026-11-08-windows-switch-native-recheck.md after
user approval. Source identity:1ac1c19e; starting documentation HEAD:baf22d43.
Task1 complete: ordinary and diagnostic cached Debug builds, --no-restore,
EnableLayoutDiagnostics=false/true respectively, each exit0 with0 warnings/errors
(8.53s /8.21s). Ordinary-token guards passed; both assets files existed; initially
no demo process/window. Existing PowerShell7 verify-layout-artifacts.ps1 PASS:
ordinary has no diagnostic payload, diagnostic has9 required types/3 resources.
Metadata inspection did not load/execute the managed App assembly.

Fresh SHA256 identities (under clients/windows/Talenro.Demo.App):

| Variant | Directory | Talenro.Demo.App.dll | Talenro.Demo.App.exe |
| --- | --- | --- | --- |
| O | bin/Debug/net10.0-windows10.0.19041.0/win-x64 | 845DABA84381CA3A952175543D9D7EC18AD2997522D332EA1F25F67CE5D86612 | 4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 |
| D | bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64 | 7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB | 4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 |

Task2 complete. Each row launched the exact resolved executable above directly,
with ordinary-token guard, one at a time; process handles supplied actual exit
codes. All startup observations were within30s, interactions below10minutes,
normal Alt+F4 closes below30s. Start times below retain the host's +04:00 offset.

| Row | Variant / arguments | PID / start | Observed result | Exit |
| --- | --- | --- | --- | --- |
| S1 | O / none | 6868 /20:13:38.1497164+04:00 | PASS: Chinese Login/offline notice; no diagnostic marker; disconnected, 关→开 (JP)→关, 国家锁定已关闭 | 0 |
| S2 | D / --layout-test --layout-language=en-US --layout-profile=baseline | 6852 /20:15:14.2860515+04:00 | PASS: English Login/offline notice and en-US/baseline marker; disconnected, Off→On (JP)→Off, Country lock off; refresh/reset checks below | 0 |
| S3 | D / --layout-test --layout-language=zh-Hans --layout-profile=baseline | 7100 /20:18:14.5081566+04:00 | PASS: Chinese Login/offline notice and zh-Hans/baseline marker; disconnected, 关→开 (JP)→关, 国家锁定已关闭; refresh/reset checks below | 0 |
| S4 | D / none | 8180 /20:21:20.4179570+04:00 | PASS: Chinese Login/default disconnected/off state after explicit-language sessions; no marker; 关→开 (JP)→关, 国家锁定已关闭 | 0 |

All rows selected jp-1 on Nodes and applied while disconnected before enabling
CountryLock. On/Off text and JP/lock-off status were inspected visually and in
the accessibility tree; scrolled to reveal status below the initial viewport.
S2/S3 additionally enabled lock again, visited Account then Connection and
observed On/开 plus JP retained in the chosen language. Disabled lock, visited
Account, scrolled to Exit demo/reset, returned to Login and re-entered: still
chosen language, disconnected and Off/关 with localized lock-off status. Tab
after the switch showed a visible focus outline on the next scenario button
in both languages; it was not activated. Offline notice persisted throughout.
No credentials, real account/network action, OS settings or window resizing.

Observation note: S2's first indexed click on the switch left it Off; a refreshed
screenshot showed the pointer beside the state label. Clicking the observed
slider toggled On immediately. Subsequent slider clicks used fresh screenshots.
This records an input-targeting no-op, not a failed launch or a language failure.
Accessibility focus reporting stayed at the pane while the screenshot showed
the Tab outline; no screen-reader/focus-metadata correctness claim is made.

Budget closed:2/2 builds,4/4 launches; zero restore/test-suite/retry invocations.
Post-close Get-Process found no Talenro.Demo.App process. Historical226/226 tests
were not repeated. No source changes in this verification-only checkpoint.

Ruling: use PowerShell and portable execution records instead of Bash task
helpers, and no new implementation/TDD/reviewer cycle — approved verification-only
scope and Windows requirement — this supplies scoped observations, not a fresh
independent whole-branch review or renewed automated-suite assurance.

Only the V3 CountryLock language finding is CLOSED by S2/S3; S1/S4 supply scoped
ordinary-path regression evidence. Broader V3 English dialog route, V4–V7 profiles,
remaining V8 checks, exact sizes/resize, text enlargement, screen-reader speech,
contrast and fallback remain unverified/not run. Task11/12 and44/34/6/217 remain
NOT_ACCEPTED; no main merge. Next propose a bounded continuation of the remaining
Stage B routes, using these fresh identities (or re-identifying rebuilt outputs).
No runtime allowance remains after this handoff.

The sections below are historical checkpoints, superseded by this latest result.

Planning follow-up after1ac1c19e: prepared
docs/superpowers/plans/2026-11-08-windows-switch-native-recheck.md (proposed2 builds,
4 launches). Await approval; no test/build/launch in the planning turn. Existing
source/test success below remains distinct from pending native defect closure.

## Latest — source correction verified, native recheck pending

After5d6039d6 user approved minimal implementation plus one targeted GREEN and
one full suite. Added CountryLock.OnContent/OffContent resources (en-US On/Off,
zh-Hans 开/关), both required-key entries and explicit typed ToggleSwitch setters.
Seven production lines in four files; no Core/state/handler changes. Shared resw
also supplies these explicit values to ordinary x:Uid localization; diagnostic
bindings apply the selected language without changing OS language or switch state.

Used the already observed valid RED (4 missing-resource failures/31 passes) without
rerunning it. Fresh approved-context verification, ordinary token False each time,
Debug/--no-restore, unchanged default timeout:
- targeted LayoutPresentationTests|LayoutViewContractTests:35/35 passed,0 skipped,
  exit0, no compiler warnings; local receipt layout-switch-green.trx.
- full Windows demo suite:226/226 passed,0 skipped,exit0,no compiler warnings;
  local receipt layout-switch-full.trx.
Both receipts are under the ignored TestResults directory. Authorized1/1 targeted
and1/1 full calls consumed. No App build/restore/native launch/settings changes.

Coverage: real resource Copy and missing-key validation plus existing declaration
binding coverage. These tests do not compile/run native WinUI setters. Review of
the seven-line diff found no extra changes; no fresh independent review claimed.
Production source correction is saved, but V3 native defect closure remains
PENDING until new ordinary/diagnostic builds and observed English/Chinese On/Off
transitions. Existing App output hashes describe pre-fix artifacts and MUST NOT
be reused to claim this fix was tested natively. Stage B V4–V8 remain NOT_RUN.

Next prepare a bounded build/artifact/native recheck scope; no launch permission
is implied by this checkpoint. Wider NOT_ACCEPTED and no-main-merge unchanged.
Historical investigation and failure records below are retained unchanged.

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
