param([Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,[string]$FrozenSHA256)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p2a-repaired-input-diagnostic'
$manifest=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
$target=Join-Path $PSScriptRoot 'diagnostic.tests.ps1'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 24),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$SHA256){if([string]::IsNullOrEmpty($SHA256) -or (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $SHA256){throw ('Pin drift: '+$Path)}}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
$principal=[Security.Principal.WindowsPrincipal]::new($identity)
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'New packet only; no retry/overwrite'}
    $s=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
    if($s.Approval -cne 'User approved ONE new repaired-version diagnostic; no fix/retry/reviewer'){throw 'Approval scope'}
    Assert-Hash $s.OriginalBuilder.Path $s.OriginalBuilder.SHA256
    Assert-Hash (Join-Path $PSScriptRoot 'fixture-builder.ps1') $s.OriginalBuilder.SHA256
    Assert-Hash $s.OriginalProduct.Path $s.OriginalProduct.SHA256
    Assert-Hash $s.OriginalTest.Path $s.OriginalTest.SHA256
    Assert-Hash (Join-Path $PSScriptRoot 'offline-timebase-plan.ps1') $s.OriginalProduct.SHA256
    Assert-Hash $s.Engine $s.EngineSHA256
    $anchors=@()
    foreach($a in $s.Anchors.PSObject.Properties){
        $path=Join-Path $s.OriginalPacket $a.Name;Assert-Hash $path $a.Value
        $anchors+=,[pscustomobject]@{Path=$path;VerificationPath=$path;Role='previous-packet-anchor';SHA256=$a.Value}
    }
    if($anchors.Count -ne 10){throw 'Anchor inventory'}
    $old=Get-Content -LiteralPath (Join-Path $s.OriginalPacket 'frozen.json') -Raw|ConvertFrom-Json
    if(@($old.Pins).Count -ne 811){throw 'Inherited pin inventory'}
    $pins=@($old.Pins)+$anchors
    $pins+=,[pscustomobject]@{Path=$s.OriginalProduct.Path;VerificationPath=$s.OriginalProduct.Path;Role='unchanged-original-product';SHA256=$s.OriginalProduct.SHA256}
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 7){throw 'Seven portable files required'}
    foreach($file in $files){
        if([IO.File]::ReadAllText($file.FullName).Contains("`r")){throw 'Physical LF required'}
        $pins+=,[pscustomobject]@{Path=$file.FullName;VerificationPath=$file.FullName;Role='diagnostic-input';SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}
    }
    if($pins.Count -ne 829){throw 'Common pin inventory'}
    foreach($pin in $pins){Assert-Hash $pin.VerificationPath $pin.SHA256}
    $prefix=@($old.Expected.full[0..19])
    $previous=[IO.File]::ReadAllLines((Join-Path $s.OriginalPacket 'full.stdout.bin'))
    if($previous.Count -ne 21){throw 'Original failure inventory'}
    for($i=0;$i -lt 20;$i++){if($previous[$i] -cne $prefix[$i]){throw 'Original prefix mismatch'}}
    if($prefix[0] -cne ('CONTEXT: '+$identity.Name+' admin=False')){throw 'Same ordinary context required'}
    $f=[pscustomobject]@{Schema=1;BaseCommit=$s.BaseCommit;Approval=$s.Approval;Pins=$pins;CommonCount=829;OriginalCommonCount=811;Engine=$s.Engine;EngineSHA256=$s.EngineSHA256;Target=$target;Prefix=$prefix;MaxHosts=1;MaxReviewers=0;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;MaxDiagnosticBytes=65536;Acceptance=$false}
    $null=New-Item -Path $packet -ItemType Directory
    Save-Json $manifest $f
    (Get-FileHash -LiteralPath $manifest).Hash
    return
}
Assert-Hash $manifest $FrozenSHA256
$f=Get-Content -LiteralPath $manifest -Raw|ConvertFrom-Json
function Check-Pins{Assert-Hash $manifest $FrozenSHA256;Assert-Hash $f.Engine $f.EngineSHA256;foreach($pin in $f.Pins){Assert-Hash $pin.VerificationPath $pin.SHA256}}
Check-Pins
if(Test-Path -LiteralPath $index){throw 'Single attempt already consumed; no retry'}
foreach($name in @('stdout.bin','stderr.bin')){if(Test-Path -LiteralPath (Join-Path $packet $name)){throw 'Existing capture forbidden'}}
$row=[ordered]@{
    Schema=1;AttemptConsumed=$true;MaxHosts=1;ChildLaunched=$false;PID=$null;ParentAdministrator=$false
    Target=$target;TargetSHA256=(Get-FileHash -LiteralPath $target).Hash;Engine=$f.Engine;EngineSHA256=$f.EngineSHA256
    StartedUTC=[DateTime]::UtcNow.ToString('o');TimedOut=$false;ExitCode=$null;ElapsedMs=$null
    CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false;StderrBytes=$null;LineCount=0
    Outcome='INCOMPLETE';Diagnostic=$null;ControllerError=$null;ProductChanged=$false;RepairPerformed=$false;Acceptance=$false
}
Save-Json $index ([pscustomobject]$row) # Persist consumption BEFORE launch.
$psi=[Diagnostics.ProcessStartInfo]::new();$psi.FileName=$f.Engine
$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
foreach($arg in @('-NoProfile','-NonInteractive','-File',$target,'-Group','all')){$psi.ArgumentList.Add($arg)}
$p=[Diagnostics.Process]::new();$p.StartInfo=$psi
$of=$null;$ef=$null;$ot=$null;$et=$null;$watch=[Diagnostics.Stopwatch]::new()
try{
    $of=[IO.File]::Open((Join-Path $packet 'stdout.bin'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $ef=[IO.File]::Open((Join-Path $packet 'stderr.bin'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $watch.Start();if(-not $p.Start()){throw 'Launch failed'}
    $row.ChildLaunched=$true;$row.PID=$p.Id;Save-Json $index ([pscustomobject]$row)
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
    foreach($stream in @('stdout','stderr')){$path=Join-Path $packet ($stream+'.bin');if(Test-Path -LiteralPath $path){$row[$stream+'SHA256']=(Get-FileHash -LiteralPath $path).Hash}}
    try{Check-Pins;$row.PostPinsIntact=$true}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}
    $p.Dispose();Save-Json $index ([pscustomobject]$row)
}
if($row.StreamsComplete){try{
    $raw=[IO.File]::ReadAllText((Join-Path $packet 'stdout.bin'),[Text.UTF8Encoding]::new($false,$true))
    $canonical=$raw.Replace("`r`n","`n")
    if($canonical.Contains("`r") -or -not $canonical.EndsWith("`n")){throw 'Stream framing'}
    $lines=@($canonical.Substring(0,$canonical.Length-1) -split "`n");$row.LineCount=$lines.Count
    $row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet 'stderr.bin')).Length
    $valid=$row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.PostPinsIntact -and $row.StderrBytes -eq 0 -and $null -eq $row.ControllerError
    $diags=@($lines|Where-Object{$_.StartsWith('DIAG: ',[StringComparison]::Ordinal)})
    if($diags.Count -ne 1){throw 'Exactly one diagnostic record required'}
    $json=$diags[0].Substring(6)
    if([Text.UTF8Encoding]::new($false).GetByteCount($json) -gt 65536){throw 'Oversized diagnostic'}
    $diag=$json|ConvertFrom-Json
    if(@($diag.PSObject.Properties).Count -ne 18 -or $diag.Schema -ne 1 -or $diag.ProductChanged -ne $false -or $diag.RepairPerformed -ne $false -or $diag.Acceptance -ne $false){throw 'Diagnostic schema/flags'}
    $row.Diagnostic=$diag
    $prefix=$lines.Count -ge 20
    for($i=0;$i -lt [Math]::Min(20,$lines.Count);$i++){if($lines[$i] -cne $f.Prefix[$i]){$prefix=$false;break}}
    $site=($diag.OriginalOperationLine -ge 1 -and $diag.OriginalOperationLine -le 167 -and $diag.DiagnosticLine -gt 0 -and -not [string]::IsNullOrEmpty($diag.Stack) -and -not [string]::IsNullOrEmpty($diag.ExceptionType) -and -not [string]::IsNullOrEmpty($diag.ErrorId))
    if($valid -and $prefix -and $site -and $lines.Count -eq 22 -and $lines[20] -ceq $diags[0] -and $lines[21] -ceq 'ASSERTION FAILED: P2AInputMutation' -and $row.ExitCode -eq 1 -and $diag.Case -ceq 'InputMutation' -and $diag.Outcome -ceq 'INPUT_MUTATION_EXCEPTION'){$row.Outcome='REPRODUCED_SITE_CAPTURED'}
    elseif($valid -and $prefix -and $lines.Count -eq 22 -and $lines[20] -ceq 'PASS: P2AInputMutation' -and $lines[21] -ceq $diags[0] -and $row.ExitCode -eq 2 -and $diag.Case -ceq 'InputMutation' -and $diag.Outcome -ceq 'NOT_REPRODUCED'){$row.Outcome='NOT_REPRODUCED'}
    elseif($valid -and $site -and $row.ExitCode -eq 1 -and $diag.Outcome -ceq 'EARLIER_FAILURE'){
        $priorCount=$lines.Count-2;$ordered=$priorCount -ge 2 -and $priorCount -lt 20
        for($i=0;$i -lt [Math]::Min($priorCount,20);$i++){if($lines[$i] -cne $f.Prefix[$i]){$ordered=$false}}
        if($ordered -and $lines[-2] -ceq $diags[0] -and $lines[-1] -ceq ('ASSERTION FAILED: P2A'+$diag.Case)){$row.Outcome='EARLIER_FAILURE'}
    }
}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
Save-Json $index ([pscustomobject]$row)
[pscustomobject]$row|ConvertTo-Json -Depth 24 -Compress
if($row.Outcome -ceq 'INCOMPLETE'){throw 'STOP: diagnostic incomplete; allowance spent, no retry/repair'}
# All outcomes stop here. No success, repair, further test or review authorization.
