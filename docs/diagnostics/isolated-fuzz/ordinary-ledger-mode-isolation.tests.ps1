# Breaks caught: new policy downgrading to legacy; diagnostic rows reused as ordinary.
. (Join-Path $PSScriptRoot 'ordinary-ledger-repair-fixture.ps1')
foreach($schema in @(0,2,3)){
    $legacy=[pscustomobject]@{TimedOut=$false;CleanupVerified=$true;ActiveAfterCleanup=0;DescendantsRemaining=0;RootId=1;RootExited=$true;NaturalShutdownVerified=$true;ShutdownEvidenceComplete=$true;TerminationRequested=$false;ActiveBeforeTermination=0;MembersBeforeTermination=@();MembersAfterCleanup=@();NaturalObservationMs=0;ShutdownElapsedMs=0;MemberEvidenceTruncated=$false;ShutdownMemberEvidence=@();MembersAtRootWait=@()}
    if($schema -ne 0){$legacy|Add-Member ShutdownSchema $schema;$legacy|Add-Member ShutdownStatePolicy $(if($schema -eq 2){'owned-signaled-exit-image-optional-v1'}else{'owned-final-signaled-exit-image-optional-v2'})}
    Assert-ProcessCompletion $legacy;Check-Ordinary $true ('GenuineLegacy-'+$schema)
    $bad=Copy-OrdinaryRepairEnvelope $legacy
    if($schema -eq 0){$bad|Add-Member ShutdownStatePolicy 'ordinary-owned-ledger-final-exit-v1'}else{$bad.ShutdownStatePolicy='ordinary-owned-ledger-final-exit-v1'}
    $refused=$false;try{Assert-ProcessCompletion $bad}catch{$refused=$true}
    Check-Ordinary $refused ('PolicyOnlyDowngrade-'+$schema)
}
$good=New-OrdinaryRootEnvelope
foreach($container in @('MemberEvidence','ShutdownMemberEvidence')){
    $empty=Copy-OrdinaryRepairEnvelope $good;$empty|Add-Member $container @()
    Check-Ordinary (Test-OrdinaryLedgerEvidence $empty).EvidenceComplete ('EmptyLegacyContainer-'+$container)
    $bad=Copy-OrdinaryRepairEnvelope $good;$bad|Add-Member $container @([pscustomobject]@{Id=2;DiagnosticShutdown=[pscustomobject]@{Attempted=$true;State='exited'}})
    Check-Ordinary (-not (Test-OrdinaryLedgerEvidence $bad).EvidenceComplete) ('DiagnosticRowRefused-'+$container)
}
Write-Output 'LIMIT: pure fake gate/evaluator checks, no workload acceptance.'
