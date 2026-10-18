param(
    [Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
    [ValidateSet('full','ownership')][string]$Slot,
    [string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p1-owned-return-regression-v2'
$mf=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$prior='.superpowers/sdd/2026-10-18-p1-inline-ownership-timing'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$Hash){if((Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(Test-Path -LiteralPath $packet){throw 'New packet only; no retry/overwrite'}
    $null=New-Item -Path $packet -ItemType Directory
    if((Test-Path -LiteralPath $mf) -or (Test-Path -LiteralPath $index)){throw 'Already frozen/consumed; no overwrite'}
    $anchors=@{
        'frozen.json'='803585E86355FD1664C36A29495BFA8BFF9C1FCC3596E8ACDE403ED8C699FD62'
        'invocation.json'='C012FE83EB47D9FB7736A12A78BC45FD61BEF4EA68FB9AD6114B1CB4A2D268C9'
        'stdout.txt'='442DE40B2FE3137AC14C62C5C3EB79B427B138A2B6022B819B7B2E16777B925B'
        'stderr.txt'='E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'
        'analysis.json'='65110E040EE6A0083CB64BB8B1EFE3BDD8CFFC64E878A1B6750C051727F3EBD6'
    }
    foreach($name in $anchors.Keys){Assert-Hash (Join-Path $prior $name) $anchors[$name]}
    $old=Get-Content -LiteralPath (Join-Path $prior 'frozen.json') -Raw|ConvertFrom-Json
    if(@($old.Pins).Count -ne 666){throw 'Inherited pin inventory'}
    $pins=@($old.Pins)
    foreach($name in $anchors.Keys){$pins+=,[pscustomobject]@{Path=(Join-Path $prior $name);SHA256=$anchors[$name]}}
    $identity=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
    foreach($item in @($identity.Candidate,$identity.FullTest,$identity.OwnershipTest)){
        Assert-Hash $item.Path $item.SHA256
    }
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 5){throw 'Expected five portable files'}
    foreach($file in $files){$pins+=,[pscustomobject]@{Path=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    Assert-Hash $engine $engineHash
    foreach($pin in $pins){Assert-Hash $pin.Path $pin.SHA256}
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    if(@($oracle.full).Count -ne 38 -or @($oracle.ownership).Count -ne 22){throw 'Success inventory counts'}
    $context='CONTEXT: '+[Security.Principal.WindowsIdentity]::GetCurrent().Name+' admin=False'
    $expected=[ordered]@{}
    foreach($suite in @('full','ownership')){$expected[$suite]=@($oracle.$suite|ForEach-Object{if($_ -ceq '{Context}'){$context}else{$_}})}
    $specs=@(
        [pscustomobject]@{Name='full';Target=(Join-Path $root $identity.FullTest.Path);Group='all';Suite='full'}
        [pscustomobject]@{Name='ownership';Target=(Join-Path $root $identity.OwnershipTest.Path);Group='regression';Suite='ownership'}
    )
    $f=[pscustomobject]@{Schema=1;PreparedUTC=[DateTime]::UtcNow.ToString('o');Pins=$pins;Expected=[pscustomobject]$expected;Order=@($specs.Name);Slots=$specs;Subject=(Join-Path $root $identity.Candidate.Path);Engine=$engine;EngineSHA256=$engineHash;MaxHosts=2;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false}
    Save-Json $mf $f
    (Get-FileHash -LiteralPath $mf).Hash
    return
}
Assert-Hash $mf $FrozenSHA256
$f=Get-Content -LiteralPath $mf -Raw|ConvertFrom-Json
function Check-Pins{Assert-Hash $mf $FrozenSHA256;Assert-Hash $engine $engineHash;foreach($pin in $f.Pins){Assert-Hash $pin.Path $pin.SHA256}}
Check-Pins
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
if($rows.Count -ge 2 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered max-two allowance; no retry'}
foreach($previous in $rows){
    if(-not $previous.SuitePassed){throw 'Previous failure freezes later slots'}
    foreach($s in @('stdout','stderr')){Assert-Hash (Join-Path $packet ($previous.Slot+'.'+$s+'.txt')) $previous.($s+'SHA256')}
}
$spec=@($f.Slots|Where-Object{$_.Name -ceq $Slot});if($spec.Count -ne 1){throw 'Slot specification'};$spec=$spec[0]
foreach($stream in @('stdout','stderr')){if(Test-Path -LiteralPath (Join-Path $packet ($Slot+'.'+$stream+'.txt'))){throw 'Existing capture; no overwrite'}}
$row=[ordered]@{
    Number=$rows.Count+1;Slot=$Slot;AttemptConsumed=$true;StartedUTC=[DateTime]::UtcNow.ToString('o');Engine=$engine;EngineSHA256=$engineHash
    Target=$spec.Target;TargetSHA256=(Get-FileHash -LiteralPath $spec.Target).Hash;ModuleSHA256=(Get-FileHash -LiteralPath $f.Subject).Hash;Group=$spec.Group
    ChildLaunched=$false;PID=$null;ParentAdministratorRole=$false;TimedOut=$false;ExitCode=$null;ElapsedMs=$null
    CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false;ExactSuccessInventory=$false;SuitePassed=$false
    Outcome='EXECUTION_INCOMPLETE';ChildOutcome=$null;LineCount=0;StderrBytes=$null;FirstMismatch=$null;ControllerError=$null;FormalAcceptance=$false
}
function Save-Index{Save-Json $index @($rows+[pscustomobject]$row)}
Save-Index # Persist consumption before any launch attempt.
$psi=[Diagnostics.ProcessStartInfo]::new()
$psi.FileName=$engine;$psi.UseShellExecute=$false;$psi.CreateNoWindow=$true;$psi.WorkingDirectory=$root
$psi.RedirectStandardOutput=$true;$psi.RedirectStandardError=$true
$psi.Arguments='-NoProfile -NonInteractive -File "'+$spec.Target+'"'
if($spec.Group.Length -gt 0){$psi.Arguments+=' -Group '+$spec.Group}
$row.Arguments=$psi.Arguments;Save-Index
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
        if(-not [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)){throw 'Bounded stream drain incomplete'}
        [void]$ot.GetAwaiter().GetResult();[void]$et.GetAwaiter().GetResult();$row.StreamsComplete=$true
    }catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}}
    $watch.Stop();$row.ElapsedMs=$watch.ElapsedMilliseconds
    if($null -ne $of){$of.Dispose()};if($null -ne $ef){$ef.Dispose()}
    foreach($stream in @('stdout','stderr')){$path=Join-Path $packet ($Slot+'.'+$stream+'.txt');if(Test-Path -LiteralPath $path){$row[$stream+'SHA256']=(Get-FileHash -LiteralPath $path).Hash}}
    try{Check-Pins;$row.PostPinsIntact=$true}catch{$row.ControllerError=([string]$row.ControllerError+' | '+$_.Exception.Message)}
    $p.Dispose();Save-Index
}
if($row.StreamsComplete){
    $stdout=[IO.File]::ReadAllText((Join-Path $packet ($Slot+'.stdout.txt')),[Text.UTF8Encoding]::new($false,$true))
    $lines=@($stdout -split '\r?\n'|Where-Object{$_ -cne ''})
    $row.LineCount=$lines.Count;$row.StderrBytes=(Get-Item -LiteralPath (Join-Path $packet ($Slot+'.stderr.txt'))).Length
    $expected=@($f.Expected.($spec.Suite));$exact=$lines.Count -eq $expected.Count
    for($i=0;$i -lt [Math]::Min($lines.Count,$expected.Count);$i++){if($lines[$i] -cne $expected[$i]){$exact=$false;$row.FirstMismatch=$i;break}}
    $row.ExactSuccessInventory=$exact
    $valid=($row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.StderrBytes -eq 0 -and $row.PostPinsIntact -and $null -eq $row.ControllerError)
    if($valid -and $exact -and $row.ExitCode -eq 0){$row.Outcome='PASSED_WITHIN_SCOPE';$row.SuitePassed=$true}
    if($lines.Count -gt 0 -and $lines[-1] -cmatch '^RESULT: (FIXTURE_INVALID|OWNERSHIP_GAP_OBSERVED|EXECUTION_INCOMPLETE) case=([A-Za-z0-9-]+) detail=([A-Za-z0-9.:-]+) calls=([0-9]+)$'){
        $row.ChildOutcome=$Matches[1];if($valid -and $row.ExitCode -eq 1){$row.Outcome=$row.ChildOutcome}
    }
}
Save-Index
[pscustomobject]$row|ConvertTo-Json -Depth 12 -Compress
if(-not $row.SuitePassed){throw 'STOP: failed/incomplete; later slots frozen, no retry or repair'}
