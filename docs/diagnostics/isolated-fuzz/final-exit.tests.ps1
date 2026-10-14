$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'same-handle.tests.ps1')
function Check-FinalExit([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
function New-PendingFinalRow {
    $row=New-Object 'IsolatedFuzzProcess+MemberObservation'
    $row.Id=42;$row.Owned=$true;$row.QueryError=0;$row.ExitCodeRead=$true;$row.ExitCode=0;$row.State='running';$row.InitialState='running';$row.ConfirmationState='running';$row.ConfirmationAttempted=$true;$row.ImageError=5
    return $row
}
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/final-exit-tests-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root | Out-Null
function Test-RealFinalExit([bool]$Short) {
    $label=if($Short){'short'}else{'long'}
    $ready=Join-Path $root ($label+'-ready');$release=Join-Path $root ($label+'-release')
    $job=[ConfirmationTestJob]::CreateJobObject([IntPtr]::Zero,$null)
    if($job -eq [IntPtr]::Zero){throw 'Test job creation failed'}
    $process=New-Object Diagnostics.Process;$handle=[IntPtr]::Zero;$assigned=$false
    try {
        $ms=if($Short){150}else{30000}
        $code="[IO.File]::WriteAllText('$ready','ready');while(-not [IO.File]::Exists('$release')){[Threading.Thread]::Sleep(5)};[Threading.Thread]::Sleep($ms);exit 7"
        $process.StartInfo.FileName=(Get-Process -Id $PID).Path
        $process.StartInfo.Arguments='-NoProfile -NonInteractive -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
        $process.StartInfo.UseShellExecute=$false;$process.StartInfo.CreateNoWindow=$true
        if(-not $process.Start()){throw 'Test child creation failed'}
        $handle=$process.Handle
        if(-not [ConfirmationTestJob]::AssignProcessToJobObject($job,$handle)){throw 'Test child assignment failed'}
        $assigned=$true;$owned=$false
        if(-not [ConfirmationTestJob]::IsProcessInJob($handle,$job,[ref]$owned) -or -not $owned){throw 'Test ownership unverified'}
        $setup=[Diagnostics.Stopwatch]::StartNew()
        while(-not [IO.File]::Exists($ready)){if($setup.ElapsedMilliseconds -ge 10000){throw 'Test readiness expired'};[Threading.Thread]::Sleep(10)}
        $row=New-PendingFinalRow;$row.Id=$process.Id;$row.Owned=$owned
        $clock=[Diagnostics.Stopwatch]::StartNew();[IO.File]::WriteAllText($release,'release')
        [IsolatedFuzzProcess]::VerifyMemberFinalExit($handle,$row,$clock)
        Check-FinalExit ($row.State -eq 'running' -and $row.InitialState -eq 'running' -and $row.ConfirmationState -eq 'running' -and $row.ExitCode -eq 0 -and $row.ImageError -eq 5) 'final exit evidence preserves core first and confirming reads'
        Check-FinalExit ($row.FinalShutdown.Attempted -and $row.FinalShutdown.WaitBudgetMs -gt 0 -and $row.FinalShutdown.WaitBudgetMs -le 500 -and $null -eq $row.DiagnosticShutdown) 'final evidence uses one bounded same-handle wait not diagnostic data'
        if($Short) {
            Check-FinalExit ($row.FinalShutdown.State -eq 'exited' -and $row.FinalShutdown.ExitCodeRead -and $row.FinalShutdown.ExitCode -eq 7) 'later signaled exit refreshes code only in final evidence'
            Check-FinalExit ([IsolatedFuzzProcess]::IsFinalShutdownStateVerified($row)) 'verified final exit can satisfy new policy despite preserved image error'
            Check-FinalExit (-not [IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'legacy snapshot predicate still rejects running missing image'
            $script:validFinalRow=$row
        } else {
            Check-FinalExit (-not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($row)) 'live member after shared window still refused'
            $next=New-PendingFinalRow
            [IsolatedFuzzProcess]::VerifyMemberFinalExit($handle,$next,$clock)
            Check-FinalExit (-not $next.FinalShutdown.Attempted -and $next.FinalShutdown.SkipReason -eq 'window-expired') 'another member never gets a new final-exit window'
        }
        $refused=$false;try {[IsolatedFuzzProcess]::VerifyMemberFinalExit($handle,$row,$clock)}catch {$refused=$true}
        Check-FinalExit $refused 'same final evidence cannot be overwritten by retry'
    } finally {
        if($handle -ne [IntPtr]::Zero -and -not $assigned){[ConfirmationTestJob]::TerminateProcess($handle,0xE003) | Out-Null}
        [ConfirmationTestJob]::TerminateJobObject($job,0xE003) | Out-Null
        [ConfirmationTestJob]::CloseHandle($job) | Out-Null;$process.Dispose()
    }
}
Test-RealFinalExit $true
Test-RealFinalExit $false
foreach($case in @(@('Owned',$false),@('QueryError',5),@('ExitCodeRead',$false),@('State','unknown'))) {
    $row=New-PendingFinalRow;$row.($case[0])=$case[1]
    [IsolatedFuzzProcess]::VerifyMemberFinalExit([IntPtr]::Zero,$row,[Diagnostics.Stopwatch]::StartNew())
    Check-FinalExit (-not $row.FinalShutdown.Attempted -and -not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($row)) ('unverified '+$case[0]+' cannot obtain final exit evidence')
}
$row=New-PendingFinalRow
[IsolatedFuzzProcess]::ObserveMemberShutdown([IntPtr]::Zero,$row,[Diagnostics.Stopwatch]::StartNew())
$refused=$false;try {[IsolatedFuzzProcess]::VerifyMemberFinalExit([IntPtr]::Zero,$row,[Diagnostics.Stopwatch]::StartNew())}catch {$refused=$true}
Check-FinalExit $refused 'diagnostic observation cannot be repurposed into final evidence'
$receipt=[pscustomobject]@{ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';ShutdownDiagnosticOnly=$false;RootId=1;RootExited=$true;TimedOut=$false;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;MemberEvidenceTruncated=$false;CleanupVerified=$true;ActiveAfterCleanup=0;TerminationRequested=$false;ActiveBeforeTermination=0;MembersAtRootWait=@($validFinalRow.Id);MembersBeforeTermination=@();MembersAfterCleanup=@();NaturalObservationMs=200;ShutdownElapsedMs=200;ShutdownMemberEvidence=@($validFinalRow)}
$serialized=$receipt | ConvertTo-Json -Depth 7 | ConvertFrom-Json
$accepted=$true;try {Assert-ProcessCompletion $serialized}catch {$accepted=$false}
Check-FinalExit $accepted 'new receipt accepts serialized same-handle final exit only with empty natural job'
foreach($case in @(
    @('TerminationRequested',$true),@('ShutdownDiagnosticOnly',$true),@('ActiveBeforeTermination',1),@('MembersBeforeTermination',@(42)),@('MembersAfterCleanup',@(42)),@('NaturalObservationMs',501),@('MemberEvidenceTruncated',$true),@('ShutdownStatePolicy','unknown'),@('ShutdownMemberEvidence',@()),@('ShutdownSchema',2)
)) {
    $r=$receipt|ConvertTo-Json -Depth 7|ConvertFrom-Json;$r.($case[0])=$case[1]
    $refused=$false;try {Assert-ProcessCompletion $r}catch {$refused=$true}
    Check-FinalExit $refused ('new receipt rejects '+$case[0])
}
foreach($case in @(@('State','running'),@('QueryError',5),@('ExitCodeRead',$false),@('SameHandleVerified',$false),@('MemberId',0),@('EndElapsedMs',501),@('WaitBudgetMs',501),@('StartElapsedMs',-1))) {
    $r=$receipt|ConvertTo-Json -Depth 7|ConvertFrom-Json;$r.ShutdownMemberEvidence[0].FinalShutdown.($case[0])=$case[1]
    $refused=$false;try {Assert-ProcessCompletion $r}catch {$refused=$true}
    Check-FinalExit $refused ('new receipt rejects final '+$case[0])
}
$r=$receipt|ConvertTo-Json -Depth 7|ConvertFrom-Json;$r.ShutdownMemberEvidence[0].FinalShutdown=$null
$refused=$false;try {Assert-ProcessCompletion $r}catch {$refused=$true}
Check-FinalExit $refused 'running receipt cannot pass with only empty accounting and no final handle evidence'
$r=$receipt|ConvertTo-Json -Depth 7|ConvertFrom-Json;$r.ShutdownMemberEvidence[0].DiagnosticShutdown=$r.ShutdownMemberEvidence[0].FinalShutdown
$refused=$false;try {Assert-ProcessCompletion $r}catch {$refused=$true}
Check-FinalExit $refused 'new receipt cannot borrow diagnostic member observation'
$shell=(Get-Process -Id $PID).Path
$eventName='Local\TalenroFinal-'+[guid]::NewGuid().ToString('N')
$childCode='$ready=[Threading.EventWaitHandle]::OpenExisting("'+$eventName+'");$ready.Set()|Out-Null;[Threading.Thread]::Sleep(150);exit 0'
$childEncoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($childCode))
$parentCode='$ready=New-Object Threading.EventWaitHandle($false,[Threading.EventResetMode]::ManualReset,"'+$eventName+'");Start-Process -FilePath "'+$shell+'" -ArgumentList "-NoProfile -EncodedCommand '+$childEncoded+'" -WindowStyle Hidden;if(-not $ready.WaitOne(5000)){exit 9};exit 0'
$parentEncoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parentCode))
$actual=Invoke-RecordedProcess $shell @('-NoProfile','-EncodedCommand',$parentEncoded) @{} (Join-Path $root 'native-final-short-child') 10
Check-FinalExit ($actual.ShutdownSchema -eq 3 -and $actual.NaturalShutdownVerified -and -not $actual.TerminationRequested) 'native final-policy process naturally completes under new receipt schema'
$pending=@($actual.ShutdownMemberEvidence|Where-Object {$_.Id -ne $actual.RootId -and $_.State -eq 'running'})
Check-FinalExit ($pending.Count -gt 0 -and @($pending|Where-Object {-not [IsolatedFuzzProcess]::IsFinalShutdownStateVerified($_)}).Count -eq 0) 'native producer records final same-handle evidence for observed running members'
$firstFinal=@($pending|Where-Object {$_.FinalShutdown.Attempted}|Sort-Object {$_.FinalShutdown.StartUTC}|Select-Object -First 1)
Check-FinalExit ($firstFinal.Count -eq 1 -and @($actual.ShutdownMemberEvidence|Where-Object {$_.Id -ne $actual.RootId -and (-not $_.CoreObservedUTC -or [DateTimeOffset]::Parse($_.CoreObservedUTC) -gt [DateTimeOffset]::Parse($firstFinal[0].FinalShutdown.StartUTC))}).Count -eq 0) 'all native batch core handles are captured before the first final wait'
Assert-ProcessCompletion (Get-Content -Raw (Join-Path $root 'native-final-short-child/receipt.json')|ConvertFrom-Json)
Check-FinalExit $true 'actual saved native final receipt validates after serialization'
