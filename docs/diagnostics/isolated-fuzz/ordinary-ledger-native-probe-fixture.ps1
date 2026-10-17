# Hand-derived receipt facts only: no type loading, process APIs or source loader.
function Check-Ordinary([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
function New-OrdinaryProbeTestApi {
    [pscustomobject]@{Succeeded=$true;Error=0;StartTimestamp=100;EndTimestamp=100;Frequency=1000;StartUTC='2026-10-17T00:00:00Z';EndUTC='2026-10-17T00:00:00Z'}
}
function New-OrdinaryProbeTestQuery([int]$Count){
    $q=New-OrdinaryProbeTestApi
    foreach($pair in @(@('Assigned',$Count),@('Returned',$Count),@('Capacity',32),@('BufferBytes',264))){$q|Add-Member -NotePropertyName $pair[0] -NotePropertyValue $pair[1]}
    return $q
}
function New-OrdinaryProbeTestEnvelope([switch]$Child){
    $count=if($Child){1}else{0};$entries=@()
    if($Child){
        $membership=New-OrdinaryProbeTestApi;$membership|Add-Member Owned $true
        $initial=[pscustomobject]@{RunIdentity='run-a';EntryId=1;SameHandleVerified=$true;Attempted=$true;Signaled=$false;State='running';Error=0;ExitCodeRead=$true;ExitCode=259;ImageError=0;ImagePath='fixture.exe';SkipReason=$null;TerminationRequested=$false;StartTimestamp=100;EndTimestamp=100;Frequency=1000;StartUTC='2026-10-17T00:00:00Z';EndUTC='2026-10-17T00:00:00Z'}
        $final=$initial|ConvertTo-Json|ConvertFrom-Json;$final.Signaled=$true;$final.State='exited';$final.ExitCode=0
        $entries=@([pscustomobject]@{RunIdentity='run-a';EntryId=1;Pid=2;SourceQuery=1;SourceQueryEvidence=(New-OrdinaryProbeTestQuery 1);Open=(New-OrdinaryProbeTestApi);Membership=$membership;Initial=$initial;Final=$final;ReleaseAttempted=$true;ReleaseSucceeded=$true;ReleaseError=0})
    }
    $ledger=[pscustomobject]@{Schema=1;RunIdentity='run-a';Mode='ordinary-identity-ledger-v1';Coverage='sampled';RootId=1;Incomplete=$false;Ambiguous=$false;CapExhausted=$false;TeardownAttempted=$true;Sealed=$true;Errors=@();Attempts=$count;Admitted=$count;QueryCount=3;ReleaseAttempts=$count;ReleaseFailures=0;Entries=$entries;FirstQuery=(New-OrdinaryProbeTestQuery $count);LastQuery=(New-OrdinaryProbeTestQuery 0)}
    $accounting=New-OrdinaryProbeTestApi;$accounting|Add-Member TotalProcesses (1+$count);$accounting|Add-Member ActiveProcesses 0;$accounting|Add-Member TotalTerminatedProcesses 0
    $frame=[pscustomobject]@{EvidenceComplete=$true;NaturalShutdownVerified=$true;CountReconciled=$true;Errors=@();FinalMembers=@();ShutdownStartTimestamp=100;EndTimestamp=100;Frequency=1000;NaturalObservationMs=0;FinalAccounting=$accounting;FinalMemberQuery=(New-OrdinaryProbeTestQuery 0)}
    [pscustomobject]@{RunIdentity='run-a';RootId=1;RootAssigned=$true;RootExited=$true;TimedOut=$false;TerminationRequested=$false;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();ActiveBeforeTermination=0;MembersBeforeTermination=@();ShutdownSchema=4;ShutdownStatePolicy='ordinary-owned-ledger-final-exit-v1';LifetimeDiagnosticOnly=$false;ShutdownDiagnosticOnly=$false;LiveSampling=$false;MemberEvidenceTruncated=$false;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;NaturalObservationMs=0;ShutdownElapsedMs=0;MemberIdentityLedger=$ledger;LedgerFinalOutcome=$frame}
}
