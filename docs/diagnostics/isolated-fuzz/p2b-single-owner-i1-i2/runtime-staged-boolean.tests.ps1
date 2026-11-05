param([ValidateSet('v3','v4')][string]$Version='v4')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$t=$null;$e=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot "runtime-staged-integrity-review-$Version.ps1"),[ref]$t,[ref]$e)
if($e.Count){throw 'parse failure'}
foreach($name in @('Need','Test-SIBool')){
 $defs=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst]},$false)|Where-Object Name -CEQ $name)
 if($defs.Count -eq 1){. ([scriptblock]::Create($defs[0].Extent.Text))}
}
# Execute the actual role predicate, not a copied approximation or whole verifier.
$commands=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Need' -and $n.Extent.Text.Contains('"$role owned exit0"')},$true))
if($commands.Count -ne 1){throw 'role predicate selection failure'}
$predicate=[scriptblock]::Create($commands[0].Extent.Text);$role='Writer';$count=0
function RoleCase($field,$value,$expected){
 $p=@{StartAttempted=$true;Associated=$true;OwnedExit=$true;ExitCode=0;KillRequested=$false;Error=$null}
 $p[$field]=$value;$accepted=$true
 try{& $predicate}catch{if($_.Exception.Message -notmatch 'owned exit0'){throw};$accepted=$false}
 if($accepted -ne $expected){throw "BOOLEAN_FAIL: $field value '$value' accepted=$accepted expected=$expected"}
 $script:count++
}
RoleCase Associated 'false' $false
foreach($field in @('StartAttempted','Associated','OwnedExit','KillRequested')){
 foreach($value in @('true','false','',0,1,$null,@(),@(1),@{})){RoleCase $field $value $false}
 RoleCase $field $true ($field -cne 'KillRequested')
 RoleCase $field $false ($field -ceq 'KillRequested')
}
foreach($expected in @($true,$false)){
 foreach($raw in @('"false"','"true"','0','1','null','[]','[true]','{}','""')){
  $value=ConvertFrom-Json -InputObject $raw -NoEnumerate
  if(Test-SIBool $value $expected){throw "BOOLEAN_FAIL: generic coercion $raw"};$count++
 }
 if(-not(Test-SIBool $expected $expected) -or (Test-SIBool (-not $expected) $expected)){throw 'BOOLEAN_FAIL: typed Boolean semantics'};$count++
}
"BOOLEAN_VERIFIED: $count assertions; actual role predicate and typed matcher"
