Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$helper=Join-Path $PSScriptRoot runtime-parent-audit-fixtures.ps1
if(-not(Test-Path $helper)){throw 'PARENT_BOOTSTRAP_RED: missing helper'}
. $helper
function Check($ok,$why){if(-not $ok){throw "PARENT_BOOTSTRAP: $why"}}
function Reject([scriptblock]$body){$caught=$false;try{& $body|Out-Null}catch{$caught=$true};Check $caught 'unsafe input accepted'}
$driver=[IO.File]::ReadAllText((Join-Path $PSScriptRoot runtime-staged-integrity-driver.ps1))
$common=[IO.File]::ReadAllText((Join-Path $PSScriptRoot runtime-copy-transport-common.ps1))
$slice=Get-ParentAuditSlice $driver $common
Reject {Get-ParentAuditSlice ($driver.Replace('Save-IW supervision.json','Save-IW other.json')) $common}
Reject {Get-ParentAuditSlice $driver ($common+'#changed')}
$t=$null;$e=$null;$ast=[Management.Automation.Language.Parser]::ParseInput($driver,[ref]$t,[ref]$e)
foreach($n in @('Need-IW','Save-IW','Map-IW','Journal-IW')){
 $definition=@($ast.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.FunctionDefinitionAst] -and $_.Name -ceq $n})[0]
 Check ($slice.Contains($definition.Extent.Text)) "original definition $n"
}
$outer=@($ast.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.TryStatementAst]})[0]
$first=@($outer.Body.Statements|Where-Object {$_.Extent.Text -ceq 'Save-IW supervision.json $retained'})[0]
$last=$outer.Body.Statements[-1]
$tail=$driver.Substring($first.Extent.StartOffset,$last.Extent.EndOffset-$first.Extent.StartOffset)
Check ($slice.Contains($tail) -and $slice.Contains($outer.CatchClauses[0].Extent.Text)) 'tail/catch identity'
Check (-not $slice.Contains('RunAsync') -and -not $slice.Contains('LoadFrom') -and -not $slice.Contains('Verify-IW')) 'operational entry excluded'
Reject {Assert-PAPath 'D:\outside-parent-test'}
Reject {Assert-PAPath ((Get-PABase)+'\..\outside')}
Reject {Assert-PAPath 'relative'}
$workspace=Join-Path (Get-PABase) bootstrap
$one=Invoke-ParentAuditCase P01 $workspace '"isolated"; exit 1'
Check ($one.Terminal -and 'isolated' -in $one.Output) 'exit isolation'
Reject {Invoke-ParentAuditCase P01 $workspace 'throw "must not run"'}
Reject {Invoke-ParentAuditCase P11 $workspace 'throw "must not run"'}
$two=Invoke-ParentAuditCase P02 $workspace '"success"'
Check ($two.Terminal -and -not $two.HadErrors -and 'success' -in $two.Output) 'fresh runspace'
'PARENT_AUDIT_BOOTSTRAP_VERIFIED'
