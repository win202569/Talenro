param([ValidateSet('v2','v3')][string]$Version='v3')
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$source=Join-Path $PSScriptRoot "runtime-staged-integrity-review-$Version.ps1"
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$tokens,[ref]$errors)
if($errors.Count){throw 'test source parse failed'}
# Load only actual parser functions, never the verifier body or business work.
foreach($name in @('ConvertFrom-SIUniqueJson','ReadJson')){
 $defs=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst]},$false)|Where-Object Name -CEQ $name)
 if($defs.Count -eq 1){. ([scriptblock]::Create($defs[0].Extent.Text))}
}
$base=Join-Path (Split-Path (Split-Path (Split-Path (Split-Path $PSScriptRoot)))) ".superpowers/sdd/2026-11-05-staged-evidence-fixtures-v3/parser-$Version"
if(Test-Path $base){throw 'occupied parser test root'}
$null=[IO.Directory]::CreateDirectory($base)
$utf8=[Text.UTF8Encoding]::new($false)
function Case($name,$raw,$reject){
 $p=Join-Path $base "$name.json"
 $s=[IO.File]::Open($p,[IO.FileMode]::CreateNew)
 try{$b=$utf8.GetBytes($raw);$s.Write($b,0,$b.Length)}finally{$s.Dispose()}
 $errorText=$null;$value=$null
 try{$value=ReadJson $p}catch{$errorText=$_.Exception.Message}
 if($reject){if($null -eq $errorText -or $errorText -notmatch 'duplicate|JSON'){throw "UNIQUE_JSON_FAIL: $name accepted or wrong refusal: $errorText"}}
 else{if($null -ne $errorText -or $value.State -cne 'VERIFIED' -or $value.Enabled -isnot [bool] -or $value.Enabled){throw "UNIQUE_JSON_FAIL: $name changed valid data"}}
 Write-Output "PASS: $name"
}
Case duplicate-state '{"State":"FAILED","State":"VERIFIED"}' $true
Case reversed-state '{"State":"VERIFIED","State":"FAILED"}' $true
Case nested '{"Audit":{"State":"FAILED","State":"VERIFIED"}}' $true
Case array-nested '{"Audits":[{"State":0,"State":1}]}' $true
Case escaped-name '{"State":0,"St\u0061te":1}' $true
Case case-alias '{"State":0,"state":1}' $true
Case malformed '{"State":' $true
Case valid-siblings '{"State":"VERIFIED","Enabled":false,"A":{"State":0},"B":{"State":1},"Items":[{"State":2},{"State":3}]}' $false
Case key-text-in-string '{"State":"VERIFIED","Enabled":false,"Text":"\"State\":0,\"State\":1"}' $false
'UNIQUE_JSON_VERIFIED: 9 cases'
