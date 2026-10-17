# Test-only timing regression/mutation controls. No product files are modified.
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'ordinary-ledger-native-probe.tests.ps1') -SetupOnly
$script:TimingMode='none';$script:TimingInjected=$false
$script:TimingObserved=New-Object 'System.Collections.Generic.List[string]'
$timingWriter=(Get-Command Write-OrdinaryLedgerProbeJSON).ScriptBlock
$timingSaved=(Get-Command Get-OrdinaryLedgerProbeSaved).ScriptBlock
$timingRunner=(Get-Command Invoke-OrdinaryLedgerProbe).ScriptBlock
function Assert-TimingFixtureLocked([string]$Stage){
    $denied=$false
    try{
        $writer=[IO.File]::Open($script:ProbeExecutableOverride,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read)
        $writer.Dispose()
    }catch [IO.IOException]{$denied=$true}
    if(-not $denied){throw ('ASSERTION FAILED: TimingLockHeld-'+$Stage)}
    $script:TimingObserved.Add($Stage)
}
function Write-OrdinaryLedgerProbeJSON([string]$Path,$Value){
    $leaf=[IO.Path]::GetFileName($Path)
    if($script:TimingMode -ceq 'finalization' -and $leaf -cin @('receipt.json','result.json','evidence-manifest.json','completed.json')){
        Assert-TimingFixtureLocked $leaf
    }
    & $timingWriter $Path $Value
    if($script:TimingMode -ceq 'drift' -and $leaf -ceq 'plan.json'){
        # Both initial binding and full plan assertion already succeeded.
        $script:ShellMetadataMode='identity';$script:TimingInjected=$true
    }
}
function Get-OrdinaryLedgerProbeSaved([string]$RunRoot){
    if($script:TimingMode -ceq 'finalization'){Assert-TimingFixtureLocked 'saved-validation'}
    & $timingSaved $RunRoot
}
function New-TimingScope([string]$Name){
    New-ProbeTestScope $Name
    $script:ShellMetadataMode='pass';$script:ShellSignatureMode='pass'
    $script:TimingMode='none';$script:TimingInjected=$false;$script:TimingObserved.Clear()
    $script:ProbeExecutableOverride=Join-Path $script:ProbeTestScope 'timing-shell.txt'
    [IO.File]::WriteAllText($script:ProbeExecutableOverride,'plaintext fixture;never executed')
}
function Test-TimingLateDrift([string]$Name){
    New-TimingScope $Name;$plan=New-ProbeTestPlan
    $script:TimingMode='drift';$message=''
    try{Invoke-ProbeTestRun $plan|Out-Null}catch{$message=$_.Exception.Message}
    finally{$script:ShellMetadataMode='pass';$script:TimingMode='none'}
    Check-Ordinary ($script:ProbeCalls -eq 0) 'TimingLateDriftNeverEntersNative'
    Check-Ordinary ($script:TimingInjected -and $message -ceq 'Shell identity/signature/hash/link drift') 'TimingLateDriftRejectedAtEntry'
    Check-Ordinary ((Test-Path -LiteralPath (Join-Path $plan.ControlRoot 'consumed.json')) -and (Test-Path -LiteralPath (Join-Path $plan.RunRoot 'stopped.json')) -and -not(Test-Path -LiteralPath (Join-Path $plan.RunRoot 'completed.json'))) 'TimingLateDriftPreservesConsumedFailure'
}
function Test-TimingFinalization([string]$Name){
    New-TimingScope $Name;$plan=New-ProbeTestPlan
    $baseline=[IO.File]::Open($script:ProbeExecutableOverride,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read);$baseline.Dispose()
    $script:TimingMode='finalization'
    try{$result=Invoke-ProbeTestRun $plan}finally{$script:TimingMode='none'}
    Check-Ordinary ($result.Finalized -and $result.ObservationPassed -and $script:ProbeCalls -eq 1) 'TimingFinalizationCompletesOnce'
    foreach($stage in @('receipt.json','result.json','evidence-manifest.json','completed.json','saved-validation')){
        Check-Ordinary ($script:TimingObserved.Contains($stage)) ('TimingLockObserved-'+$stage)
    }
    $released=[IO.File]::Open($script:ProbeExecutableOverride,[IO.FileMode]::Open,[IO.FileAccess]::ReadWrite,[IO.FileShare]::Read);$released.Dispose()
    Check-Ordinary $true 'TimingBindingReleasedOnlyAfterFinalization'
}
function Test-TimingMutant([string]$From,[string]$To,[scriptblock]$Scenario,[string]$ExpectedAssertion,[string]$Name){
    $text=$timingRunner.ToString()
    if([regex]::Matches($text,[regex]::Escape($From)).Count -ne 1){throw 'Mutation target must be unique'}
    $mutant=[scriptblock]::Create($text.Replace($From,$To));$message=''
    try{
        Set-Item Function:script:Invoke-OrdinaryLedgerProbe -Value $mutant
        try{& $Scenario|Out-Null}catch{$message=$_.Exception.Message}
    }finally{Set-Item Function:script:Invoke-OrdinaryLedgerProbe -Value $timingRunner}
    Check-Ordinary ($message -ceq ('ASSERTION FAILED: '+$ExpectedAssertion)) ('TimingMutationCaught-'+$Name)
}
# Observe each assertion fail against the targeted in-memory regression first.
# Dummy MemoryStream only replaces the removed entry binding's disposal slot.
Test-TimingMutant '$entryBinding=Open-OrdinaryLedgerProbeShellBinding $Plan.ShellBinding' '$entryBinding=[pscustomobject]@{Stream=(New-Object IO.MemoryStream)}' {Test-TimingLateDrift 'mutant-entry'} 'TimingLateDriftNeverEntersNative' 'RemovedEntryRecheck'
Test-TimingMutant "Write-OrdinaryLedgerProbeJSON (Join-Path `$process 'receipt.json') `$receipt" "foreach(`$file in `$bindings){`$file.Stream.Dispose()};Write-OrdinaryLedgerProbeJSON (Join-Path `$process 'receipt.json') `$receipt" {Test-TimingFinalization 'mutant-finalization'} 'TimingLockHeld-receipt.json' 'PrematureBindingRelease'
# Unmodified production runner must then satisfy both scenarios.
Test-TimingLateDrift 'real-entry'
Test-TimingFinalization 'real-finalization'
Write-Output 'LIMIT: in-memory mutants,plaintext fixtures and fake OS/workload seams only;no native probe.'
