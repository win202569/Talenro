. (Join-Path $PSScriptRoot 'run.ps1')
# Explicit entry only: dot-sourcing never launches an observation.
function Get-RetainedChildObservationRoot([string]$RunRoot) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^retained-child-observe-[0-9]{3}$'){throw 'Owned retained-child identity required'}
    return $root
}
function Get-RetainedChildObservationPlan([string]$RunRoot) {
    $root=Get-RetainedChildObservationRoot $RunRoot
    if(Test-Path -LiteralPath $root){throw 'New retained-child identity required; no replay'}
    $shell=Get-IsolatedFullPath (Join-Path ([Environment]::GetFolderPath('System')) 'WindowsPowerShell/v1.0/powershell.exe')
    if(-not (Test-Path -LiteralPath $shell -PathType Leaf)){throw 'Existing Windows PowerShell required'}
    $nonce=[guid]::NewGuid().ToString('N')
    # Keep historical short-child helper semantics untouched. Markers record
    # readiness/release/exit intent, never OS exit times or guaranteed capture.
    $common=@'
$ErrorActionPreference='Stop'
$fixtureRoot='__ROOT__'
$fixtureNonce='__NONCE__'
function Write-Stage([string]$Name) {
    $text=[pscustomobject]@{Nonce=$fixtureNonce;ProcessId=$PID;UTC=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Compress
    $target=Join-Path $fixtureRoot $Name
    [IO.File]::WriteAllText(($target+'.tmp'),$text)
    [IO.File]::Move(($target+'.tmp'),$target)
}
'@
    $common=$common.Replace('__ROOT__',$root.Replace("'","''")).Replace('__NONCE__',$nonce)
    $child=$common+@'

Write-Stage 'child-ready.json'
$clock=[Diagnostics.Stopwatch]::StartNew()
while(-not [IO.File]::Exists((Join-Path $fixtureRoot 'release.json'))) {
    if($clock.ElapsedMilliseconds -ge 10000){throw 'Child release deadline exceeded'}
    Start-Sleep -Milliseconds 5
}
$release=[IO.File]::ReadAllText((Join-Path $fixtureRoot 'release.json'))|ConvertFrom-Json
if($release.Nonce -cne $fixtureNonce){throw 'Wrong release nonce'}
Start-Sleep -Milliseconds 150
Write-Stage 'child-exit.json'
'@
    $encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($child))
    $parent=$common+@'

Start-Process -FilePath '__SHELL__' -ArgumentList '-NoProfile -NonInteractive -EncodedCommand __CHILD__' -WindowStyle Hidden
$clock=[Diagnostics.Stopwatch]::StartNew()
while(-not [IO.File]::Exists((Join-Path $fixtureRoot 'child-ready.json'))) {
    if($clock.ElapsedMilliseconds -ge 10000){throw 'Parent readiness deadline exceeded'}
    Start-Sleep -Milliseconds 5
}
$ready=[IO.File]::ReadAllText((Join-Path $fixtureRoot 'child-ready.json'))|ConvertFrom-Json
if($ready.Nonce -cne $fixtureNonce -or $ready.ProcessId -le 0 -or $ready.ProcessId -eq $PID){throw 'Wrong child readiness identity'}
Write-Stage 'release.json'
Write-Stage 'parent-exit.json'
'@
    $parent=$parent.Replace('__SHELL__',$shell.Replace("'","''")).Replace('__CHILD__',$encoded)
    $binding=@(foreach($name in @('retained-child-observation.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs')){
        $path=Get-IsolatedFullPath (Join-Path $PSScriptRoot $name)
        if([IsolatedFuzzFileInfo]::Links($path) -ne 1){throw 'Source link not independent'}
        [pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash}
    })
    [pscustomobject]@{Schema=1;RunRoot=$root;Nonce=$nonce;DiagnosticOnly=$true;Acceptance=$false;Coverage='sampled';NoRetry=$true;Executable=$shell;ExecutableSHA256=(Get-FileHash -LiteralPath $shell -Algorithm SHA256).Hash;SourceBinding=$binding;LimitSeconds=20;ChildDelayMs=150;ReadyLimitMs=10000;ParentScript=$parent;ChildScript=$child;Arguments=@('-NoProfile','-NonInteractive','-EncodedCommand',[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parent)))}
}
function Write-RetainedChildObservationJSON([string]$Path,$Value) {
    $target=Get-IsolatedFullPath $Path
    $bytes=[Text.Encoding]::UTF8.GetBytes(($Value|ConvertTo-Json -Depth 8 -WarningAction Stop))
    $stream=[IO.File]::Open($target,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
}
function Get-RetainedChildObservationEvidence([object]$Plan,[object]$Receipt) {
    $root=Get-RetainedChildObservationRoot $Plan.RunRoot
    $stages=@{};$stageComplete=$false;$stageError=$null
    try {
        $begin=[DateTimeOffset]$Receipt.StartUTC;$end=[DateTimeOffset]$Receipt.EndUTC
        if($end -lt $begin){throw 'Reversed receipt interval'}
        foreach($name in @('child-ready.json','release.json','parent-exit.json','child-exit.json')) {
            $path=Get-IsolatedFullPath (Join-Path $root $name)
            if(-not (Test-Path -LiteralPath $path -PathType Leaf) -or (Get-Item -LiteralPath $path).Length -gt 4096 -or [IsolatedFuzzFileInfo]::Links($path) -ne 1){throw ('Missing/unsafe stage: '+$name)}
            $row=[IO.File]::ReadAllText($path)|ConvertFrom-Json
            if($row.Nonce -cne $Plan.Nonce -or $row.ProcessId -le 0 -or $row.ProcessId -gt [uint32]::MaxValue){throw 'Wrong stage identity'}
            $time=[DateTimeOffset]$row.UTC
            if($time -lt $begin -or $time -gt $end){throw 'Stage outside receipt interval'}
            $stages[$name]=$row
        }
        $ready=$stages['child-ready.json'];$release=$stages['release.json'];$parent=$stages['parent-exit.json'];$child=$stages['child-exit.json']
        if($ready.ProcessId -eq $Receipt.RootId -or $ready.ProcessId -ne $child.ProcessId -or $release.ProcessId -ne $Receipt.RootId -or $parent.ProcessId -ne $Receipt.RootId){throw 'Stage process identities disagree'}
        if([DateTimeOffset]$ready.UTC -gt [DateTimeOffset]$release.UTC -or [DateTimeOffset]$release.UTC -gt [DateTimeOffset]$parent.UTC -or [DateTimeOffset]$release.UTC -gt [DateTimeOffset]$child.UTC){throw 'Stage order invalid'}
        $stageComplete=$true
    }catch{$stageError=$_.Exception.Message}
    $core=$false;$coreError=$null
    try {
        # These are diagnostic prerequisites, not a substitute acceptance guard.
        # Never clear mode flags, merge registry evidence, or repair core rows.
        if($Receipt.LifetimeDiagnosticOnly -isnot [bool] -or -not $Receipt.LifetimeDiagnosticOnly -or $Receipt.ShutdownDiagnosticOnly -isnot [bool] -or -not $Receipt.ShutdownDiagnosticOnly -or $Receipt.LiveSampling){throw 'Wrong diagnostic mode'}
        if($Receipt.ShutdownSchema -ne 3 -or $Receipt.ShutdownStatePolicy -cne 'owned-final-signaled-exit-image-optional-v2' -or $Receipt.RootId -le 0 -or $Receipt.ExitCode -ne 0 -or $Receipt.TimedOut -or -not $Receipt.RootExited -or $Receipt.TerminationRequested -or -not $Receipt.CleanupVerified -or -not $Receipt.NaturalShutdownVerified -or -not $Receipt.ShutdownEvidenceComplete -or $Receipt.MemberEvidenceTruncated){throw 'Core natural completion incomplete'}
        if($Receipt.ActiveBeforeTermination -ne 0 -or $Receipt.ActiveAfterCleanup -ne 0 -or $Receipt.DescendantsRemaining -ne 0 -or $Receipt.MembersBeforeTermination.Count -ne 0 -or $Receipt.MembersAfterCleanup.Count -ne 0 -or $Receipt.NaturalObservationMs -lt 0 -or $Receipt.NaturalObservationMs -gt 500 -or $Receipt.ShutdownElapsedMs -lt 0 -or $Receipt.ShutdownElapsedMs -gt 5000){throw 'Core cleanup bounds incomplete'}
        $coreRows=@($Receipt.ShutdownMemberEvidence)
        if($coreRows.Count -gt 32){throw 'Core member evidence exceeds bound'}
        $coreIds=New-Object 'System.Collections.Generic.HashSet[UInt64]'
        foreach($row in $coreRows){
            if($null -eq $row -or $row.Id -le 0 -or $row.Id -gt [uint32]::MaxValue -or -not $coreIds.Add([uint64]$row.Id) -or $null -ne $row.DiagnosticShutdown){throw 'Core member identity incomplete'}
            if($row.Id -eq $Receipt.RootId){continue}
            $member=New-Object 'IsolatedFuzzProcess+MemberObservation'
            foreach($field in @('Id','Owned','QueryError','ExitCodeRead','ExitCode','State','ConfirmationAttempted','ConfirmationState')){$member.$field=$row.$field}
            if($null -ne $row.FinalShutdown){
                $final=New-Object 'IsolatedFuzzProcess+ShutdownObservation'
                foreach($field in @('MemberId','SameHandleVerified','Attempted','State','QueryError','ExitCodeRead','ExitCode','WaitBudgetMs','StartElapsedMs','EndElapsedMs')){$final.$field=$row.FinalShutdown.$field}
                $member.FinalShutdown=$final
            }
            # Reuse only the pure same-handle predicate, never the acceptance
            # guard or a retained-entry/PID join to fill missing core evidence.
            if(-not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($member)){throw 'Core member final evidence unverified'}
        }
        foreach($id in $Receipt.MembersAtRootWait){if(-not $coreIds.Contains([uint64]$id)){throw 'Core initial member missing final evidence'}}
        $core=$true
    }catch{$coreError=$_.Exception.Message}
    $lifetime=$false;$lifetimeError=$null;$childEntry=$null
    try {
        $d=$Receipt.LifetimeDiagnostics
        if($d.Schema -ne 1 -or $d.Coverage -cne 'sampled' -or $d.Incomplete -or $d.AdmissionBudgetExhausted -or $d.TimedOut -or -not $d.RootExited -or -not $d.TeardownAttempted -or $d.ReleaseFailures -ne 0 -or $d.Errors.Count -ne 0){throw 'Incomplete lifetime diagnostics'}
        if($d.Attempts -lt 1 -or $d.Attempts -gt 32 -or $d.Admitted -ne $d.Attempts -or $d.Entries.Count -ne $d.Attempts -or $d.ReleaseAttempts -ne $d.Admitted -or $d.QueryCount -lt 1){throw 'Lifetime counters or release attempts incomplete'}
        if(($d|ConvertTo-Json -Depth 6 -WarningAction Stop) -match '"[^"\r\n]*(Handle|Registry|Api)[^"\r\n]*"\s*:'){throw 'Internal lifetime property forbidden'}
        $frequency=$d.FirstQuery.Frequency
        $ids=New-Object 'System.Collections.Generic.HashSet[UInt64]';$entries=New-Object 'System.Collections.Generic.HashSet[int]'
        foreach($e in $d.Entries){
            if($e.EntryId -lt 1 -or $e.EntryId -gt $d.Attempts -or -not $entries.Add([int]$e.EntryId) -or $e.Pid -le 0 -or $e.Pid -gt [uint32]::MaxValue -or $e.Pid -eq $Receipt.RootId -or -not $ids.Add([uint64]$e.Pid) -or $e.SourceQuery -lt 1 -or $e.SourceQuery -gt $d.QueryCount){throw 'Lifetime entry identity ambiguous'}
            $previous=-1L
            foreach($row in @($e.SourceQueryEvidence,$e.Open,$e.Membership,$e.Admission,$e.RootWaitEnd,$e.Final)){
                if($null -eq $row -or $frequency -le 0 -or $row.Frequency -ne $frequency -or $row.StartTimestamp -lt $previous -or $row.EndTimestamp -lt $row.StartTimestamp){throw 'Missing or unordered lifetime phase clocks'}
                $start=[DateTimeOffset]$row.StartUTC;$finish=[DateTimeOffset]$row.EndUTC
                if($start -lt [DateTimeOffset]$Receipt.StartUTC -or $finish -gt [DateTimeOffset]$Receipt.EndUTC -or $finish -lt $start){throw 'Lifetime phase outside receipt interval'}
                $previous=$row.EndTimestamp
            }
            foreach($q in @($e.SourceQueryEvidence,$d.FirstQuery,$d.LastQuery)){
                if(-not $q.Succeeded -or $q.Error -ne 0 -or $q.Frequency -ne $frequency -or $q.StartTimestamp -lt 0 -or $q.EndTimestamp -lt $q.StartTimestamp -or $q.Assigned -ne $q.Returned -or $q.Returned -gt 32 -or $q.Capacity -ne 32){throw 'Lifetime source query incomplete'}
            }
            if(-not $e.Open.Succeeded -or $e.Open.Error -ne 0 -or -not $e.Membership.Succeeded -or $e.Membership.Error -ne 0 -or -not $e.Membership.Owned){throw 'Lifetime ownership incomplete'}
            foreach($r in @($e.Admission,$e.RootWaitEnd,$e.Final)){
                if(-not $r.Attempted -or -not $r.SameEntry -or $r.Ambiguous -or $r.TerminationRequested -or $r.Error -ne 0 -or -not $r.ExitCodeRead -or $r.State -cnotin @('running','exited') -or -not [string]::IsNullOrEmpty($r.SkipReason)){throw 'Lifetime observation unknown or ambiguous'}
            }
            if($e.Final.State -cne 'exited' -or $e.Final.ExitCode -ne 0){throw 'Lifetime final exit unverified'}
            if($stageComplete -and $e.Pid -eq $stages['child-ready.json'].ProcessId){
                if($e.Admission.State -cne 'running'){throw 'Intended child not captured running at admission'}
                $childEntry=$e.EntryId
            }
        }
        if(-not $stageComplete -or $null -eq $childEntry){throw 'Intended child not captured by retained entry'}
        $lifetime=$true
    }catch{$lifetimeError=$_.Exception.Message}
    [pscustomobject]@{RunRoot=$root;DiagnosticOnly=$true;Acceptance=$false;OriginalTestPassed=$false;Coverage='sampled';NoRetry=$true;StageEvidenceComplete=$stageComplete;StageError=$stageError;Stages=$stages;CoreNaturalComplete=$core;CoreError=$coreError;LifetimeEvidenceComplete=$lifetime;LifetimeError=$lifetimeError;ChildEntryId=$childEntry;ObservationPassed=($stageComplete -and $core -and $lifetime)}
}
function Invoke-RetainedChildObservationNative($Plan,[string]$Stdout,[string]$Stderr) {
    # This sole external boundary always selects the explicit retained-only mode.
    [IsolatedFuzzProcess]::RunWithRetainedHandleDiagnostics($Plan.Executable,$Plan.Arguments,@{},$Plan.RunRoot,$Stdout,$Stderr,20)
}
function Invoke-RetainedChildObservation([string]$RunRoot) {
    $plan=Get-RetainedChildObservationPlan $RunRoot
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    $readyToFinalize=$false
    try {
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'plan.json') $plan
        foreach($row in $plan.SourceBinding){if((Get-FileHash -LiteralPath (Get-IsolatedFullPath (Join-Path $PSScriptRoot $row.Path)) -Algorithm SHA256).Hash -cne $row.SHA256){throw 'Source drift before launch'}}
        if((Get-FileHash -LiteralPath (Get-IsolatedFullPath $plan.Executable) -Algorithm SHA256).Hash -cne $plan.ExecutableSHA256){throw 'Shell drift before launch'}
        $process=Join-Path $plan.RunRoot 'process';New-Item -ItemType Directory -Path $process -ErrorAction Stop|Out-Null
        $stdout=Join-Path $process 'stdout.txt';$stderr=Join-Path $process 'stderr.txt'
        try{$native=Invoke-RetainedChildObservationNative $plan $stdout $stderr}catch{
            Write-RetainedChildObservationJSON (Join-Path $process 'launch-error.json') ([pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;CleanupVerified=$false;Acceptance=$false})
            throw
        }
        $receipt=ConvertTo-RecordedProcessReceipt $native $stdout $stderr $plan.Executable $plan.Arguments @{}
        Write-RetainedChildObservationJSON (Join-Path $process 'receipt.json') $receipt
        $result=Get-RetainedChildObservationEvidence $plan $receipt
        # result.json is provisional. Durable completion also requires a valid
        # manifest and its separately written completed.json hash binding.
        $result|Add-Member FinalizationRequired $true
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'result.json') $result
        if(-not $result.ObservationPassed){throw 'Retained-child observation incomplete; stop without retry'}
        $readyToFinalize=$true
    }catch{
        Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'stopped.json') ([pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;NoRetry=$true;Acceptance=$false})
        throw
    }finally{
        try {
            $files=@(Get-TreeManifest $plan.RunRoot)
            $manifest=Join-Path $plan.RunRoot 'evidence-manifest.json'
            Write-RetainedChildObservationJSON $manifest ([pscustomobject]@{Schema=1;Acceptance=$false;NoRetry=$true;Files=$files})
            if($readyToFinalize){
                Write-RetainedChildObservationJSON (Join-Path $plan.RunRoot 'completed.json') ([pscustomobject]@{Schema=1;Finalized=$true;ObservationPassed=$true;Acceptance=$false;OriginalTestPassed=$false;NoRetry=$true;ManifestSHA256=(Get-FileHash -LiteralPath $manifest -Algorithm SHA256 -ErrorAction Stop).Hash})
            }
        }catch{
            $stopped=Join-Path $plan.RunRoot 'stopped.json'
            if(-not (Test-Path -LiteralPath $stopped)){
                Write-RetainedChildObservationJSON $stopped ([pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');Error=$_.Exception.Message;Stage='finalization';NoRetry=$true;Acceptance=$false})
            }
            throw
        }
    }
    $result|Add-Member Finalized $true
    return $result
}
