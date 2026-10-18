$ErrorActionPreference='Stop'
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
Write-Output 'STAGE: ReflectionReuse'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.reflection-reuse.tests.ps1')
if(-not $?){exit 1}
Write-Output 'STAGE: ExistingAll'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.ownership.tests.ps1') -Group all
if(-not $?){exit 1}
Write-Output 'PASS: ReflectionReuseRegression'
