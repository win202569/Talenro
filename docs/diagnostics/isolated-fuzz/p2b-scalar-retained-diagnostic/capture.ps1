# One fresh retained-descriptor diagnostic host; no old slot reuse.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-retained-diagnostic'
if(Test-Path -LiteralPath $packet){throw 'Diagnostic consumed; no replay'}
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-scalar-decoration-diagnostic'
function Assert-Hash($Path,$Hash){if($Hash -cnotmatch '^[A-F0-9]{64}$' -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin mismatch: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash (Join-Path $old 'frozen.json') '7EDB4EFB27D168B544268904645B5C7736DF64C91B7D1B780C87DF03B05C2D76'
Assert-Hash (Join-Path $old 'receipt.json') 'F9CA78FA3F78BEFFCB7B446613386AE454F99C3B9F226819CEEC08FD8F226024'
$f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
if($f.Pins.Count -ne 1020){throw 'Inherited inventory'}
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$prior=Get-Content -LiteralPath (Join-Path $old 'receipt.json') -Raw|ConvertFrom-Json
if($prior.Outcome -cne 'INCOMPLETE' -or $prior.Consumed -ne 1){throw 'Expected failed prior scope'}
foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $old ($stream+'.bin')) $prior.($stream+'SHA256')}
$subject=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-sdk-structural-compatibility/control-frame.ps1'
$subjectHash='B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
$pins=@($f.Pins)
foreach($dir in @($old,$PSScriptRoot)){foreach($file in Get-ChildItem -LiteralPath $dir -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}}
$reportPath=Join-Path $root 'docs/roadmap/2026-10-18-p2b-scalar-decoration-diagnostic.md'
$pins+=,[pscustomobject]@{VerificationPath=$reportPath;SHA256=(Get-FileHash -LiteralPath $reportPath).Hash;Role='prior-failure-explanation'}
if($pins.Count -ne 1030){throw '1030mapped pins required'}
$target=Join-Path $PSScriptRoot 'diagnostic.ps1'
$null=New-Item -ItemType Directory -Path $packet
Save-Json (Join-Path $packet 'frozen.json') ([pscustomobject]@{Schema=1;Pins=$pins;BaseCommit='e66fe5059f326402e1f6e3032d476f58adfda1c6';MaxHosts=1;MaxMs=10000;ExitWaitMs=2000;DrainMs=2000;ExpectedRows=35;Scope='scalar-retained-evidence-only';RuntimeApproved=$false})
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
foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash;$receipt.PostPinsIntact=$true
foreach($stream in @('stdout','stderr')){$receipt.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($stream+'.bin'))).Hash}
Save-Json $receiptPath $receipt
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet 'stdout.bin')))
$diagnostic=$raw|ConvertFrom-Json
if(-not $receipt.TimedOut -and $receipt.OwnedExit -and $receipt.StreamsComplete -and $receipt.PostPinsIntact -and $receipt.ExitCode -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length -eq 0 -and $diagnostic.Schema -eq 1 -and $diagnostic.Mode -ceq 'p2b-scalar-retained-diagnostic-only' -and $diagnostic.Captured -and $diagnostic.FixtureObservable -and @($diagnostic.Rows).Count -eq 35 -and @($diagnostic.Rows|Where-Object{$_.Phase -cne 'Completed'}).Count -eq 0 -and -not $diagnostic.Admin -and $diagnostic.ProductImported -and $diagnostic.ProductCalled -and $diagnostic.CompilePerformed -and -not $diagnostic.FullSuiteRun -and $null -eq $diagnostic.Error -and $diagnostic.Stage -ceq 'Captured' -and -not $diagnostic.RuntimeApproved){$receipt.Outcome='DIAGNOSTIC_CAPTURED'}
Save-Json $receiptPath $receipt;$receipt|ConvertTo-Json
[pscustomobject]@{Stage=$diagnostic.Stage;Rows=@($diagnostic.Rows).Count;AcceptedDecorated=@($diagnostic.Rows|Where-Object{$_.ContractWouldRequireRefusal -and $null -ne $_.Result -and $_.Result.Accepted}).Count;Error=$diagnostic.Error}|ConvertTo-Json -Depth 10
if($receipt.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Diagnostic incomplete; no retry'}
