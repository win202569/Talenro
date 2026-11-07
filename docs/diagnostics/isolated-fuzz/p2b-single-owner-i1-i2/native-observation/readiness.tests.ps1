param([ValidateSet('Red','Green')][string]$Phase='Green')
$ErrorActionPreference='Stop'
if(Test-Path (Join-Path $PSScriptRoot 'readiness.ps1')){. (Join-Path $PSScriptRoot 'readiness.ps1')}
if(-not(Get-Command Get-C12ObservationReadiness -ErrorAction SilentlyContinue)){throw 'ASSERT_READINESS_POLICY_ABSENT'}
# Fixture evidence tests the decision contract, not actual machine readiness.
function New-Fixture {
 $e=@{}
 foreach($k in @('Profile','Ownership','StartStopBounds','StorageEnforcement','OfflineAnalysis','PrivacyAcknowledgement','PrivilegeSeparation')){$e[$k]=@{Status='Verified';Kind='Observed';Source='fixture-contract-only';Issues=@()}}
 return $e
}
$count=0
function Assert-Decision($Name,$Evidence,$Expected) {
 $r=Get-C12ObservationReadiness $Evidence
 if($r.Status -cne $Expected -or $r.RuntimeAuthorized -isnot [bool] -or $r.RuntimeAuthorized){throw "ASSERT_$Name"}
 if($Expected -ceq 'BLOCKED' -and @($r.Blockers).Count -eq 0){throw "ASSERT_REASON_$Name"}
 if($Expected -ceq 'READY_FOR_RUNTIME_PLAN' -and @($r.Blockers).Count -ne 0){throw 'ASSERT_UNEXPECTED_BLOCKER'}
 $script:count++;"PASS $Name"
}
Assert-Decision 'complete-contract-never-authorizes' (New-Fixture) 'READY_FOR_RUNTIME_PLAN'
foreach($key in @('Profile','Ownership','StartStopBounds','StorageEnforcement','OfflineAnalysis','PrivacyAcknowledgement','PrivilegeSeparation')){
 foreach($variant in @('missing','failed','unknown-status','empty-source','synthetic','unknown-field')){
  $e=New-Fixture
  switch($variant){missing {$e.Remove($key)} failed {$e[$key].Status='Failed'} unknown-status {$e[$key].Status='maybe'} empty-source {$e[$key].Source=''} synthetic {$e[$key].Kind='Synthetic'} unknown-field {$e[$key].Extra='unknown'}}
  Assert-Decision "$key-$variant" $e 'BLOCKED'
 }
}
foreach($case in @(
 @('Ownership','INSTANCE_NAME_ONLY'),@('StorageEnforcement','BUFFER_ONLY'),@('OfflineAnalysis','HELP_ONLY'),
 @('PrivacyAcknowledgement','APPROVAL_MISSING'),@('StartStopBounds','CLI_EXIT_COLLECTOR_UNKNOWN'),
 @('StartStopBounds','GO_EXIT_COLLECTOR_UNKNOWN'),@('StartStopBounds','COLLECTOR_STOP_GO_UNKNOWN'),
 @('Ownership','PID_WITHOUT_CREATION_TIME'),@('OfflineAnalysis','LOST_EVENTS'),@('OfflineAnalysis','UNMAPPED_EVENTS')
)){$e=New-Fixture;$e[$case[0]].Issues=@($case[1]);Assert-Decision $case[1] $e 'BLOCKED'}
$e=New-Fixture;$e.Ownership.Kind='Static';Assert-Decision 'static-ownership' $e 'BLOCKED'
$e=New-Fixture;$e.StorageEnforcement.Kind='Static';Assert-Decision 'static-storage' $e 'BLOCKED'
$e=New-Fixture;$e.OfflineAnalysis.Kind='Static';Assert-Decision 'static-analysis' $e 'BLOCKED'
$e=New-Fixture;$e.Profile.Kind='Static';Assert-Decision 'static-profile-contract' $e 'READY_FOR_RUNTIME_PLAN'
$e=New-Fixture;$e.Extra=@{};Assert-Decision 'unknown-key' $e 'BLOCKED'
Assert-Decision 'null' $null 'BLOCKED'
"PASS readiness cases=$count phase=$Phase"
