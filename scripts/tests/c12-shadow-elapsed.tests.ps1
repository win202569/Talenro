param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=Split-Path (Split-Path $PSScriptRoot)
$root=Join-Path $repo ('.superpowers/sdd/2026-11-06-c12-shadow-elapsed-v1/'+$Phase)
if(Test-Path -LiteralPath $root){throw 'occupied phase; no retry'}
$null=[IO.Directory]::CreateDirectory($root)
$utf8=[Text.UTF8Encoding]::new($false)
function Save-Elapsed($path,$value){
 $bytes=$utf8.GetBytes((ConvertTo-Json -InputObject $value -Depth 15))
 $stream=[IO.File]::Open($path,[IO.FileMode]::CreateNew)
 try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
}
function Check-Elapsed($ok,$message){if(-not $ok){throw "ELAPSED_TEST: $message"}}
Save-Elapsed (Join-Path $root 'claim.json') @{Phase=$Phase;Consumed=$true;Overall='NOT_ACCEPTED'}
$results=[Collections.Generic.List[object]]::new();$case='load'
try{
 . (Join-Path $repo 'scripts/private/c12-shadow-audit.ps1')
 foreach($case in @('matched','drift','incomplete','occupied','deadline')){
  $source=Join-Path $root $case;$null=[IO.Directory]::CreateDirectory((Join-Path $source 'testdata/c12'))
  [IO.File]::WriteAllText((Join-Path $source 'a.txt'),$(if($case -eq 'drift'){'changed'}else{'abc'}),$utf8)
  Save-Elapsed (Join-Path $source 'testdata/c12/shadow-source-manifest.v1.json') @{Schema=1;Mode='SHADOW';Files=@(@{Path='a.txt';SHA256='BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD'})}
  $outer=[Diagnostics.Stopwatch]::StartNew()
  $session=New-C12ShadowSession $source -Synthetic $true
  $lower=[Diagnostics.Stopwatch]::StartNew()
  $deadline=[datetime]::UtcNow.AddMinutes(2)
  $group=Start-C12ShadowGroup $session $source @{GroupID='g';Profile='base';Package='p';RunSuffix='r';CandidateTree=$null} $deadline
  $null=Complete-C12ShadowGroup $group $false $false $true $deadline
  $summaryPath=Join-Path $session.EvidenceRoot 'run.json'
  if($case -eq 'occupied'){Save-Elapsed $summaryPath @{Sentinel='preserve'};$sentinel=(Get-FileHash -LiteralPath $summaryPath).Hash}
  if($case -eq 'deadline'){$deadline=[datetime]::UtcNow.AddSeconds(-1)}
  $minimum=$lower.Elapsed.TotalMilliseconds
  $result=Complete-C12ShadowSession $session 0 ($case -eq 'incomplete') $true $deadline
  $maximum=$outer.Elapsed.TotalMilliseconds
  if($case -in @('occupied','deadline')){
   Check-Elapsed ($result.State -ceq 'INCOMPLETE' -and $null -eq $result.Reference) 'failure fabricated reference'
   if($case -eq 'occupied'){Check-Elapsed ((Get-FileHash -LiteralPath $summaryPath).Hash -ceq $sentinel) 'occupied summary overwritten'}
   else{Check-Elapsed ($result.Reason -ceq 'DEADLINE' -and -not(Test-Path -LiteralPath $summaryPath)) 'expired deadline ignored'}
   $elapsed=$null
  }else{
   $want=@{matched='MATCHED';drift='DRIFT';incomplete='INCOMPLETE'}[$case]
   Check-Elapsed ($result.State -ceq $want -and $null -ne $result.Reference) 'state changed'
   $saved=Get-Content -LiteralPath $summaryPath -Raw|ConvertFrom-Json -AsHashtable
   Check-Elapsed ($saved.Contains('ElapsedMilliseconds')) 'missing final ElapsedMilliseconds'
   $elapsed=$saved.ElapsedMilliseconds
   Check-Elapsed (($elapsed -is [double] -or $elapsed -is [long] -or $elapsed -is [int]) -and [double]::IsFinite([double]$elapsed) -and $elapsed -ge $minimum -and $elapsed -le $maximum) 'elapsed type or measurement bounds'
   Check-Elapsed ($saved.State -ceq $want -and $saved.FormalExitCode -eq 0 -and (Get-FileHash -LiteralPath $summaryPath).Hash -ceq $result.Reference.SHA256) 'persisted reference mismatch'
  }
  $results.Add(@{Case=$case;Matched=$true;ElapsedMilliseconds=$elapsed;Minimum=$minimum;Maximum=$maximum})
  Write-Output "PASS: $case"
 }
 Save-Elapsed (Join-Path $root 'result.json') @{Status='ELAPSED_VERIFIED';Results=$results.ToArray();CoverageBlocked=@('reparse');Overall='NOT_ACCEPTED'}
 'ELAPSED_VERIFIED'
}catch{
 Save-Elapsed (Join-Path $root 'result.json') @{Status='STOPPED';Case=$case;Reason=$_.Exception.Message;Results=$results.ToArray();NoRetry=$true;Overall='NOT_ACCEPTED'}
 throw
}
