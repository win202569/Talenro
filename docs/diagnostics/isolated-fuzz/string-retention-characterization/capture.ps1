param(
    [Parameter(Mandatory)][ValidateSet('Prepare','Run')][string]$Action,
    [string]$FrozenSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
Set-Location -LiteralPath $root
$packet=Join-Path $root '.superpowers/sdd/2026-10-18-p1-string-retention-characterization'
$mf=Join-Path $packet 'frozen.json';$index=Join-Path $packet 'invocation.json'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$prior='.superpowers/sdd/2026-10-18-p1-owned-return-repair'
function Save-Json($Path,$Value){[IO.File]::WriteAllText($Path,($Value|ConvertTo-Json -Depth 20),[Text.UTF8Encoding]::new($false))}
function Assert-Hash($Path,[string]$Hash){if((Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash -cne $Hash){throw ('Pin drift: '+$Path)}}
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
if($principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Ordinary controller token required'}
if($Action -ceq 'Prepare'){
    if(-not (Test-Path -LiteralPath $packet -PathType Container)){throw 'Create ledger packet before preparation'}
    if((Test-Path -LiteralPath $mf) -or (Test-Path -LiteralPath $index)){throw 'Already frozen/consumed; no overwrite'}
    $anchors=@{
        'frozen.json'='EA27F7D75EFCAF291A4B59F15B531122B41ED19AACAF6EA1850031010FC613F8'
        'invocations.json'='5C6820047BBB033AA43AC5F92766CE39E0F7CE5B421850E5F164DEFBF1032783'
        'focused.stdout.txt'='259CC202BD1004450F0EECCD4FE99EFB3535278D267CB9A135A16A8DFE45C192'
        'focused.stderr.txt'='E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'
    }
    foreach($name in $anchors.Keys){Assert-Hash (Join-Path $prior $name) $anchors[$name]}
    $old=Get-Content -LiteralPath (Join-Path $prior 'frozen.json') -Raw|ConvertFrom-Json
    if(@($old.Pins).Count -ne 548){throw 'Inherited pin inventory'}
    $pins=@($old.Pins)
    foreach($name in $anchors.Keys){$pins+=,[pscustomobject]@{Path=(Join-Path $prior $name);SHA256=$anchors[$name]}}
    $files=@(Get-ChildItem -LiteralPath $PSScriptRoot -File -Force)
    if($files.Count -ne 5){throw 'Expected five portable files'}
    foreach($file in $files){$pins+=,[pscustomobject]@{Path=$file.FullName;SHA256=(Get-FileHash -LiteralPath $file.FullName).Hash}}
    Assert-Hash $engine $engineHash
    foreach($pin in $pins){Assert-Hash $pin.Path $pin.SHA256}
    $schema=Get-Content -LiteralPath (Join-Path $PSScriptRoot 'output-schema.json') -Raw|ConvertFrom-Json
    if(@($schema.Cases).Count -ne 4 -or $schema.CompleteLineCount -ne 6 -or @($schema.Properties).Count -ne 19){throw 'Schema inventory'}
    $f=[pscustomobject]@{Schema=1;PreparedUTC=[DateTime]::UtcNow.ToString('o');Pins=$pins;OutputSchema=$schema;Target=(Join-Path $PSScriptRoot 'probe.ps1');Engine=$engine;EngineSHA256=$engineHash;MaxHosts=1;ProductCalls=0;MaxCases=4;MaxMs=30000;ExitWaitMs=2000;DrainMs=2000;FormalAcceptance=$false}
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
    Target=$f.Target;TargetSHA256=(Get-FileHash -LiteralPath $f.Target).Hash
    ChildLaunched=$false;PID=$null;ParentAdministratorRole=$false;TimedOut=$false;ExitCode=$null;ElapsedMs=$null
    CleanupStatus='unknown';StreamsComplete=$false;PostPinsIntact=$false;SchemaValidated=$false;DiagnosticComplete=$false
    Outcome='INCOMPLETE';ChildOutcome=$null;LineCount=0;StderrBytes=$null;FirstMismatch=$null;ControllerError=$null;FormalAcceptance=$false
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
    $valid=($row.ChildLaunched -and -not $row.TimedOut -and $row.ElapsedMs -le 34000 -and $row.CleanupStatus -ceq 'owned-host-exited' -and $row.StderrBytes -eq 0 -and $row.PostPinsIntact -and $null -eq $row.ControllerError)
    if($valid -and $row.ExitCode -eq 0){
        try{
            $schema=$f.OutputSchema
            if($lines.Count -ne 6 -or $lines[0] -cne $schema.Host -or $lines[-1] -cne $schema.Terminal){throw 'Diagnostic envelope'}
            for($i=0;$i -lt 4;$i++){
                if(-not $lines[$i+1].StartsWith('FACT: ',[StringComparison]::Ordinal)){throw 'Fact prefix'}
                $fact=$lines[$i+1].Substring(6)|ConvertFrom-Json
                $props=@($fact.PSObject.Properties)
                if($props.Count -ne $schema.Properties.Count){throw 'Fact field count'}
                for($j=0;$j -lt $props.Count;$j++){if($props[$j].Name -cne $schema.Properties[$j]){throw 'Fact field order/name'}}
                foreach($name in $schema.Booleans){if($fact.$name -isnot [bool]){throw 'Fact Boolean type'}}
                foreach($name in $schema.NullableBooleans){if($null -ne $fact.$name -and $fact.$name -isnot [bool]){throw 'Fact nullable Boolean type'}}
                if($fact.Case -cne $schema.Cases[$i] -or $fact.ProductCalls -ne 0 -or ($fact.ProductCalls -isnot [int] -and $fact.ProductCalls -isnot [long])){throw 'Case/call identity'}
                foreach($name in @('BaseType','StoredTypeBefore','StoredTypeAfter','ViewType','Capability')){if($fact.$name -isnot [string]){throw 'Fact String type'}}
                $wantBase=if($i -eq 1){'System.Int32'}else{'System.String'}
                if($fact.BaseType -cne $wantBase -or $fact.StoredTypeBefore -cne $fact.StoredTypeAfter -or
                    $schema.StoredTypes -cnotcontains $fact.StoredTypeBefore -or $schema.ViewTypes -cnotcontains $fact.ViewType -or
                    $schema.Capabilities -cnotcontains $fact.Capability){throw 'Fact type/capability bounds'}
                if($fact.StoredTypeBefore -cne $wantBase -and $fact.StoredTypeBefore -cne 'System.Management.Automation.PSObject'){throw 'Stored/base type disagreement'}
                if(-not ($fact.BaseIdentitySame -and $fact.FieldDescriptorSame -and $fact.HeldDescriptorRetained -and $fact.HeldPayloadSame -and $fact.HeldMutationRetained)){throw 'Required fidelity'}
                if($fact.Control -ne ($i -lt 2)){throw 'Control identity'}
                if($fact.Control -and -not $fact.RereadDescriptorRetained){throw 'Positive control failed'}
                if($i -eq 0 -and -not ($fact.HeldViewSameOnReread -and $fact.MemberCollectionSameOnReread -and $fact.PostViewSameAsReread)){throw 'Held control view identity'}
                if($fact.MemberCollectionSameOnReread -and -not $fact.RereadDescriptorRetained){throw 'Inconsistent member identity'}
                if($fact.RereadDescriptorRetained){
                    if($fact.Capability -cne 'RETAINED' -or $fact.RereadPayloadSame -ne $true -or $fact.RereadMutationRetained -ne $true){throw 'Retained fact consistency'}
                }elseif($fact.Capability -cne 'NOT_RETAINED' -or $null -ne $fact.RereadPayloadSame -or $null -ne $fact.RereadMutationRetained){throw 'Absence fact consistency'}
            }
            $row.SchemaValidated=$true;$row.DiagnosticComplete=$true;$row.Outcome='DIAGNOSTIC_COMPLETE'
        }catch{$row.ControllerError=$_.Exception.Message}
    }
    if($lines.Count -gt 0 -and $lines[-1] -cmatch '^RESULT: FIXTURE_INVALID case=([a-z0-9-]+) detail=([a-z0-9-]+) productCalls=0$'){
        $row.ChildOutcome='FIXTURE_INVALID'
        if($valid -and $row.ExitCode -eq 1){$row.Outcome='FIXTURE_INVALID'}
    }
}
Save-Index
[pscustomobject]$row|ConvertTo-Json -Depth 12 -Compress
if(-not $row.DiagnosticComplete){throw 'STOP: diagnostic invalid/incomplete; attempt spent, no retry'}
