Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$helper=Join-Path $PSScriptRoot runtime-staged-integrity.ps1
if(-not(Test-Path $helper)){throw 'SI_TEST: missing helper'}
. $helper
$token='11111111-2222-3333-4444-555555555555';$hash='A'*64
$phases=@('task.returned','role.save.begin','role.save.end','exit.ready')
$utf8=[Text.UTF8Encoding]::new($false,$true);$script:assertions=0;$script:cases=0
function Assert([bool]$Ok,[string]$Why){$script:assertions++;if(-not $Ok){throw "SI_TEST: $Why"}}
function Case([string]$Name,[scriptblock]$Body){& $Body;$script:cases++;[Console]::Out.WriteLine("PASS: $Name")}
function Fixture([int]$Count=4){
 $text=''
 for($i=0;$i -lt $Count;$i++){$text+=('{"Schema":1,"Token":"'+$token+'","InputSHA256":"'+$hash+'","Role":"writer-staged-timing","Seq":'+($i+1)+',"Phase":"'+$phases[$i]+'","ElapsedTicks":'+(($i+1)*10)+',"Frequency":1000}'+[char]10)}
 return $text
}
function Parse([string]$Text){Read-SITrace ($utf8.GetBytes($Text)) $token $hash}
function Invalid([string]$Text){Assert ((Parse $Text).Status -ceq 'INVALID') 'reject malformed record'}
function GoodFacts{return @{TransportVerified=$true;TerminalVerified=$true;BusinessVerified=$true;AuditPersisted=$true;AuditIdentityVerified=$true;AuditMapsVerified=$true;WriterDelegationVerified=$true;RoleVerified=$true;TimingVerified=$true;OutputVerified=$true}}
# T01 catches changed wire identity/order, missing emission or wrong interval pairing.
Case T01 {
 $s=New-SIState $token $hash 1000;$m=[IO.MemoryStream]::new()
 try{for($i=0;$i -lt 4;$i++){Assert (Write-SIRecord $s $m $phases[$i] (($i+1)*10)) 'write'}
 Assert ($utf8.GetString($m.ToArray()) -ceq (Fixture)) 'exact wire bytes'
 $r=Read-SITrace $m.ToArray() $token $hash;Assert ($r.Status -ceq 'COMPLETE' -and $r.Records.Count -eq 4) 'complete'
 Assert ($r.Intervals.Count -eq 1 -and $r.Intervals[0].ElapsedTicks -eq 10 -and $r.Intervals[0].Milliseconds -eq 10) 'paired local interval'
 }finally{$m.Dispose()}
}
# T02 catches integer overflow, excess emissions or budget refunds.
Case T02 {
 $s=New-SIState $token $hash ([long]::MaxValue);$bytes=ConvertTo-SIRecord $s task.returned ([long]::MaxValue)
 Assert ($bytes.Length -le 384 -and $bytes[-1] -eq 10 -and $s.Seq -eq 0) 'pure max integer encoding'
 $s=New-SIState $token $hash 1000;$m=[IO.MemoryStream]::new()
 try{foreach($p in $phases){Assert (Write-SIRecord $s $m $p 10) 'normal write'};Assert (Write-SIRecord $s $m error 10) 'fifth terminal error';$n=$m.Length
 Assert (-not(Write-SIRecord $s $m error 10)) 'sixth refused';Assert ($m.Length -eq $n -and $s.Seq -eq 5 -and $s.Bytes -le 1920) 'bounded'}finally{$m.Dispose()}
}
# T03 catches retry after a real stream write failure.
Case T03 {
 $s=New-SIState $token $hash 1000;$m=[IO.MemoryStream]::new([byte[]]@(0),$false)
 try{Assert (-not(Write-SIRecord $s $m task.returned 10)) 'real write failure';Assert ($s.Failed -and $s.Seq -eq 1 -and $s.Bytes -gt 0) 'admission retained';$n=$s.Bytes
 Assert (-not(Write-SIRecord $s $m error 11)) 'latched';Assert ($s.Seq -eq 1 -and $s.Bytes -eq $n -and $m.ToArray()[0] -eq 0) 'no retry mutation'}finally{$m.Dispose()}
}
# T04 catches permissive/coercive identity, key, order and numeric parsing.
Case T04 {
 $one=Fixture 1;$two=Fixture 2
 foreach($bad in @(
 $one.Replace('writer-staged-timing','writer-timing'),$two.Replace('role.save.begin','pins.begin'),
 $one.Replace($token,'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee'),$one.Replace($hash,('B'*64)),
 $one.Replace('"Schema":1','"Schema":2'),$one.Replace('"Schema":1','"Schema":1,"Schema":1'),
 $one.Replace('"Schema":1','"Schema":1,"Extra":0'),$one.Replace('"Schema":1,',''),
 $one.Replace('"Seq":1','"Seq":"1"'),$one.Replace('"Seq":1','"Seq":true'),$one.Replace('"Seq":1','"Seq":1.0'),
 $one.Replace('"Seq":1','"Seq":2'),($one+$one),$two.Replace('"ElapsedTicks":20','"ElapsedTicks":0'),
 $one.Replace('"Frequency":1000','"Frequency":0'))){Invalid $bad}
 $lines=$two.Split([char]10);Invalid ($lines[0]+[char]10+$lines[1].Replace('"Frequency":1000','"Frequency":1001')+[char]10)
}
Case T05 {
 Assert ((Parse '').Status -ceq 'PARTIAL') 'empty'
 $tail='{"Schema":';$r=Parse ((Fixture 2)+$tail)
 Assert ($r.Status -ceq 'PARTIAL' -and $r.Records.Count -eq 2 -and $r.RawTailBase64 -ceq [Convert]::ToBase64String($utf8.GetBytes($tail))) 'retain tail'
 Assert ((Parse (Fixture 3)).Status -ceq 'PARTIAL') 'missing final phase'
}
Case T06 {
 Assert ((Read-SITrace ([byte[]]@(255,10)) $token $hash).Status -ceq 'INVALID') 'invalid UTF8'
 Invalid ((Fixture 1).Replace("`n","`r`n"));Invalid ((' '*384)+"`n");Invalid (' '*1921)
}
Case T07 {
 $error=(Fixture 1).Replace('task.returned','error');Assert ((Parse $error).Status -ceq 'ERROR') 'terminal error'
 Invalid ($error+(Fixture 1));Assert ((Parse ((Fixture)+'x')).Status -ceq 'PARTIAL') 'tail prevents complete'
}
# Gate tests catch acceptance with any absent, coerced or false prerequisite.
Case G01 {$r=Test-SIObservation (GoodFacts);Assert ($r.Passed -and $r.Reasons.Count -eq 0) 'all verified'}
Case G02 {
 foreach($key in @((GoodFacts).Keys)){
  foreach($bad in @($false,'true',1,$null)){$facts=GoodFacts;$facts[$key]=$bad;Assert (-not(Test-SIObservation $facts).Passed) "reject $key"}
  $facts=GoodFacts;$facts.Remove($key);Assert (-not(Test-SIObservation $facts).Passed) "missing $key"
 }
}
Case G03 {foreach($key in @('AuditIdentityVerified','AuditMapsVerified','WriterDelegationVerified')){$facts=GoodFacts;$facts[$key]=$false;Assert (-not(Test-SIObservation $facts).Passed) 'stale/invalid evidence'}}
Case G04 {foreach($key in @('RoleVerified','TimingVerified')){$facts=GoodFacts;$facts[$key]=$false;Assert (-not(Test-SIObservation $facts).Passed) 'late/incomplete failure'}}
Assert ($script:cases -eq 11) 'eleven cases'
[Console]::Out.WriteLine('STAGED_INTEGRITY_MEMORY_VERIFIED')
[ordered]@{Cases=$script:cases;Assertions=$script:assertions;NoChildProcesses=$true;NoCompilation=$true}|ConvertTo-Json -Compress
