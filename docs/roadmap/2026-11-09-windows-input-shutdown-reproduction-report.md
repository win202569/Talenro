# Windows input-associated shutdown stall — reproduction report

Status: evidence report, not a fix or a standalone minimal sample. Prepared from
saved observations; no new runtime attempt and no external issue submission.

Follow-up: an isolated single-TextBox prototype was subsequently built and observed.
Automatic input closed normally; manual close also signaled but exact-value
verification missed its route deadline. The prototype has not reproduced the stall.
See [prototype observation](2026-11-09-windows-textbox-prototype-observation.md).
The report body below describes the earlier demo evidence; its historical statements
that a reduced sample had not yet been built are superseded by this follow-up.

## Summary

In the existing offline WinUI demo, automated literal-text entry of `jp-1` followed
by normal window close repeatedly left the original process handle unsignaled.
Task Manager showed a wait on a service host later mapped to `cbdhsvc_8b240`
(local template group `ClipboardSvcGroup`). An adjacent physical-keyboard input
run exited cleanly. This establishes an observed input-associated difference,
not the responsible component or a causal mechanism.

## Environment and artifact identity

- Windows x64, ordinary user token. Exact Windows build and active IME version
  were not captured in the adjacent-run receipts; do not infer them from SDK targets.
- Observer: PowerShell 7.6.9, .NET 10.0.14. This is the observer runtime, not a
  measured version of every self-contained application component.
- Source project: `clients/windows/Talenro.Demo.App/Talenro.Demo.App.csproj`;
  net10.0-windows10.0.19041.0, unpackaged/self-contained WinUI x64;
  declared Microsoft.WindowsAppSDK package `[2.6.5]` (source declaration only).
- Adjacent-run baseline: `cf0f605dddd738ef336ba099f0b126b8751b1d48`.
- Artifact path relative to repository:
  `clients/windows/Talenro.Demo.App/bin/layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.exe`.
- EXE SHA256: `4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95`.
- DLL SHA256: `7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB`.
- Arguments: `--layout-test --layout-language=zh-Hans --layout-profile=long-country`.
- Input interface used in failing run: bundled `@oai/sky` `type_text`, computer-use
  package directory version `26.1103.32239`. Its documented contract does not
  establish whether clipboard transport is used internally.

The ignored binaries are not included in Git. A fresh clone preserves source and
evidence, not these exact artifacts; any rebuild produces a new artifact identity
and must not be silently treated as the observed binary.

## Observed reproduction recipe (not permission to rerun)

1. Confirm ordinary token, expected artifact/recorder identities and no existing
   demo process/window. Use a fresh observer host with no preloaded native helper.
2. Launch once using the artifact and arguments above; keep the original Process
   and SafeHandle. No output redirection or PID-based reopening.
3. Enter demo -> Nodes -> click SearchBox, confirm caret/focus. Do not select a
   node, apply a change, connect, or open a dialog.
4. Automatic condition: one `sky.type_text({window, text:'jp-1'})`. Manual
   condition: user physically types the same text, without paste or automated
   text/value input. Confirm exact value and filtered JP group/jp-1 node.
5. Send one normal Alt+F4. Record close UTC; do not send a second close or kill.
6. Read native then managed observations approximately every5s from the retained
   handle; verify creation identity each time. Outer waits are at most1000ms.
   Startup30s, route120s, close criterion30s, absolute360s from launch. Stop on
   identity/native read failure; preserve disagreements instead of replacing them.
7. After observation, dispose caller resources and independently check for remaining
   process/window state. Do not start another condition with a leftover demo.

Observer modules: `clients/windows/tests/process-observation.psm1`,
`native-process-observation.psm1`, `native-process-observation.cs`. Protocol details:
[paired observation plan](../superpowers/plans/2026-11-08-windows-paired-native-observation.md).
The historical launch command was session-local; this report does not provide a
newly tested turnkey runner. No runtime allowance remains.

## Expected and observed

Expected: normal close yields an actual process signal and successful exit code0
within30s. A vanished window, managed HasExited alone or observer exit0 is insufficient.

