# C4 CIM access — read-only environment diagnosis

Date2026-11-10, Asia/Tbilisi. Baseline f1556c9d6eeedbaaa6c60fff15dbc8e8adad16e6.
User authorized the next proposed step: read-only environment diagnosis, no app
launch or system/settings changes. No live observation allowance was renewed.

## Evidence

Both execution contexts used bundled PowerShell7.6.9 and CimCmdlets7.0.0.0.
Ordinary token checks returned ADMIN=False. Queries used local root/cimv2,
OperationTimeoutSec5 and limited properties; no private process inventory printed.

| Read-only probe | Restricted command context | Approved ordinary host context |
| --- | --- | --- |
| Win32_Process for querying shell's own PID | CimException, NativeErrorCode2 | One row |
| Win32_OperatingSystem, Version only | Same CimException | One row |
| Service inspection | Cannot open Service Control Manager | Four named services readable |

Restricted CIM message: `无法从客户端中访问 CIM 资源。` HResult -2146233088.
Initial per-name service query claimed names unavailable; the later explicit service
inventory error demonstrates access failure, so those messages are NOT evidence that
Windows services are absent.

In the approved ordinary host context, DcomLaunch, RpcSs and Winmgmt were Running /
Automatic. WinRM was Stopped / Manual, while the local CIM probes still succeeded.
No service was started/stopped/reconfigured. No elevation of the Windows user token
occurred: tool execution outside its restricted context still reported ADMIN=False.

Finally, extracted only Get-C4Policy, Assert-C4OrdinaryToken, Invoke-C4Git and
Assert-C4Preflight from the committed entry via PowerShell AST. Executed that
read-only preflight in the approved ordinary host context: PASS. This includes
ordinary token, expected clean branch, consumed source blobs, fixed artifact hashes
and empty demo/prototype CIM inventory. No writer/backend/native observer imports,
receipt creation or target launch was invoked. No test suite or build was run.

## Diagnosis and next boundary

The reproducible difference isolates the observed blocker to the command execution
context's access restrictions, rather than evidence of a broken WMI service or
C4 product defect. The exact underlying restriction/permission implementation was
not inspected; do not claim a particular ACL or Windows security-setting cause.
No repair is indicated by these results, and no source/product changes were made.

For a future separately approved C4 run, request execution of the unchanged fixed
entry in the ordinary host context, retaining ADMIN=False and every existing gate.
Do not weaken guards, bypass approval, enable WinRM, change UAC or retry automatically.
This preflight is point-in-time evidence; the real entry must repeat it at launch.
The earlier attempt stays BLOCKED_BEFORE_LAUNCH; UI/exit checks remain NOT_RUN.
Next step requires renewed one-launch/zero-retry bounded approval. No main merge or
broader acceptance. Only report/status/recovery documents are preserved remotely.
