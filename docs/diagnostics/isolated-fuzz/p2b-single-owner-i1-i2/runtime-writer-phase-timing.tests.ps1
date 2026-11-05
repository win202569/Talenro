Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$helper=Join-Path $PSScriptRoot runtime-writer-phase-timing.ps1
if(-not(Test-Path $helper)){throw 'WPT_TEST: missing helper'}
. $helper
$token='11111111-2222-3333-4444-555555555555';$hash='A'*64
$phases=@('task.returned','pins.begin','pins.end','map.save.begin','map.save.end','map.meta.end','role.save.begin','role.save.end','exit.ready')
$utf8=[Text.UTF8Encoding]::new($false,$true);$script:assertions=0;$script:cases=0
function Assert([bool]$Ok,[string]$Why){$script:assertions++;if(-not $Ok){throw "WPT_TEST: $Why"}}
function Case([string]$Name,[scriptblock]$Body){& $Body;$script:cases++;[Console]::Out.WriteLine("PASS: $Name")}
function Fixture([int]$Count=9){
 $text=''
 for($i=0;$i -lt $Count;$i++){$text+=('{"Schema":1,"Token":"'+$token+'","InputSHA256":"'+$hash+'","Role":"writer-timing","Seq":'+($i+1)+',"Phase":"'+$phases[$i]+'","ElapsedTicks":'+(($i+1)*10)+',"Frequency":1000}'+[char]10)}
 return $text
}
function Parse([string]$Text){Read-WPTTrace ($utf8.GetBytes($Text)) $token $hash}
function Invalid([string]$Text){Assert ((Parse $Text).Status -ceq 'INVALID') 'reject malformed record'}
function Throws([scriptblock]$Body){$threw=$false;try{& $Body|Out-Null}catch{$threw=$true};Assert $threw 'must reject'}
Case T01 {
 $s=New-WPTState $token $hash 1000;$m=[IO.MemoryStream]::new()
 try{for($i=0;$i -lt 9;$i++){Assert (Write-WPTRecord $s $m $phases[$i] (($i+1)*10)) 'write'}
 Assert ($utf8.GetString($m.ToArray()) -ceq (Fixture)) 'exact wire bytes'
 $r=Read-WPTTrace $m.ToArray() $token $hash;Assert ($r.Status -ceq 'COMPLETE' -and $r.Records.Count -eq 9) 'complete'
 Assert ($r.Intervals.Count -eq 3 -and $r.Intervals[0].ElapsedTicks -eq 10 -and $r.Intervals[0].Milliseconds -eq 10) 'paired local intervals'
 }finally{$m.Dispose()}
}
Case T02 {$s=New-WPTState $token $hash ([long]::MaxValue);$bytes=ConvertTo-WPTRecord $s task.returned ([long]::MaxValue);Assert ($bytes.Length -le 384 -and $bytes[-1] -eq 10) 'max integer record bound';Assert ($s.Seq -eq 0 -and $s.Bytes -eq 0) 'encode pure'}
Case T03 {
 $s=New-WPTState $token $hash 1000;$m=[IO.MemoryStream]::new()
 try{foreach($p in $phases){Assert (Write-WPTRecord $s $m $p 10) 'normal write'};Assert (Write-WPTRecord $s $m error 10) 'tenth terminal error';$n=$m.Length
 Assert (-not(Write-WPTRecord $s $m error 10)) 'eleventh refused';Assert ($m.Length -eq $n -and $s.Seq -eq 10 -and $s.Bytes -le 3840) 'no extra bytes'}finally{$m.Dispose()}
}
Case T04 {
 Throws {New-WPTState bad $hash 1000};Throws {New-WPTState $token bad 1000};Throws {New-WPTState $token $hash 0}
 $s=New-WPTState $token $hash 1000;Throws {ConvertTo-WPTRecord $s bad 1};Throws {ConvertTo-WPTRecord $s task.returned -1}
}
Case T05 {
 $s=New-WPTState $token $hash 1000;$m=[IO.MemoryStream]::new([byte[]]@(0),$false)
 try{Assert (-not(Write-WPTRecord $s $m task.returned 10)) 'real write failure';Assert ($s.Failed -and $s.Seq -eq 1 -and $s.Bytes -gt 0) 'admission retained';$seq=$s.Seq;$n=$s.Bytes
 Assert (-not(Write-WPTRecord $s $m error 11)) 'latched';Assert ($s.Seq -eq $seq -and $s.Bytes -eq $n -and $m.ToArray()[0] -eq 0) 'no retry mutation'}finally{$m.Dispose()}
}
Case T06 {Assert ((Parse '').Status -ceq 'PARTIAL') 'empty';$r=Parse (Fixture 2);Assert ($r.Status -ceq 'PARTIAL' -and $r.Records.Count -eq 2) 'valid prefix'}
Case T07 {$tail='{"Schema":';$r=Parse ((Fixture 2)+$tail);Assert ($r.Status -ceq 'PARTIAL' -and $r.Records.Count -eq 2 -and $r.RawTailBase64 -ceq [Convert]::ToBase64String($utf8.GetBytes($tail))) 'retain tail'}
Case T08 {
 Assert ((Read-WPTTrace ([byte[]]@(255,10)) $token $hash).Status -ceq 'INVALID') 'invalid UTF8'
 Invalid ((Fixture 1).Replace("`n","`r`n"));Invalid ((' '*384)+"`n");Invalid (' '*3841)
}
Case T09 {
 $one=Fixture 1
 Invalid ($one.Replace('"Schema":1','"Schema":1,"Schema":1'));Invalid ($one.Replace('"Schema":1','"Schema":1,"Extra":0'))
 Invalid ($one.Replace('"Schema":1,',''));Invalid ($one.Replace('"Seq":1','"Seq":"1"'));Invalid ($one.Replace('"Seq":1','"Seq":true'));Invalid ($one.Replace('"Seq":1','"Seq":1.0'))
}
Case T10 {
 $one=Fixture 1
 Invalid ($one.Replace($token,'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee'));Invalid ($one.Replace($hash,('B'*64)));Invalid ($one.Replace('writer-timing','producer'));Invalid ($one.Replace('"Schema":1','"Schema":2'))
}
Case T11 {
 $one=Fixture 1;$two=Fixture 2
 Invalid ($one.Replace('"Seq":1','"Seq":2'));Invalid ($one+$one);Invalid ($one.Replace('task.returned','pins.begin'))
 Invalid ($two.Replace('"ElapsedTicks":20','"ElapsedTicks":0'));Invalid ($two.Replace('"Frequency":1000}','"Frequency":0}'))
 $lines=$two.Split([char]10);Invalid ($lines[0]+[char]10+$lines[1].Replace('"Frequency":1000','"Frequency":1001')+[char]10)
}
Case T12 {
 $error=(Fixture 1).Replace('task.returned','error');$r=Parse $error;Assert ($r.Status -ceq 'ERROR') 'error trace';Invalid ($error+(Fixture 1))
 Assert ((Parse ((Fixture)+'x')).Status -ceq 'PARTIAL') 'tail prevents complete';Assert ((Parse (Fixture 8)).Status -ceq 'PARTIAL') 'missing phase prevents complete'
}
Assert ($script:cases -eq 12) 'twelve cases'
[Console]::Out.WriteLine('WRITER_PHASE_MEMORY_TESTS_VERIFIED')
[ordered]@{Cases=$script:cases;Assertions=$script:assertions;NoChildProcesses=$true;NoCompilation=$true}|ConvertTo-Json -Compress
