param([Parameter(Mandatory)][ValidateSet('focused','full')][string]$Slot,[string]$TracePath)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
# This wrapper shares the single approved target process; it spawns no host.
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$OutputEncoding=[Text.UTF8Encoding]::new($false)
$leaf=Join-Path $PSScriptRoot '../p2b-scalar-raw-repair'
if($Slot -ceq 'focused'){& (Join-Path $leaf 'scalar.tests.ps1')}
else{& (Join-Path $leaf 'control-frame.tests.ps1') -Group all -TracePath $TracePath}
exit $LASTEXITCODE
