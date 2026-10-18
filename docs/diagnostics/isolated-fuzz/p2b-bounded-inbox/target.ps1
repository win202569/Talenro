param([ValidateSet('red','full','capacity','sequence','final')][string]$Slot,[string]$TracePath)
$group=if($Slot -ceq 'capacity'){'capacity'}elseif($Slot -ceq 'sequence'){'sequence'}else{'all'}
& (Join-Path $PSScriptRoot 'inbox.tests.ps1') -Group $group -TracePath $TracePath
exit $LASTEXITCODE
