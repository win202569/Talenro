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
Write-Output 'LIMIT: synthetic receipts only; no process query or native Run invoked.'
