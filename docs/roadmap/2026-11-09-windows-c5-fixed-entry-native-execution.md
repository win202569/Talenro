# C5 fixed-entry native observation — 2026-11-09

Status: PARTIAL / NOT_ACCEPTED. One newly approved launch consumed; zero remaining.
Baseline4f50c581. Fixed observer completed with verified native/managed exit0.
The final Active restoration step was NOT_RUN to preserve the close-by300s limit.
This is not a shutdown-root-cause experiment or broader acceptance.

## Authorization and preflight

User explicitly answered yes to one launch, zero retries, at most6min, normal close,
no rebuild/settings/forced termination. Followed the existing C5 route in
`docs/superpowers/plans/2026-11-09-windows-unknown-usage-native-check.md`, using
`clients/windows/tests/observe-c5.ps1 -RunApproved` rather than the old ad-hoc observer.
The shorthand approval question called the third state “过期”; the existing C5 plan
and executed control are Stale / 信息陈旧, NOT the distinct Expired / 过期 option.
No Expired behavior is claimed.

Ordinary token confirmed (Administrator=false). Initial restricted CIM read was
inaccessible; a separately approved execution-boundary elevation allowed the read
while the OS token remained ordinary. No application was started during preflight.
Expected clean development branch, observer source blobs, EXE/DLL hashes and zero
demo/prototype processes passed. Supported UI inventory found zero target windows.
UI operations used supported computer-use APIs only, no terminal UI or raw UIA.

## Identity and evidence

- RunId: `425352a69f0a40f59fe9913cea119062`; PID20612.
- Original Process start: `2026-11-09T19:19:14.0413035Z`.
- Fixed artifact: `clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
- EXE SHA256: `4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95`.
- DLL SHA256: `7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB`.
- Arguments: `--layout-test --layout-language=zh-Hans --layout-profile=unknown-usage`.
- Raw relative to artifact directory: `process-observations/425352a69f0a40f59fe9913cea119062/samples.jsonl`.
- Raw SHA256: `A61B785FF6603402A5FFD83259D41B0BB9941F1FC7611C24C164712DF10949BB`.
- 148 JSONL records: start, first native identity,145 pairs, provisional outcome.
- All145 pairs complete/session-identity-verified, with native Wait/ExitCode/Times
  status ok and managed PID/start matching. The fixed observer passed its native/CIM
  guard; raw format does not independently persist the initial CIM row.

Raw JSONL stays locally ignored and unchanged. Screenshots/accessibility observations
were returned by the UI tool in this chat; not exported or committed. The portable
report transcribes visible synthetic values; it does not claim screenshot-file
recovery from Git. The historical empty C5 receipt remains untouched.

## Observed route

| Step | Result | Direct observation |
| --- | --- | --- |
| Startup/Login | PASS | First window observation began19:19:31.351Z, within30s; offline notice visible, unknown-usage inactive |
| Enter demo, initial Account | PASS | 有效（模拟）; `12.4 / 100 GB（虚构示例）`; `2 / 5（虚构已激活设备）`; inactive title |
| Unknown on Connection | PASS | Existing selector confirmed by keyboard; notice 演示额度未知，暂不能开始连接; title inactive |
| Unknown on Account | PASS | 额度未知（模拟）; `不可用 / Unavailable — no numeric usage is known`; devices 不可用; active title |
| Return to Connection | PASS | Unknown selection visible, title inactive |
| Stale on Connection | PASS | Selected 信息陈旧（模拟）; notice 演示额度信息已陈旧，暂不能开始连接; title inactive |
| Stale on Account | PASS | 信息陈旧（模拟）; exact same unavailable usage/device values; active title |
| Return to Connection | PASS | Title inactive before normal close |
| Restore Active, revisit Account | NOT_RUN | Stopped functional route before300s close deadline; cannot claim numeric restoration |

Window title reported root1136x620 DIP, page916x534 after demo entry, render1/text1.
No resize or settings change. No target-viewport acceptance claimed. Unknown/Stale
usage text was fully visible at this viewport. Lower Account actions were below the
visible page and not tested. Popup selection coordinates were outside window bounds
and were rejected by the tool; no input at those rejected points. Refreshed UI, then
used focused menu keyboard navigation (End/Up/Return for Unknown, Down/Return for
Stale). No text entry, clipboard access, connection request or model injection.
Initial wheel operation did not move the page; later observed scroll reached380.
UI refresh often showed the pre-action tree immediately; fresh observations, not
those stale snapshots, supplied the table verdicts. Navigation consumed time, so the
final restoration was omitted rather than extending the budget or starting again.

## Close and observer outcome

Normal titlebar close requested `2026-11-09T19:24:02.917Z` (~288.9s after process
creation; within300s launch-clock budget). Original-handle native exit time and
managed ExitTime agreed: `2026-11-09T19:24:03.1713612Z`, about254ms after request.
Final pair observed at19:24:04.682Z: native signal, successful code0, valid exit time;
managed WaitSignaled/HasExited true and ExitCode0. No forced close or retry.

Raw provisional outcome: sequence148, exited, ExitVerified=true,
ElapsedMs290684.3078, Finalized=false. After cleanup the information/success output
reported Finalized=true, ElapsedMs290697.36, OverrunMs0, reason agreement; observer
host exit0. This distinguishes persisted provisional evidence from cleanup success.
Subsequent process and window inventories both had zero targets. EXE/DLL hashes
unchanged. No builds, tests, source repair, service/system changes or second launch.

## Disposition and recovery

C5 is PARTIAL / NOT_ACCEPTED because the restoration step remains unobserved.
This one run validates the fixed observer's successful live path on this machine;
it does not certify failure paths, reliability, all platforms, C4/RTL, C6, Task11/12
or44/34/6/217. Historical shutdown failure remains unresolved; current success is
not a fix. No automatic continuation or additional launch allowance.

Next work can consolidate the remaining acceptance gap without launching. Any
future restoration-only observation needs a separate bounded approval; do not
repeat completed stages merely to obtain another successful shutdown. Recover from
latest development branch, not main; ignored artifact/raw files are not in Git.
