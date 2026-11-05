Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot runtime-staged-evidence-fixtures-v2.ps1)
$workspace=Get-SEBase;$caseWorkspace=Join-Path $workspace cases
$repo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$d='.superpowers/sdd/2026-11-05-runtime-staged-integrity-v1'
$source=[IO.File]::ReadAllText((Join-Path $PSScriptRoot runtime-staged-integrity-review-v2.ps1))
$utf8=[Text.UTF8Encoding]::new($false,$true)
function Save-SE([string]$Path,$Value){Write-SEBytes $Path ($utf8.GetBytes((ConvertTo-Json -InputObject $Value -Depth 100)))}
function Read-SE([string]$Path){Get-Content -LiteralPath $Path -Raw|ConvertFrom-Json -AsHashtable}
function Edit-SE([string]$Path,[byte[]]$Bytes){
 $p=Assert-SEPath $Path
 Write-SEBytes ($p+'.baseline') ([IO.File]::ReadAllBytes($p))
 [IO.File]::WriteAllBytes($p,$Bytes)
}
function Edit-SEJson([string]$Path,$Value){Edit-SE $Path ($utf8.GetBytes((ConvertTo-Json -InputObject $Value -Depth 100)))}
function Omit-SE([string]$Path){
 $p=Assert-SEPath $Path;$to=Assert-SEPath ($p+'.omitted')
 if(Test-Path -LiteralPath $to){throw 'SE occupied omission'}
 Move-Item -LiteralPath $p -Destination $to
}
$results=[Collections.Generic.List[object]]::new();$invoked=0;$current=$null
try{
 if((Get-Location).Path -cne $repo){throw 'SE controller root'}
 Save-SE (Join-Path $workspace matrix.claim.json) ([ordered]@{MaxEvaluations=16;AttemptConsumed=$true;BusinessStarts=0;Counters=@(44,34,6,217)})
 $pins=[ordered]@{}
 foreach($name in @('runtime-staged-evidence-fixtures-v2.ps1','runtime-staged-evidence-payload-v2.tests.ps1','runtime-staged-evidence-matrix-v2.ps1','runtime-staged-integrity-review-v2.ps1','portable-runtime-staged-integrity-observation-final.json','portable-runtime-staged-integrity-review-v2.json')){
  $p=Join-Path $PSScriptRoot $name;$pins[$p]=(Get-FileHash -LiteralPath $p).Hash
 }
 Save-SE (Join-Path $workspace source-pins.json) $pins
 $oldArchive=Read-SE (Join-Path $PSScriptRoot portable-runtime-staged-integrity-observation-final.json)
 foreach($entry in $oldArchive.Files){if((Get-FileHash -LiteralPath (Join-Path $repo $entry.Path)).Hash -cne $entry.SHA256){throw 'SE original drift before matrix'}}
 $categories=@('','integrity-review|Cannot find','persisted full audit','JSON|ConvertFrom-Json|Unexpected|Invalid','audit identity','audit identity','duplicate|ambiguous|JSON|audit','owned exit0|Boolean|type','after-map|Cannot find','map ','owned exit0','terminal supervision','terminal supervision','four-phase timing','stable length','already exists|exist|verification.json')
 for($number=1;$number -le 16;$number++){
  $current='C{0:D2}' -f $number
  $fixture=New-SEFixture $current $caseWorkspace
  $root=$fixture.Root;$data=Join-Path $root $d;$review=Join-Path $root $fixture.ReviewRelative
  $scriptText=ConvertTo-SEFixtureVerifier $source $root $fixture.ReviewRelative
  Write-SEBytes (Join-Path $root rebound-verifier.ps1) ($utf8.GetBytes($scriptText))
  $auditPath=Join-Path $data integrity-review.json;$supervisionPath=Join-Path $data supervision.json;$rowPath=Join-Path $data result.json
  switch($current){
   C02 {Omit-SE $auditPath}
   C03 {$a=Read-SE $auditPath;$a.State='FAILED';$a.After=$null;$a.Error='fixture failure';Edit-SEJson $auditPath $a}
   C04 {Edit-SE $auditPath ($utf8.GetBytes('{invalid'))}
   C05 {$a=Read-SE $auditPath;$a.InputSHA256='B'*64;Edit-SEJson $auditPath $a}
   C06 {$a=Read-SE $auditPath;$a.IntentSHA256='B'*64;Edit-SEJson $auditPath $a}
   C07 {
    $raw=[IO.File]::ReadAllText($auditPath)
    $changed=[regex]::Replace($raw,'"State"\s*:\s*"VERIFIED"','"State":"FAILED","State":"VERIFIED"')
    if($changed -ceq $raw){throw 'SE duplicate-key mutation absent'}
    Edit-SE $auditPath ($utf8.GetBytes($changed))
   }
   C08 {$s=Read-SE $supervisionPath;$s.Writer.Associated='false';$row=Read-SE $rowPath;$row.Result=$s;Edit-SEJson $supervisionPath $s;Edit-SEJson $rowPath $row}
   C09 {Omit-SE (Join-Path $data supervisor.after-map.json)}
   C10 {
    $mapPath=Join-Path $data supervisor.after-map.json;$map=Read-SE $mapPath;$key=@($map.Keys)[0];$map[$key]='B'*64;Edit-SEJson $mapPath $map
    $a=Read-SE $auditPath;$a.After.SHA256=(Get-FileHash $mapPath).Hash;$a.After.Length=(Get-Item $mapPath).Length
    $row=Read-SE $rowPath;$row.After=$a.After;Edit-SEJson $auditPath $a;Edit-SEJson $rowPath $row
   }
   C11 {$s=Read-SE $supervisionPath;$s.Writer.ExitCode=1;$row=Read-SE $rowPath;$row.Result=$s;Edit-SEJson $supervisionPath $s;Edit-SEJson $rowPath $row}
   C12 {$s=Read-SE $supervisionPath;$s.PendingResources=$true;$row=Read-SE $rowPath;$row.Result=$s;Edit-SEJson $supervisionPath $s;Edit-SEJson $rowPath $row;Omit-SE (Join-Path $data data/stdout.bin);Omit-SE (Join-Path $data data/stderr.bin)}
   C13 {$s=Read-SE $supervisionPath;$s.ExitUnconfirmed=$true;$row=Read-SE $rowPath;$row.Result=$s;Edit-SEJson $supervisionPath $s;Edit-SEJson $rowPath $row;Omit-SE (Join-Path $data data/stdout.bin);Omit-SE (Join-Path $data data/stderr.bin)}
   C14 {
    $s=Read-SE $supervisionPath;$lines=$utf8.GetString([Convert]::FromBase64String($s.DiagnosticBase64)).Split([char]10)
    $s.DiagnosticBase64=[Convert]::ToBase64String($utf8.GetBytes(($lines[0..2] -join "`n")+"`n"));$row=Read-SE $rowPath;$row.Result=$s;Edit-SEJson $supervisionPath $s;Edit-SEJson $rowPath $row
   }
   C15 {Edit-SE (Join-Path $data data/stderr.bin) ([byte[]]@(1))}
   C16 {Write-SEBytes (Join-Path $review verification.json) ($utf8.GetBytes('SENTINEL-NOT-A-VERDICT'))}
  }
  $before=[ordered]@{}
  foreach($file in Get-ChildItem -LiteralPath $root -File -Recurse){$before[[IO.Path]::GetRelativePath($root,$file.FullName)]=(Get-FileHash -LiteralPath $file.FullName).Hash}
  Save-SE (Join-Path $root fixture-inputs.json) $before
  $invoked++
  $run=Invoke-SEFixture $scriptText $root $fixture.InputSHA256
  Save-SE (Join-Path $root invocation.json) $run
  if(-not $run.Terminal){throw 'SE uncertain runspace completion'}
  $verdict=$null;$verdictError=$null;$verdictPath=Join-Path $review verification.json
  if(Test-Path -LiteralPath $verdictPath){try{$verdict=Read-SE $verdictPath}catch{$verdictError=$_.Exception.Message}}
  $passed=$null -ne $verdict -and $verdict.Contains('ObservationPassed') -and $verdict.ObservationPassed -is [bool] -and $verdict.ObservationPassed
  $reason=if($null -ne $verdict -and $verdict.Contains('Reason')){[string]$verdict.Reason}else{($run.Errors -join '|')+'|'+$verdictError}
  $ok=$false
  if($number -eq 1){
   $ok=$passed -and $verdict.Status -ceq 'STAGED_PREFIX_VERIFIED' -and -not $run.HadErrors -and $verdict.Facts.Count -eq 10
   if($ok){foreach($value in $verdict.Facts.Values){if($value -isnot [bool] -or -not $value){$ok=$false}}}
  }elseif($number -eq 16){$ok=-not $passed -and [IO.File]::ReadAllText($verdictPath) -ceq 'SENTINEL-NOT-A-VERDICT' -and $run.HadErrors}
  else{$ok=-not $passed -and $null -ne $verdict -and $verdict.Status -ceq 'STAGED_PREFIX_FAILED' -and $reason -match $categories[$number-1]}
  foreach($p in $before.Keys){if((Get-FileHash -LiteralPath (Join-Path $root $p)).Hash -cne $before[$p]){throw 'SE fixture input modified by verifier'}}
  foreach($entry in $oldArchive.Files){if((Get-FileHash -LiteralPath (Join-Path $repo $entry.Path)).Hash -cne $entry.SHA256){throw 'SE original evidence modified'}}
  $result=[ordered]@{Case=$current;Expected=if($number -eq 1){'ACCEPT'}else{'REJECT'};ObservationPassed=$passed;Matched=$ok;Reason=$reason;Terminal=$run.Terminal;OriginalEvidenceUnchanged=$true}
  $results.Add($result);Save-SE (Join-Path $root case-result.json) $result
  [Console]::Out.WriteLine("$current matched=$ok observationPassed=$passed")
  if(-not $ok){throw "SE unexpected outcome $current"}
 }
 Save-SE (Join-Path $workspace matrix-result.json) ([ordered]@{Status='MATRIX_VERIFIED';Invocations=$invoked;Results=$results.ToArray();NotRun=@();NewChildProcesses=0;Counters=@(44,34,6,217);Overall='NOT_ACCEPTED'})
}catch{
 $remaining=@(1..16|ForEach-Object {'C{0:D2}' -f $_}|Where-Object {$_ -notin @($results|ForEach-Object Case)})
 $failure=[ordered]@{Status='MATRIX_STOPPED';Reason=$_.Exception.Message;Stack=$_.ScriptStackTrace;CurrentCase=$current;Invocations=$invoked;Results=$results.ToArray();Unfinished=$remaining;NewChildProcesses=0;Counters=@(44,34,6,217);Overall='NOT_ACCEPTED';NoRetry=$true}
 Save-SE (Join-Path $workspace matrix-result.json) $failure
 $failure|ConvertTo-Json -Depth 12
 exit 1
}
