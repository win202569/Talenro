# C5 Unknown to Active restoration — 2026-11-09

Result: scoped restoration PASS. User approved one additional launch, zero retries,
max360s, no rebuild/settings changes, no repeated Stale scenario. Allowance1/1 used.
Baseline19b8d856; unchanged fixed observer and preserved artifact. No code changes.

## Preflight and route

Fixed preflight passed: ordinary token, expected clean development branch, source
blobs, preserved EXE/DLL hashes, no demo/prototype processes. Supported UI inventory
also had no targets. Only launch: `pwsh -NoProfile -File clients/windows/tests/observe-c5.ps1 -RunApproved`.
Arguments remained `--layout-test --layout-language=zh-Hans --layout-profile=unknown-usage`.

First window observation19:45:17.145Z, within30s of process creation. Initial input
reported coordinate geometry unavailable; fresh window/screenshot recovery succeeded
before one input retry, without a second application launch. Immediate post-action
trees sometimes lagged; refreshed observations determined results.

| Step | Result | Observed evidence |
| --- | --- | --- |
| Enter offline demo | PASS | Permanent offline warning, existing Connection page |
| Select Unknown | PASS | Scrolled existing page to380; menu Down navigation then Return; no text/clipboard input |
| Account in Unknown | PASS | 额度未知（模拟）; 不可用 / Unavailable — no numeric usage is known; devices 不可用; title active |
| Return Connection | PASS | Title inactive, existing selector used |
| Select Active | PASS | Menu Home focused 有效（模拟）, Return applied it; prior notice cleared |
| Account restored | PASS | 有效（模拟）; `12.4 / 100 GB（虚构示例）`; `2 / 5（虚构已激活设备）`; title inactive |
| Normal titlebar close | PASS | Original native/managed exit0 and valid matching exit evidence |

Root1136x620 DIP/page916x534/render1/text1, unchanged window. Restored usage/device
values directly visible in screenshots and accessibility text. No stale unavailable
sample remained. No connection, account submission, Stale selection, resize, system
setting change, build or test-suite run. Screenshots returned in chat, not exported
or committed; portable report is a transcription, not recoverable screenshot files.

## Process evidence

- RunId `8e1af832106f4ebfb4a5fdcb8b435b3b`, PID8396.
- Start `2026-11-09T19:44:59.9750203Z`.
- Close request `2026-11-09T19:48:38.244Z` (~218.3s after creation, before300s).
- Exit `2026-11-09T19:48:38.4845343Z` (~240.5ms after request).
-113 records: start, initial native identity,110 complete pairs, provisional outcome.
-110 pairs carry verified session identity and native API status ok; managed PID8396.
  Observer enforced creation identity and managed agreement; final native signaled/code0.
- Raw outcome sequence113: exited, ExitVerified=true, ElapsedMs220118.4511,
  Finalized=false. Post-cleanup terminal result: Finalized=true,
  ElapsedMs220130.8161, OverrunMs0, agreement; observer host exit0.
- Post-run target process and window counts both0. No forced exit or relaunch.

Artifact directory:
`clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64`.
EXE SHA256 `4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95`;
DLL SHA256 `7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB`;
both unchanged after observation.
Raw file below artifact directory:
`process-observations/8e1af832106f4ebfb4a5fdcb8b435b3b/samples.jsonl`;
SHA256 `6F770DCD5D1550876FCB10DC0B20562712CF3500BA0A31BAD989CBC5D6C3B313`.
Raw file remains ignored/unaltered, not included in Git.

## Acceptance boundary and next step

This closes the separately approved Unknown→Active recovery check. Together with
the prior partial report, evidence now covers initial Active, Unknown and Stale
presentation, off-page inactivity, and Unknown→Active restoration across two runs.
Do not rewrite these as a single continuous Active→Unknown→Stale→Active run:
Stale→Active in the original continuous route was not observed. No blanket C5,
viewport, RTL/C4, C6, Task11/12 or44/34/6/217 acceptance promotion, main merge,
shutdown root-cause claim or reliability claim follows. No launch allowance remains.
Next step is read-only acceptance-evidence reconciliation, not another automatic run.
Recover from latest development branch; binaries/raw observations are not restored
by cloning. Preserve the previous report and its original partial verdict.
