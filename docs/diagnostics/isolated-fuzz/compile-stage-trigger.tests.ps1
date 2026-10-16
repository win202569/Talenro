$ErrorActionPreference='Stop'
function Check-Stage([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$source=Join-Path $PSScriptRoot 'compile-stage-trigger.cs'
Check-Stage (Test-Path $source) 'CompileStageTrigger exists'
Add-Type -Path @((Join-Path $PSScriptRoot 'retained-lifetime.cs'),$source,(Join-Path $PSScriptRoot 'compile-stage-session.cs'),(Join-Path $PSScriptRoot 'retained-lifetime-fixture.cs'),(Join-Path $PSScriptRoot 'compile-stage-fixture.cs'))
$compiler='C:/Fixed SDK/pkg/tool/windows_amd64/compile.exe';$directory='C:/Repo Space/internal/nodecontrol/nodeleafbinding'
$line='"C:/Fixed SDK/pkg/tool/windows_amd64/compile.exe" -o "$WORK/b127/_pkg_.a" -p talenro.local/platform/internal/nodecontrol/nodeleafbinding -pack "C:/Repo Space/internal/nodecontrol/nodeleafbinding/a.go" "C:/Repo Space/internal/nodecontrol/nodeleafbinding/a_test.go"'
function New-Stage([string]$Text=''){$b=New-Object FakeCompileStageBytes;$c=New-Object FakeCompileStageClock;$b.Append([Text.Encoding]::UTF8.GetBytes($Text));[pscustomobject]@{Bytes=$b;Clock=$c;Trigger=(New-Object CompileStageTrigger($b,$c,$compiler,$directory))}}
function Stage-Hash([byte[]]$Bytes){$sha=[Security.Cryptography.SHA256]::Create();try{([BitConverter]::ToString($sha.ComputeHash($Bytes))).Replace('-','')}finally{$sha.Dispose()}}
function Stage-Root($s,[uint32]$Status=258){$r=New-Object CompileStageRootCheck;$r.Succeeded=$true;$r.WaitResult=$Status;$r.Frequency=1000;$r.StartTimestamp=$s.Clock.Now;$r.EndTimestamp=$s.Clock.Now;$r.StartUTC=$s.Clock.UtcNow;$r.EndUTC=$s.Clock.UtcNow;return $r}
$s=New-Stage $line;$s.Trigger.Observe(100000)
Check-Stage (-not $s.Trigger.HasCandidate) 'FragmentedLine before LF'
$s.Bytes.Append([byte[]]@(10));$s.Trigger.Observe(100000)
Check-Stage ($s.Trigger.HasCandidate -and $s.Trigger.Report.State -ceq 'Waiting') 'FragmentedLine complete candidate not activation'
Check-Stage ($s.Trigger.ResolveCandidate((Stage-Root $s),100000)) 'unsignaled activation'
Check-Stage (-not $s.Trigger.ResolveCandidate((Stage-Root $s),100000)) 'activate once'
$calls=$s.Bytes.ReadCalls;$s.Trigger.Observe(100000);Check-Stage ($s.Bytes.ReadCalls -eq $calls) 'no scanning after activation'
$s.Trigger.Dispose();$s.Trigger.Dispose();Check-Stage ($s.Bytes.DisposeCalls -eq 1) 'dispose once'
$raw="noise`r`n"+$line+"`r`n";$s=New-Stage ($raw+"suffix`n");$s.Trigger.Observe(100000);$r=$s.Trigger.Report
Check-Stage ($s.Trigger.HasCandidate -and $r.LineStart -eq 7 -and $r.LineEnd -eq [Text.Encoding]::UTF8.GetByteCount($raw) -and $r.LineNumber -eq 2) 'PrefixStopsAtMatch offsets'
Check-Stage ($r.PrefixSHA256 -ceq (Stage-Hash ([Text.Encoding]::UTF8.GetBytes($raw)))) 'PrefixStopsAtMatch digest'
Check-Stage ($r.LineSHA256 -ceq (Stage-Hash ([Text.Encoding]::UTF8.GetBytes($line+"`r`n"))) -and [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($r.RawLineBase64)) -ceq ($line+"`r`n")) 'raw line binding CRLF'
$s.Trigger.Dispose()
$bad=@($line.Replace(' -p ',' -p wrong -p '),$line.Replace('nodeleafbinding -pack','nodeleafbindingextra -pack'),$line.Replace('Fixed SDK','Other SDK'),$line.Replace(' -pack ',' -pack -pack '),$line.Replace('a_test.go','a.go'),$line.Replace(' -pack ',' -pack -o '),$line.Replace('a_test.go','../nodeleafbinding/a_test.go'),$line.Replace(' -pack ',' -pack "text" '),'vet {"Package":"nodeleafbinding"}', $line.Replace('compile.exe','link.exe'),$line.Replace(' -pack ',' -o "'+$directory+'/fake_test.go" -pack ').Replace('a_test.go','a.go'))
$i=0;foreach($text in $bad){$s=New-Stage ($text+"`n");$s.Trigger.Observe(100000);Check-Stage (-not $s.Trigger.HasCandidate) ('ExactGrammar negative '+(++$i));$s.Trigger.Dispose()}
$escaped=$line.Replace($compiler,$compiler.Replace('/',([string][char]92)*2)).Replace($directory,$directory.Replace('/',([string][char]92)*2));$s=New-Stage ($escaped+"`n");$s.Trigger.Observe(100000);Check-Stage $s.Trigger.HasCandidate 'quoted escaped Windows backslashes';$s.Trigger.Dispose()
$s=New-Stage ($line.Replace('a_test.go',([char]0x4e2d)+'_test.go')+"`n");$s.Bytes.MaxRead=1
for($i=0;$i -lt 500 -and -not $s.Trigger.HasCandidate;$i++){$s.Trigger.Observe(100000)}
Check-Stage $s.Trigger.HasCandidate 'short reads split UTF8';$s.Trigger.Dispose()
foreach($text in @($line.Replace('a_test.go','a\q_test.go'),$line.Substring(0,$line.Length-1))){$s=New-Stage ($text+"`n");$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed') 'malformed compiler candidate fails';$s.Trigger.Dispose()}
$s=New-Stage;$s.Bytes.Append([Text.Encoding]::UTF8.GetBytes($line.Substring(0,$line.Length-1)));$s.Bytes.Append([byte[]]@(255,34,10));$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed') 'invalid candidate UTF8';$s.Trigger.Dispose()
$complete=$line+"`n";$n=[Text.Encoding]::UTF8.GetByteCount($complete)
$s=New-Stage ((' '* (65536-$n))+$complete);$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.HasCandidate -and $s.Trigger.Report.LineEnd -eq 65536) 'ExactLimitCandidate line bound';$s.Trigger.Dispose()
$s=New-Stage ('x'*65536);$s.Trigger.Observe(100000);$calls=$s.Bytes.ReadCalls;$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed' -and $s.Bytes.ReadCalls -eq $calls) 'line bound no LF stops without extra read';$s.Trigger.Dispose()
$s=New-Stage;$s.Bytes.Append([FakeCompileStageBytes]::Padding(16777216-$n));$s.Bytes.Append([Text.Encoding]::UTF8.GetBytes($complete))
for($i=0;$i -lt 256;$i++){$s.Trigger.Observe(100000)}
Check-Stage ($s.Trigger.HasCandidate -and $s.Trigger.Report.LineEnd -eq 16777216 -and $s.Trigger.Report.PollCount -eq 256) 'ExactLimitCandidate total bound';$s.Trigger.Dispose()
$s=New-Stage;$s.Bytes.Append([FakeCompileStageBytes]::Padding(16777216));for($i=0;$i -lt 256;$i++){$s.Trigger.Observe(100000)};$calls=$s.Bytes.ReadCalls;$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed' -and $s.Bytes.ReadCalls -eq $calls) 'total bound no match';$s.Trigger.Dispose()
$s=New-Stage;for($i=0;$i -lt 60000;$i++){$s.Trigger.Observe(100000)};$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed' -and $s.Trigger.Report.PollCount -eq 60000 -and $s.Bytes.ReadCalls -eq 0) 'poll bound including empty EOF';$s.Trigger.Dispose()
$s=New-Stage "hello`n";$s.Trigger.Observe(100000);$s.Bytes.VisibleLength=1;$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed') 'length regression';$s.Trigger.Dispose()
$s=New-Stage $complete;$s.Bytes.ThrowRead=$true;$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed') 'read exception';$s.Trigger.Dispose()
$s=New-Stage;$s.Clock.Now=999;$s.Trigger.Observe(100000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed') 'backwards clock';$s.Trigger.Dispose()
$s=New-Stage $complete;$s.Trigger.Observe(1000);Check-Stage ($s.Trigger.Report.State -ceq 'Failed' -and $s.Bytes.ReadCalls -eq 0) 'expired deadline no read';$s.Trigger.Dispose()
$s=New-Stage $complete;$s.Trigger.Observe(100000);$null=$s.Trigger.ResolveCandidate((Stage-Root $s 0),100000);Check-Stage ($s.Trigger.Report.State -ceq 'RootExitedWithoutActivation') 'signaled root cannot activate';$s.Trigger.Dispose()
$s=New-Stage $line;$s.Trigger.RootExited();$s.Bytes.Append([byte[]]@(10));$s.Trigger.Observe(100000);Check-Stage (-not $s.Trigger.HasCandidate -and $s.Bytes.ReadCalls -eq 0) 'no retroactive activation';$s.Trigger.Dispose()
$s=New-Stage;for($i=0;$i -lt 20;$i++){$s.Trigger.Fail('z'*600)};Check-Stage ($s.Trigger.Report.Errors.Count -eq 16 -and @($s.Trigger.Report.Errors|Where-Object {$_.Length -ne 512}).Count -eq 0) 'bounded errors';$s.Bytes.ThrowDispose=$true;$s.Trigger.Dispose();Check-Stage ($s.Trigger.Report.Errors.Count -eq 16) 'dispose failure latched without unbounded errors'
$temporary=Join-Path ([IO.Path]::GetTempPath()) ('stage-reader-'+[guid]::NewGuid().ToString('N')+'.txt')
try{[IO.File]::WriteAllBytes($temporary,[byte[]]@(65));$reader=New-Object FileCompileStageByteSource($temporary);$buffer=New-Object byte[] 2;Check-Stage ($reader.Read($buffer,0,2) -eq 1 -and $buffer[0] -eq 65) 'production reader actual bytes';$writer=[IO.File]::Open($temporary,[IO.FileMode]::Open,[IO.FileAccess]::Write,[IO.FileShare]::ReadWrite);$writer.Position=1;$writer.WriteByte(66);$writer.Dispose();Check-Stage ($reader.Length -eq 2 -and $reader.Read($buffer,0,2) -eq 1 -and $buffer[0] -eq 66) 'production reader append visibility';$reader.Dispose();$reader.Dispose()}finally{if(Test-Path $temporary){Remove-Item -LiteralPath $temporary}}
