param([Parameter(Mandatory)][ValidateSet('Prepare','SealRepair','Run')][string]$Action,
    [ValidateSet('red','full','phase-mutant','expiry-mutant','final')][string]$Slot,
    [string]$FrozenSHA256,[string]$RepairSealSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'));Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2a-type-receiver-repair'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json'
$builder=Join-Path $PSScriptRoot 'full.tests.ps1';$module=Join-Path $PSScriptRoot 'offline-timebase-plan.ps1'
$seal=Join-Path $packet 'repair-seal.json'
$oldBuilderCopy=Join-Path $packet 'old-tests.ps1';$productCopy=Join-Path $packet 'product-baseline.ps1'
$goodBuilderCopy=Join-Path $packet 'repaired-tests.ps1'
$phaseGood='Get-TimebasePlanDelta $snapshot.Frequency 500 $false';$phaseBad='Get-TimebasePlanDelta $snapshot.Frequency 500 $true'
$expiryGood='($snapshot.ObservedTick -ge $deadline)';$expiryBad='($snapshot.ObservedTick -gt $deadline)'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$SHA256){if([string]::IsNullOrEmpty($SHA256) -or (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $SHA256){throw ('Pin drift: '+$Path)}}
function Text-Hash([string]$Text){[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.UTF8Encoding]::new($false).GetBytes($Text)))}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'New packet only; no retry/overwrite'}
    $s=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
    if($s.Approval -cne 'User approved bounded test receiver repair; five NEW ordinary hosts plus one conditional readonly review'){throw 'Approval scope'}
    Assert-Hash $s.OriginalProduct.Path $s.OriginalProduct.SHA256
    Assert-Hash $s.OriginalTests.Path $s.OriginalTests.SHA256
    Assert-Hash $module $s.OriginalProduct.SHA256;Assert-Hash $builder $s.OldBuilderSHA256
    Assert-Hash $s.Engine $s.EngineSHA256
    if([IO.File]::ReadAllText($builder) -cne [IO.File]::ReadAllText((Join-Path $root $s.OriginalTests.Path))){throw 'Exact old full suite required'}
    $original=[IO.File]::ReadAllText((Join-Path $root $s.OriginalTests.Path))
    if($original.Split($s.TestBefore,[StringSplitOptions]::None).Count -ne 2){throw 'Unique exact test receiver anchor'}
    $anchors=@()
    foreach($a in $s.Anchors.PSObject.Properties){
        $path=Join-Path $s.InheritedPacket $a.Name;Assert-Hash $path $a.Value
        $anchors+=,[pscustomobject]@{Path=$path;VerificationPath=$path;Role='previous-packet-anchor';SHA256=$a.Value}
    }
    if($anchors.Count -ne 4){throw 'Anchor inventory'}
    $previous=Get-Content -LiteralPath (Join-Path $s.InheritedPacket 'frozen.json') -Raw|ConvertFrom-Json
    if(@($previous.Pins).Count -ne 829){throw 'Inherited inventory'}
    $pins=@($previous.Pins)+$anchors
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 8){throw 'Eight portable files required'}
    foreach($file in $files){
        if([IO.File]::ReadAllText($file.FullName).Contains("`r")){throw 'Physical LF required'}
        if($file.Name -cin @('full.tests.ps1','offline-timebase-plan.ps1')){continue}
        $pins+=,[pscustomobject]@{Path=$file.FullName;VerificationPath=$file.FullName;Role='immutable-repair-input';SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}
    }
    if($pins.Count -ne 839){throw 'Common inventory'}
    foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $good=[IO.File]::ReadAllText($module)
    if($good.Split($phaseGood,[StringSplitOptions]::None).Count -ne 2 -or $good.Split($expiryGood,[StringSplitOptions]::None).Count -ne 2){throw 'Unique mutant anchors'}
    $o=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    if(@($o.red).Count -ne 21 -or @($o.full).Count -ne 27 -or @($o.'phase-mutant').Count -ne 4 -or @($o.'expiry-mutant').Count -ne 4){throw 'Oracle inventory'}
    $expected=[ordered]@{};$context='CONTEXT: '+$identity.Name+' admin=False'
    foreach($name in @('red','full','phase-mutant','expiry-mutant')){$expected[$name]=@($o.$name|ForEach-Object{if($_ -ceq '{Context}'){$context}else{$_}})}
    $f=[pscustomobject]@{Schema=1;BaseCommit=$s.BaseCommit;Approval=$s.Approval;TestBefore=$s.TestBefore;TestAfter=$s.TestAfter;Pins=$pins;CommonCount=839;InheritedCount=829;Engine=$s.Engine;EngineSHA256=$s.EngineSHA256;Expected=[pscustomobject]$expected;Order=$s.Order;OldBuilderSHA256=$s.OldBuilderSHA256;ProductGoodSHA256=$s.OriginalProduct.SHA256;PhaseSHA256=(Text-Hash $good.Replace($phaseGood,$phaseBad));ExpirySHA256=(Text-Hash $good.Replace($expiryGood,$expiryBad));MaxHosts=5;MaxReviewers=1;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false}
    $null=New-Item -Path $packet -ItemType Directory
    [IO.File]::WriteAllText($oldBuilderCopy,[IO.File]::ReadAllText($builder),[Text.UTF8Encoding]::new($false))
    [IO.File]::WriteAllText($productCopy,$good,[Text.UTF8Encoding]::new($false))
    Save-Json $manifest $f
    (Get-FileHash -LiteralPath $manifest).Hash
    return
}
Assert-Hash $manifest $FrozenSHA256
$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
function Check-Pins{
    Assert-Hash $manifest $FrozenSHA256;Assert-Hash $f.Engine $f.EngineSHA256
    foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    Assert-Hash $oldBuilderCopy $f.OldBuilderSHA256;Assert-Hash $productCopy $f.ProductGoodSHA256
}
Check-Pins
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
foreach($previous in $rows){
    if(-not $previous.ExactExpectedOutcome){throw 'Earlier unexpected outcome freezes all remaining work'}
    foreach($stream in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($previous.Slot+'.'+$stream+'.bin')) $previous.($stream+'SHA256')}
}
if($Action -ceq 'SealRepair'){
    if($rows.Count -ne 1 -or $rows[0].Slot -cne 'red' -or (Test-Path -LiteralPath $seal) -or (Test-Path -LiteralPath $goodBuilderCopy)){throw 'Seal once after exact RED'}
    Assert-Hash $module $f.ProductGoodSHA256
    $new=[IO.File]::ReadAllText($builder)
    if($new.Contains("`r") -or $new -cne [IO.File]::ReadAllText($oldBuilderCopy).Replace($f.TestBefore,$f.TestAfter)){throw 'Exact one-expression test helper repair required'}
    [IO.File]::WriteAllText($goodBuilderCopy,$new,[Text.UTF8Encoding]::new($false))
    Save-Json $seal ([pscustomobject]@{Schema=1;FrozenSHA256=$FrozenSHA256;OldBuilderSHA256=$f.OldBuilderSHA256;GoodBuilderSHA256=(Text-Hash $new);ProductGoodSHA256=$f.ProductGoodSHA256;FormalAcceptance=$false})
    (Get-FileHash -LiteralPath $seal).Hash
    return
}
if([string]::IsNullOrEmpty($Slot) -or $rows.Count -ge 5 -or $Slot -cne $f.Order[$rows.Count]){throw 'Five ordered slots only; no retry'}
$builderSHA=$f.OldBuilderSHA256;$moduleSHA=$f.ProductGoodSHA256;$mutantDiff=$null
if($Slot -cne 'red'){
    Assert-Hash $seal $RepairSealSHA256;$repair=Get-Content -LiteralPath $seal -Raw|ConvertFrom-Json
    if($repair.FrozenSHA256 -cne $FrozenSHA256){throw 'Seal manifest mismatch'}
    $builderSHA=$repair.GoodBuilderSHA256;Assert-Hash $goodBuilderCopy $builderSHA
}
if($Slot -ceq 'phase-mutant'){$moduleSHA=$f.PhaseSHA256;$mutantDiff=[pscustomobject]@{Before=$phaseGood;After=$phaseBad}}
if($Slot -ceq 'expiry-mutant'){$moduleSHA=$f.ExpirySHA256;$mutantDiff=[pscustomobject]@{Before=$expiryGood;After=$expiryBad}}
Assert-Hash $builder $builderSHA;Assert-Hash $module $moduleSHA
$focused=$false
$target=Join-Path $PSScriptRoot $(if($focused){'fixture-isolation.tests.ps1'}else{'full.tests.ps1'})
$group=if($focused){''}elseif($Slot -ceq 'phase-mutant'){'phase-rounding'}elseif($Slot -ceq 'expiry-mutant'){'deadline-equality'}else{'all'}
$suite=if($Slot -ceq 'final'){'full'}else{$Slot};$expectedExit=if($Slot -cin @('red','phase-mutant','expiry-mutant')){1}else{0}
foreach($stream in @('stdout','stderr')){if(Test-Path -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.bin'))){throw 'Existing capture forbidden'}}
$row=[ordered]@{
    Number=$rows.Count+1;Slot=$Slot;AttemptConsumed=$true;ChildLaunched=$false;PID=$null;ParentAdministrator=$false
    Target=$target;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;BuilderSHA256=$builderSHA;ModuleSHA256=$moduleSHA;MutantDiff=$mutantDiff
    Engine=$f.Engine;EngineSHA256=$f.EngineSHA256;Group=$group;StartedUTC=[DateTime]::UtcNow.ToString('o')
    TimedOut=$false;ExitCode=$null;ElapsedMs=$null;CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false
    StderrBytes=$null;LineCount=0;ExactLines=$false;ExactExpectedOutcome=$false;Outcome='INCOMPLETE';FirstMismatch=$null;ControllerError=$null;FormalAcceptance=$false
}
function Save-Index{Save-Json $index @($rows+[pscustomobject]$row)}
Save-Index # Durable preconsumption includes helper/module/mutant identities.
$psi=[Diagnostics.ProcessStartInfo]::new();$psi.FileName=$f.Engine
$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target)){$psi.ArgumentList.Add($arg)}
if($group.Length -gt 0){$psi.ArgumentList.Add('-Group');$psi.ArgumentList.Add($group)}
$p=[Diagnostics.Process]::new();$p.StartInfo=$psi
$of=$null;$ef=$null;$ot=$null;$et=$null;$watch=[Diagnostics.Stopwatch]::new()
try{
    $of=[IO.File]::Open((Join-Path $packet ($Slot+'.stdout.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $ef=[IO.File]::Open((Join-Path $packet ($Slot+'.stderr.bin')),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $watch.Start();if(-not $p.Start()){throw 'Launch failed'}
    $row.ChildLaunched=$true;$row.PID=$p.Id;Save-Index
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($of);$et=$p.StandardError.BaseStream.CopyToAsync($ef)
    if(-not $p.WaitForExit([Math]::Max(0,30000-[int]$watch.ElapsedMilliseconds))){$row.TimedOut=$true}
}catch{$row.ControllerError=$_.Exception.Message}finally{
    if($row.ChildLaunched){try{
        if(-not $p.HasExited){$p.Kill();if(-not $p.WaitForExit(2000)){throw 'Owned exit unknown'}}
        $row.CleanupStatus='owned-host-exited';$row.ExitCode=$p.ExitCode
        if($null -eq $ot -or $null -eq $et){throw 'Missing stream tasks'}
        if(-not [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)){throw 'Bounded drain incomplete'}
        [void]$ot.GetAwaiter().GetResult();[void]$et.GetAwaiter().GetResult();$row.StreamsComplete=$true
    }catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
    $watch.Stop();$row.ElapsedMs=$watch.ElapsedMilliseconds
    if($null -ne $of){$of.Dispose()};if($null -ne $ef){$ef.Dispose()}
    foreach($stream in @('stdout','stderr')){$path=Join-Path $packet ($Slot+'.'+$stream+'.bin');if(Test-Path -LiteralPath $path){$row[$stream+'SHA256']=(Get-FileHash -LiteralPath $path).Hash}}
    try{Check-Pins;Assert-Hash $builder $builderSHA;Assert-Hash $module $moduleSHA;if($Slot -cne 'red'){Assert-Hash $seal $RepairSealSHA256;Assert-Hash $goodBuilderCopy $builderSHA};$row.PostPinsIntact=$true}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}
    $p.Dispose();Save-Index
}
if($row.StreamsComplete){try{
    $raw=[IO.File]::ReadAllText((Join-Path $packet ($Slot+'.stdout.bin')),[Text.UTF8Encoding]::new($false,$true))
    $canonical=$raw.Replace("`r`n","`n")
    if($canonical.Contains("`r") -or -not $canonical.EndsWith("`n")){throw 'Stream framing'}
    $lines=@($canonical.Substring(0,$canonical.Length-1) -split "`n");$row.LineCount=$lines.Count
    $row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.bin'))).Length
    $expected=@($f.Expected.$suite);$exact=$lines.Count -eq $expected.Count
    for($i=0;$i -lt [Math]::Min($lines.Count,$expected.Count);$i++){if($lines[$i] -cne $expected[$i]){$exact=$false;$row.FirstMismatch=$i;break}}
    $row.ExactLines=$exact
    $valid=$row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.StderrBytes -eq 0 -and $row.PostPinsIntact -and $null -eq $row.ControllerError
    if($valid -and $exact -and $row.ExitCode -eq $expectedExit){$row.ExactExpectedOutcome=$true;$row.Outcome='EXPECTED_WITHIN_SCOPE'}
}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
Save-Index
[pscustomobject]$row|ConvertTo-Json -Depth 12 -Compress
if(-not $row.ExactExpectedOutcome){throw 'STOP: unexpected/incomplete; no later slot, repair or retry'}
