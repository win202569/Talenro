# One combined invocation. Git reads reproduce checkout bytes; no live runner.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=Split-Path (Split-Path $PSScriptRoot)
$workspace=Join-Path $repo '.superpowers/sdd/2026-11-06-c12-shadow-audit-integration-v1'
$suiteRoot=Join-Path $workspace 'OfflineCombined'
if(Test-Path -LiteralPath $suiteRoot){throw 'occupied combined suite; no retry'}
$null=[IO.Directory]::CreateDirectory($suiteRoot)
function Save-Suite($name,$value){
 $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $value -Depth 20))
 $stream=[IO.File]::Open((Join-Path $suiteRoot $name),[IO.FileMode]::CreateNew)
 try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
}
Save-Suite 'claim.json' @{Consumed=$true;MaximumInvocations=1;Synthetic=$true;CoverageBlocked=@('reparse');Overall='NOT_ACCEPTED'}
$stage='checkout';$checks=[Collections.Generic.List[object]]::new()
try{
 # PowerShell >=7.4 preserves raw bytes when native stdout is redirected to a file.
 if($PSVersionTable.PSVersion -lt [version]'7.4'){throw 'raw native redirection requires PowerShell7.4+'}
 $manifest=Get-Content -LiteralPath (Join-Path $repo 'testdata/c12/shadow-source-manifest.v1.json') -Raw|ConvertFrom-Json
 $serial=0
 foreach($entry in $manifest.Files){
  $serial++;$destination=Join-Path $suiteRoot ('checkout-{0:D2}.bin' -f $serial)
  if(Test-Path -LiteralPath $destination){throw 'occupied checkout evidence'}
  # Exactly six read-only Git invocations, explicit Windows checkout policy.
  & git -C $repo -c "safe.directory=$repo" -c core.autocrlf=true cat-file --filters "--path=$($entry.Path)" "HEAD:$($entry.Path)" > $destination
  if($LASTEXITCODE -ne 0){throw "Git checkout read failed: $($entry.Path)"}
  $checkoutHash=(Get-FileHash -LiteralPath $destination).Hash
  $workingHash=(Get-FileHash -LiteralPath (Join-Path $repo $entry.Path)).Hash
  if($checkoutHash -cne $entry.SHA256 -or $workingHash -cne $entry.SHA256){throw "checkout pin mismatch: $($entry.Path)"}
  $checks.Add(@{Path=$entry.Path;CheckoutSHA256=$checkoutHash;WorkingSHA256=$workingHash;Matched=$true})
 }
 $stage='module-definition';$tokens=$null;$errors=$null
 $moduleAST=[Management.Automation.Language.Parser]::ParseFile((Join-Path $repo 'scripts/private/c12-shadow-audit.ps1'),[ref]$tokens,[ref]$errors)
 if($errors.Count -or @($moduleAST.EndBlock.Statements|Where-Object {$_ -isnot [Management.Automation.Language.FunctionDefinitionAst]}).Count){throw 'module is not definition-only'}
 $stage='modules'
 & (Join-Path $PSScriptRoot 'c12-shadow-audit.tests.ps1') -Phase Combined
 $moduleResult=Get-Content -LiteralPath (Join-Path $workspace 'Combined/result.json') -Raw|ConvertFrom-Json
 $passed=@($moduleResult.Results|Where-Object {$_.Matched -is [bool] -and $_.Matched})
 $blocked=@($moduleResult.Results|Where-Object {-not $_.Matched})
 if($moduleResult.Status -cne 'SHADOW_RECORDS_VERIFIED_WITH_COVERAGE_BLOCKER' -or $passed.Count -ne 21 -or $blocked.Count -ne 1 -or $blocked[0].Case -cne 'reparse' -or $blocked[0].Status -cne 'BLOCKED'){throw 'module results unexpected'}
 $stage='hooks'
 & (Join-Path $PSScriptRoot 'c12-shadow-hooks.tests.ps1') -Phase Combined
 $hookResult=Get-Content -LiteralPath (Join-Path $workspace 'HooksCombined/result.json') -Raw|ConvertFrom-Json
 if($hookResult.Status -cne 'SHADOW_HOOKS_VERIFIED' -or $hookResult.Results.Count -ne 12 -or @($hookResult.Results|Where-Object {$_.Matched -isnot [bool] -or -not $_.Matched}).Count){throw 'hook results unexpected'}
 Save-Suite 'result.json' @{Status='C12_SHADOW_OFFLINE_VERIFIED_WITH_COVERAGE_BLOCKER';CheckoutChecks=$checks.ToArray();ModulePassed=21;ModuleBlocked=1;HooksPassed=12;DefinitionOnly=$true;CoverageBlocked=@('reparse');Synthetic=$true;Overall='NOT_ACCEPTED'}
 'C12_SHADOW_OFFLINE_VERIFIED_WITH_COVERAGE_BLOCKER'
}catch{
 Save-Suite 'result.json' @{Status='STOPPED';Stage=$stage;Reason=$_.Exception.Message;Stack=$_.ScriptStackTrace;CheckoutChecks=$checks.ToArray();NoRetry=$true;Overall='NOT_ACCEPTED'}
 throw
}
