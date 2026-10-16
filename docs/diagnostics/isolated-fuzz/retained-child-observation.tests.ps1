$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
$helper=Join-Path $PSScriptRoot 'retained-child-observation.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
function Check-RetainedChild([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
Check-RetainedChild ($null -ne (Get-Command Get-RetainedChildObservationPlan -ErrorAction SilentlyContinue)) 'dedicated retained observation entry exists'
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
$unit=Join-Path $scope ('retained-child-unit-'+[guid]::NewGuid().ToString('N'))
$root=Join-Path $unit 'retained-child-observe-901'
$plan=Get-RetainedChildObservationPlan $root
Check-RetainedChild (-not (Test-Path -LiteralPath $root) -and $plan.LimitSeconds -eq 20 -and $plan.ChildDelayMs -eq 150 -and $plan.ReadyLimitMs -eq 10000 -and $plan.DiagnosticOnly -and -not $plan.Acceptance) 'planning is read only with fixed diagnostic limits'
Check-RetainedChild ($plan.SourceBinding.Count -eq 8 -and @($plan.SourceBinding.Path) -contains 'compile-stage-trigger.cs' -and @($plan.SourceBinding.Path) -contains 'compile-stage-session.cs' -and $plan.ExecutableSHA256 -match '^[A-F0-9]{64}$') 'plan binds eight transitive production sources and existing shell'
Check-RetainedChild (($plan.Arguments[0..2] -join '|') -ceq '-NoProfile|-NonInteractive|-EncodedCommand' -and [Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($plan.Arguments[3])) -ceq $plan.ParentScript) 'fixed encoded parent body survives argument construction'
foreach($invalid in @((Join-Path $unit 'short-child-observe-901'),(Join-Path $workspace 'retained-child-observe-901'))){
    $caught=$false;try{Get-RetainedChildObservationPlan $invalid|Out-Null}catch{$caught=$true}
    Check-RetainedChild $caught ('wrong scope or identity refused: '+[IO.Path]::GetFileName($invalid))
}
New-Item -ItemType Directory -Path $root -Force|Out-Null
$caught=$false;try{Get-RetainedChildObservationPlan $root|Out-Null}catch{$caught=$true}
Check-RetainedChild $caught 'planning refuses consumed identity'
# Execute fixture bodies in this process only, replacing their process/sleep boundary.
$calls=New-Object 'System.Collections.Generic.List[object]'
function Start-Process {
    param($FilePath,$ArgumentList,$WindowStyle)
    $calls.Add([pscustomobject]@{Executable=$FilePath;Arguments=$ArgumentList;WindowStyle=$WindowStyle})
    [pscustomobject]@{Nonce=$plan.Nonce;ProcessId=23;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $root 'child-ready.json')
}
& ([scriptblock]::Create($plan.ParentScript))
Check-RetainedChild ($calls.Count -eq 1 -and $calls[0].Executable -ceq $plan.Executable -and $calls[0].WindowStyle -ceq 'Hidden' -and [Text.Encoding]::Unicode.GetString([Convert]::FromBase64String(($calls[0].Arguments -split ' ')[-1])) -ceq $plan.ChildScript) 'parent uses exactly one fixed hidden child boundary'
$release=Get-Content (Join-Path $root 'release.json') -Raw|ConvertFrom-Json
Check-RetainedChild ($release.Nonce -ceq $plan.Nonce -and $release.ProcessId -eq $PID) 'parent release preserves nonce and self identity'
$childRoot=Join-Path $unit 'retained-child-observe-902';$childPlan=Get-RetainedChildObservationPlan $childRoot
New-Item -ItemType Directory -Path $childRoot|Out-Null
[pscustomobject]@{Nonce=$childPlan.Nonce;ProcessId=$PID;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $childRoot 'release.json')
$sleeps=New-Object 'System.Collections.Generic.List[int]'
function Start-Sleep {param([int]$Milliseconds);$sleeps.Add($Milliseconds)}
& ([scriptblock]::Create($childPlan.ChildScript))
Check-RetainedChild ($sleeps.Count -eq 1 -and $sleeps[0] -eq 150 -and (Test-Path (Join-Path $childRoot 'child-exit.json'))) 'child records intent after one requested 150ms delay'
# Complete literal synthetic native DTO: no OS job, PID lookup or native Run.
function New-RetainedChildNative {
    $n=New-Object 'IsolatedFuzzProcess+Result'
    $n.RootId=17;$n.StartUTC='2026-10-16T00:00:00Z';$n.EndUTC='2026-10-16T00:00:10Z'
    $n.RootExited=$true;$n.CleanupVerified=$true;$n.NaturalShutdownVerified=$true;$n.ShutdownEvidenceComplete=$true
    $n.ShutdownDiagnosticOnly=$true;$n.LifetimeDiagnosticOnly=$true;$n.NaturalObservationMs=200;$n.ShutdownElapsedMs=210
    # Root-wait facts are intentionally nonzero; final accounting/list are empty.
    # A schema3 receipt must not confuse these distinct lifecycle phases.
    $n.TotalProcesses=2;$n.ActiveAtRootWait=1;$n.DescendantsRemaining=1
    $n.MembersAtRootWait=[uint64[]]@(23);$n.MembersBeforeTermination=[uint64[]]@();$n.MembersAfterCleanup=[uint64[]]@();$n.MemberEvidence=@();$n.ShutdownMemberEvidence=@()
    $coreRow=New-Object 'IsolatedFuzzProcess+MemberObservation';$coreRow.Id=23;$coreRow.Owned=$true;$coreRow.ExitCodeRead=$true;$coreRow.State='exited'
    $n.MemberEvidence=@($coreRow);$n.ShutdownMemberEvidence=@($coreRow)
    $d=New-Object LifetimeReport;$n.LifetimeDiagnostics=$d
    $d.Attempts=1;$d.Admitted=1;$d.QueryCount=2;$d.RootExited=$true;$d.TeardownAttempted=$true;$d.ReleaseAttempts=1
    $e=New-Object LifetimeEntry;$e.EntryId=1;$e.Pid=23;$e.SourceQuery=1
    $q=New-Object LifetimeQueryEvidence;$q.Succeeded=$true;$q.Assigned=2;$q.Returned=2;$q.Capacity=32;$q.BufferBytes=264;$q.Frequency=1000;$q.StartTimestamp=100;$q.EndTimestamp=110;$q.StartUTC='2026-10-16T00:00:00.100Z';$q.EndUTC='2026-10-16T00:00:00.110Z'
    $d.FirstQuery=$q;$d.LastQuery=$q;$e.SourceQueryEvidence=$q
    $e.Open=New-Object LifetimeApiEvidence;$e.Open.Succeeded=$true;$e.Open.Frequency=1000;$e.Open.StartTimestamp=120;$e.Open.EndTimestamp=130;$e.Open.StartUTC='2026-10-16T00:00:00.120Z';$e.Open.EndUTC='2026-10-16T00:00:00.130Z'
    $e.Membership=New-Object LifetimeMembership;$e.Membership.Succeeded=$true;$e.Membership.Owned=$true;$e.Membership.Frequency=1000;$e.Membership.StartTimestamp=140;$e.Membership.EndTimestamp=150;$e.Membership.StartUTC='2026-10-16T00:00:00.140Z';$e.Membership.EndUTC='2026-10-16T00:00:00.150Z'
    foreach($phase in @('Admission','RootWaitEnd','Final')){
        $r=New-Object LifetimeRead;$r.Attempted=$true;$r.SameEntry=$true;$r.ExitCodeRead=$true;$r.Frequency=1000
        $r.State=if($phase -ceq 'Final'){'exited'}else{'running'};$r.ExitCode=if($phase -ceq 'Final'){0}else{259}
        $r.StartTimestamp=switch($phase){Admission{160} RootWaitEnd{300} Final{500}};$r.EndTimestamp=$r.StartTimestamp+10
        $r.StartUTC=([DateTimeOffset]'2026-10-16T00:00:00Z').AddMilliseconds($r.StartTimestamp).ToString('o');$r.EndUTC=([DateTimeOffset]'2026-10-16T00:00:00Z').AddMilliseconds($r.EndTimestamp).ToString('o')
        $e.$phase=$r
    }
    $d.Entries.Add($e);return $n
}
function Write-RetainedStages($Plan){
    foreach($name in @('child-ready.json','release.json','parent-exit.json','child-exit.json')){
        $id=if($name.StartsWith('child')){23}else{17}
        [pscustomobject]@{Nonce=$Plan.Nonce;ProcessId=$id;UTC='2026-10-16T00:00:00.200Z'}|ConvertTo-Json -Compress|Set-Content -LiteralPath (Join-Path $Plan.RunRoot $name)
    }
}
Write-RetainedStages $plan
$native=New-RetainedChildNative
$receipt=ConvertTo-RecordedProcessReceipt $native 'out' 'err' 'fixture' @() @{}
$goodJSON=$receipt|ConvertTo-Json -Depth 6 -WarningAction Stop
$observedReceipt=$goodJSON|ConvertFrom-Json
$good=Get-RetainedChildObservationEvidence $plan $observedReceipt
Check-RetainedChild ($good.ObservationPassed -and $good.CoreNaturalComplete) 'root wait descendants do not imply final residual'
Check-RetainedChild ($observedReceipt.DescendantsRemaining -eq 1 -and $observedReceipt.ActiveAtRootWait -eq 1 -and $observedReceipt.MembersAtRootWait[0] -eq 23 -and $observedReceipt.ActiveAfterCleanup -eq 0 -and $observedReceipt.MembersAfterCleanup.Count -eq 0) 'observation preserves distinct root wait and final facts'
Check-RetainedChild ($good.ObservationPassed -and $good.StageEvidenceComplete -and $good.LifetimeEvidenceComplete -and $good.CoreNaturalComplete -and $good.ChildEntryId -eq 1 -and -not $good.Acceptance -and -not $good.OriginalTestPassed -and $good.Coverage -ceq 'sampled') 'complete same-entry fixture stays sampled and non acceptance'
$refused=$false;try{Assert-ProcessCompletion ($goodJSON|ConvertFrom-Json)}catch{$refused=$true}
Check-RetainedChild $refused 'ordinary guard still refuses observed diagnostic receipt'
# Each mutation names a wrong success branch the real observer must reject.
foreach($mutation in @('missing','wrong-child','duplicate','ambiguous','budget','close','forced','timeout','open87','source','clock','phase','final-code','flags','core87','core-missing','core-unowned','core-raw87','core-duplicate','core-running','cap','error','raw-handle','final-active','final-list','pretermination-active','pretermination-list','natural-overrun','cleanup-overrun','cleanup-unverified')){
    $bad=$goodJSON|ConvertFrom-Json;$d=$bad.LifetimeDiagnostics;$e=$d.Entries[0]
    switch($mutation){
        missing{$d.Entries=@();$d.Attempts=0;$d.Admitted=0;$d.ReleaseAttempts=0}
        wrong-child{$e.Pid=24}
        duplicate{$d.Entries=@($e,$e);$d.Attempts=2;$d.Admitted=2;$d.ReleaseAttempts=2}
        ambiguous{$e.RootWaitEnd.Ambiguous=$true}
        budget{$e.Final.Attempted=$false;$e.Final.SkipReason='budget-exhausted'}
        close{$d.ReleaseFailures=1}
        forced{$bad.TerminationRequested=$true;$e.Final.TerminationRequested=$true}
        timeout{$bad.TimedOut=$true}
        open87{$e.Open.Succeeded=$false;$e.Open.Error=87}
        source{$e.SourceQueryEvidence=$null}
        clock{$e.Membership.EndTimestamp=119}
        phase{$e.RootWaitEnd=$null}
        final-code{$e.Final.ExitCode=7}
        flags{$bad.LifetimeDiagnosticOnly=$false}
        core87{$bad.NaturalShutdownVerified=$false;$bad.ShutdownEvidenceComplete=$false;$bad.MemberEvidence=@([pscustomobject]@{Id=99;Owned=$false;QueryError=87})}
        core-missing{$bad.ShutdownMemberEvidence=@()}
        core-unowned{$bad.ShutdownMemberEvidence[0].Owned=$false}
        core-raw87{$bad.ShutdownMemberEvidence[0].QueryError=87}
        core-duplicate{$bad.ShutdownMemberEvidence=@($bad.ShutdownMemberEvidence[0],$bad.ShutdownMemberEvidence[0])}
        core-running{$bad.ShutdownMemberEvidence[0].State='running'}
        cap{$d.Attempts=33}
        error{$d.Incomplete=$true;$d.Errors=@('list-overrun:-1')}
        raw-handle{$e|Add-Member Handle 99}
        final-active{$bad.ActiveAfterCleanup=1}
        final-list{$bad.MembersAfterCleanup=@(23)}
        pretermination-active{$bad.ActiveBeforeTermination=1}
        pretermination-list{$bad.MembersBeforeTermination=@(23)}
        natural-overrun{$bad.NaturalObservationMs=501}
        cleanup-overrun{$bad.ShutdownElapsedMs=5001}
        cleanup-unverified{$bad.CleanupVerified=$false}
    }
    $observed=Get-RetainedChildObservationEvidence $plan $bad
    Check-RetainedChild (-not $observed.ObservationPassed -and -not $observed.Acceptance) ('observer refuses '+$mutation)
}
[pscustomobject]@{Nonce=('0'*32);ProcessId=23;UTC='2026-10-16T00:00:00.200Z'}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $root 'child-exit.json')
Check-RetainedChild (-not (Get-RetainedChildObservationEvidence $plan ($goodJSON|ConvertFrom-Json)).StageEvidenceComplete) 'wrong stage nonce cannot supply coverage'
Write-RetainedStages $plan
[pscustomobject]@{Nonce=$plan.Nonce;ProcessId=23;UTC='2026-10-17T00:00:00Z'}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $root 'child-exit.json')
Check-RetainedChild (-not (Get-RetainedChildObservationEvidence $plan ($goodJSON|ConvertFrom-Json)).ObservationPassed) 'stage outside receipt interval cannot pass'
function Test-RetainedController([string]$Mode,[int]$Suffix){
    $controllerRoot=Join-Path $unit ('retained-child-observe-'+$Suffix)
    $nativeCalls=New-Object 'System.Collections.Generic.List[object]'
    # Only replace the unavoidable external native process boundary; conversion,
    # validation, hashing and persistence remain the real implementation.
    function Invoke-RetainedChildObservationNative($Plan,$Stdout,$Stderr){
        $nativeCalls.Add($Plan)
        if($Plan.LimitSeconds -ne 20 -or ($Plan.Arguments[0..2] -join '|') -cne '-NoProfile|-NonInteractive|-EncodedCommand' -or (Split-Path -Parent $Stdout) -cne (Join-Path $Plan.RunRoot 'process')){throw 'Wrong native launch contract'}
        [IO.File]::WriteAllText($Stdout,'fixture stdout');[IO.File]::WriteAllText($Stderr,'fixture stderr')
        if($Mode -ceq 'throw'){throw 'synthetic launch exception'}
        Write-RetainedStages $Plan
        $n=New-RetainedChildNative
        if($Mode -ceq 'refuse'){$n.LifetimeDiagnostics.ReleaseFailures=1}
        return $n
    }
    $failure=$null;$returned=$null
    try{$returned=Invoke-RetainedChildObservation $controllerRoot}catch{$failure=$_.Exception.Message}
    Check-RetainedChild ($nativeCalls.Count -eq 1) ('one native boundary call only: '+$Mode)
    $manifest=Get-Content (Join-Path $controllerRoot 'evidence-manifest.json') -Raw|ConvertFrom-Json
    $hashesOK=$manifest.Files.Count -ge 5
    foreach($row in $manifest.Files){$path=Join-Path $controllerRoot $row.Path;if((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash -cne $row.SHA256 -or (Get-Item -LiteralPath $path).Length -ne $row.Length){$hashesOK=$false}}
    Check-RetainedChild ($hashesOK -and -not $manifest.Acceptance -and $manifest.NoRetry) ('persisted artifact hashes match bytes: '+$Mode)
    $savedPlan=Get-Content (Join-Path $controllerRoot 'plan.json') -Raw|ConvertFrom-Json
    Check-RetainedChild ($savedPlan.SourceBinding.Count -eq 8 -and @($savedPlan.SourceBinding.Path) -contains 'compile-stage-trigger.cs' -and @($savedPlan.SourceBinding.Path) -contains 'compile-stage-session.cs' -and -not $savedPlan.Acceptance) ('recoverable source binding saved: '+$Mode)
    if($Mode -ceq 'pass'){
        $saved=Get-Content (Join-Path $controllerRoot 'result.json') -Raw|ConvertFrom-Json
        $savedReceipt=Get-Content (Join-Path $controllerRoot 'process/receipt.json') -Raw|ConvertFrom-Json
        Check-RetainedChild ($null -eq $failure -and $returned.ObservationPassed -and $saved.ObservationPassed -and $savedReceipt.DescendantsRemaining -eq 1 -and $savedReceipt.LifetimeDiagnosticOnly -and $savedReceipt.LifetimeDiagnostics.Entries[0].Final.State -ceq 'exited' -and -not $saved.Acceptance) 'controller persists native diagnostic metadata without acceptance'
        $completed=Get-Content (Join-Path $controllerRoot 'completed.json') -Raw|ConvertFrom-Json
        Check-RetainedChild ($saved.FinalizationRequired -and $returned.Finalized -and $completed.Finalized -and $completed.ObservationPassed -and -not $completed.Acceptance -and $completed.ManifestSHA256 -ceq (Get-FileHash -LiteralPath (Join-Path $controllerRoot 'evidence-manifest.json') -Algorithm SHA256).Hash) 'durable completion separately binds finalized artifact manifest'
    }else{
        $stopped=Get-Content (Join-Path $controllerRoot 'stopped.json') -Raw|ConvertFrom-Json
        Check-RetainedChild ($null -ne $failure -and $stopped.NoRetry -and -not $stopped.Acceptance) ('failure recorded without retry: '+$Mode)
        if($Mode -ceq 'refuse'){Check-RetainedChild (-not (Get-Content (Join-Path $controllerRoot 'result.json') -Raw|ConvertFrom-Json).ObservationPassed) 'refused result preserved before throwing'}
        else{Check-RetainedChild (Test-Path (Join-Path $controllerRoot 'process/launch-error.json')) 'launch exception never invents a receipt'}
    }
    $duplicate=$false;try{Invoke-RetainedChildObservation $controllerRoot|Out-Null}catch{$duplicate=$true}
    Check-RetainedChild ($duplicate -and $nativeCalls.Count -eq 1) ('consumed identity never reopens native boundary: '+$Mode)
}
Test-RetainedController 'pass' 903
Test-RetainedController 'refuse' 904
Test-RetainedController 'throw' 905
function Test-RetainedFinalizationFailure([string]$Mode,[int]$Suffix){
    $target=Join-Path $unit ('retained-child-observe-'+$Suffix)
    $locks=New-Object 'System.Collections.Generic.List[IDisposable]'
    $count=[pscustomobject]@{Calls=0}
    function Invoke-RetainedChildObservationNative($Plan,$Stdout,$Stderr){
        $count.Calls++;Write-RetainedStages $Plan
        [IO.File]::WriteAllText($Stdout,'stdout');[IO.File]::WriteAllText($Stderr,'stderr')
        if($Mode -ceq 'manifest-write'){New-Item -ItemType Directory -Path (Join-Path $Plan.RunRoot 'evidence-manifest.json')|Out-Null}
        else{$locks.Add([IO.File]::Open($Stdout,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::None))}
        return (New-RetainedChildNative)
    }
    $failure=$null;try{Invoke-RetainedChildObservation $target|Out-Null}catch{$failure=$_.Exception.Message}finally{foreach($stream in $locks){$stream.Dispose()}}
    Check-RetainedChild ($null -ne $failure -and $count.Calls -eq 1) ('real finalization filesystem failure stops: '+$Mode)
    Check-RetainedChild (Test-Path -LiteralPath (Join-Path $target 'stopped.json') -PathType Leaf) ('finalization failure persists stopped state: '+$Mode)
    $stopped=Get-Content (Join-Path $target 'stopped.json') -Raw|ConvertFrom-Json
    $provisional=Get-Content (Join-Path $target 'result.json') -Raw|ConvertFrom-Json
    Check-RetainedChild ($stopped.NoRetry -and -not $stopped.Acceptance -and $provisional.FinalizationRequired -and -not (Test-Path (Join-Path $target 'completed.json'))) ('unfinalized result is never durable success: '+$Mode)
}
Test-RetainedFinalizationFailure 'manifest-write' 906
Test-RetainedFinalizationFailure 'hash-read' 907
Write-Output 'LIMIT: fake native/process boundaries only; no child, PID query, Go or matrix execution.'
