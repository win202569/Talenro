$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-Observation([bool]$Value,[string]$Name) {if(-not $Value){throw "ASSERTION FAILED: $Name"};Write-Output "PASS: $Name"}
$receipt=[pscustomobject]@{ShutdownSchema=2;ShutdownDiagnosticOnly=$true;RootExited=$true;TimedOut=$false;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;CleanupVerified=$true;ActiveAfterCleanup=0;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();NaturalObservationMs=0;ShutdownElapsedMs=0}
$refused=$false;try {Assert-ProcessCompletion $receipt}catch {$refused=$true}
Check-Observation $refused 'diagnostic-only receipt cannot become acceptance even with empty natural group'
. (Join-Path $PSScriptRoot 'same-handle.tests.ps1')
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$root=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/observation-tests-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root | Out-Null
function New-ObservationRow {
    $row=New-Object 'IsolatedFuzzProcess+MemberObservation'
    $row.Owned=$true;$row.ExitCodeRead=$true;$row.ExitCode=0;$row.State='running';$row.InitialState='running';$row.ConfirmationState='running';$row.ConfirmationAttempted=$true;$row.ImageError=5
    return $row
}
function Test-RealObservation([bool]$Short) {
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
        if(-not [ConfirmationTestJob]::IsProcessInJob($handle,$job,[ref]$owned) -or -not $owned){throw 'Test child ownership unverified'}
        $setup=[Diagnostics.Stopwatch]::StartNew()
        while(-not [IO.File]::Exists($ready)){if($setup.ElapsedMilliseconds -ge 10000){throw 'Test readiness bound exceeded'};[Threading.Thread]::Sleep(10)}
        $row=New-ObservationRow;$row.Id=$process.Id;$row.Owned=$owned
        $clock=[Diagnostics.Stopwatch]::StartNew();[IO.File]::WriteAllText($release,'release')
        [IsolatedFuzzProcess]::ObserveMemberShutdown($handle,$row,$clock)
        $d=$row.DiagnosticShutdown
        Check-Observation ($d.Attempted -and $d.WaitBudgetMs -gt 0 -and $d.WaitBudgetMs -le 500) 'observation uses only the remaining shared window'
        Check-Observation ($row.State -eq 'running' -and $row.InitialState -eq 'running' -and $row.ConfirmationState -eq 'running' -and $row.ExitCode -eq 0 -and $row.ImageError -eq 5) 'diagnostic wait never rewrites first or confirming reads'
        Check-Observation (-not [IsolatedFuzzProcess]::IsShutdownStateVerified($row)) 'later observation never upgrades rejected core evidence'
        if($Short) {
            Check-Observation ($d.State -eq 'exited' -and $d.ExitCodeRead -and $d.ExitCode -eq 7 -and $d.QueryError -eq 0) 'short-lived owned handle records later signaled exit and actual code'
        } else {
            Check-Observation ($d.State -eq 'running' -and -not $d.ExitCodeRead -and $d.QueryError -eq 0) 'long-lived owned handle remains running without inferred exit'
            $next=New-ObservationRow
            [IsolatedFuzzProcess]::ObserveMemberShutdown($handle,$next,$clock)
            Check-Observation (-not $next.DiagnosticShutdown.Attempted -and $next.DiagnosticShutdown.SkipReason -eq 'window-expired') 'second member cannot receive a fresh per-member window'
        }
        $refused=$false;try {[IsolatedFuzzProcess]::ObserveMemberShutdown($handle,$row,$clock)}catch {$refused=$true}
        Check-Observation $refused 'a member observation cannot be replaced by another attempt'
    } finally {
        if($handle -ne [IntPtr]::Zero -and -not $assigned){[ConfirmationTestJob]::TerminateProcess($handle,0xE003) | Out-Null}
        [ConfirmationTestJob]::TerminateJobObject($job,0xE003) | Out-Null
        [ConfirmationTestJob]::CloseHandle($job) | Out-Null;$process.Dispose()
    }
}
Test-RealObservation $true
Test-RealObservation $false
foreach($case in @(@('Owned',$false),@('QueryError',5),@('ExitCodeRead',$false),@('State','unknown'))) {
    $row=New-ObservationRow;$row.($case[0])=$case[1]
    [IsolatedFuzzProcess]::ObserveMemberShutdown([IntPtr]::Zero,$row,[Diagnostics.Stopwatch]::StartNew())
    Check-Observation (-not $row.DiagnosticShutdown.Attempted -and $row.DiagnosticShutdown.SkipReason -eq 'unverified-state') ('unverified '+$case[0]+' cannot query a handle')
}
$clock=[Diagnostics.Stopwatch]::StartNew();[Threading.Thread]::Sleep(510)
$row=New-ObservationRow
[IsolatedFuzzProcess]::ObserveMemberShutdown([IntPtr]::Zero,$row,$clock)
Check-Observation (-not $row.DiagnosticShutdown.Attempted -and $row.DiagnosticShutdown.SkipReason -eq 'window-expired') 'already-expired shared clock prevents any handle query'
$shell=(Get-Process -Id $PID).Path
$receipt=Invoke-RecordedProcess -Executable $shell -Arguments @('-NoProfile','-NonInteractive','-Command','exit 0') -Environment @{} -ReceiptDir (Join-Path $root 'diagnostic-run') -LimitSeconds 10 -CaptureShutdownDiagnostic $true
Check-Observation ($receipt.ShutdownDiagnosticOnly -and $receipt.ExitCode -eq 0 -and -not $receipt.TimedOut -and $receipt.CleanupVerified) 'opt-in native run records diagnostic-only flag and owned cleanup'
$saved=Get-Content -Raw -LiteralPath (Join-Path $root 'diagnostic-run/receipt.json') | ConvertFrom-Json
Check-Observation ($saved.ShutdownDiagnosticOnly) 'diagnostic-only flag survives receipt serialization'
$refused=$false;try {Assert-ProcessCompletion $receipt}catch {$refused=$true}
Check-Observation $refused 'actual diagnostic run cannot serve as completed arm preparation'
$eventName='Local\TalenroObserve-'+[guid]::NewGuid().ToString('N')
$childCode='$ready=[Threading.EventWaitHandle]::OpenExisting("'+$eventName+'");$ready.Set()|Out-Null;[Threading.Thread]::Sleep(30000);exit 0'
$childEncoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($childCode))
$parentCode='$ready=New-Object Threading.EventWaitHandle($false,[Threading.EventResetMode]::ManualReset,"'+$eventName+'");Start-Process -FilePath "'+$shell+'" -ArgumentList "-NoProfile -EncodedCommand '+$childEncoded+'" -WindowStyle Hidden;if(-not $ready.WaitOne(5000)){exit 9};exit 0'
$parentEncoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($parentCode))
$receipt=Invoke-RecordedProcess -Executable $shell -Arguments @('-NoProfile','-EncodedCommand',$parentEncoded) -Environment @{} -ReceiptDir (Join-Path $root 'diagnostic-live-child') -LimitSeconds 10 -CaptureShutdownDiagnostic $true
$observed=@($receipt.MemberEvidence | Where-Object {$_.Owned -and $_.State -eq 'running' -and $_.DiagnosticShutdown})
Check-Observation ($observed.Count -gt 0 -and $receipt.CleanupVerified) 'native diagnostic path observes and cleans real owned live child'
Check-Observation (@($observed | Where-Object {$null -eq $_.PSObject.Properties['CoreObservedUTC'] -or -not $_.CoreObservedUTC -or [DateTimeOffset]::Parse($_.CoreObservedUTC) -gt [DateTimeOffset]::Parse($_.DiagnosticShutdown.StartUTC)}).Count -eq 0) 'native path records distinct core snapshot time before diagnostic observation'
