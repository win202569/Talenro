Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-wrapper-reference-diagnostic'
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-sdk-compatibility'
if(Test-Path -LiteralPath $packet){throw 'One host consumed; no replay'}
function Assert-Hash($Path,$Hash){if($Hash -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash (Join-Path $old 'frozen.json') '204E6E0652F8D12E97E51CAD23B6688B9392F15A69D0D5608B3BFE7A465E700E'
Assert-Hash (Join-Path $old 'invocations.json') 'FB1405516977527148C73A738187D1311F53A88AF3EF36B3EEFF035F63686C72'
$f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
if($f.Pins.Count -ne 1078){throw 'Inherited inventory'}
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$prior=@(Get-Content -LiteralPath (Join-Path $old 'invocations.json') -Raw|ConvertFrom-Json)
if($prior.Count -ne 1 -or $prior[0].Slot -cne 'red' -or $prior[0].Outcome -cne 'INCOMPLETE'){throw 'Preserve stopped prior scope'}
foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $old ('red.'+$stream+'.bin')) $prior[0].($stream+'SHA256')}
$subject=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-scalar-sdk-compatibility/control-frame.ps1'
$subjectHash='B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
$pins=@($f.Pins)
foreach($dir in @($old,$PSScriptRoot)){foreach($file in Get-ChildItem -LiteralPath $dir -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}}
foreach($name in @('2026-10-18-p2b-scalar-wrapper-readonly.md','2026-10-18-p2b-scalar-sdk-compatibility.md')){
    $path=Join-Path $root ('docs/roadmap/'+$name);$pins+=,[pscustomobject]@{VerificationPath=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash}
}
$pins+=,[pscustomobject]@{VerificationPath=$subject;SHA256=$subjectHash;Role='live-baseline-unchanged'}
if($pins.Count -ne 1090){throw '1090 mapped pins required'}
$target=Join-Path $PSScriptRoot 'diagnostic.ps1'
$null=New-Item -ItemType Directory -Path $packet
Save-Json (Join-Path $packet 'frozen.json') ([pscustomobject]@{Schema=1;Pins=$pins;BaseCommit='f3890d526228e3147a7dcd51f71cc9134e94cd00';MaxHosts=1;MaxMs=60000;ExitWaitMs=2000;DrainMs=2000;ExpectedCases=12;Scope='controlled-sdk-reference-evidence-only';RuntimeApproved=$false})
$receipt=[ordered]@{Schema=1;Consumed=1;Maximum=1;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;EngineSHA256=$engineHash;FrozenSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'frozen.json')).Hash;PinCount=$pins.Count;Launched=$false;PID=$null;TimedOut=$false;ExitCode=$null;ElapsedMs=$null;StreamsComplete=$false;OwnedExit=$false;PostPinsIntact=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;FormalAcceptance=$false}
$receiptPath=Join-Path $packet 'receipt.json';Save-Json $receiptPath $receipt
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target)){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet 'stdout.bin'));$stderr=[IO.File]::Create((Join-Path $packet 'stderr.bin'));$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$receipt.Launched=$true;$receipt.PID=$p.Id;Save-Json $receiptPath $receipt
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit(60000)){$receipt.TimedOut=$true;$p.Kill()}
    $receipt.OwnedExit=$p.WaitForExit(2000);if($receipt.OwnedExit){$receipt.ExitCode=$p.ExitCode}
    $receipt.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),2000)
}finally{$timer.Stop();$receipt.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $receiptPath $receipt}
foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash;$receipt.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$receipt.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($stream+'.bin'))).Hash}
Save-Json $receiptPath $receipt
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'stdout.bin')))
$d=$raw|ConvertFrom-Json
$complete=$d.Schema -eq 1 -and $d.Mode -ceq 'p2b-wrapper-reference-diagnostic' -and $d.Captured -and $d.Stage -ceq 'Captured' -and $null -eq $d.Error -and $d.HostVersion -ceq '7.6.5' -and $d.Core -and $d.X64 -and -not $d.Admin -and $d.Cases.Count -eq 12 -and -not $d.ProductImported -and -not $d.ProductCalled -and -not $d.FixtureCompiled -and -not $d.RuntimeApproved
foreach($case in $d.Cases){$complete=$complete -and $case.Complete -and $case.Phases.Count -eq 3;foreach($phase in $case.Phases){$complete=$complete -and $phase.Complete}}
if(-not $receipt.TimedOut -and $receipt.OwnedExit -and $receipt.StreamsComplete -and $receipt.PostPinsIntact -and $receipt.ExitCode -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length -eq 0 -and $complete){$receipt.Outcome='DIAGNOSTIC_CAPTURED'}
Save-Json $receiptPath $receipt;$receipt|ConvertTo-Json
[pscustomobject]@{Stage=$d.Stage;Cases=$d.Cases.Count;References=$d.References.Count;Error=$d.Error}|ConvertTo-Json -Depth 12
if($receipt.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Diagnostic incomplete; no retry'}
