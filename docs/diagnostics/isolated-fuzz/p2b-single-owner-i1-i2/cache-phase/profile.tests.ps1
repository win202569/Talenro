param([ValidateSet('OwnerRed','OwnerGreen')][string]$Phase,[Parameter(Mandatory)][string]$Root)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'owner.ps1')
$repo='D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator'
$root=New-C12WaitRoot $Root
$cases=[Collections.Generic.List[object]]::new();$session='a'*32
$digest='h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='
function Check($b,$message){if(-not $b){throw $message}}
function Reject($body,$code){$m=$null;try{& $body}catch{$m=$_.Exception.Message};Check ($null -ne $m -and $m.Contains($code)) "Expected $code;actual $m"}
function Case($name,$body){& $body;$cases.Add(@{Name=$name;Status='PASS'})}
function Events($ids){
 $list=[Collections.Generic.List[object]]::new()
 foreach($id in $ids){foreach($p in @('PREPARE','ZIP','DIRECTORY')){foreach($kind in @('START','END')){
  $list.Add(@{Schema=1;Sequence=$list.Count+1;Session=$session;Module=$id.Module;Version=$id.Version;Phase=$p;Kind=$kind;Status=$(if($kind -ceq 'START'){'STARTED'}else{'OK'});Digest=$(if($kind -ceq 'END' -and $p -cne 'PREPARE'){$digest}else{''});ElapsedNs=0})
 }}}
 return ,$list.ToArray()
}
function Lines($events){return (($events|ForEach-Object {($_|ConvertTo-Json -Compress)+"`n"}) -join '')}
function File($name,$text){$p=Join-Path $root $name;[IO.File]::WriteAllText($p,$text,[Text.UTF8Encoding]::new($false));return $p}
function Append($w,$text){$b=[Text.Encoding]::UTF8.GetBytes($text);$w.Write($b,0,$b.Length);$w.Flush()}
$libc=Events @(@{Module='modernc.org/libc';Version='v1.72.1'})
$paired=Events @(@{Module='golang.org/x/text';Version='v0.40.0'},@{Module='modernc.org/sqlite';Version='v1.49.1'})
$libcPath=File 'libc.jsonl' (Lines $libc);$pairedPath=File 'paired.jsonl' (Lines $paired)
function Owned($name,$mode,$inputPath,$deadline,$raw=$false){
 $pwsh=[Diagnostics.Process]::GetCurrentProcess().MainModule.FileName
 $params=@{Executable=$pwsh;Arguments=@('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'profile-fixture.ps1'),'-InputPath',$inputPath,'-Mode',$mode);WorkingDirectory=$repo;Environment=@{};Root=(New-C12WaitRoot (Join-Path $root $name));DeadlineMs=$deadline}
 if(-not $raw){$params.Session=$session;$params.Profile='libc-v1'}
 return Invoke-C12PhaseOwned @params
}
try{
 Case 'libc-six-events' {$v=Read-C12PhaseEvents -Path $libcPath -Session $session -Profile 'libc-v1';Check ($v.Complete -and $v.Events.Count -eq 6 -and $v.Phases.Count -eq 3) 'Libc completion'}
 if($Phase -ceq 'OwnerRed'){throw 'UNEXPECTED_RED_PASS'}
 Case 'paired-default-and-explicit' {
  $a=Read-C12PhaseEvents -Path $pairedPath -Session $session
  $b=Read-C12PhaseEvents -Path $pairedPath -Session $session -Profile 'paired-v1'
  Check ($a.Complete -and $b.Complete -and $a.Events.Count -eq 12 -and $b.Phases.Count -eq 6) 'Paired regression'
 }
 Case 'resolver-identities-and-copy' {
  $a=@(Get-C12PhaseProfile);Check ($a.Count -eq 2 -and $a[0].Module -ceq 'golang.org/x/text' -and $a[0].Version -ceq 'v0.40.0' -and $a[1].Module -ceq 'modernc.org/sqlite' -and $a[1].Version -ceq 'v1.49.1') 'Paired IDs'
  $a[0].Module='changed';$b=@(Get-C12PhaseProfile);Check ($b[0].Module -ceq 'golang.org/x/text') 'Shared state'
  $c=@(Get-C12PhaseProfile 'libc-v1');Check ($c.Count -eq 1 -and $c[0].Module -ceq 'modernc.org/libc' -and $c[0].Version -ceq 'v1.72.1') 'Libc IDs'
 }
 Case 'invalid-profile-before-io' {
  foreach($p in @('','unknown','PAIRED-V1','LIBC-V1')){
   Reject {Get-C12PhaseProfile $p} 'PROFILE'
   Reject {Read-C12PhaseEvents -Path (Join-Path $root 'absent') -Session $session -Profile $p} 'PROFILE'
   Reject {Invoke-C12PhaseOwned -Executable 'missing.exe' -Root (Join-Path $root 'absent') -DeadlineMs 100 -Profile $p} 'PROFILE'
  }
  Reject {Invoke-C12PhaseOwned -Executable 'missing.exe' -Root (Join-Path $root 'absent') -DeadlineMs 100 -Profile 'libc-v1'} 'PROFILE_SESSION'
 }
 Case 'cross-profile-both-directions' {
  Reject {Read-C12PhaseEvents -Path $pairedPath -Session $session -Profile 'libc-v1'} 'ORDER'
  Reject {Read-C12PhaseEvents -Path $libcPath -Session $session} 'ORDER'
 }
 Case 'identity-binding-order' {
  foreach($key in @('Version','Session','Sequence','Phase')){
   $bad=$libc[0].Clone();$bad[$key]=switch($key){Version{'v1.72.0'} Session{'b'*32} Sequence{2} Phase{'ZIP'}}
   $p=File "$key.jsonl" (Lines @($bad));$code=if($key -in @('Session','Sequence')){'BINDING'}else{'ORDER'}
   Reject {Read-C12PhaseEvents -Path $p -Session $session -Profile 'libc-v1'} $code
  }
 }
 Case 'extra-terminal' {Reject {Read-C12PhaseEvents -Path (File 'extra.jsonl' ((Lines $libc)+(Lines @($libc[0])))) -Session $session -Profile 'libc-v1'} 'ORDER_AFTER_TERMINAL'}
 Case 'missing-end' {$v=Read-C12PhaseEvents -Path (File 'start.jsonl' (Lines @($libc[0]))) -Session $session -Profile 'libc-v1';Check (-not $v.Complete -and $v.Phases[0].Status -ceq 'Incomplete' -and $v.Phases[1].Status -ceq 'NotRun') 'Incomplete lost'}
 Case 'open-writer-partial-append' {
  $p=Join-Path $root 'open.jsonl';$w=[IO.File]::Open($p,'CreateNew','Write','Read')
  try{
   $line=Lines @($libc[1]);Append $w ((Lines @($libc[0]))+$line.Substring(0,15))
   $v=Read-C12PhaseEvents -Path $p -Session $session -Profile 'libc-v1';Check ($v.Partial -and $v.Events.Count -eq 1) 'Partial accepted'
   $blocked=$false;try{$other=[IO.File]::Open($p,'Open','Write','ReadWrite');$other.Dispose()}catch [IO.IOException]{$blocked=$true};Check $blocked 'Second writer allowed'
   Append $w $line.Substring(15);$v=Read-C12PhaseEvents -Path $p -Session $session -Profile 'libc-v1';Check (-not $v.Partial -and $v.Events.Count -eq 2 -and $v.Phases[0].Status -ceq 'Complete') 'Append missing'
  }finally{$w.Dispose()}
 }
 Case 'failed-event-stops' {
  foreach($status in @('ERROR','MISMATCH')){$bad=$libc[1].Clone();$bad.Status=$status;$p=File "$status.jsonl" (Lines @($libc[0],$bad));$v=Read-C12PhaseEvents -Path $p -Session $session -Profile 'libc-v1';Check ($v.Failed -and -not $v.Complete) 'Failure accepted';Reject {Read-C12PhaseEvents -Path (File "$status-extra.jsonl" (Lines @($libc[0],$bad,$libc[2]))) -Session $session -Profile 'libc-v1'} 'ORDER_AFTER_TERMINAL'}
 }
 Case 'missing-digest' {$bad=$libc[3].Clone();$bad.Digest='';Reject {Read-C12PhaseEvents -Path (File 'digest.jsonl' (Lines @($libc[0],$libc[1],$libc[2],$bad))) -Session $session -Profile 'libc-v1'} 'EVENT_DIGEST'}
 Case 'caps' {
  Reject {Read-C12PhaseEvents -Path (File 'event-limit.jsonl' ('x'*65537)) -Session $session -Profile 'libc-v1'} 'EVENT_LIMIT'
  $quota=Join-Path (Split-Path $root -Parent) "$Phase-quota";[void][IO.Directory]::CreateDirectory($quota);$p=Join-Path $quota 'zeros.raw';$w=[IO.File]::Open($p,'CreateNew','Write','Read');try{$w.SetLength(4194305)}finally{$w.Dispose()}
  Reject {Read-C12PhaseEvents -Path $p -Session $session -Profile 'libc-v1'} 'TOTAL_LIMIT'
 }
 Case 'occupied-record' {$r=New-C12WaitRoot (Join-Path $root 'occupied');$v=Write-C12WaitRecord $r 'sentinel.json' @{Value='original'};Reject {Write-C12WaitRecord $r 'sentinel.json' @{Value='changed'}} 'OCCUPIED';Check ((Get-FileHash $v.Path).Hash -ceq $v.SHA256) 'Changed sentinel'}
 Case 'deadline-boundary' {Check ((Get-C12WaitDecision 1000 1000 $false $true) -ceq 'STOP') 'Boundary';Check ((Get-C12WaitDecision 999 1000 $false $false) -ceq 'WAIT') 'Early stop'}
 Case 'owned-libc-complete' {$r=Owned 'complete' 'complete' $libcPath 5000;Check ($r.Started -and $r.ExitConfirmed -and $r.ExitCode -eq 0 -and -not $r.TimedOut -and $null -eq $r.Failure) "Complete: $($r.Failure)";$v=Read-C12PhaseEvents $r.StdoutPath $session 'libc-v1';Check ($v.Complete -and $v.Events.Count -eq 6) 'Final parse';Check ((Get-FileHash $libcPath).Hash -ceq (Get-FileHash $r.StdoutPath).Hash) 'Bytes changed'}
 Case 'owned-wrong-profile-stop' {$r=Owned 'wrong' 'wait' $pairedPath 5000;Check ($r.Started -and $r.ExitConfirmed -and -not $r.TimedOut -and $r.EndReason -ceq 'ERROR' -and $r.Failure -match 'ORDER') 'Protocol stop'}
 Case 'owned-partial-cutoff' {$p=File 'cutoff-input.jsonl' (Lines @($libc[0]));$r=Owned 'cutoff' 'wait' $p 3000;Check ($r.Started -and $r.ExitConfirmed -and $r.TimedOut -and $r.EndReason -ceq 'DEADLINE' -and $null -eq $r.Failure) 'Cutoff';$v=Read-C12PhaseEvents $r.StdoutPath $session 'libc-v1';Check (-not $v.Complete -and $v.Phases[0].Status -ceq 'Incomplete' -and $v.Phases[1].Status -ceq 'NotRun') 'Cutoff evidence'}
 Case 'owned-native-raw' {$p=File 'raw-input.txt' "native output`n";$r=Owned 'raw' 'complete' $p 5000 $true;Check ($r.Started -and $r.ExitConfirmed -and $r.ExitCode -eq 0 -and -not $r.TimedOut -and $null -eq $r.Failure) 'Raw regression';Check ((Get-FileHash $p).Hash -ceq (Get-FileHash $r.StdoutPath).Hash) 'Raw bytes'}
 $v=@{Status='PASS';Phase=$Phase;Cases=$cases.ToArray();Count=$cases.Count};$null=Write-C12WaitRecord $root 'test-result.json' $v;$v|ConvertTo-Json -Depth 10 -Compress
}catch{
 $v=@{Status='STOP';Phase=$Phase;Error=$_.Exception.Message;Cases=$cases.ToArray();Count=$cases.Count};$null=Write-C12WaitRecord $root 'test-result.json' $v;$v|ConvertTo-Json -Depth 10 -Compress;exit 1
}
