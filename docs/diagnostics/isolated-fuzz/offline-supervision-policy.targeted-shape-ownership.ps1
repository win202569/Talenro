$ErrorActionPreference='Stop'
# Targeted semantic evidence only: deliberately excludes original35/full acceptance.
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
Write-Output 'STAGE: ShapeLookup'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.shape-lookup.tests.ps1')
if(-not $?){exit 1}
Write-Output 'STAGE: OwnershipRegression'
& (Join-Path $PSScriptRoot 'offline-supervision-policy.ownership.tests.ps1') -Group regression
if(-not $?){exit 1}
Write-Output 'PASS: TargetedShapeOwnership'
Write-Output 'LIMIT: targeted shape and ownership semantics only; original35, full 30s gate and runtime acceptance excluded.'