| Adjacent condition, 2026-11-09 UTC | Automatic | Manual |
| --- | --- | --- |
| Demo PID | 8516 | 23160 |
| Launch | 06:06:29.1238800 | 06:12:57.9423263 |
| Close request | 06:07:47.746 | 06:14:40.806 |
| Search/filter | PASS | PASS (accessibility/user confirmation) |
| Valid paired receipts | 73 | 22 |
| Final native wait / code | 258 / successful0 | 0 signaled / successful0 |
| Final native exit FILETIME | high0/low0 | high31289361/low2602630162 |
| Final managed wait / HasExited | false / true | true / true |
| Final managed exit time | 1601-01-01 | 2026-11-09T06:14:41.0368018Z |
| Outcome | Timeout360.012s from launch | Signal about0.250s after close |

Automatic post-close57 receipts preserved the same disagreement. User screenshot:
`Talenro.Demo.App.exe PID8516/thread17368 -> svchost.exe PID9888/thread1812`.
Later service enumeration returned `cbdhsvc_8b240`, Running; host creation before/
after enumeration was `2026-11-08T17:11:36.0749030Z`. Earlier PID1772/thread10816
also waited on the same target identifiers. Service enumeration is later evidence,
not an exact capture-time mapping, and thread creation identity was not established.

After each adjacent observer ended, independent queries found no remaining demo
or observer process/window. No forced termination or service operation occurred.
This disappearance does not prove observer disposal caused termination.

## Limits and useful next investigation

- Fixed run order, unequal timing, foreground changes, IME behavior and state
  carried between runs remain confounders. Both manual final captures showed chat
  pixels while app accessibility returned the correct search result; do not claim
  screenshot-based manual-route verification.
- A previous manual run also signaled in about0.248s. This is not a randomized
  sample or reliability rate. It does not clear historical intermittent failure.
- No stack, specific blocking API, tool-internal transport or actual clipboard
  content was captured. Do not label this a proven Windows, WinUI, service or
  automation-tool defect, deadlock, or completed fix.
- Source review found synchronous fixture filtering and no direct clipboard call
  on the inspected search path; framework/native interactions remain possible.
- Before more runs, seek input-transport clarification from available documented
  interfaces or prepare a smaller standalone TextBox reproduction as a separately
  scoped change. Such a sample has not been built or shown to reproduce this issue.
- Any debugger/dump, service modification, new launches or external submission
  needs its own scope. Do not disable clipboard services or force exit as a fix.

All Task11/12 and44/34/6/217 NOT_ACCEPTED boundaries remain. No secrets, raw process
inventory, screenshots or memory dumps are included in this report.

## Source records

### Local input-interface inspection follow-up

On2026-11-09, searched readable md/js/ts/json/cs/py files under the installed
computer-use26.1103.32239 package for clipboard/type_text/paste. Matches provided
only the documented literal-text contract, not transport implementation. Package
root exposed docs/skills/assets/metadata. In the existing node runtime,
import.meta.resolve('@oai/sky') returned the bare specifier, not a source path.
These checks did not establish clipboard use or prove implementation source is
unavailable everywhere. No helper executable search, reverse engineering, custom
transport, UI input, target launch or clipboard-content access was performed.

This local documentation route is exhausted within the inspected scope. Useful
next work is a separately approved isolated single-TextBox WinUI sample, preserving
the existing application unchanged; build and runtime budgets must be scoped before
execution. Its purpose would be to test whether the application-specific search
and layout code is necessary for the symptom, not to prove a vendor defect.
Alternatively, a support inquiry can ask whether type_text uses clipboard/paste,
how it restores clipboard state, and what non-sensitive diagnostics are supported.
No inquiry has been sent and no reduced sample has been implemented.

- [Adjacent comparison](2026-11-09-windows-adjacent-input-comparison.md).
- [Earlier manual input](2026-11-09-windows-manual-input-observation-execution.md).
- [Wait-chain screenshot and service follow-up](2026-11-08-windows-wait-chain-observation-execution.md).
