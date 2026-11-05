param([ValidateSet('Bindings','All')][string]$Stage='Bindings')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
function Need([bool]$Ok,[string]$Why){if(-not $Ok){throw "BUSINESS_STATIC: $Why"}}
function Read([string]$Name){
 $p=Join-Path $PSScriptRoot $Name
 Need (Test-Path -LiteralPath $p) "missing $Name"
 return (Get-Content -LiteralPath $p -Raw).Replace("`r`n","`n").TrimEnd()
}
# Break caught: changes to business guards/body hidden inside a transport rebinding.
$capture=Read runtime-isolated-writer-business-capture.ps1
$expected=(Read runtime-copy-transport-capture.ps1).Replace('.superpowers/sdd/2026-11-04-runtime-copy-transport-v1','.superpowers/sdd/2026-11-05-runtime-isolated-writer-business-prefix-v1').Replace('runtime-copy-transport-common.ps1','runtime-isolated-writer-business-common.ps1')
Need ($capture -ceq $expected) 'capture binding-only delta'
$common=Read runtime-isolated-writer-business-common.ps1
$expected=(Read runtime-copy-transport-common.ps1).Replace('$intent.Computation -ne 29','$intent.Computation -ne 42').Replace('$intent.Ceiling -ne 29','$intent.Ceiling -ne 42').Replace('$intent.MaxProcesses -ne 1','$intent.MaxProcesses -ne 2')
Need ($common -ceq $expected) 'common allowance-only delta'
$role=Read runtime-isolated-writer-business-role.ps1
Need ($role.Contains("[ValidateSet('BusinessPrefix42')]") -and $role.Contains("if(`$Role -ceq 'Producer')")) 'fixed dispatch'
Need (-not $role.Contains('Producer]::RunAsync') -and -not $role.Contains('ProducerOptions')) 'no synthetic producer'
Need ($role.Contains('& (Join-Path $PSScriptRoot runtime-isolated-writer-business-capture.ps1)') -and $role.Contains('exit $LASTEXITCODE')) 'capture owns terminal receipt before exit'
$names=@('runtime-isolated-writer-business-capture.ps1','runtime-isolated-writer-business-common.ps1','runtime-isolated-writer-business-role.ps1')
if($Stage -ceq 'All'){
 $driver=Read runtime-isolated-writer-business-driver.ps1
 Need ([regex]::Matches($driver,'Supervisor\]::RunAsync\(\$options\)').Count -eq 1) 'exactly one supervisor invocation'
 Need ($driver.Contains('Computation=42') -and $driver.Contains('MaxStarts=2') -and $driver.Contains('MaxProcesses=2')) 'finite intent'
 Need (-not $driver.Contains('for(') -and -not $driver.Contains('foreach($i')) 'no scenario retry loop'
 Need ($driver.Contains("SupervisionOutcome -ceq 'SUCCESS'") -and $driver.Contains('PendingResources')) 'terminal guard'
 $names+='runtime-isolated-writer-business-driver.ps1'
}
foreach($name in $names){
 $tokens=$null;$errors=$null
 $ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot $name),[ref]$tokens,[ref]$errors)
 Need ($errors.Count -eq 0) "parse $name"
 $forbidden=@($ast.FindAll({param($a)$a -is [Management.Automation.Language.CommandAst] -and $a.GetCommandName() -cin @('Add-Type','Start-Process','Start-Job','Invoke-Expression','pwsh','powershell')},$true))
 Need ($forbidden.Count -eq 0) "no alternate launch/compile $name"
}
[Console]::Out.WriteLine('BUSINESS_PREFIX_STATIC_VERIFIED')
