param([Parameter(Mandatory)][ValidateSet('red','focused','full')][string]$Slot,[string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$OutputEncoding=[Text.UTF8Encoding]::new($false)
if($Slot -ceq 'full'){& (Join-Path $PSScriptRoot 'control-frame.tests.ps1') -Group all -TracePath $TracePath}
elseif($Slot -ceq 'red'){& (Join-Path $PSScriptRoot 'scalar.tests.ps1') -Probe}
else{& (Join-Path $PSScriptRoot 'scalar.tests.ps1')}
exit $LASTEXITCODE
