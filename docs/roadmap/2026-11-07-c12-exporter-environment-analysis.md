# Exporter environment investigation — exact registry target identified

Source0208b1b1. User requested a separate environment investigation. This turn
used source/file/environment reads and metadata-only PE inspection;no exporter,
WPA,WPR or Go launch,no registry access,no security/configuration change.
Systematic debugging used;observation remains BLOCKED,44/34/6/217,NOT_ACCEPTED.

## New findings

1. Launcher source uses absolute executable path,UseShellExecute=false,
   CreateNoWindow=true,redirected output and ArgumentList. It does not select a
   username or elevation verb. It removes GO/GIT_ variables,then applies supplied
   overrides;other environment variables are inherited. This is source evidence,
   not proof of the child's Windows token/registry permissions.
2. Current PowerShell7.6.5 is X64 on X64 Windows. Twelve specifically inspected
   variables were limited to .NET startup/root/dependency/profiling and compatibility
   controls;eleven .NET/profiling variables were absent. __COMPAT_LAYER is present
   as DetectorsAppHealth. No supported causal explanation for that value was
   established. Do not clear/change it as a speculative fix. These are current
   process observations,not a captured environment from the two failed invocations.
3. The exporter executable reports only the default data stream in this inspection;
   no Zone.Identifier stream was listed. This does not prove policy absence or
   trusted provenance. Exporter EXE/DLL hashes remain as previously recorded.
4. Searching installed DLL metadata for KitsCeipOptedIn identified
   Microsoft.Performance.Base.dll. Metadata-only PEReader inspection read its
   get_KitsCeipOptedIn and GetValue method bodies;no assembly was loaded for
   execution and no static initializer was invoked.

## Exact static target

Base DLL SHA256:
2A324D03D9AB11B4CE992AF741127ADCA76A8FE0899057000D54A3F059955C6E.
Tool6038fa read getter IL:
`28FE02000A28FF02000A2D02162A162001000080727E11007072C811007016286000002B16FE032A`.
User-string tokens at instruction offsets20/25 decode to
`Software\Microsoft\Windows Kits\CEIP` and `OptIn`.
Tool1399f9 resolved member references:Windows platform check,OpenBaseKey,
OpenSubKey(String),GetValue(String,Object),Dispose. The getter pushes registry
view0 and hive0x80000001 (CurrentUser),then invokes a generic integer GetValue
with default0 and returns whether the result is positive. The helper opens the
base hive and subkey,reads the value if available and disposes keys. Its observed
IL has disposal finally paths,not a catch converting access denial to default0.

Therefore the installed getter's Windows path targets
`HKEY_CURRENT_USER\Software\Microsoft\Windows Kits\CEIP`,value `OptIn`.
This static path matches the saved exception's getter/helper method names and
narrows the next investigation. It does not establish that the key exists,its
value,ACL,which user hive the failed process saw,or which policy denied access.
No registry value or access-control data was read in this turn.

Microsoft's [OpenSubKey documentation](https://learn.microsoft.com/en-us/dotnet/api/microsoft.win32.registrykey.opensubkey?view=netframework-4.8.1)
distinguishes insufficient access from a missing key;the string-only overload is
read-only. That supports investigating read access,not requesting write/admin
access. This API reference is general guidance,not attestation of the installed
exporter's runtime implementation or a diagnosis of this machine's policy.

## Next narrowly scoped decision

Propose checking only the exact current-user key's read accessibility in the
current execution context,without changing its value or permissions,enumerating
other registry locations,or launching the exporter. Prefer an existence/access
result without publishing the OptIn value. If access is denied,retain the denial
and stop;do not automatically elevate or alter protections. A successful read
would still not reproduce the original child's token/context and would require
careful comparison before any new exporter run.

This is a proposal,not executed here. Broader registry inspection,token changes,
compatibility-variable changes,telemetry preference changes and installation repair
are not authorized by this report. No new capture/analysis capability demonstrated.
The original300s Go timeout and other collector/storage/privacy gates remain open.
Author review only;only this report and recovery/status documents are changed.
