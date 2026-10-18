param([Parameter(Mandatory)][ValidateSet('red','full','capacity','sequence','final')][string]$Slot,[string]$TracePath,[string]$OriginalTracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
if($Slot -ceq 'red'){& (Join-Path $PSScriptRoot 'tree.tests.ps1') -Group red -TracePath $TracePath;if(-not $?){exit 1};exit 0}
if($Slot -cin @('full','final')){
    & (Join-Path $PSScriptRoot 'tree.tests.ps1') -Group all -TracePath $TracePath
    if(-not $?){exit 1}
    & (Join-Path $PSScriptRoot 'inbox.tests.ps1') -Group all -TracePath $OriginalTracePath
    if(-not $?){exit 1}
    exit 0
}
& (Join-Path $PSScriptRoot 'inbox.tests.ps1') -Group $Slot -TracePath $TracePath
if(-not $?){exit 1}
exit 0
