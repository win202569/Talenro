# Readonly PS5 command discovery diagnostic; no native adapter or workload.
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
. (Join-Path $PSScriptRoot 'run.ps1')
$PSVersionTable | Format-List | Out-String | Write-Output
Write-Output ('PSModulePath='+$env:PSModulePath)
Get-Module Microsoft.PowerShell.Utility | Select-Object Name,Path,ModuleType,@{n='ExportsHash';e={$_.ExportedCommands.ContainsKey('Get-FileHash')}} | Format-List | Out-String | Write-Output
Get-Module -ListAvailable Microsoft.PowerShell.Utility | Select-Object Name,Path,CompatiblePSEditions | Format-List | Out-String | Write-Output
$found=Get-Command Get-FileHash -ErrorAction SilentlyContinue
if($null -eq $found){Write-Output 'HashCommand=MISSING'}else{$found|Select-Object Name,Source,ModuleName,CommandType|Format-List|Out-String|Write-Output}
Write-Output 'LIMIT: readonly module/command discovery; no imports or workload.'
