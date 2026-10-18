param(
    [Parameter(Mandatory)][ValidateSet('Prepare','SealProduct','Run')][string]$Action,
    [ValidateSet('red','full','phase-mutant','expiry-mutant','final')][string]$Slot,
    [string]$FrozenSHA256,[string]$ProductSealSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2a-timebase-plan'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json'
$seal=Join-Path $packet 'product-seal.json';$goodCopy=Join-Path $packet 'good-product.ps1'
$module=Join-Path $PSScriptRoot 'offline-timebase-plan.ps1'
$target=Join-Path $PSScriptRoot 'offline-timebase-plan.tests.ps1'
$phaseGood='Get-TimebasePlanDelta $snapshot.Frequency 500 $false'
$phaseBad='Get-TimebasePlanDelta $snapshot.Frequency 500 $true'
$expiryGood='($snapshot.ObservedTick -ge $deadline)'
$expiryBad='($snapshot.ObservedTick -gt $deadline)'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$SHA256){if([string]::IsNullOrEmpty($SHA256) -or (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $SHA256){throw ('Pin drift: '+$Path)}}
function Text-Hash([string]$Text){[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.UTF8Encoding]::new($false).GetBytes($Text)))}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary token required'}
if($Action -ceq 'Prepare'){
    if((Test-Path -LiteralPath $packet) -or (Test-Path -LiteralPath $module)){throw 'New packet and absent product required'}
    $source=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
    Assert-Hash $source.ApprovedPlan $source.ApprovedPlanSHA256
    Assert-Hash $source.Spec $source.SpecSHA256
    Assert-Hash $source.Engine $source.EngineSHA256
    $pins=@()
    foreach($anchor in $source.Anchors.PSObject.Properties){
        $path=Join-Path $source.InheritedPacket $anchor.Name;Assert-Hash $path $anchor.Value
        $pins+=,[pscustomobject]@{Path=$path;VerificationPath=$path;Role='previous-packet-anchor';SHA256=$anchor.Value}
    }
    $old=Get-Content -LiteralPath (Join-Path $source.InheritedPacket 'frozen.json') -Raw|ConvertFrom-Json
    if(@($old.Pins).Count -ne 762 -or $source.InheritedPins -ne 762){throw 'Inherited inventory'}
    # Preserve the previous explicit mapping; never reinterpret historical paths.
    $pins=@($old.Pins)+$pins
    $pins+=,[pscustomobject]@{Path=$source.Spec;VerificationPath=$source.Spec;Role='design';SHA256=$source.SpecSHA256}
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 6){throw 'Six pre-product portable files required'}
    foreach($file in $files){
        $raw=[IO.File]::ReadAllText($file.FullName,[Text.UTF8Encoding]::new($false,$true))
        if($raw.Contains("`r")){throw 'Portable file must use physical LF'}
        $pins+=,[pscustomobject]@{Path=$file.FullName;VerificationPath=$file.FullName;Role='new-frozen-input';SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}
    }
    if($pins.Count -ne 775){throw 'Common pin inventory'}
    foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    if(@($oracle.red).Count -ne 3 -or @($oracle.full).Count -ne 27 -or @($oracle.'phase-mutant').Count -ne 4 -or @($oracle.'expiry-mutant').Count -ne 4){throw 'Oracle counts'}
    $context='CONTEXT: '+$identity.Name+' admin=False';$expected=[ordered]@{}
    foreach($name in @('red','full','phase-mutant','expiry-mutant')){$expected[$name]=@($oracle.$name|ForEach-Object{if($_ -ceq '{Context}'){$context}else{$_}})}
    $f=[pscustomobject]@{
        Schema=1;BaseCommit=$source.BaseCommit;ApprovedPlan=$source.ApprovedPlan;ApprovedPlanSHA256=$source.ApprovedPlanSHA256
        Pins=$pins;InheritedCount=762;CommonCount=775;Expected=[pscustomobject]$expected
        Engine=$source.Engine;EngineSHA256=$source.EngineSHA256;Target=$target;Subject=$module
        Order=@('red','full','phase-mutant','expiry-mutant','final');MaxHosts=5;MaxReviewers=1
        MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false
    }
    $null=New-Item -Path $packet -ItemType Directory
    Save-Json $manifest $f
    (Get-FileHash -LiteralPath $manifest).Hash
    return
}
Assert-Hash $manifest $FrozenSHA256
$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
function Check-Pins{
    Assert-Hash $manifest $FrozenSHA256;Assert-Hash $f.Engine $f.EngineSHA256
    foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
}
Check-Pins
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($previous in $rows){
    if(-not $previous.ExactExpectedOutcome){throw 'Previous unexpected outcome freezes all later work'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($previous.Slot+'.'+$stream+'.txt')) $previous.($stream+'SHA256')}
}
if($Action -ceq 'SealProduct'){
    if($rows.Count -ne 1 -or $rows[0].Slot -cne 'red' -or (Test-Path -LiteralPath $seal) -or (Test-Path -LiteralPath $goodCopy)){throw 'Seal once after exact RED'}
    $good=[IO.File]::ReadAllText($module,[Text.UTF8Encoding]::new($false,$true))
    if($good.Contains("`r") -or $good.Split($phaseGood,[StringSplitOptions]::None).Count -ne 2 -or $good.Split($expiryGood,[StringSplitOptions]::None).Count -ne 2){throw 'Exact good mutant anchors required'}
    [IO.File]::WriteAllText($goodCopy,$good,[Text.UTF8Encoding]::new($false))
    Save-Json $seal ([pscustomobject]@{Schema=1;GoodSHA256=(Get-FileHash -LiteralPath $module).Hash;GoodCopy=$goodCopy;FullManifestSHA256=$FrozenSHA256;PhaseSHA256=(Text-Hash $good.Replace($phaseGood,$phaseBad));ExpirySHA256=(Text-Hash $good.Replace($expiryGood,$expiryBad));FormalAcceptance=$false})
    (Get-FileHash -LiteralPath $seal).Hash
    return
}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 5 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered max-five allowance; no retries'}
$sourceSHA=$null;$mutantDiff=$null
if($Slot -ceq 'red'){
    if(Test-Path -LiteralPath $module){throw 'RED requires absent module'}
}else{
    Assert-Hash $seal $ProductSealSHA256;$s=Get-Content -LiteralPath $seal -Raw|ConvertFrom-Json
    if($s.FullManifestSHA256 -cne $FrozenSHA256){throw 'Product seal manifest mismatch'}
    Assert-Hash $goodCopy $s.GoodSHA256
    $sourceSHA=$s.GoodSHA256
    if($Slot -ceq 'phase-mutant'){$sourceSHA=$s.PhaseSHA256;$mutantDiff=[pscustomobject]@{Before=$phaseGood;After=$phaseBad}}
    if($Slot -ceq 'expiry-mutant'){$sourceSHA=$s.ExpirySHA256;$mutantDiff=[pscustomobject]@{Before=$expiryGood;After=$expiryBad}}
    Assert-Hash $module $sourceSHA
}
$group=switch -CaseSensitive ($Slot){'phase-mutant'{'phase-rounding'} 'expiry-mutant'{'deadline-equality'} default{'all'}}
$suite=if($Slot -ceq 'final'){'full'}else{$Slot}
$expectedExit=if($Slot -in @('full','final')){0}else{1}
foreach($stream in @('stdout','stderr')){if(Test-Path -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.txt'))){throw 'Existing capture forbidden'}}
$row=[ordered]@{
    Number=$rows.Count+1;Slot=$Slot;AttemptConsumed=$true;ChildLaunched=$false;PID=$null
    Target=$target;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;ModuleSHA256=$sourceSHA;MutantDiff=$mutantDiff
    Engine=$f.Engine;EngineSHA256=$f.EngineSHA256;Group=$group;ParentAdministrator=$false
    StartedUTC=[DateTime]::UtcNow.ToString('o');TimedOut=$false;ExitCode=$null;ElapsedMs=$null
    CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false;ExactLines=$false
    ExactExpectedOutcome=$false;Outcome='INCOMPLETE';LineCount=0;StderrBytes=$null;FirstMismatch=$null;ControllerError=$null;FormalAcceptance=$false
}
function Save-Index{Save-Json $index @($rows+[pscustomobject]$row)}
Save-Index # Durable consumption, including mutant identity/diff, before launch.
$psi=[Diagnostics.ProcessStartInfo]::new();$psi.FileName=$f.Engine
$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
foreach($argument in @('-NoProfile','-NonInteractive','-File',$target,'-Group',$group)){$psi.ArgumentList.Add($argument)}
$p=[Diagnostics.Process]::new();$p.StartInfo=$psi
$of=$null;$ef=$null;$ot=$null;$et=$null;$watch=[Diagnostics.Stopwatch]::new()
try{
    $of=[IO.File]::Open((Join-Path $packet ($Slot+'.stdout.txt')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $ef=[IO.File]::Open((Join-Path $packet ($Slot+'.stderr.txt')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $watch.Start();if(-not $p.Start()){throw 'Launch failed'}
    $row.ChildLaunched=$true;$row.PID=$p.Id;Save-Index
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($of);$et=$p.StandardError.BaseStream.CopyToAsync($ef)
    if(-not $p.WaitForExit([Math]::Max(0,30000-[int]$watch.ElapsedMilliseconds))){$row.TimedOut=$true}
}catch{$row.ControllerError=$_.Exception.Message}finally{
    if($row.ChildLaunched){try{
        if(-not $p.HasExited){$p.Kill();if(-not $p.WaitForExit(2000)){throw 'Owned exit unknown'}}
        $row.CleanupStatus='owned-host-exited';$row.ExitCode=$p.ExitCode
        if($null -eq $ot -or $null -eq $et){throw 'Stream tasks unavailable'}
        if(-not [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)){throw 'Bounded drain incomplete'}
        [void]$ot.GetAwaiter().GetResult();[void]$et.GetAwaiter().GetResult();$row.StreamsComplete=$true
    }catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
    $watch.Stop();$row.ElapsedMs=$watch.ElapsedMilliseconds
    if($null -ne $of){$of.Dispose()};if($null -ne $ef){$ef.Dispose()}
    foreach($stream in @('stdout','stderr')){
        $path=Join-Path $packet ($Slot+'.'+$stream+'.txt')
        if(Test-Path -LiteralPath $path){$row[$stream+'SHA256']=(Get-FileHash -LiteralPath $path).Hash}
    }
    try{
        Check-Pins
        if($Slot -ceq 'red'){if(Test-Path -LiteralPath $module){throw 'Product appeared during RED'}}else{Assert-Hash $seal $ProductSealSHA256;Assert-Hash $goodCopy $s.GoodSHA256;Assert-Hash $module $sourceSHA}
        $row.PostPinsIntact=$true
    }catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}
    $p.Dispose();Save-Index
}
if($row.StreamsComplete){try{
    $stdout=[IO.File]::ReadAllText((Join-Path $packet ($Slot+'.stdout.txt')),[Text.UTF8Encoding]::new($false,$true))
    $canonical=$stdout.Replace("`r`n","`n")
    if($canonical.Contains("`r") -or -not $canonical.EndsWith("`n")){throw 'Unexpected stream framing'}
    $canonical=$canonical.Substring(0,$canonical.Length-1);$lines=@($canonical -split "`n")
    $row.LineCount=$lines.Count;$row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.txt'))).Length
    $expected=@($f.Expected.$suite);$exact=$lines.Count -eq $expected.Count
    for($i=0;$i -lt [Math]::Min($lines.Count,$expected.Count);$i++){if($lines[$i] -cne $expected[$i]){$exact=$false;$row.FirstMismatch=$i;break}}
    $row.ExactLines=$exact
    $valid=$row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.StderrBytes -eq 0 -and $row.PostPinsIntact -and $null -eq $row.ControllerError
    if($valid -and $exact -and $row.ExitCode -eq $expectedExit){$row.ExactExpectedOutcome=$true;$row.Outcome='EXPECTED_WITHIN_SCOPE'}
}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
Save-Index
[pscustomobject]$row|ConvertTo-Json -Depth 12 -Compress
if(-not $row.ExactExpectedOutcome){throw 'STOP: unexpected/incomplete; no retry, repair or later slot'}
