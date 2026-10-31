param([Parameter(Mandatory)][ValidateSet('RunOnce')][string]$Action)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$packet=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-record-chain-diagnosis'
$oldPacket=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-review-repair'
$seed=Join-Path $oldPacket seed.frozen.json
$seedHash='9091825652E469001E51D8A510B12E0F84013C38509D740C649942DC2EAAFF6B'
$engine='C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe'
$engineHash='362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139'
$sma=Join-Path ([IO.Path]::GetDirectoryName($engine)) System.Management.Automation.dll
$smaHash='2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B'
$plan=Join-Path $repo 'docs/superpowers/plans/2026-10-31-offline-record-chain-diagnosis.md'
$names=@('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Capture-Failure')
function RCD-Need([bool]$Value,[string]$Message){if(-not $Value){throw $Message}}
function RCD-Hash([string]$Path){(Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash}
function RCD-Once([string]$Path,[byte[]]$Bytes){
    $stream=$null;$primary=$null
    try{$stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None);$stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()}
    catch{$primary=$_}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{if($null -eq $primary){$primary=$_}else{throw}}}}
    if($null -ne $primary){throw $primary}
}
function RCD-JSON([string]$Path,$Value){RCD-Once $Path ([Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 35)))}
function RCD-Verify($Pins){foreach($pin in $Pins){RCD-Need ((RCD-Hash $pin.Path) -ceq $pin.SHA256) ('Input drift: '+$pin.Path)}}
RCD-Need (-not (Test-Path -LiteralPath $packet)) 'Diagnostic packet already exists; no retry'
RCD-Need ((RCD-Hash $plan) -ceq '953AEDFEFA0C2BF3C48ED8E3064696C76729764D512AD09AE4D6982FDCC4D76A') 'Approved diagnostic plan identity'
RCD-Need ((RCD-Hash $seed) -ceq $seedHash) 'Seed hash'
RCD-Need ((RCD-Hash $engine) -ceq $engineHash -and (RCD-Hash $sma) -ceq $smaHash) 'Pinned engine/SMA'
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
RCD-Need ($identity.Name -ceq 'DESKTOP-H4SVCLU\Lenovo' -and -not ([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) 'Fixed non-admin identity'
$manifest=Get-Content -LiteralPath $seed -Raw|ConvertFrom-Json
RCD-Need ($manifest.Schema -eq 2 -and $manifest.Pins.Count -eq 2317) 'Exact seed inventory'
$pins=[Collections.Generic.List[object]]::new()
foreach($pin in $manifest.Pins){RCD-Need ($pin.ExpectedState -ceq 'Present') 'Present pins only';$pins.Add([ordered]@{Path=$pin.VerificationPath;SHA256=$pin.SHA256})}
$sources=@('README.md','observe.ps1','driver.ps1')
$dependencies=@($sources|ForEach-Object {Join-Path $PSScriptRoot $_})+@($seed,$engine,$sma,$plan,(Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-review-repair/portable-stopped.json'))+@(Get-ChildItem -LiteralPath $oldPacket -Recurse -File|Sort-Object FullName|ForEach-Object {$_.FullName})
foreach($path in $dependencies){$pins.Add([ordered]@{Path=$path;SHA256=(RCD-Hash $path)})}
foreach($file in @('observe.ps1','driver.ps1')){
    $tokens=$null;$errors=$null
    $null=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot $file),[ref]$tokens,[ref]$errors)
    RCD-Need ($errors.Count -eq 0) ('Diagnostic AST invalid: '+$file)
}
RCD-Verify $pins
$null=[IO.Directory]::CreateDirectory($packet)
$snapshot=Join-Path $packet snapshots
$null=[IO.Directory]::CreateDirectory($snapshot)
foreach($file in $sources){RCD-Once (Join-Path $snapshot $file) ([IO.File]::ReadAllBytes((Join-Path $PSScriptRoot $file)))}
RCD-Once (Join-Path $packet approved-plan.snapshot.md) ([IO.File]::ReadAllBytes($plan))
RCD-JSON (Join-Path $packet pins.json) ($pins.ToArray())
RCD-JSON (Join-Path $packet intent.json) ([ordered]@{Action=$Action;Authorization='User answered yes to the proposed single scope-only stage after aac17117';PlanSHA256=(RCD-Hash $plan);SeedSHA256=$seedHash;StartAttemptAllowance=1;Targets=0;Compilations=0;Reviews=0;OriginalRepairRemainder='FROZEN';NoWriterInvocation=$true;NoCollectorInvocation=$true})
RCD-Once (Join-Path $packet progress.md) ([Text.UTF8Encoding]::new($false).GetBytes(@'
# SDD ledger — plan: docs/superpowers/plans/2026-10-31-offline-record-chain-diagnosis.md

Baseline aac17117cbcefb40db8b6496c2bd70d6d110528d; user explicitly approved single scope-only stage.
Pre-flight: Task 1 produces actual packet/trace; Task 2 preserves and explains it, never reruns it.
Ruling: Approved finite native diagnostic overrides skill-default setup/Bash/TDD/full-suite/task-done reruns/review/deletion; no product fix is performed. Cost: observer is not regression-verified and findings are limited to this observation.
Existing linked development worktree verified; no installs or baseline runs.
Task 1: diagnostic inputs snapshotted once; seed inputs and stopped packet pinned; launch at most once.
Original repair remainder 6 hosts / 7 targets / 1 compile / conditional 1 review remains FROZEN.
'@))
$row=[ordered]@{Started=$false;PID=$null;OwnedExit=$false;StreamsComplete=$false;ExitCode=$null;TimedOut=$false;TerminationRequested=$false;Errors=@();PostPinsIntact=$false;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
$errors=[Collections.Generic.List[object]]::new()
function RCD-Error($Record){$errors.Add([ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace})}
$process=[Diagnostics.Process]::new();$out=$null;$err=$null;$ot=$null;$et=$null
$timer=[Diagnostics.Stopwatch]::StartNew()
try{
    $out=[IO.File]::Open((Join-Path $packet stdout.bin),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $err=[IO.File]::Open((Join-Path $packet stderr.bin),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    $process.StartInfo=[Diagnostics.ProcessStartInfo]::new()
    $process.StartInfo.FileName=$engine;$process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
    $process.StartInfo.RedirectStandardOutput=$true;$process.StartInfo.RedirectStandardError=$true;$process.StartInfo.WorkingDirectory=$repo
    foreach($arg in @('-NoLogo','-NoProfile','-NonInteractive','-File',(Join-Path $PSScriptRoot observe.ps1),'-PacketRoot',$packet,'-SeedSHA256',$seedHash)){$process.StartInfo.ArgumentList.Add($arg)}
    # Persist consumption before Start; failure never refunds or retries.
    RCD-JSON (Join-Path $packet start-attempt.json) ([ordered]@{AttemptConsumed=$true;MaxAttempts=1})
    if(-not $process.Start()){throw 'Diagnostic host did not start'}
    $row.Started=$true;$row.PID=$process.Id
    $ot=$process.StandardOutput.BaseStream.CopyToAsync($out);$et=$process.StandardError.BaseStream.CopyToAsync($err)
    if(-not $process.WaitForExit([Math]::Max(0,(120000-[int]$timer.ElapsedMilliseconds)))){$row.TimedOut=$true}
}catch{RCD-Error $_}
finally{
    $associated=$false
    try{$property=[Diagnostics.Process].GetProperty('Id',([Reflection.BindingFlags]::Public -bor [Reflection.BindingFlags]::Instance));$id=$property.GetValue($process);if($id -is [int] -and $id -gt 0){$associated=$true;$row.PID=$id}}
    catch{if($row.Started){RCD-Error $_}}
    if($associated){
        $exited=$false
        try{$exited=$process.HasExited}catch{RCD-Error $_}
        if(-not $exited){$row.TerminationRequested=$true;try{$process.Kill()}catch{RCD-Error $_}}
        try{$row.OwnedExit=$process.WaitForExit(2000);if($row.OwnedExit){$row.ExitCode=$process.ExitCode}}catch{RCD-Error $_}
    }
    if($null -ne $ot -and $null -ne $et){try{$row.StreamsComplete=[Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($ot,$et),2000)}catch{RCD-Error $_}}
    foreach($stream in @($out,$err)){if($null -ne $stream){try{$stream.Dispose()}catch{RCD-Error $_}}}
    try{$process.Dispose()}catch{RCD-Error $_};$timer.Stop()
}
try{RCD-Verify $pins;$row.PostPinsIntact=$true}catch{RCD-Error $_}
$row.Errors=$errors.ToArray()
# Save actual exit/drain/raw metadata before parsing the trace.
RCD-JSON (Join-Path $packet runtime.final.json) $row
$assessment=[ordered]@{TraceState='UNKNOWN';CommandShadowing='UNKNOWN';SentinelVisibility='UNKNOWN';ChangedCommands=@();Reason=$null;HistoricalFirstCause='UNKNOWN';OriginalRepairRemainder='FROZEN';SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
try{
    RCD-Need ($row.Started -and $row.OwnedExit -and $row.StreamsComplete -and $row.ExitCode -eq 0 -and -not $row.TimedOut -and -not $row.TerminationRequested -and $row.PostPinsIntact -and $errors.Count -eq 0) 'Runtime/exit/drain/pins incomplete'
    RCD-Need ((Get-Item -LiteralPath (Join-Path $packet stdout.bin)).Length -eq 0 -and (Get-Item -LiteralPath (Join-Path $packet stderr.bin)).Length -eq 0) 'Unexpected raw output'
    $tracePath=Join-Path $packet trace.jsonl
    RCD-Need ((Get-Item -LiteralPath $tracePath).Length -gt 0 -and (Get-Item -LiteralPath $tracePath).Length -le 1048576) 'Trace size'
    $stages=@('outer-before','module-created','module-scriptblock','outer-after','complete')
    $rows=@(Get-Content -LiteralPath $tracePath|ForEach-Object {ConvertFrom-Json -InputObject $_ -AsHashtable})
    RCD-Need ($rows.Count -eq 5) 'Exact stage count'
    $stateKeys=@('PrimaryPresent','PrimaryNull','SecondaryPresent','SecondaryCount','SentinelPresent','SentinelValue')
    for($i=0;$i -lt 5;$i++){
        $r=$rows[$i]
        RCD-Need (($r.Keys -join ',') -ceq 'Sequence,Stage,Commands,OuterState,ModuleState' -and $r.Sequence -is [long] -and $r.Sequence -eq $i -and $r.Stage -ceq $stages[$i]) 'Ordered exact stage schema'
        RCD-Need ($r.Commands.Count -eq 4) 'Four command observations'
        for($j=0;$j -lt 4;$j++){
            $c=$r.Commands[$j]
            RCD-Need (($c.Keys -join ',') -ceq 'Name,SameOriginal,ModuleName,ModuleGuid' -and $c.Name -ceq $names[$j] -and $c.SameOriginal -is [bool] -and $c.ModuleName -is [string]) 'Command schema/type/order'
            RCD-Need (($null -eq $c.ModuleGuid -and $c.ModuleName -ceq '') -or ($c.ModuleGuid -is [string] -and $c.ModuleName.Length -gt 0 -and $c.ModuleGuid -match '^[a-fA-F0-9-]{36}$')) 'Module identity projection'
            if($i -eq 0){RCD-Need ($c.SameOriginal -and $c.ModuleName -ceq '' -and $null -eq $c.ModuleGuid) 'Outer baseline identity'}
        }
        foreach($side in @('OuterState','ModuleState')){
            $s=$r[$side]
            if($i -eq 0 -and $side -ceq 'ModuleState'){RCD-Need ($null -eq $s) 'No module before creation';continue}
            RCD-Need (($s.Keys -join ',') -ceq ($stateKeys -join ',')) 'State schema'
            foreach($key in @('PrimaryPresent','SecondaryPresent','SentinelPresent')){RCD-Need ($s[$key] -is [bool]) 'State boolean'}
            RCD-Need (($s.PrimaryPresent -and $s.PrimaryNull -is [bool] -and $s.PrimaryNull) -or (-not $s.PrimaryPresent -and $null -eq $s.PrimaryNull)) 'Primary projection'
            RCD-Need (($s.SecondaryPresent -and $s.SecondaryCount -is [long] -and $s.SecondaryCount -eq 0) -or (-not $s.SecondaryPresent -and $null -eq $s.SecondaryCount)) 'Secondary projection'
            RCD-Need (($s.SentinelPresent -and $s.SentinelValue -ceq 'scope-only-20261031') -or (-not $s.SentinelPresent -and $null -eq $s.SentinelValue)) 'Sentinel projection'
        }
    }
    $assessment.ChangedCommands=@($rows[1].Commands|Where-Object {-not $_.SameOriginal}|ForEach-Object {$_.Name})
    $assessment.CommandShadowing=if($assessment.ChangedCommands.Count -gt 0){'observed'}else{'not-observed'}
    $outer=$rows[3].OuterState.SentinelPresent;$inner=$rows[3].ModuleState.SentinelPresent
    $assessment.SentinelVisibility=if($outer){'outer-visible'}elseif($inner){'module-only'}else{'other'}
    $assessment.TraceState='COMPLETE'
}catch{$assessment.Reason=[ordered]@{Message=$_.Exception.Message;ErrorId=$_.FullyQualifiedErrorId;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}}
RCD-JSON (Join-Path $packet assessment.json) $assessment
[ordered]@{Runtime=$row;Assessment=$assessment;Packet=$packet}|ConvertTo-Json -Depth 15
if($assessment.TraceState -cne 'COMPLETE'){throw 'Diagnostic UNKNOWN; no rerun/repair/resume'}
