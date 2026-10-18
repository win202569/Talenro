# Exactly one separately approved ordinary diagnostic host; no retry.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-sequence-binding-diagnostic'
if(Test-Path -LiteralPath $packet){throw 'Diagnostic consumed; no replay'}
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-passive-descriptor-repair'
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
function Assert-Hash($Path,$Hash){if((Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
Assert-Hash (Join-Path $old 'frozen.json') '89A67BB759A52B121CBCB7493E8C570FD116A433FD6432181B900CDC96AE301C'
Assert-Hash (Join-Path $old 'invocations.json') '6D35BDE10C3559668564C06830F4C034947AF20CA30B191D6FA6FF9C12253ED6'
$f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$oldRow=Get-Content -LiteralPath (Join-Path $old 'invocations.json') -Raw|ConvertFrom-Json
if(@($oldRow).Count -ne 3 -or $oldRow[2].Outcome -cne 'INCOMPLETE'){throw 'Expected stopped prior scope'}
foreach($row in $oldRow){
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $old ($row.Slot+'.'+$stream+'.bin')) $row.($stream+'SHA256')}
    Assert-Hash (Join-Path $old ($row.Slot+'.trace.json')) $row.TraceSHA256
}
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139';Assert-Hash $engine $engineHash
$subject=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-passive-descriptor-repair/control-frame.ps1';Assert-Hash $subject 'F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0'
if($f.Pins.Count -ne 963){throw 'Inherited inventory'}
$pins=@($f.Pins)
$pins+=,[pscustomobject]@{VerificationPath=$subject;SHA256='F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0';Role='unchanged-live-candidate'}
foreach($dir in @($old,$PSScriptRoot)){foreach($file in Get-ChildItem -LiteralPath $dir -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}}
if($pins.Count -ne 981){throw '981mapped pins required'}
$target=Join-Path $PSScriptRoot 'diagnostic.ps1'
$null=New-Item -ItemType Directory -Path $packet
Save-Json (Join-Path $packet 'frozen.json') ([pscustomobject]@{Schema=1;Pins=$pins;MaxHosts=1;BaseCommit='5431f76019c37dfce9b3dbdb31eb501e007e9fdd';MaxMs=10000;ExitWaitMs=2000;DrainMs=2000;Scope='sequence-binding-evidence-only';RuntimeApproved=$false})
$receipt=[ordered]@{Schema=1;Consumed=1;Maximum=1;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;EngineSHA256=$engineHash;FrozenSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'frozen.json')).Hash;PinCount=$pins.Count;Launched=$false;PID=$null;TimedOut=$false;ExitCode=$null;ElapsedMs=$null;StreamsComplete=$false;OwnedExit=$false;PostPinsIntact=$false;Outcome='INCOMPLETE';stdoutSHA256=$null;stderrSHA256=$null;FormalAcceptance=$false}
$receiptPath=Join-Path $packet 'receipt.json';Save-Json $receiptPath $receipt
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target)){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet 'stdout.bin'));$stderr=[IO.File]::Create((Join-Path $packet 'stderr.bin'));$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$receipt.Launched=$true;$receipt.PID=$p.Id;Save-Json $receiptPath $receipt
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit(10000)){$receipt.TimedOut=$true;$p.Kill()}
    $receipt.OwnedExit=$p.WaitForExit(2000);if($receipt.OwnedExit){$receipt.ExitCode=$p.ExitCode}
    $receipt.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),2000)
}finally{$timer.Stop();$receipt.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $receiptPath $receipt}
foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256};Assert-Hash $engine $engineHash;Assert-Hash $subject 'F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0';$receipt.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$receipt.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($stream+'.bin'))).Hash}
Save-Json $receiptPath $receipt
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'stdout.bin')))
$diagnostic=$raw|ConvertFrom-Json
if(-not $receipt.TimedOut -and $receipt.OwnedExit -and $receipt.StreamsComplete -and $receipt.PostPinsIntact -and $receipt.ExitCode -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length -eq 0 -and $diagnostic.Schema -eq 1 -and $diagnostic.Mode -ceq 'p2b-sequence-binding-diagnostic-only' -and $diagnostic.Captured -and -not $diagnostic.Admin -and $diagnostic.ProductImported -and $diagnostic.ProductCalled -and -not $diagnostic.CompilePerformed -and -not $diagnostic.FullSuiteRun -and $null -eq $diagnostic.Error -and $diagnostic.Stage -ceq 'Captured' -and -not $diagnostic.RuntimeApproved){$receipt.Outcome='DIAGNOSTIC_CAPTURED'}
Save-Json $receiptPath $receipt;$receipt|ConvertTo-Json;[Console]::Out.Write($raw)
if($receipt.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Diagnostic incomplete; no retry'}
