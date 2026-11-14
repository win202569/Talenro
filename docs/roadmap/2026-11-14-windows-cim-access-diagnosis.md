# CIM access environment diagnosis — read-only checks

Date2026-11-14 (client Asia/Tbilisi). User authorized read-only diagnosis only,
no security/settings changes and no App launch. Baseline dce3a161.

## Evidence

Default restricted execution, PowerShell7.6.9 / CimCmdlets7.0.0.0:
- Windows principal IsAdministrator=false; impersonation None.
- Win32_Process self query fails: CimException, NativeErrorCode=AccessDenied,
  HResult=-2146233088, message `无法从客户端中访问 CIM 资源。`.
- Get-Service cannot resolve Winmgmt/RpcSs/DcomLaunch in that context.
- Read-only sc.exe query Winmgmt fails OpenService5 Access is denied. Thus the
  Get-Service messages do NOT establish absent services.

Permission-reviewed host execution of read-only checks, same PowerShell7.6.9:
- IsAdministrator=false (requesting execution permission did not create an
  administrator user token).
- Winmgmt, RpcSs, DcomLaunch all Running (numeric4), Automatic (numeric2).
- Win32_Process self query succeeds,1 row.
- Exact existing observer inventory filter succeeds,0 rows:
  Name LIKE 'Talenro.Demo.%' OR Name LIKE '%Prototype%'.

## Conclusion and limits

The contrast isolates the observed denial to the execution context's access
boundary, not evidence of missing/stopped WMI/RPC services or administrator
rights being necessary. It does not identify the exact ACL/token restriction.
No services restarted, settings changed, installs, build, restore, test suite,
UI automation or App launch. All checks were read-only; only this sanitized
record/status/recovery documentation is saved under standing checkpoint policy.

Do not alter UAC, WMI permissions or service configuration. A future separately
approved observation can request the same permission-reviewed host execution,
provided its own unchanged ordinary-token, source/hash/path and inventory gates
all pass. Do not substitute a separate query for that entry's preflight. This
diagnosis grants no new launch/retry budget and does not reuse yesterday's run
authorization. Current0-process evidence is point-in-time, not permanent.
