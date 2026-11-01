Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
. (Join-Path $repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1')
# Publication has no collector. Its live outcome belongs to capture alone.
$script:DefaultFactory=(Get-Command New-ODRecordStream).ScriptBlock
Export-ModuleMember
