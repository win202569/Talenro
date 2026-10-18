param(
    [Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
    [ValidateSet('focused','full','ownership')][string]$Slot,
    [string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p1-view-aware-ownership'
$mf=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocations.json'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$prior='.superpowers/sdd/2026-10-18-p1-string-retention-fixed-getter'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$Hash){if((Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(-not (Test-Path -LiteralPath $packet -PathType Container)){throw 'Create ledger packet before preparation'}
    if((Test-Path -LiteralPath $mf) -or (Test-Path -LiteralPath $index)){throw 'Already frozen/consumed; no overwrite'}
    $anchors=@{
        'frozen.json'='7EA1FBB567215A231E1D49649D8631AF3029BF551C6DDF04EAD0FA4717E47C15'
        'invocation.json'='443D954016ECFEA58322CA290F0F1C1C68D5CADE53EACA8A9FDD91C980393B02'
        'stdout.txt'='AAD79341B92C1B4779DB469871CA2198F32AE52F11194C8C727C75FFD003BCB8'
        'stderr.txt'='E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'
    }
    foreach($name in $anchors.Keys){Assert-Hash (Join-Path $prior $name) $anchors[$name]}
    $old=Get-Content -LiteralPath (Join-Path $prior 'frozen.json') -Raw|ConvertFrom-Json
    if(@($old.Pins).Count -ne 566){throw 'Inherited pin inventory'}
    $pins=@($old.Pins)
    foreach($name in $anchors.Keys){$pins+=,[pscustomobject]@{Path=(Join-Path $prior $name);SHA256=$anchors[$name]}}
    $identity=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
    foreach($source in $identity.Sources){Assert-Hash $source.Path $source.SHA256;$pins+=,[pscustomobject]@{Path=$source.Path;SHA256=$source.SHA256}}
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 6){throw 'Expected six portable files'}
    foreach($file in $files){$pins+=,[pscustomobject]@{Path=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    Assert-Hash $engine $engineHash
    foreach($pin in $pins){Assert-Hash $pin.Path $pin.SHA256}
    $oracle=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'expected-output.json') -Raw|ConvertFrom-Json
    if(@($oracle.focused).Count -ne 153 -or @($oracle.Facts).Count -ne 53 -or @($oracle.FactProperties).Count -ne 11 -or @($oracle.full).Count -ne 38 -or @($oracle.ownership).Count -ne 22){throw 'Success inventory counts'}
    $context='CONTEXT: '+[Security.Principal.WindowsIdentity]::GetCurrent().Name+' admin=False'
    $expected=[ordered]@{}
    foreach($suite in @('focused','full','ownership')){$expected[$suite]=@($oracle.$suite|ForEach-Object{if($_ -ceq '{Context}'){$context}else{$_}})}
    $specs=@(
        [pscustomobject]@{Name='focused';Target=(Join-Path $PSScriptRoot 'owned-return.tests.ps1');Group='';Suite='focused'}
        [pscustomobject]@{Name='full';Target=(Join-Path $PSScriptRoot '../scalar-owned-return-candidate/offline-supervision-policy.tests.ps1');Group='all';Suite='full'}
        [pscustomobject]@{Name='ownership';Target=(Join-Path $PSScriptRoot '../scalar-owned-return-candidate/offline-supervision-policy.ownership.tests.ps1');Group='regression';Suite='ownership'}
    )
    $f=[pscustomobject]@{Schema=1;PreparedUTC=[DateTime]::UtcNow.ToString('o');Pins=$pins;Expected=[pscustomobject]$expected;Order=@($specs.Name);Slots=$specs;Subject=(Join-Path $PSScriptRoot '../scalar-owned-return-candidate/offline-supervision-policy.ps1');Engine=$engine;EngineSHA256=$engineHash;MaxHosts=3;FocusedMaxCalls=80;FocusedMaxCases=39;DirectScalarCalls=29;DirectMetadataCalls=1;FactSchema=$oracle;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false}
    Save-Json $mf $f
    (Get-FileHash -LiteralPath $mf).Hash
    return
}
Assert-Hash $mf $FrozenSHA256
$f=Get-Content -LiteralPath $mf -Raw|ConvertFrom-Json
function Check-Pins{Assert-Hash $mf $FrozenSHA256;Assert-Hash $engine $engineHash;foreach($pin in $f.Pins){Assert-Hash $pin.Path $pin.SHA256}}
Check-Pins
$rows=@();if(Test-Path -LiteralPath $index){$rows=@(Get-Content -LiteralPath $index -Raw|ConvertFrom-Json)}
if($rows.Count -ge 3 -or $Slot -cne $f.Order[$rows.Count]){throw 'Ordered max-three allowance; no retry'}
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
    for($i=0;$i -lt [Math]::Min($lines.Count,$expected.Count);$i++){
        $matched=$lines[$i] -ceq $expected[$i]
        if($spec.Suite -ceq 'focused' -and $expected[$i] -cmatch '^\{Fact:([a-zA-Z0-9-]+)\}$'){
            $caseName=$Matches[1];$matched=$false
            try{
                if(-not $lines[$i].StartsWith('FACT: ',[StringComparison]::Ordinal)){throw 'Missing fact'}
                $fact=$lines[$i].Substring(6)|ConvertFrom-Json
                $names=@($fact.PSObject.Properties.Name)
                $schema=@($f.FactSchema.Facts|Where-Object{$_.Case -ceq $caseName})
                if($schema.Count -ne 1 -or $names.Count -ne 11 -or ($names -join '|') -cne ($f.FactSchema.FactProperties -join '|')){throw 'Fact inventory'}
                $schema=$schema[0]
                if($fact.Case -cne $caseName -or $fact.BaseType -cne $schema.BaseType -or $fact.StoredType -cnotin $schema.StoredTypes){throw 'Storage or case mismatch'}
                foreach($name in @('HeldDescriptorRetained','HeldPayloadSame','HeldMutationRetained','FieldBaseUnchanged','CrossOwnerClean','IndependentBase')){
                    if($fact.$name -isnot [bool] -or -not $fact.$name){throw 'Required true fact'}
                }
                if($fact.RereadDescriptorRetained -isnot [bool]){throw 'Reread fact type'}
                if($fact.StoredType -cne 'System.String' -and -not $fact.RereadDescriptorRetained){throw 'Required reread retention'}
                if($fact.RereadDescriptorRetained){
                    if($fact.RereadMutationRetained -isnot [bool] -or -not $fact.RereadMutationRetained){throw 'Reread mutation'}
                }elseif($null -ne $fact.RereadMutationRetained){throw 'Absent reread must be null'}
                $matched=$true
            }catch{$matched=$false}
        }
        if(-not $matched){$exact=$false;$row.FirstMismatch=$i;break}
    }
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
