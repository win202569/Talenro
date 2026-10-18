param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
    [ValidateSet('reserved-tail','sequence-equality','final')][string]$Slot,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-codec-negative-controls'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json'
$previous=Join-Path $root '.superpowers/sdd/2026-10-18-p2b-payloadlength-repair'
$candidate=Join-Path $root 'docs/diagnostics/isolated-fuzz/p2b-payloadlength-repair'
$subject=Join-Path $candidate 'control-frame.ps1';$target=Join-Path $candidate 'control-frame.tests.ps1'
$goodCopy=Join-Path $packet 'good-product.ps1'
$goodHash='FA6047A4DF30AE97F3CD02F689CAC12A073020682571E34587A2E74D53AB65E6'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$tailGood='Assert-P2BCZeroBytes $snapshot 224 32';$tailBad='Assert-P2BCZeroBytes $snapshot 224 0'
$seqGood='$message.Sequence -ne $expected.NextSequence';$seqBad='$message.Sequence -lt $expected.NextSequence'
function Assert-Hash($Path,$Hash){if([string]::IsNullOrEmpty($Hash) -or (Get-FileHash -LiteralPath $Path).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
function Text-Hash([string]$Text){[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.UTF8Encoding]::new($false).GetBytes($Text)))}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary parent required'}
Assert-Hash $engine $engineHash
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'New packet required; no replay'}
    Assert-Hash (Join-Path $previous 'frozen.json') '6BF854E076D81112130688BF68200CE9F7F7CB4CE09393F04454440FB7FCD048'
    Assert-Hash (Join-Path $previous 'receipt.json') 'D9FA8125F2F2991C6BB04AB11B36CE6511B8046E4F019E92B2C828C4B3E39D70'
    $prior=Get-Content -LiteralPath (Join-Path $previous 'receipt.json') -Raw|ConvertFrom-Json
    if($prior.Outcome -cne 'EXPECTED_WITHIN_SCOPE' -or $prior.ExitCode -ne 0 -or -not $prior.ExactLines){throw 'Prerequisite full must be exact'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $previous ($stream+'.bin')) $prior.($stream+'SHA256')}
    $old=Get-Content -LiteralPath (Join-Path $previous 'frozen.json') -Raw|ConvertFrom-Json
    if($old.Pins.Count -ne 888){throw 'Inherited inventory'}
    foreach($pin in $old.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    Assert-Hash $subject $goodHash;Assert-Hash $target '239E80EFDE155F2950A8C8AE9633680D549CF7B825EBCCB5D65D54327375B597'
    $good=[IO.File]::ReadAllText($subject)
    if($good.Split($tailGood,[StringSplitOptions]::None).Count -ne 2 -or $good.Split($seqGood,[StringSplitOptions]::None).Count -ne 2){throw 'Unique mutation anchors required'}
    $pins=@();$mapped=0
    foreach($pin in $old.Pins){
        if([StringComparer]::OrdinalIgnoreCase.Equals([IO.Path]::GetFullPath($pin.VerificationPath),[IO.Path]::GetFullPath($subject))){
            $mapped++;$pins+=,[pscustomobject]@{SourceVerificationPath=$pin.VerificationPath;VerificationPath=$goodCopy;SHA256=$pin.SHA256;Role='temporary-mutant-baseline-mapping'}
        }else{$pins+=,$pin}
    }
    if($mapped -ne 1){throw 'Exactly one mutable subject mapping required'}
    foreach($dir in @($previous,$PSScriptRoot)){foreach($file in Get-ChildItem -LiteralPath $dir -File -Force){$pins+=,[pscustomobject]@{VerificationPath=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}}
    if($pins.Count -ne 895){throw '895mapped pins required'}
    $oracle=Get-Content -LiteralPath (Join-Path $candidate 'expected-output.json') -Raw|ConvertFrom-Json
    $expected=[ordered]@{}
    foreach($name in @('reserved-tail','sequence-equality','full')){$expected[$name]=@($oracle.$name|ForEach-Object{if($_ -ceq '{Context}'){'CONTEXT: '+$identity.Name+' admin=False'}else{$_}})}
    if($expected['reserved-tail'].Count -ne 4 -or $expected['sequence-equality'].Count -ne 4 -or $expected.full.Count -ne 35){throw 'Oracle inventory'}
    $null=New-Item -ItemType Directory -Path $packet
    [IO.File]::WriteAllText($goodCopy,$good,[Text.UTF8Encoding]::new($false));Assert-Hash $goodCopy $goodHash
    Save-Json $manifest ([pscustomobject]@{Schema=1;BaseCommit='07269905d7ba84ca4891be08246c3573aad99bf6';Pins=$pins;Expected=$expected;Order=@('reserved-tail','sequence-equality','final');GoodSHA256=$goodHash;TailSHA256=(Text-Hash $good.Replace($tailGood,$tailBad));SequenceSHA256=(Text-Hash $good.Replace($seqGood,$seqBad));MaxHosts=3;MaxReviewers=1;MaxMs=60000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false})
    (Get-FileHash -LiteralPath $manifest).Hash;return
}
Assert-Hash $manifest $FrozenSHA256;$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($row in $rows){if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Previous unexpected result freezes remaining work'};foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($row.Slot+'.'+$stream+'.bin')) $row.($stream+'SHA256')}}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 3 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered max-three allowance; no retry'}
$subjectHash=if($Slot -ceq 'reserved-tail'){$f.TailSHA256}elseif($Slot -ceq 'sequence-equality'){$f.SequenceSHA256}else{$f.GoodSHA256}
Assert-Hash $subject $subjectHash
$group=if($Slot -ceq 'final'){'all'}else{$Slot};$want=if($Slot -ceq 'final'){@($f.Expected.full)}else{@($f.Expected.$Slot)};$wantExit=if($Slot -ceq 'final'){0}else{1}
$row=[pscustomobject][ordered]@{Number=($rows.Count+1);Slot=$Slot;AttemptConsumed=$true;Launched=$false;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;ModuleSHA256=$subjectHash;Group=$group;TimedOut=$false;ExitCode=$null;StreamsComplete=$false;OwnedExit=$false;PostPinsIntact=$false;ExactLines=$false;Outcome='INCOMPLETE';FormalAcceptance=$false;PID=$null;ElapsedMs=$null;LineCount=$null;StderrBytes=$null;stdoutSHA256=$null;stderrSHA256=$null}
$rows+=,$row;Save-Json $index $rows
$p=[Diagnostics.Process]::new();$p.StartInfo=[Diagnostics.ProcessStartInfo]::new()
$p.StartInfo.FileName=$engine;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true;$p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target,'-Group',$group)){$p.StartInfo.ArgumentList.Add($arg)}
$stdout=[IO.File]::Create((Join-Path $packet ($Slot+'.stdout.bin')));$stderr=[IO.File]::Create((Join-Path $packet ($Slot+'.stderr.bin')))
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $null=$p.Start();$row.Launched=$true;$row.PID=$p.Id;Save-Json $index $rows
    $outTask=$p.StandardOutput.BaseStream.CopyToAsync($stdout);$errTask=$p.StandardError.BaseStream.CopyToAsync($stderr)
    if(-not $p.WaitForExit($f.MaxMs)){$row.TimedOut=$true;$p.Kill()}
    $row.OwnedExit=$p.WaitForExit($f.ExitWaitMs);if($row.OwnedExit){$row.ExitCode=$p.ExitCode}
    $row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($outTask,$errTask),$f.DrainMs)
}finally{$timer.Stop();$row.ElapsedMs=$timer.ElapsedMilliseconds;$stdout.Dispose();$stderr.Dispose();$p.Dispose();Save-Json $index $rows}
Assert-Hash $manifest $FrozenSHA256;Assert-Hash $subject $subjectHash;Assert-Hash $engine $engineHash
foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256};$row.PostPinsIntact=$true
$raw=[IO.File]::ReadAllText((Join-Path $packet ($Slot+'.stdout.bin')))
$lines=@($raw.Replace("`r`n","`n").TrimEnd("`n").Split("`n"));$equal=$lines.Count -eq $want.Count
if($equal){for($i=0;$i -lt $want.Count;$i++){if($lines[$i] -cne $want[$i]){$equal=$false;break}}}
$row.ExactLines=$equal;$row.LineCount=$lines.Count;$row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length
foreach($stream in @('stdout','stderr')){$row.($stream+'SHA256')=(Get-FileHash -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))).Hash}
if(-not $row.TimedOut -and $row.OwnedExit -and $row.StreamsComplete -and $row.PostPinsIntact -and $row.ExactLines -and $row.ExitCode -eq $wantExit -and $row.StderrBytes -eq 0){$row.Outcome='EXPECTED_WITHIN_SCOPE'}
Save-Json $index $rows;$row|ConvertTo-Json;[Console]::Out.Write($raw)
if($row.Outcome -cne 'EXPECTED_WITHIN_SCOPE'){throw 'Unexpected result; restore subject and freeze remaining work'}
