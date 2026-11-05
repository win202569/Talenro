Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot runtime-parent-audit-fixtures.ps1)
$base=Get-PABase;$workspace=Join-Path $base cases
$driverPath=Join-Path $PSScriptRoot runtime-staged-integrity-driver.ps1
$commonPath=Join-Path $PSScriptRoot runtime-copy-transport-common.ps1
$slice=Get-ParentAuditSlice ([IO.File]::ReadAllText($driverPath)) ([IO.File]::ReadAllText($commonPath))
$sourcePins=@{};foreach($p in @($driverPath,$commonPath,(Join-Path $PSScriptRoot runtime-parent-audit-fixtures.ps1),(Join-Path $PSScriptRoot runtime-parent-audit-fixtures.tests.ps1),$PSCommandPath)){$sourcePins[$p]=(Get-FileHash $p).Hash}
Save-PA (Join-Path $base source-pins.json) $sourcePins
Save-PA (Join-Path $base matrix.claim.json) @{MaxEvaluations=10;AttemptConsumed=$true;BusinessStarts=0;Counters=@(44,34,6,217)}
function Need-PA($ok,$why){if(-not $ok){throw "PA_MATRIX: $why"}}
function Read-PA($path){Get-Content -LiteralPath $path -Raw|ConvertFrom-Json -AsHashtable}
$results=[Collections.Generic.List[object]]::new();$invocations=0;$current=$null
try{
 foreach($number in 1..10){
  $current='P{0:D2}' -f $number;$root=Join-Path $workspace $current
  $console=[IO.StringWriter]::new();$originalConsole=[Console]::Out
  try{[Console]::SetOut($console);$invocations++;$run=Invoke-ParentAuditCase $current $workspace $slice}
  finally{[Console]::SetOut($originalConsole);$consoleText=$console.ToString();$console.Dispose()}
  Save-PA (Join-Path $root console.json) @{Text=$consoleText}
  Need-PA $run.Terminal 'runspace not terminal'
  $success=$consoleText.Contains('BUSINESS_PREFIX_TRANSPORT_COMPLETED_REVIEW_REQUIRED')
  $stopPath=Join-Path $root execution.stop.json;$auditPath=Join-Path $root integrity-review.json;$resultPath=Join-Path $root result.json;$mapPath=Join-Path $root supervisor.after-map.json
  $stop=if(Test-Path $stopPath){Read-PA $stopPath}else{$null}
  if($number -eq 1){Need-PA ($success -and $null -eq $stop -and -not $run.HadErrors) 'normal completion'}
  else{Need-PA (-not $success -and $null -ne $stop -and $stop.Status -ceq 'STOPPED' -and $stop.NoRetry -is [bool] -and $stop.NoRetry) 'failure must stop without success marker'}
  $sentinels=@{P04='supervisor.after-map.json';P05='integrity-review.json';P06='result.json';P07='supervision.json'}
  if($sentinels.ContainsKey($current)){Need-PA ([IO.File]::ReadAllText((Join-Path $root $sentinels[$current])) -ceq 'SENTINEL-NOT-EVIDENCE') 'sentinel changed'}
  if($number -ne 7){
   $supervision=Read-PA (Join-Path $root supervision.json);$context=Read-PA (Join-Path $root context.json)
   Need-PA (($supervision|ConvertTo-Json -Depth 20 -Compress) -ceq ($context.Retained|ConvertTo-Json -Depth 20 -Compress)) 'supervision changed'
  }
  if($number -in @(1,6,10)){
   $audit=Read-PA $auditPath;$map=Read-PA $mapPath
   Need-PA ($audit.State -ceq 'VERIFIED' -and $null -eq $audit.Error -and $null -ne $audit.After) 'verified audit required'
   Need-PA ($map.Count -eq 1 -and $map['source.txt'] -ceq (Get-FileHash (Join-Path $root source.txt)).Hash) 'real map hash'
   Need-PA ($audit.After.SHA256 -ceq (Get-FileHash $mapPath).Hash -and $audit.After.Length -eq (Get-Item $mapPath).Length -and $audit.After.Count -eq 1) 'after metadata'
  }
  if($number -in @(2,3,4)){
   $audit=Read-PA $auditPath
   Need-PA ($audit.State -ceq 'FAILED' -and $null -eq $audit.After -and -not [string]::IsNullOrEmpty($audit.Error)) 'failed audit required'
   $pattern=@{2='input drift';3='absent pin';4='exist'}[$number]
   Need-PA ($audit.Error -match $pattern -and $stop.Message -match 'post-exit integrity audit failed') 'wrong audit refusal'
   Need-PA (-not(Test-Path $resultPath)) 'unexpected result'
   if($number -ne 4){Need-PA (-not(Test-Path $mapPath)) 'unexpected after map'}
  }
  if($number -in @(5,6,7)){Need-PA ($stop.Message -match 'exist') 'exclusive creation refusal'}
  if($number -eq 5){Need-PA (-not(Test-Path $resultPath)) 'result after audit save failure'}
  if($number -in @(7,8,9)){
   foreach($p in @($mapPath,$auditPath,$resultPath)){Need-PA (-not(Test-Path $p)) 'postprocessing despite early refusal'}
  }
  if($number -in @(8,9)){Need-PA ($stop.Message -match 'pending resource or unknown exit' -and $stop.Message -notmatch 'input drift') 'terminal gate ordering'}
  if($number -in @(1,10)){
   $result=Read-PA $resultPath
   Need-PA ($result.Status -ceq 'NOT_ACCEPTED' -and $result.BusinessVerified -is [bool] -and -not $result.BusinessVerified) 'result acceptance widened'
  }
  if($number -eq 10){Need-PA ($stop.Message -match 'transport did not succeed') 'transport failure upgraded'}
  foreach($p in $sourcePins.Keys){Need-PA ((Get-FileHash -LiteralPath $p).Hash -ceq $sourcePins[$p]) 'source changed'}
  $result=[ordered]@{Case=$current;Matched=$true;Terminal=$run.Terminal;SuccessMarker=$success;StopReason=$(if($null -ne $stop){$stop.Message}else{$null});OriginalSourcesUnchanged=$true;Synthetic=$true}
  $results.Add($result);Save-PA (Join-Path $root case-result.json) $result
  [Console]::Out.WriteLine("$current matched=True")
 }
 Save-PA (Join-Path $base matrix-result.json) @{Status='PARENT_AUDIT_MATRIX_VERIFIED';Invocations=$invocations;Results=$results.ToArray();NotRun=@();Counters=@(44,34,6,217);Overall='NOT_ACCEPTED';BusinessStarts=0;NewChildProcesses=0}
 'PARENT_AUDIT_MATRIX_VERIFIED'
}catch{
 $failure=[ordered]@{Status='PARENT_AUDIT_MATRIX_STOPPED';Reason=$_.Exception.Message;Stack=$_.ScriptStackTrace;CurrentCase=$current;Invocations=$invocations;Results=$results.ToArray();Unfinished=@(1..10|ForEach-Object {'P{0:D2}' -f $_}|Where-Object {$_ -notin @($results|ForEach-Object Case)});NoRetry=$true;Counters=@(44,34,6,217);Overall='NOT_ACCEPTED';BusinessStarts=0;NewChildProcesses=0}
 Save-PA (Join-Path $base matrix-result.json) $failure
 $failure|ConvertTo-Json -Depth 15
 exit 1
}
