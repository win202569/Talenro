param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-first-snapshot-diagnostic'
$previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-bounded-inbox-tick-correction'
$manifest=Join-Path $packet 'frozen.json';$receipt=Join-Path $packet 'invocation.json'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
function Assert-Hash($Path,$Hash){if((Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))}
$id=[Security.Principal.WindowsIdentity]::GetCurrent()
if(([Security.Principal.WindowsPrincipal]::new($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'One-shot packet already exists'}
    Assert-Hash (Join-Path $previous 'frozen.json') '6A8A65225274D7FAE99FF329716DD21024514E56B7534F72F3BAEB560756D5EE'
    Assert-Hash (Join-Path $previous 'invocations.json') '8EB4CF553EBE12B6D6050F7852E07CE5940EA00D35668F3E7AB2C87F269F7216'
    Assert-Hash (Join-Path $previous 'product-seal.json') '9750B67EBC55C598BE80485B7F08522EA0BE122F490057868B7E87BC6AC346E5'
    $f=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($f.Pins.Count -ne 1262){throw 'Prior mapped inventory'}
    $pins=@($f.Pins);foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $extra=@(Get-ChildItem -LiteralPath $previous -File -Force|ForEach-Object{$_.FullName})
    $leaf=@(Get-ChildItem -LiteralPath (Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-bounded-inbox-tick-correction') -File -Force)
    if($leaf.Count -ne 11){throw 'Prior portable inventory'};$extra+=@($leaf|ForEach-Object{$_.FullName})
    $extra+=,(Join-Path $root 'docs/roadmap/2026-10-18-p2b-bounded-inbox-tick-correction.md')
    $inputs=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force);if($inputs.Count -ne 4){throw 'Four diagnostic inputs required'}
    $extra+=@($inputs|ForEach-Object{$_.FullName})
    foreach($path in $extra){$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash;Role='diagnostic-frozen-input'}}
    $null=New-Item -ItemType Directory -Path $packet
    Save-Json $manifest ([pscustomobject]@{Schema=1;BaseCommit='27e6dc45aeb853a989aa919c223b9420e8510153';Pins=$pins;MaxHosts=1;MaxMs=120000;ExitWaitMs=2000;DrainMs=2000;OriginalHostsSpent=2;OriginalRemainingFrozen=3;ReviewerEligible=$false;RuntimeApproved=$false})
    [pscustomobject]@{FrozenSHA256=(Get-FileHash -LiteralPath $manifest).Hash;Pins=$pins.Count}|ConvertTo-Json;return
}
Assert-Hash $manifest $FrozenSHA256;$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
if(Test-Path -LiteralPath $receipt){throw 'Single approved host consumed; no retry'}
$r=[pscustomobject][ordered]@{Schema=1;AttemptConsumed=$true;Launched=$false;PID=$null;ElapsedMs=$null;TimedOut=$false;ExitCode=$null;OwnedExit=$false;StreamsComplete=$false;PostPinsIntact=$false;Captured=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;RuntimeApproved=$false}
Save-Json $receipt $r
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot 'diagnostic.ps1'))){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet 'stdout.bin'));$stderr=[IO.File]::Create((Join-Path $packet 'stderr.bin'))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$r.Launched=$true;$r.PID=$p.Id;Save-Json $receipt $r
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$et=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$r.TimedOut=$true;$p.Kill()}
    $r.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($r.OwnedExit){$r.ExitCode=$p.ExitCode}
    $r.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),$f.DrainMs)
}finally{$timer.Stop();$r.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $receipt $r}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256};$r.PostPinsIntact=$true
$r.stdoutSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'stdout.bin')).Hash;$r.stderrSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'stderr.bin')).Hash;Save-Json $receipt $r
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'stdout.bin')))
$d=$raw|ConvertFrom-Json
$r.Captured=$d.Schema -eq 1 -and $d.Mode -ceq 'p2b-first-snapshot-diagnostic' -and $d.HostVersion -ceq '7.6.5' -and $d.Core -and $d.X64 -and -not $d.Admin -and $d.Created.Accepted -and $d.Created.Reason -ceq 'created' -and $d.SnapshotCalls -eq 1 -and $d.Captured -and $d.BreakpointRemoved -and $null -eq $d.Error -and -not $d.ProductChanged -and -not $d.FullSuiteRun -and -not $d.RuntimeApproved
if($r.ExitCode -eq 0 -and -not $r.TimedOut -and $r.OwnedExit -and $r.StreamsComplete -and $r.PostPinsIntact -and $r.Captured -and (Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length -eq 0){$r.Outcome='DIAGNOSTIC_CAPTURED'}
Save-Json $receipt $r;$r|ConvertTo-Json;[Console]::Out.Write($raw)
if($r.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Diagnostic incomplete; no retry or expanded work'}
