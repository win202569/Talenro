# Exactly one separately approved ordinary diagnostic host; no retry.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-rawingress-diagnostic'
if(Test-Path -LiteralPath $packet){throw 'Diagnostic consumed; no replay'}
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-descriptor-ingress'
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
function Assert-Hash($Path,$Hash){if((Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
Assert-Hash (Join-Path $old 'frozen.json') '86CC3D95CA69BDE36670EE6B70A7EE51EE7E4449D638894D29F487EE1521E310'
Assert-Hash (Join-Path $old 'invocations.json') '9F1BCDE92CC7E7D5FA6484CA8E7D3414838A026C4CC09C87C691CA755E18B37B'
$f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$oldRow=Get-Content -LiteralPath (Join-Path $old 'invocations.json') -Raw|ConvertFrom-Json
foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $old ('red.'+$stream+'.bin')) $oldRow.($stream+'SHA256')}
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139';Assert-Hash $engine $engineHash
$subject=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-descriptor-ingress/control-frame.ps1';Assert-Hash $subject 'FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6'
$pins=@($f.Pins)
foreach($dir in @($old,$PSScriptRoot)){foreach($file in Get-ChildItem -LiteralPath $dir -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}}
if($pins.Count -ne 924){throw '924mapped pins required'}
$target=Join-Path $PSScriptRoot 'diagnostic.ps1'
$null=New-Item -ItemType Directory -Path $packet
Save-Json (Join-Path $packet 'frozen.json') ([pscustomobject]@{Schema=1;Pins=$pins;MaxHosts=1;Scope='rawingress-evidence-only';RuntimeApproved=$false})
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
foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256};Assert-Hash $subject 'FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6';$receipt.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$receipt.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($stream+'.bin'))).Hash}
Save-Json $receiptPath $receipt
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'stdout.bin')))
$diagnostic=$raw|ConvertFrom-Json
if(-not $receipt.TimedOut -and $receipt.OwnedExit -and $receipt.StreamsComplete -and $receipt.PostPinsIntact -and $receipt.ExitCode -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length -eq 0 -and $diagnostic.Schema -eq 1 -and $diagnostic.Mode -ceq 'p2b-rawingress-diagnostic-only' -and $diagnostic.Captured -and -not $diagnostic.Admin -and -not $diagnostic.ProductImported -and -not $diagnostic.ProductCalled -and -not $diagnostic.RuntimeApproved){$receipt.Outcome='DIAGNOSTIC_CAPTURED'}
Save-Json $receiptPath $receipt;$receipt|ConvertTo-Json;[Console]::Out.Write($raw)
if($receipt.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Diagnostic incomplete; no retry'}
