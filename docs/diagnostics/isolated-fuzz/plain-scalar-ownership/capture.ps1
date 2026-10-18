param(
    [Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
    [string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p1-plain-scalar-ownership'
$mf=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$review='.superpowers/sdd/2026-10-18-p1-token-readonly-review/manifest.json'
$reviewHash='3AEC22EAA09003FC2BE6B03C32AD4C9728DFC4E660AC0340A6D124D48CBCC0B2'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$Hash){if((Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(-not (Test-Path -LiteralPath $packet -PathType Container)){throw 'Create ledger packet before preparation'}
    if((Test-Path -LiteralPath $mf) -or (Test-Path -LiteralPath $index)){throw 'Already frozen/consumed; no overwrite'}
    Assert-Hash $review $reviewHash
    $r=Get-Content -LiteralPath $review -Raw|ConvertFrom-Json
    Assert-Hash $r.InheritedManifest $r.InheritedManifestSHA256
    $inherited=Get-Content -LiteralPath $r.InheritedManifest -Raw|ConvertFrom-Json
    if(@($r.ScopePins).Count -ne 51 -or @($inherited.Pins).Count -ne 476){throw 'Historical pin inventory'}
    $pins=@($r.ScopePins)+@($inherited.Pins)
    $pins+=,[pscustomobject]@{Path=$review;SHA256=$reviewHash}
    $pins+=,[pscustomobject]@{Path=$r.InheritedManifest;SHA256=$r.InheritedManifestSHA256}
    $subject='docs/diagnostics/isolated-fuzz/integer-token-local-candidate/offline-supervision-policy.ps1'
    Assert-Hash $subject 'F9B4E408D9A4FF9C7A36EFD6E1F2472F17DACB3F76F463FF4973DAB4139CC82A'
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 5){throw 'Expected five portable files'}
    foreach($file in $files){$pins+=,[pscustomobject]@{Path=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    $proposal='docs/roadmap/2026-10-18-p1-plain-scalar-ownership-proposal.md'
    $pins+=,[pscustomobject]@{Path=$proposal;SHA256=(Get-FileHash -LiteralPath $proposal).Hash}
    Assert-Hash $engine $engineHash
    foreach($pin in $pins){Assert-Hash $pin.Path $pin.SHA256}
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    if(@($oracle.Lines).Count -ne 41){throw 'Expected 41 full-success output lines'}
    $f=[pscustomobject]@{Schema=1;PreparedUTC=[DateTime]::UtcNow.ToString('o');Pins=$pins;Expected=@($oracle.Lines);Target=(Join-Path $PSScriptRoot 'ownership.tests.ps1');Subject=$subject;Engine=$engine;EngineSHA256=$engineHash;MaxHosts=1;MaxCalls=38;MaxCases=18;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false}
    Save-Json $mf $f
    (Get-FileHash -LiteralPath $mf).Hash
    return
}
Assert-Hash $mf $FrozenSHA256
$f=Get-Content -LiteralPath $mf -Raw|ConvertFrom-Json
function Check-Pins{
    Assert-Hash $mf $FrozenSHA256;Assert-Hash $engine $engineHash
    foreach($pin in $f.Pins){Assert-Hash $pin.Path $pin.SHA256}
}
Check-Pins
foreach($name in @('invocation.json','stdout.txt','stderr.txt')){if(Test-Path -LiteralPath (Join-Path $packet $name)){throw 'One attempt bound: no overwrite/retry'}}
$row=[ordered]@{
    Schema=1;AttemptConsumed=$true;StartedUTC=[DateTime]::UtcNow.ToString('o');Engine=$engine;EngineSHA256=$engineHash
    Target=$f.Target;TargetSHA256=(Get-FileHash -LiteralPath $f.Target).Hash;ModuleSHA256=(Get-FileHash -LiteralPath $f.Subject).Hash
    ChildLaunched=$false;PID=$null;ParentAdministratorRole=$false;TimedOut=$false;ExitCode=$null;ElapsedMs=$null
    CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false;ExactSuccessInventory=$false;SuitePassed=$false
    Outcome='EXECUTION_INCOMPLETE';ChildOutcome=$null;LineCount=0;StderrBytes=$null;FirstMismatch=$null;ControllerError=$null;FormalAcceptance=$false
}
function Save-Index{Save-Json $index ([pscustomobject]$row)}
# Persist the consumed allowance before any launch attempt.
Save-Index
$psi=[Diagnostics.ProcessStartInfo]::new()
$psi.FileName=$engine;$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
$psi.Arguments='-NoProfile -NonInteractive -File "'+$f.Target+'"'
$row.Arguments=$psi.Arguments;Save-Index
$p=[Diagnostics.Process]::new();$p.StartInfo=$psi
$of=$null;$ef=$null;$ot=$null;$et=$null;$watch=[Diagnostics.Stopwatch]::new()
try{
    $of=[IO.File]::Open((Join-Path $packet 'stdout.txt'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $ef=[IO.File]::Open((Join-Path $packet 'stderr.txt'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $watch.Start();if(-not $p.Start()){throw 'Launch failed'}
    $row.ChildLaunched=$true;$row.PID=$p.Id;Save-Index
    $ot=$p.StandardOutput.BaseStream.CopyToAsync($of);$et=$p.StandardError.BaseStream.CopyToAsync($ef)
    if(-not $p.WaitForExit([Math]::Max(0,30000-[int]$watch.ElapsedMilliseconds))){$row.TimedOut=$true}
}catch{$row.ControllerError=$_.Exception.Message}finally{
    if($row.ChildLaunched){try{
        if(-not $p.HasExited){$p.Kill();if(-not $p.WaitForExit(2000)){throw 'Owned exit unknown'}}
        $row.CleanupStatus='owned-host-exited';$row.ExitCode=$p.ExitCode
        if($null -eq $ot -or $null -eq $et){throw 'Stream tasks unavailable'}
        if(-not [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)){throw 'Bounded stream drain incomplete'}
        [void]$ot.GetAwaiter().GetResult();[void]$et.GetAwaiter().GetResult();$row.StreamsComplete=$true
    }catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
    $watch.Stop();$row.ElapsedMs=$watch.ElapsedMilliseconds
    if($null -ne $of){$of.Dispose()};if($null -ne $ef){$ef.Dispose()}
    foreach($stream in @('stdout','stderr')){$path=Join-Path $packet ($stream+'.txt');if(Test-Path -LiteralPath $path){$row[$stream+'SHA256']=(Get-FileHash -LiteralPath $path).Hash}}
    try{Check-Pins;$row.PostPinsIntact=$true}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}
    $p.Dispose();Save-Index
}
if($row.StreamsComplete){
    $stdout=[IO.File]::ReadAllText((Join-Path $packet 'stdout.txt'),[Text.UTF8Encoding]::new($false,$true))
    $lines=@($stdout -split '\r?\n'|Where-Object{$_ -cne ''})
    $row.LineCount=$lines.Count;$row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet 'stderr.txt')).Length
    $expected=@($f.Expected);$exact=$lines.Count -eq $expected.Count
    for($i=0;$i -lt [Math]::Min($lines.Count,$expected.Count);$i++){if($lines[$i] -cne $expected[$i]){$exact=$false;$row.FirstMismatch=$i;break}}
    $row.ExactSuccessInventory=$exact
    $valid=($row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.StderrBytes -eq 0 -and $row.PostPinsIntact -and $null -eq $row.ControllerError)
    if($valid -and $exact -and $row.ExitCode -eq 0){$row.Outcome='ISOLATED_WITHIN_MATRIX';$row.SuitePassed=$true}
    # Named child diagnostics do not become isolation success. Save them for readonly adjudication.
    if($lines.Count -gt 0 -and $lines[-1] -cmatch '^RESULT: (FIXTURE_INVALID|OWNERSHIP_GAP_OBSERVED|EXECUTION_INCOMPLETE) case=([A-Za-z0-9-]+) detail=([A-Za-z0-9.:-]+) calls=([0-9]+)$'){
        $row.ChildOutcome=$Matches[1]
        if($valid -and $row.ExitCode -eq 1){$row.Outcome=$row.ChildOutcome}
    }
}
Save-Index
[pscustomobject]$row|ConvertTo-Json -Depth 12 -Compress
if(-not $row.SuitePassed){throw 'STOP: characterization failed/incomplete; attempt spent, no retry'}
