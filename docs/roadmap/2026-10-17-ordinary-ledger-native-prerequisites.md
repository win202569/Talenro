# Native probe prerequisites — blocked, no runtime entry

Read-only inspection on 2026-10-17, beginning 15:36:15 +04:00, against
`a02baace2a8c8a9e391bdf95079ee15c6e1d320b` on
`codex/c12-b01-task9-coordinator`. This report is not runtime approval,
security clearance, a test run or acceptance.

## Decisive prerequisite failure

The fixed executable `C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe`
has two hardlink names, as reported by the read-only `fsutil hardlink list`:

```text
\Windows\System32\WindowsPowerShell\v1.0\powershell.exe
\Windows\WinSxS\amd64_microsoft-windows-powershell-exe_31bf3856ad364e35_10.0.26100.9549_none_04d05158b23eb255\powershell.exe
```

The helper's `Get-OrdinaryLedgerProbeFile` requires link count exactly one
(line 61); the planner applies it to this exact executable (line 83).
Therefore the observed system file does not satisfy the existing planner
precondition. This is a static source/OS-metadata conclusion, not an observed
helper exception: the helper was deliberately not imported or invoked.
The Stage A fake link boundary did not establish compatibility with this file.

Do not remove a system hardlink, copy/replace the executable, substitute another
shell or relax the guard to obtain a pass. A separately reviewed design decision
is needed before any implementation or runtime allowance. No new test budget
is implied. Keep executable identity/binding and process-safety requirements.

## Other observed prerequisites

| Check | Observation and limit |
| --- | --- |
| Current diagnostic host | PowerShell Core 7.6.5; 64-bit process and OS; Windows 10.0.26200.0. Not the required fresh Desktop 5.1 control host. No Desktop host was launched. |
| Fixed executable | 495616 bytes; file version 10.0.26100.8875 (WinBuild.160101.0800); PE signature 00004550, machine 8664 (x64). |
| Executable SHA256 | `60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F` |
| Signature | Authenticode Status Valid, signer Microsoft Windows / Microsoft Corporation / US. This is a point-in-time check, not a held launch binding. |
| Inbox module | Microsoft.PowerShell.Utility.psd1 exists under the system WindowsPowerShell module directory. Import/runtime resolution not checked. |
| Sources | All 11 names returned by the source-name function were read and SHA256 calculated; each leaf reported no ReparsePoint attribute. Ancestor/link-count completeness and runtime load binding were not established by this check. |
| Helper SHA256 | `28D74C60D6A2581F9E958AD986F7D086B99D63DC2F5E888A97FFD62791963080` |
| Future identities | All four fixed `ledger-native-root-001`, `ledger-native-root-001-control`, `ledger-native-child-001`, `ledger-native-child-001-control` paths absent in the prescribed diagnostic scope at inspection. They were not created or consumed. |

These are individual observations, not an atomic snapshot or approval plan.
No nonce, encoded command, fresh runtime host or outer watchdog was prepared.

## Security status — existing gates remain

Sandboxed CIM/event access was unavailable; one elevated **read-only** query
succeeded. AMServiceEnabled, AntivirusEnabled, RealTimeProtectionEnabled,
BehaviorMonitorEnabled and IsTamperProtected were all True. WinDefend and
SecurityHealthService were Running. SignatureLastUpdated displayed
2026-10-17 06:15:35 in the querying host's local representation.

A query for event IDs 1116/1117/5001/5007 over the preceding seven days, capped
at the newest 30 records, returned 30 informational 5007 records, from
2026-10-14 03:41:49 through 2026-10-17 13:51:24 (host-local display).
Only time/ID/level metadata was inspected; event payloads were not analyzed.
The cap can hide older matching events. This does **not** establish absence of
detections, resolve prior Defender evidence or clear I2/I3. No security settings
were changed, no exclusions added and no quarantined files restored.

## Execution accounting and handoff

No helper import, native entry, child workload, test suite, Go/Docker execution,
install, metadata continuation, main merge or acceptance activation occurred.
One combined diagnostic command had a PowerShell parse error before execution;
its corrected read-only query then ran. It was not a product/test invocation.
Stage A remains 36/36 consumed; previous records and budgets remain unchanged.

Next: decide how to reconcile the fixed inbox executable with the independent-file
policy without weakening safety, and separately resolve security gates. Stage B
is both unapproved and blocked on the currently observed prerequisites.
Recover from the development branch and recheck machine-specific facts on a new
computer; this report must not be reused as that machine's clearance.
