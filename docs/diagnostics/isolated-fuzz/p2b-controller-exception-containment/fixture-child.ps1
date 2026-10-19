param([Parameter(Mandatory)][ValidateSet('EmitAndExit','Hold')][string]$Mode)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::Out.Write('P2B-CONTAINMENT-OUT')
[Console]::Error.Write('P2B-CONTAINMENT-ERR')
if($Mode -ceq 'Hold'){[Threading.Thread]::Sleep(45000)}
exit 0
