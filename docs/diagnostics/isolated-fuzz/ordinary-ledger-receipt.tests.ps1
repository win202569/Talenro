param([switch]$DiagnoseRoundtrip)
# Breaks caught: coercion/default evidence, version stripping, forged identity joins,
# contradictory timing, diagnostic contamination and accidental acceptance rollout.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
if(-not(Test-Path -LiteralPath (Join-Path $PSScriptRoot 'ordinary-ledger-receipt.ps1'))){throw 'ASSERTION FAILED: planned ordinary receipt evaluator missing'}
. (Join-Path $PSScriptRoot 'ordinary-ledger-receipt.ps1')
. (Join-Path $PSScriptRoot 'run.ps1')
function New-OrdinaryEnvelope {
    $f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000);$f.Api.Candidates=[UInt64[]]@();$f.Api.Exited=$true
    $s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$outcome=$s.Finish($true,$false,100)
    [pscustomobject]@{RunIdentity='run-a';RootId=1;RootAssigned=$true;RootExited=$true;TimedOut=$false;TerminationRequested=$false;CleanupVerified=$true;ActiveAfterCleanup=0;MembersAfterCleanup=@();ActiveBeforeTermination=0;MembersBeforeTermination=@();ShutdownSchema=4;ShutdownStatePolicy='ordinary-owned-ledger-final-exit-v1';LifetimeDiagnosticOnly=$false;ShutdownDiagnosticOnly=$false;LiveSampling=$false;MemberEvidenceTruncated=$false;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;NaturalObservationMs=0;ShutdownElapsedMs=0;MemberIdentityLedger=$f.Ledger.Report;LedgerFinalOutcome=$outcome}
}
function Copy-OrdinaryEnvelope($r){$r|ConvertTo-Json -Depth 16|ConvertFrom-Json}
function Must-RefuseOrdinary($r,$name){$e=Test-OrdinaryLedgerEvidence $r;Check-Ordinary (-not $e.EvidenceComplete -and -not $e.Acceptance -and $e.Errors.Count -gt 0) $name}
function Must-RefuseGate($r,$name){$refused=$false;try{Assert-ProcessCompletion $r}catch{$refused=$true};Check-Ordinary $refused $name}
$good=New-OrdinaryEnvelope;$e=Test-OrdinaryLedgerEvidence $good
Check-Ordinary ($e.EvidenceComplete -and -not $e.Acceptance -and $e.Errors.Count -eq 0) 'CompleteButNotAccepted';Must-RefuseGate $good 'Schema4GateRefuses'
$round=Copy-OrdinaryEnvelope $good;$e=Test-OrdinaryLedgerEvidence $round
if($DiagnoseRoundtrip){
    Write-Output ('ROUNDTRIP RESULT: '+($e|ConvertTo-Json -Compress))
    Write-Output ('ROUNDTRIP JSON: '+($round|ConvertTo-Json -Depth 16 -Compress))
    return
}
Check-Ordinary ($e.EvidenceComplete -and -not $e.Acceptance) ('NestedRoundtrip: '+($e.Errors -join '; '))
foreach($field in @('RootAssigned','RootExited','TimedOut','TerminationRequested','CleanupVerified','LifetimeDiagnosticOnly','ShutdownDiagnosticOnly','LiveSampling','MemberEvidenceTruncated','NaturalShutdownVerified','ShutdownEvidenceComplete')){
    foreach($bad in @($null,0,1,'true','false')){$r=Copy-OrdinaryEnvelope $good;$r.$field=$bad;Must-RefuseOrdinary $r ('StrictTypes-'+$field)}
}
foreach($field in @('ShutdownSchema','ShutdownStatePolicy','RunIdentity','MemberIdentityLedger','LedgerFinalOutcome')){
    $r=Copy-OrdinaryEnvelope $good;$r.PSObject.Properties.Remove($field);Must-RefuseOrdinary $r ('Missing-'+$field);Must-RefuseGate $r 'MarkerDowngrade'
}
$r=Copy-OrdinaryEnvelope $good;$r.ShutdownSchema=3;Must-RefuseOrdinary $r 'OldSchema';Must-RefuseGate $r 'OldSchemaMarkerRefused'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries[0].Final=$null;Must-RefuseOrdinary $r 'MissingRead'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries[0].RunIdentity='run-b';Must-RefuseOrdinary $r 'WrongRunIdentity'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries=@($r.MemberIdentityLedger.Entries[0],$r.MemberIdentityLedger.Entries[0]);Must-RefuseOrdinary $r 'DuplicateEntry'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries[0].Final.EndTimestamp=99;Must-RefuseOrdinary $r 'ContradictoryTiming'
foreach($field in @('Frequency','TotalProcesses','ActiveProcesses','TotalTerminatedProcesses')){$r=Copy-OrdinaryEnvelope $good;$r.LedgerFinalOutcome.FinalAccounting.$field=-1;Must-RefuseOrdinary $r ('BadAccounting-'+$field)}
foreach($field in @('LifetimeDiagnostics','CompileStageTrigger','BuildOnlyDiagnostic')){$r=Copy-OrdinaryEnvelope $good;$r|Add-Member $field ([pscustomobject]@{});Must-RefuseOrdinary $r 'DiagnosticContamination'}
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries[0]|Add-Member Handle 900002;Must-RefuseOrdinary $r 'RawHandleRefused'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.FirstQuery.StartUTC='not-a-time';Must-RefuseOrdinary $r 'InvalidUtcMetadata'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries[0].Final.TerminationRequested=$true;Must-RefuseOrdinary $r 'ForcedExitRefused'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.Entries[0].Membership.Owned='true';Must-RefuseOrdinary $r 'NestedStrictTypes'
$r=Copy-OrdinaryEnvelope $good;$r.MemberIdentityLedger.QueryCount='3';Must-RefuseOrdinary $r 'NumericCoercionRefused'
$legacy=[pscustomobject]@{ShutdownSchema=3;ShutdownStatePolicy='owned-final-signaled-exit-image-optional-v2';RootId=1;RootExited=$true;TimedOut=$false;CleanupVerified=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveAfterCleanup=0;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();NaturalObservationMs=0;ShutdownElapsedMs=0;MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@();MembersAtRootWait=@()}
Assert-ProcessCompletion $legacy;Check-Ordinary $true 'LegacySchema3Unchanged'
$legacy|Add-Member RunIdentity 'run-a';Must-RefuseGate $legacy 'LegacyNewMarkerRefused'
Write-Output 'LIMIT: pure fake receipt evaluation only; Acceptance=false.'
