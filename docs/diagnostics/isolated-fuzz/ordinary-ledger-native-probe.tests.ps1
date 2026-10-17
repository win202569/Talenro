param([switch]$SetupOnly)
# Real planner/runner/saved validator with fake workload and file-link boundaries.
. (Join-Path $PSScriptRoot 'ordinary-ledger-native-probe-fixture.ps1')
$helper=Join-Path $PSScriptRoot 'ordinary-ledger-native-probe.ps1'
$script:ProbeCalls=0;$script:ProbeMode='pass';$script:ProbeDrift=$false;$script:ProbeArmed=$false;$script:ProbeExecutableOverride=$null;$script:ProbeCheckLock=$false;$script:ProbeLockObserved=$false
$workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$script:ProbeTestBase=Join-Path $workspace ('.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/probe-unit-'+[guid]::NewGuid().ToString('N'))
function Get-OrdinaryLedgerProbeScope {return $script:ProbeTestScope}
function Get-OrdinaryLedgerProbeFile([string]$Path){
    if($script:ProbeExecutableOverride -and $Path -ieq (Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe')){$Path=$script:ProbeExecutableOverride}
    $safe=Get-IsolatedFullPath $Path;$item=Get-Item -LiteralPath $safe -ErrorAction Stop
    if($item.PSIsContainer){throw 'Expected test file'}
    $hash=(Get-FileHash -LiteralPath $safe -Algorithm SHA256).Hash
    if($script:ProbeDrift -and $item.Name -ceq 'ordinary-ledger.cs'){$hash='0'*64}
    [pscustomobject]@{Path=$safe;SHA256=$hash;Length=$item.Length}
}
function New-ProbeTestScope([string]$Name){
    $script:ProbeTestScope=Join-Path $script:ProbeTestBase $Name
    New-Item -ItemType Directory -Path $script:ProbeTestScope -ErrorAction Stop|Out-Null
    $script:ProbeCalls=0;$script:ProbeMode='pass';$script:ProbeDrift=$false;$script:ProbeArmed=$false;$script:ProbeExecutableOverride=$null;$script:ProbeCheckLock=$false;$script:ProbeLockObserved=$false
}
function New-ProbeTestPlan([string]$Case='Root'){
    $suffix=if($Case -ceq 'Root'){'root'}else{'child'}
    Get-OrdinaryLedgerProbePlan (Join-Path $script:ProbeTestScope ('ledger-native-'+$suffix+'-001')) $Case
}
function Refuse-Probe([scriptblock]$Action,[string]$Name){$refused=$false;try{& $Action|Out-Null}catch{$refused=$true};Check-Ordinary $refused $Name}
function Invoke-OrdinaryLedgerProbeNative($Plan,[string]$Stdout,[string]$Stderr){
    if(-not $script:ProbeArmed){throw 'Unarmed fake native boundary'}
    $script:ProbeArmed=$false
    $script:ProbeCalls++
    if($script:ProbeCheckLock){
        try{$attempt=[IO.File]::Open($script:ProbeExecutableOverride,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read);$attempt.Dispose()}catch [IO.IOException]{$script:ProbeLockObserved=$true}
    }
    if($script:ProbeMode -ceq 'throw'){throw 'planned fake boundary failure'}
    if($script:ProbeMode -ceq 'missing'){return $null}
    [IO.File]::WriteAllText($Stdout,'fake stdout');[IO.File]::WriteAllText($Stderr,'')
    $receipt=New-OrdinaryProbeTestEnvelope -Child:($Plan.Case -ceq 'Child')
    foreach($pair in @(@('ExitCode',0),@('StartUTC','2026-10-17T00:00:00Z'),@('EndUTC','2026-10-17T00:00:01Z'),@('Executable',$Plan.Executable),@('Arguments',$Plan.Arguments),@('Environment',@{}),@('WorkingDirectory',$Plan.RunRoot))){$receipt|Add-Member -NotePropertyName $pair[0] -NotePropertyValue $pair[1]}
    if($script:ProbeMode -ceq 'cleanup'){$receipt.CleanupVerified=$false}
    $nonce=if($script:ProbeMode -ceq 'nonce'){'0'*32}else{$Plan.Nonce}
    $stages=if($Plan.Case -ceq 'Root'){@('root-exit.json')}else{@('child-ready.json','release.json','parent-exit.json','child-exit.json')}
    foreach($stage in $stages){$pidValue=if($stage -like 'child-*'){2}else{1};Write-OrdinaryLedgerProbeJSON (Join-Path $Plan.RunRoot $stage) ([pscustomobject]@{Nonce=$nonce;ProcessId=$pidValue;UTC='2026-10-17T00:00:00.200Z'})}
    return $receipt
}
$expectedBoundary=(Get-Command Invoke-OrdinaryLedgerProbeNative).ScriptBlock.ToString()
$expectedFileBoundary=(Get-Command Get-OrdinaryLedgerProbeFile).ScriptBlock
$expectedScopeBoundary=(Get-Command Get-OrdinaryLedgerProbeScope).ScriptBlock
if(Test-Path -LiteralPath $helper){. $helper}
$productionIndependentBoundary=(Get-Command Get-OrdinaryLedgerProbeFile).ScriptBlock
Check-Ordinary ((Get-Command Invoke-OrdinaryLedgerProbeNative).ScriptBlock.ToString() -ceq $expectedBoundary) 'ImportMustPreserveFailClosedBoundary'
Set-Item Function:Get-OrdinaryLedgerProbeFile -Value $expectedFileBoundary
Set-Item Function:Get-OrdinaryLedgerProbeScope -Value $expectedScopeBoundary
Check-Ordinary ($null -ne (Get-Command Get-OrdinaryLedgerProbeShell -ErrorAction SilentlyContinue)) 'DedicatedInboxShellValidatorRequired'
$script:ShellMetadataMode='pass';$script:ShellSignatureMode='pass'
function Get-OrdinaryLedgerProbeIdentity($Stream){
    if($script:ShellMetadataMode -ceq 'throw'){throw 'fake metadata unavailable'}
    [pscustomobject]@{Volume=7;IndexHigh=0;IndexLow=$(if($script:ShellMetadataMode -ceq 'identity'){99}else{42});Links=$(if($script:ShellMetadataMode -ceq 'links'){3}else{2})}
}
function Get-OrdinaryLedgerProbeSignature([string]$Path){
    [pscustomobject]@{Valid=($script:ShellSignatureMode -cne 'invalid');Subject=$(if($script:ShellSignatureMode -ceq 'subject'){'CN=Other'}else{'CN=Microsoft Windows, O=Microsoft Corporation, L=Redmond, S=Washington, C=US'});CertificateSHA256=$(if($script:ShellSignatureMode -ceq 'certificate'){'B'*64}else{'A'*64})}
}
$realBinding=(Get-Command Open-OrdinaryLedgerProbeFileBinding).ScriptBlock
function Open-OrdinaryLedgerProbeFileBinding([string]$Path,[string]$ExpectedSHA256=''){
    $actual=$Path
    if($script:ProbeExecutableOverride -and $Path -ieq (Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe')){$actual=$script:ProbeExecutableOverride}
    $record=& $realBinding $actual $ExpectedSHA256
    $record.Path=$Path;return $record
}
$shell=Get-OrdinaryLedgerProbeShell
Check-Ordinary ($shell.Identity.Links -eq 2 -and $shell.Identity.IndexLow -eq 42 -and $shell.Signer.CertificateSHA256 -ceq ('A'*64)) 'InboxMultiLinkBoundInsteadOfRejected'
foreach($mode in @('invalid','subject')){$script:ShellSignatureMode=$mode;Refuse-Probe {Get-OrdinaryLedgerProbeShell} ('UntrustedShellRefused-'+$mode)}
$script:ShellSignatureMode='pass';$script:ShellMetadataMode='throw'
Refuse-Probe {Get-OrdinaryLedgerProbeShell} 'MissingShellIdentityRefused'
$script:ShellMetadataMode='pass'
Refuse-Probe {Open-OrdinaryLedgerProbeShellBinding ([pscustomobject]@{Path='C:/wrong.exe'})} 'ShellExceptionCannotApplyToOtherPath'
$realIndependent=(Get-Command Get-OrdinaryLedgerProbeFile).ScriptBlock
# The general file boundary was already replaced above; test the saved production
# function below using only its link-count OS seam, never P/Invoke in this suite.
Set-Item Function:Get-OrdinaryLedgerProbeFile -Value $productionIndependentBoundary
function Get-OrdinaryLedgerProbeLinkCount([string]$Path){return 2}
Refuse-Probe {Get-OrdinaryLedgerProbeFile $helper} 'SourceMultiLinkStillRefused'
Set-Item Function:Get-OrdinaryLedgerProbeFile -Value $realIndependent
Check-Ordinary ($null -ne (Get-Command Get-OrdinaryLedgerProbePlan -ErrorAction SilentlyContinue)) 'NativeProbePlannerExists'
function Invoke-ProbeTestRun($Plan){
    $script:ProbeArmed=$true
    try{Invoke-OrdinaryLedgerProbe $Plan -RuntimeApproved -ApprovedPlanSHA256 (Get-OrdinaryLedgerProbePlanHash $Plan)}finally{$script:ProbeArmed=$false}
}
if($SetupOnly){return}
Check-Ordinary ($null -ne (Get-Command Assert-OrdinaryLedgerProbeFreshTypes -ErrorAction SilentlyContinue)) 'FreshTypeGuardExists'
Refuse-Probe {Assert-OrdinaryLedgerProbeFreshTypes} 'UnverifiedPreloadedTypesRefused'
Refuse-Probe {Invoke-OrdinaryLedgerProbeNative $null 'unused' 'unused'} 'UnarmedBoundaryRefused'
New-ProbeTestScope 'planning';$plan=New-ProbeTestPlan
$script:ProbeDrift=$true;Refuse-Probe {New-ProbeTestPlan} 'LoadedSourceDriftBeforePlanRefused';$script:ProbeDrift=$false
Check-Ordinary ($script:ProbeCalls -eq 0 -and -not(Test-Path -LiteralPath $plan.RunRoot) -and -not(Test-Path -LiteralPath $plan.ControlRoot)) 'ImportAndPlanningLaunchNothing'
Check-Ordinary ($plan.Case -ceq 'Root' -and $plan.LimitSeconds -eq 5 -and $plan.Executable -ceq (Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe') -and $plan.Arguments.Count -eq 4 -and ($plan.Arguments[0..2] -join '|') -ceq '-NoProfile|-NonInteractive|-EncodedCommand') 'FixedRootShellArguments'
$childPlan=New-ProbeTestPlan 'Child'
Check-Ordinary ($childPlan.LimitSeconds -eq 20 -and $childPlan.Case -ceq 'Child' -and -not $plan.Acceptance -and $plan.NoRetry -and $plan.SourceBinding.Count -eq 11) 'FixedChildPolicyAndBindings'
foreach($mutation in @('executable','arguments','limit','missing-source','source-hash','nonce','acceptance','shell-identity','shell-links','shell-signer')){
    $bad=$plan|ConvertTo-Json -Depth 12|ConvertFrom-Json
    switch($mutation){executable{$bad.Executable='C:/wrong.exe'} arguments{$bad.Arguments[0]='-Command'} limit{$bad.LimitSeconds=600} missing-source{$bad.SourceBinding=@($bad.SourceBinding|Select-Object -Skip 1)} source-hash{$bad.SourceBinding[0].SHA256='0'*64} nonce{$bad.Nonce='invalid'} acceptance{$bad.Acceptance=$true} shell-identity{$bad.ShellBinding.Identity.IndexLow=99} shell-links{$bad.ShellBinding.Identity.Links=3} shell-signer{$bad.ShellBinding.Signer.CertificateSHA256='B'*64}}
    Refuse-Probe {Assert-OrdinaryLedgerProbePlan $bad $true} ('PlanMutationRefused-'+$mutation)
}
Refuse-Probe {Get-OrdinaryLedgerProbePlan (Join-Path $script:ProbeTestScope 'metadata-runtime-retained-probe-001') 'Root'} 'WrongIdentityRefused'
Refuse-Probe {Invoke-OrdinaryLedgerProbe $plan} 'RuntimeNotApprovedRefused'
Refuse-Probe {Invoke-OrdinaryLedgerProbe $plan -RuntimeApproved -ApprovedPlanSHA256 ('0'*64)} 'WrongApprovalPinRefused'
Check-Ordinary ($script:ProbeCalls -eq 0 -and -not(Test-Path -LiteralPath $plan.RunRoot)) 'NoBoundaryBeforeApproval'
Refuse-Probe {Invoke-OrdinaryLedgerProbe $childPlan -RuntimeApproved -ApprovedPlanSHA256 (Get-OrdinaryLedgerProbePlanHash $childPlan)} 'ChildBeforeRootRefused'
foreach($mode in @('identity','links','throw')){
    $script:ShellMetadataMode=$mode
    Refuse-Probe {Invoke-ProbeTestRun $plan} ('ShellMetadataDriftBeforeConsumption-'+$mode)
    Check-Ordinary ($script:ProbeCalls -eq 0 -and -not(Test-Path -LiteralPath $plan.ControlRoot)) ('NoIdentityConsumedOnShellDrift-'+$mode)
}
$script:ShellMetadataMode='pass';$script:ShellSignatureMode='certificate'
Refuse-Probe {Invoke-ProbeTestRun $plan} 'SignerDriftBeforeConsumptionRefused'
$script:ShellSignatureMode='pass'
$result=Invoke-ProbeTestRun $plan
Check-Ordinary ($result.ObservationPassed -and $result.Finalized -and -not $result.Acceptance -and $script:ProbeCalls -eq 1) 'RootFakeRunFinalizedOnce'
$saved=Get-OrdinaryLedgerProbeSaved $plan.RunRoot
Check-Ordinary ($saved.Finalized -and $saved.ObservationPassed -and -not $saved.Acceptance) 'DurableSavedResultVerified'
$raw=Get-Content -Raw -LiteralPath (Join-Path $plan.RunRoot 'process/receipt.json')|ConvertFrom-Json
Refuse-Probe {Assert-ProcessCompletion $raw} 'Schema4StillNotAcceptance'
Refuse-Probe {New-ProbeTestPlan} 'ConsumedIdentityRefused'
Refuse-Probe {Invoke-OrdinaryLedgerProbe $plan -RuntimeApproved -ApprovedPlanSHA256 (Get-OrdinaryLedgerProbePlanHash $plan)} 'RunReplayRefused'
$childResult=Invoke-ProbeTestRun $childPlan
Check-Ordinary ($childResult.ObservationPassed -and $childResult.ChildEntryId -eq 1 -and -not $childResult.Acceptance -and $script:ProbeCalls -eq 2) 'ChildFakeSameRunEntryVerified'
Check-Ordinary ($childResult.ObservationClaim -ceq 'same-run-owned-handle-natural-completion' -and $childResult.ChildOutlivedRootVerified -is [bool] -and -not $childResult.ChildOutlivedRootVerified) 'NoExitOrderingClaim'
$script:ProbeDrift=$true;Refuse-Probe {Get-OrdinaryLedgerProbeSaved $plan.RunRoot} 'SavedSourceDriftRefused';$script:ProbeDrift=$false
[IO.File]::AppendAllText((Join-Path $plan.RunRoot 'result.json'),' ')
Refuse-Probe {Get-OrdinaryLedgerProbeSaved $plan.RunRoot} 'SavedByteTamperingRefused'
foreach($mode in @('throw','missing','nonce','cleanup')){
    New-ProbeTestScope $mode;$p=New-ProbeTestPlan;$script:ProbeMode=$mode
    Refuse-Probe {Invoke-ProbeTestRun $p} ('FakeFailureStopped-'+$mode)
    Check-Ordinary ($script:ProbeCalls -eq 1 -and (Test-Path -LiteralPath (Join-Path $p.ControlRoot 'consumed.json')) -and (Test-Path -LiteralPath (Join-Path $p.RunRoot 'stopped.json')) -and -not(Test-Path -LiteralPath (Join-Path $p.RunRoot 'completed.json'))) ('NoCompletionAfterFailure-'+$mode)
    Refuse-Probe {Get-OrdinaryLedgerProbeSaved $p.RunRoot} ('FailedSavedRefused-'+$mode)
}
New-ProbeTestScope 'missing-saved';$p=New-ProbeTestPlan
Invoke-ProbeTestRun $p|Out-Null
$receiptPath=Join-Path $p.RunRoot 'process/receipt.json';Move-Item -LiteralPath $receiptPath -Destination ($receiptPath+'.missing') -ErrorAction Stop
Refuse-Probe {Get-OrdinaryLedgerProbeSaved $p.RunRoot} 'MissingSavedReceiptRefused'
New-ProbeTestScope 'executable-lock'
$script:ProbeExecutableOverride=Join-Path $script:ProbeTestScope 'fake-shell.txt';[IO.File]::WriteAllText($script:ProbeExecutableOverride,'not an executable;fake boundary only')
$baseline=[IO.File]::Open($script:ProbeExecutableOverride,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read);$baseline.Dispose()
$p=New-ProbeTestPlan;$script:ProbeCheckLock=$true;Invoke-ProbeTestRun $p|Out-Null
Check-Ordinary $script:ProbeLockObserved 'ExecutableDenyWriteHeldThroughBoundary'
$released=[IO.File]::Open($script:ProbeExecutableOverride,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read);$released.Dispose()
Check-Ordinary $true 'ExecutableBindingReleasedAfterRun'
New-ProbeTestScope 'finalization';$p=New-ProbeTestPlan
$originalWriter=(Get-Command Write-OrdinaryLedgerProbeJSON).ScriptBlock
function Write-OrdinaryLedgerProbeJSON([string]$Path,$Value){if([IO.Path]::GetFileName($Path) -ceq 'completed.json'){throw 'planned completion write failure'};& $originalWriter $Path $Value}
Refuse-Probe {Invoke-ProbeTestRun $p} 'CompletionWriteFailureRefused'
Check-Ordinary (-not(Test-Path -LiteralPath (Join-Path $p.RunRoot 'completed.json')) -and (Test-Path -LiteralPath (Join-Path $p.RunRoot 'stopped.json'))) 'CompletionNotInferredFromProvisionalResult'
Write-Output 'LIMIT: fake workload and independent-file seam only; native entry and real child never invoked.'
