param([Parameter(Mandatory)][ValidateSet('Prepare','SealProduct','Run')][string]$Action,
    [ValidateSet('red','focused','full')][string]$Slot,[string]$FrozenSHA256,[string]$ProductSealSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-sdk-structural-compatibility'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json';$seal=Join-Path $packet 'product-seal.json'
$previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-sequence-binding-diagnostic'
$subject=Join-Path $PSScriptRoot 'control-frame.ps1';$baselineCopy=Join-Path $packet 'baseline-product.ps1'
$baselineHash='F76B2DB784F94569C9ABC24AF256AEE2DAFF8C21BF71EF328F12BF2BFA13DBC0'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$sdk=Join-Path ([IO.Path]::GetDirectoryName($engine)) 'System.Management.Automation.dll'
$sdkHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
function Assert-Hash($Path,$Hash){if([string]::IsNullOrEmpty($Hash) -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'New packet required; no replay'}
    Assert-Hash (Join-Path $previous 'frozen.json') '1AF2B26C9E979031D51AADA62C1F3DA1C8925B13F7539D2BB52B717C3C72A9F0'
    Assert-Hash (Join-Path $previous 'receipt.json') 'E2E6DDDD8322FF8110932154F615D206F749E5C516915DDE9D9946B94956ADF5'
    $old=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 981){throw 'Inherited inventory'}
    foreach($pin in $old.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $prior=Get-Content -LiteralPath (Join-Path $previous 'receipt.json') -Raw|ConvertFrom-Json
    if($prior.Outcome -cne 'DIAGNOSTIC_CAPTURED'){throw 'Prior diagnostic incomplete'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $previous ($stream+'.bin')) $prior.($stream+'SHA256')}
    Assert-Hash $subject $baselineHash
    $pins=@($old.Pins)
    $oldSubject=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-passive-descriptor-repair/control-frame.ps1'
    Assert-Hash $oldSubject $baselineHash
    $pins+=,[pscustomobject]@{VerificationPath=$oldSubject;SHA256=$baselineHash;Role='previous-live-module-unchanged'}
    foreach($file in Get-ChildItem -LiteralPath $previous -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    $research=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-sdk-metadata-investigation'
    Assert-Hash (Join-Path $research 'sdk-metadata.json') 'D6F18F947481CAE3C18B9870E3C6822114301254FF6382A60802441D21087E8B'
    foreach($file in Get-ChildItem -LiteralPath $research -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}

    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force|Where-Object{$_.Name -cne 'control-frame.ps1'})
    if($files.Count -ne 9){throw 'Nine immutable candidate inputs required'}
    foreach($file in $files){
        if([IO.File]::ReadAllText($file.FullName).Contains("`r")){throw 'Physical LF required'}
        $pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}
    }
    $pins+=,[pscustomobject]@{VerificationPath=$baselineCopy;SHA256=$baselineHash;Role='pre-repair-product-snapshot'}
    if($pins.Count -ne 998){throw '998mapped pins required'}
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json;$expected=[ordered]@{}
    foreach($name in @('red','focused','full')){$expected[$name]=@($oracle.$name|ForEach-Object{if($_ -ceq '{Context}'){'CONTEXT: '+$identity.Name+' admin=False'}else{$_}})}
    if($expected.red.Count -ne 6 -or $expected.focused.Count -ne 7 -or $expected.full.Count -ne 36){throw 'Oracle inventory'}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllText($baselineCopy,[IO.File]::ReadAllText($subject),[Text.UTF8Encoding]::new($false));Assert-Hash $baselineCopy $baselineHash
    Save-Json $manifest ([pscustomobject]@{Schema=1;BaseCommit='98f21970a41288ca2acbd008f46cde8ae2657135';Pins=$pins;Expected=$expected;Order=@('red','focused','full');MaxHosts=3;MaxReviewers=1;MaxMs=60000;ExitWaitMs=2000;DrainMs=2000;ManagedCounterFixtureApproved=$true;FormalAcceptance=$false})
    (Get-FileHash -LiteralPath $manifest).Hash;return
}
Assert-Hash $manifest $FrozenSHA256;$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($row in $rows){if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result freezes remaining work'};if($null -ne $row.TraceSHA256){Assert-Hash (Join-Path $packet ($row.Slot+'.trace.json')) $row.TraceSHA256};foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($row.Slot+'.'+$stream+'.bin')) $row.($stream+'SHA256')}}
if($Action -ceq 'SealProduct'){
    if($rows.Count -ne 1 -or $rows[0].Slot -cne 'red' -or (Test-Path -LiteralPath $seal)){throw 'Seal once after exact RED/probe only'}
    $hash=(Get-FileHash -LiteralPath $subject).Hash;if($hash -ceq $baselineHash){throw 'New ingress candidate required'}
    Save-Json $seal ([pscustomobject]@{ModuleSHA256=$hash;FrozenSHA256=$FrozenSHA256;RuntimeApproved=$false})
    (Get-FileHash -LiteralPath $seal).Hash;return
}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 3 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered max-three allowance; no retries'}
$subjectHash=$baselineHash
if($Slot -cne 'red'){Assert-Hash $seal $ProductSealSHA256;$s=Get-Content -LiteralPath $seal -Raw|ConvertFrom-Json;if($s.FrozenSHA256 -cne $FrozenSHA256){throw 'Seal binding'};$subjectHash=$s.ModuleSHA256}
Assert-Hash $subject $subjectHash
$target=Join-Path $PSScriptRoot $(if($Slot -ceq 'full'){'control-frame.tests.ps1'}else{'descriptor.tests.ps1'})
$want=@($f.Expected.$Slot);$wantExit=if($Slot -ceq 'red'){1}else{0}
$row=[pscustomobject][ordered]@{Number=($rows.Count+1);Slot=$Slot;AttemptConsumed=$true;Launched=$false;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;ModuleSHA256=$subjectHash;TimedOut=$false;ExitCode=$null;StreamsComplete=$false;OwnedExit=$false;PostPinsIntact=$false;ExactLines=$false;Outcome='INCOMPLETE';FormalAcceptance=$false;PID=$null;ElapsedMs=$null;LineCount=$null;StderrBytes=$null;stdoutSHA256=$null;stderrSHA256=$null;TraceSHA256=$null}
$rows+=,$row;Save-Json $index $rows
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target)){$p.StartInfo.ArgumentList.Add($arg)}
$p.StartInfo.ArgumentList.Add('-TracePath');$p.StartInfo.ArgumentList.Add((Join-Path $packet ($Slot+'.trace.json')))
if($Slot -ceq 'red'){$p.StartInfo.ArgumentList.Add('-Probe')}
if($Slot -ceq 'full'){$p.StartInfo.ArgumentList.Add('-Group');$p.StartInfo.ArgumentList.Add('all')}
$stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')));$stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$row.Launched=$true;$row.PID=$p.Id;Save-Json $index $rows
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$row.TimedOut=$true;$p.Kill()}
    $row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}
    $row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),$f.DrainMs)
}finally{$timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $index $rows}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash;Assert-Hash $sdk $sdkHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256};$row.PostPinsIntact=$true
$raw=[Text.UTF8Encoding]::new($false,$true).GetString([IO.File]::ReadAllBytes((Join-Path $packet ($Slot+'.stdout.bin'))))
$normalized=$raw.Replace("`r`n","`n");if($normalized.EndsWith("`n")){$normalized=$normalized.Substring(0,$normalized.Length-1)}
$lines=@($normalized.Split("`n"));$equal=$lines.Count -eq $want.Count
if($equal){for($i=0;$i -lt $want.Count;$i++){if($lines[$i] -cne $want[$i]){$equal=$false;break}}}
$row.ExactLines=$equal;$row.LineCount=$lines.Count;$row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length
foreach($stream in @('stdout','stderr')){$row.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))).Hash}
if(-not $row.TimedOut -and $row.OwnedExit -and $row.StreamsComplete -and $row.PostPinsIntact -and $row.ExactLines -and $row.ExitCode -eq $wantExit -and $row.StderrBytes -eq 0){$row.Outcome='EXPECTED_WITHIN_SCOPE'}
$trace=Join-Path $packet ($Slot+'.trace.json');if(Test-Path -LiteralPath $trace){$row.TraceSHA256=(Get-FileHash -LiteralPath $trace).Hash}
Save-Json $index $rows;$row|ConvertTo-Json;[Console]::Out.Write($raw)
if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result; freeze remaining work,no repair/retry/reviewer'}
