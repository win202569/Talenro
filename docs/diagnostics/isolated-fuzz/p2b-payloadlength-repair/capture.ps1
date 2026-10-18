# Single separately approved full host for the one-expression test candidate.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-payloadlength-repair'
if(Test-Path -LiteralPath $packet){throw 'One-host allowance already consumed; no replay'}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
$old=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-control-codec'
$f=Get-Content -LiteralPath (Join-Path $old 'frozen.json') -Raw|ConvertFrom-Json
function Assert-Hash($Path,$Hash){if((Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
Assert-Hash (Join-Path $old 'frozen.json') 'C240C986D107AF7C6E15C9543936499D43E547CC123159BFCB0351407021A77A'
Assert-Hash (Join-Path $old 'invocations.json') '6674962585BF6A4FFF80399283D765930FC17659D21A6E5FB7AE6A7DF591B14E'
Assert-Hash $f.Subject 'FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6'
Assert-Hash $f.Engine $f.EngineSHA256
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$prior=@(Get-Content -LiteralPath (Join-Path $old 'invocations.json') -Raw|ConvertFrom-Json)
foreach($row in $prior){foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $old ($row.Slot+'.'+$stream+'.bin')) $row.($stream+'SHA256')}}
$original=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-control-codec'
$target=Join-Path $PSScriptRoot 'control-frame.tests.ps1'
$oldText=[IO.File]::ReadAllText((Join-Path $original 'control-frame.tests.ps1'))
if($oldText.Split('@($len-1,$len+1,113)',[StringSplitOptions]::None).Count -ne 2 -or
   [IO.File]::ReadAllText($target) -cne $oldText.Replace('@($len-1,$len+1,113)','@(($len-1),($len+1),113)')){throw 'Exact one-expression delta required'}
foreach($name in @('control-frame.ps1','fixtures.json','expected-output.json')){Assert-Hash (Join-Path $PSScriptRoot $name) (Get-FileHash -LiteralPath (Join-Path $original $name)).Hash}
$pins=@($f.Pins)
foreach($dir in @($old,(Join-Path $root '.superpowers/sdd/2026-10-18-p2b-payloadlength-diagnostic'),$PSScriptRoot)){
    foreach($file in Get-ChildItem -LiteralPath $dir -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
}
$oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
$expected=@($oracle.full|ForEach-Object{if($_ -ceq '{Context}'){'CONTEXT: '+$identity.Name+' admin=False'}else{$_}})
if($expected.Count -ne 35){throw 'Original full oracle must have35lines'}
$null=New-Item -ItemType Directory -Path $packet
[IO.File]::WriteAllText((Join-Path $packet 'frozen.json'),([pscustomobject]@{Pins=$pins;Expected=$expected;MaxHosts=1;RuntimeApproved=$false}|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
$receipt=[ordered]@{Schema=1;Consumed=1;Maximum=1;Scope='test-only-one-expression-full';TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;ModuleSHA256=(Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'control-frame.ps1')).Hash;EngineSHA256=$f.EngineSHA256;FrozenSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'frozen.json')).Hash;PinCount=$pins.Count;OriginalAllowance='2/5 consumed;3 frozen;review0/1 ineligible';Launched=$false;TimedOut=$false;ExitCode=$null;StreamsComplete=$false;OwnedExit=$false;PostPinsIntact=$false;ExactLines=$false;Outcome='INCOMPLETE';FormalAcceptance=$false}
function Save-Receipt{[IO.File]::WriteAllText((Join-Path $packet 'receipt.json'),($receipt|ConvertTo-Json),[Text.UTF8Encoding]::new($false))}
Save-Receipt
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$f.Engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target,'-Group','all')){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet 'stdout.bin'));$stderr=[IO.File]::Create((Join-Path $packet 'stderr.bin'))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$receipt.Launched=$true;$receipt.PID=$p.Id;Save-Receipt
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit(60000)){$receipt.TimedOut=$true;$p.Kill()}
    $receipt.OwnedExit=$p.WaitForExit(2000)
    if($receipt.OwnedExit){$receipt.ExitCode=$p.ExitCode}
    $receipt.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),2000)
}finally{
    $timer.Stop();$receipt.ElapsedMs=$timer.ElapsedMilliseconds
    $stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Receipt
}
foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256};$receipt.PostPinsIntact=$true
$lines=@([IO.File]::ReadAllText((Join-Path $packet 'stdout.bin')).Replace("`r`n","`n").TrimEnd("`n").Split("`n"))
$equal=$lines.Count -eq $expected.Count
if($equal){for($i=0;$i -lt $expected.Count;$i++){if($lines[$i] -cne $expected[$i]){$equal=$false;break}}}
$receipt.ExactLines=$equal;$receipt.LineCount=$lines.Count
$receipt.stdoutSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'stdout.bin')).Hash
$receipt.stderrSHA256=(Get-FileHash -LiteralPath (Join-Path $packet 'stderr.bin')).Hash
$receipt.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length
if(-not $receipt.TimedOut -and $receipt.OwnedExit -and $receipt.StreamsComplete -and $receipt.PostPinsIntact -and $receipt.ExitCode -eq 0 -and $receipt.ExactLines -and $receipt.StderrBytes -eq 0){$receipt.Outcome='EXPECTED_WITHIN_SCOPE'}
Save-Receipt
Get-Content -LiteralPath (Join-Path $packet 'receipt.json')
Get-Content -LiteralPath (Join-Path $packet 'stdout.bin')
if($receipt.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result; stop with no repair/retry/reviewer'}
