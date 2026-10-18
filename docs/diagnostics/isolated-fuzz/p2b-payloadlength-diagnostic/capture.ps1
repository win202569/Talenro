# One approved ordinary expression-only diagnostic host; no retry.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-payloadlength-diagnostic'
if(Test-Path -LiteralPath $packet){throw 'Diagnostic already consumed; no replay'}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-control-codec'
$f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
if((Get-FileHash -LiteralPath (Join-Path $old 'invocations.json')).Hash -cne '6674962585BF6A4FFF80399283D765930FC17659D21A6E5FB7AE6A7DF591B14E'){throw 'Original index drift'}
foreach($pin in $f.Pins){if((Get-FileHash -LiteralPath $pin.VerificationPath).Hash -cne $pin.SHA256){throw 'Frozen pin drift'}}
if((Get-FileHash -LiteralPath $f.Engine).Hash -cne $f.EngineSHA256){throw 'Engine drift'}
if((Get-FileHash -LiteralPath $f.Subject).Hash -cne 'FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6'){throw 'Module drift'}
$target=Join-Path $PSScriptRoot 'expression.ps1'
$null=New-Item -ItemType Directory -Path $packet
$receipt=[ordered]@{Schema=1;Consumed=1;Maximum=1;Scope='expression-only-no-product';TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;ControllerSHA256=(Get-FileHash -LiteralPath $PSCommandPath).Hash;EngineSHA256=$f.EngineSHA256;OriginalAllowance='2/5 consumed;3 frozen;review0/1 ineligible';Launched=$false;TimedOut=$false;ExitCode=$null;StreamsComplete=$false;OwnedExit=$false;FormalAcceptance=$false}
function Save-Receipt{[IO.File]::WriteAllText((Join-Path $packet 'receipt.json'),($receipt|ConvertTo-Json),[Text.UTF8Encoding]::new($false))}
Save-Receipt
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$f.Engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target)){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet 'stdout.bin'));$stderr=[IO.File]::Create((Join-Path $packet 'stderr.bin'))
try{
    $null=$p.Start();$receipt.Launched=$true;$receipt.PID=$p.Id;Save-Receipt
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit(10000)){$receipt.TimedOut=$true;$p.Kill()}
    $receipt.OwnedExit=$p.WaitForExit(2000)
    if($receipt.OwnedExit){$receipt.ExitCode=$p.ExitCode}
    $receipt.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),2000)
}finally{
    $stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Receipt
}
$receipt.stdoutSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'stdout.bin')).Hash
$receipt.stderrSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'stderr.bin')).Hash
Save-Receipt
Get-Content -LiteralPath (Join-Path $packet 'receipt.json')
Get-Content -LiteralPath (Join-Path $packet 'stdout.bin')
if($receipt.TimedOut -or -not $receipt.OwnedExit -or -not $receipt.StreamsComplete -or $receipt.ExitCode -ne 0 -or (Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length -ne 0){throw 'Diagnostic incomplete; no retry'}
