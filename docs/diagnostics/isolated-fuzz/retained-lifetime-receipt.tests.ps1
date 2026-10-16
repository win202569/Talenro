$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'run.ps1')
function Check-Receipt([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
function New-CleanLifetimeNative {
    $r=New-Object 'IsolatedFuzzProcess+Result'
    $r.RootId=17;$r.RootExited=$true;$r.NaturalShutdownVerified=$true;$r.ShutdownEvidenceComplete=$true;$r.CleanupVerified=$true
    $r.MembersAtRootWait=[uint64[]]@();$r.MembersBeforeTermination=[uint64[]]@();$r.MembersAfterCleanup=[uint64[]]@();$r.ShutdownMemberEvidence=@()
    return $r
}
function Must-Reject($Receipt,[string]$Name){$refused=$false;try{Assert-ProcessCompletion $Receipt}catch{$refused=$true};Check-Receipt $refused $Name}
$native=New-CleanLifetimeNative;$native.LifetimeDiagnosticOnly=$true;$native.ShutdownDiagnosticOnly=$false;$native.LifetimeDiagnostics=New-Object LifetimeReport
$saved=ConvertTo-RecordedProcessReceipt $native 'stdout' 'stderr' 'fixture' @() @{}
Check-Receipt ($null -ne $saved.PSObject.Properties['LifetimeDiagnostics']) 'OptionalLifetimeMetadataPreserved'
Must-Reject $saved 'LifetimeFlagAloneRejected'
$saved.ShutdownDiagnosticOnly=$true;Must-Reject $saved 'BothFlagsRejected'
$round=$saved|ConvertTo-Json -Depth 12|ConvertFrom-Json
Check-Receipt ($round.LifetimeDiagnostics.Schema -eq 1 -and $round.LifetimeDiagnostics.Coverage -ceq 'sampled' -and $round.LifetimeDiagnosticOnly) 'NestedRoundtrip'
$legacy=ConvertTo-RecordedProcessReceipt (New-CleanLifetimeNative) 'out' 'err' 'fixture' @() @{}
$legacy.PSObject.Properties.Remove('LifetimeDiagnosticOnly');$legacy.PSObject.Properties.Remove('LifetimeDiagnostics')
Assert-ProcessCompletion $legacy;Check-Receipt $true 'LegacyReceiptWithoutFieldsUnchanged'
$legacy|Add-Member LifetimeDiagnostics (New-Object LifetimeReport)
$legacy.NaturalShutdownVerified=$false;$legacy.ShutdownEvidenceComplete=$false;$legacy.TerminationRequested=$true
$row=New-Object 'IsolatedFuzzProcess+MemberObservation';$row.Id=42;$row.QueryError=87;$row.State='unknown';$legacy.MemberEvidence=@($row)
Must-Reject $legacy 'CoreOpen87NeverUpgraded'
Check-Receipt ($legacy.MemberEvidence[0].QueryError -eq 87 -and -not $legacy.MemberEvidence[0].Owned -and -not $legacy.MemberEvidence[0].ExitCodeRead) 'CoreFailureFieldsUnchanged'
$saved.LifetimeDiagnostics.Entries.Clear();$saved.ShutdownDiagnosticOnly=$false;Must-Reject $saved 'MissingSnapshotNeverBecomesSuccess'
$json=$saved|ConvertTo-Json -Depth 12
Check-Receipt ($json -notmatch '"Handle"|"Api"|"Registry"') 'ReceiptNoInternalHandleProperties'

# Test data only: use the real DTOs/converter/JSON/guard, never a native API.
function Set-LifetimeFixtureFields($Value,[hashtable]$Fields){foreach($name in $Fields.Keys){$Value.$name=$Fields[$name]};return $Value}
function New-LifetimeQueryFixture([long]$Start,[uint32]$Count){
    $origin=[DateTimeOffset]'2026-10-16T00:00:00Z'
    Set-LifetimeFixtureFields (New-Object LifetimeQueryEvidence) @{Succeeded=$true;Error=0;Assigned=$Count;Returned=$Count;Capacity=32;BufferBytes=264;StartTimestamp=$Start;EndTimestamp=($Start+100);Frequency=1000;StartUTC=$origin.AddMilliseconds($Start).ToString('o');EndUTC=$origin.AddMilliseconds($Start+100).ToString('o')}
}
function New-LifetimeReadFixture([string]$State,[long]$Start,[bool]$Ambiguous,[bool]$Terminated){
    $origin=[DateTimeOffset]'2026-10-16T00:00:00Z'
    Set-LifetimeFixtureFields (New-Object LifetimeRead) @{Attempted=$true;SameEntry=$true;State=$State;ExitCodeRead=$true;ExitCode=$(if($State -ceq 'running'){259}else{0});Error=0;ImagePath='fixture.exe';ImageError=0;Ambiguous=$Ambiguous;TerminationRequested=$Terminated;StartTimestamp=$Start;EndTimestamp=($Start+100);Frequency=1000;StartUTC=$origin.AddMilliseconds($Start).ToString('o');EndUTC=$origin.AddMilliseconds($Start+100).ToString('o')}
}
function Assert-PopulatedLifetimeRoundtrip($Receipt){
    $d=$Receipt.LifetimeDiagnostics
    if($d.Schema -ne 1 -or $d.Coverage -cne 'sampled' -or -not $Receipt.LifetimeDiagnosticOnly -or -not $Receipt.ShutdownDiagnosticOnly){throw 'nested mode markers lost'}
    if($d.Attempts -ne 32 -or $d.Admitted -ne 31 -or $d.QueryCount -ne 7 -or -not $d.AdmissionBudgetExhausted -or -not $d.Incomplete -or $d.Entries.Count -ne 32){throw 'nested cap counters lost'}
    if(-not $d.RootExited -or $d.TimedOut -or -not $d.TeardownAttempted -or $d.ReleaseAttempts -ne 31 -or $d.ReleaseFailures -ne 1 -or $d.Errors.Count -ne 2 -or $d.Errors[1] -cne 'close:6'){throw 'nested teardown/error facts lost'}
    $e=$d.Entries[0];$last=$d.Entries[31]
    if($e.EntryId -ne 1 -or $e.Pid -ne 1001 -or $e.SourceQuery -ne 1 -or $last.EntryId -ne 32 -or $last.Pid -ne 1032){throw 'nested entry identity lost'}
    if($null -eq $e.SourceQueryEvidence -or $e.SourceQueryEvidence.Assigned -ne 32 -or $e.SourceQueryEvidence.Returned -ne 32 -or $e.SourceQueryEvidence.Capacity -ne 32 -or $e.SourceQueryEvidence.BufferBytes -ne 264 -or $e.SourceQueryEvidence.Frequency -ne 1000 -or $e.SourceQueryEvidence.StartTimestamp -ne 1000 -or $e.SourceQueryEvidence.EndTimestamp -ne 1100){throw 'nested source-query evidence lost'}
    if($d.FirstQuery.StartTimestamp -ne 1000 -or $d.LastQuery.StartTimestamp -ne 2000 -or $d.LastQuery.Assigned -ne 4 -or $d.LastQuery.Returned -ne 4){throw 'nested first/last query facts lost'}
    if(-not $e.Open.Succeeded -or $e.Open.Error -ne 0 -or $e.Open.StartTimestamp -ne 1100 -or $e.Open.EndTimestamp -ne 1200 -or -not $e.Membership.Succeeded -or -not $e.Membership.Owned -or $e.Membership.StartTimestamp -ne 1300){throw 'nested open/membership evidence lost'}
    if($e.Admission.State -cne 'exited' -or -not $e.Admission.ExitCodeRead -or $e.Admission.ExitCode -ne 0 -or $e.Admission.StartTimestamp -ne 1500 -or $e.Admission.Ambiguous){throw 'nested admission evidence lost'}
    $rootUTC=([DateTimeOffset]$e.RootWaitEnd.StartUTC).ToUniversalTime().ToString('o')
    if(-not $e.RootWaitEnd.Attempted -or -not $e.RootWaitEnd.SameEntry -or -not $e.RootWaitEnd.Ambiguous -or $e.RootWaitEnd.State -cne 'running' -or $e.RootWaitEnd.ExitCode -ne 259 -or $e.RootWaitEnd.TerminationRequested -or $e.RootWaitEnd.StartTimestamp -ne 2500 -or $rootUTC -cne '2026-10-16T00:00:02.5000000+00:00'){throw 'nested ambiguity/raw state lost'}
    if(-not $e.Final.Ambiguous -or -not $e.Final.TerminationRequested -or $e.Final.State -cne 'exited' -or -not $e.Final.ExitCodeRead -or $e.Final.ExitCode -ne 0 -or $e.Final.StartTimestamp -ne 3500 -or $e.Final.ImageError -ne 5 -or $null -ne $e.Final.ImagePath){throw 'nested final observation lost'}
    if($last.Open.Succeeded -or $last.Open.Error -ne 87 -or $null -ne $last.Membership -or $null -ne $last.Admission -or $null -ne $last.RootWaitEnd -or $null -ne $last.Final){throw 'nested failed admission defaults upgraded'}
    if(($d|ConvertTo-Json -Depth 6 -WarningAction Stop) -match '"[^"\r\n]*(Handle|Registry|Api)[^"\r\n]*"\s*:'){throw 'nested raw handle/internal property leaked'}
}
$full=Set-LifetimeFixtureFields (New-Object LifetimeReport) @{Attempts=32;Admitted=31;QueryCount=7;AdmissionBudgetExhausted=$true;Incomplete=$true;RootExited=$true;TimedOut=$false;TeardownAttempted=$true;ReleaseAttempts=31;ReleaseFailures=1}
$full.Errors.Add('ambiguous-lifetime:-1');$full.Errors.Add('close:6')
$full.FirstQuery=New-LifetimeQueryFixture 1000 32;$full.LastQuery=New-LifetimeQueryFixture 2000 4
foreach($i in 1..32){
    $entry=Set-LifetimeFixtureFields (New-Object LifetimeEntry) @{EntryId=$i;Pid=(1000+$i);SourceQuery=1;SourceQueryEvidence=$full.FirstQuery}
    $entry.Open=Set-LifetimeFixtureFields (New-Object LifetimeApiEvidence) @{Succeeded=($i -lt 32);Error=$(if($i -eq 32){87}else{0});StartTimestamp=1100;EndTimestamp=1200;Frequency=1000;StartUTC='2026-10-16T00:00:01.100Z';EndUTC='2026-10-16T00:00:01.200Z'}
    if($i -lt 32){
        $entry.Membership=Set-LifetimeFixtureFields (New-Object LifetimeMembership) @{Succeeded=$true;Owned=$true;Error=0;StartTimestamp=1300;EndTimestamp=1400;Frequency=1000;StartUTC='2026-10-16T00:00:01.300Z';EndUTC='2026-10-16T00:00:01.400Z'}
        $entry.Admission=New-LifetimeReadFixture $(if($i -eq 1){'exited'}else{'running'}) 1500 $false $false
        $entry.RootWaitEnd=New-LifetimeReadFixture 'running' 2500 ($i -eq 1) $false
        $entry.Final=New-LifetimeReadFixture 'exited' 3500 ($i -eq 1) $true
        # Direct reflection preserves a genuine null string; PowerShell's typed
        # string-field assignment of $null otherwise produces an empty string.
        $entry.Final.GetType().GetField('ImagePath').SetValue($entry.Final,$null);$entry.Final.ImageError=5
    }
    $full.Entries.Add($entry)
}
$populatedNative=New-CleanLifetimeNative;$populatedNative.LifetimeDiagnosticOnly=$true;$populatedNative.ShutdownDiagnosticOnly=$true;$populatedNative.LifetimeDiagnostics=$full
$populatedReceipt=ConvertTo-RecordedProcessReceipt $populatedNative 'out' 'err' 'fixture' @() @{}
# Match the depth used by the actual recorded-process writer; warnings are failures.
$populatedJSON=$populatedReceipt|ConvertTo-Json -Depth 6 -WarningAction Stop
$populatedRound=$populatedJSON|ConvertFrom-Json
Assert-PopulatedLifetimeRoundtrip $populatedRound
Check-Receipt $true 'PopulatedNestedRoundtripAtWriterDepth6'
Check-Receipt (-not [object]::ReferenceEquals($full,$populatedRound.LifetimeDiagnostics)) 'PopulatedRoundtripNotAnAlias'
Must-Reject $populatedRound 'PopulatedDiagnosticStillNotAcceptance'
# In-memory loss/injection controls prove the new assertions detect regressions.
# They neither patch the production converter nor represent a product RED/fix.
foreach($mutation in @('source','ambiguity','cap','handle')){
    $bad=$populatedJSON|ConvertFrom-Json
    switch($mutation){
        'source'{$bad.LifetimeDiagnostics.Entries[0].SourceQueryEvidence=$null;$expected='source-query'}
        'ambiguity'{$bad.LifetimeDiagnostics.Entries[0].RootWaitEnd.Ambiguous=$false;$expected='ambiguity'}
        'cap'{$bad.LifetimeDiagnostics.Attempts=31;$expected='cap counters'}
        'handle'{$bad.LifetimeDiagnostics.Entries[0]|Add-Member Handle 900002;$expected='raw handle'}
    }
    $failure=$null;try{Assert-PopulatedLifetimeRoundtrip $bad}catch{$failure=$_.Exception.Message}
    Check-Receipt ($null -ne $failure -and $failure -match $expected) ('PopulatedLossControl '+$mutation)
}
Write-Output 'LIMIT: synthetic receipts only; no process query or native Run invoked.'
