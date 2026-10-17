# Native entry is inspected, never invoked. Real session/disposal use fake APIs.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
. (Join-Path $PSScriptRoot 'run.ps1')
$entry=[IsolatedFuzzProcess].GetMethod('RunWithOrdinaryIdentityLedger')
Check-Ordinary ($null -ne $entry -and $entry.GetParameters().Count -eq 7) 'ExplicitEntryOnly'
$native=New-Object 'IsolatedFuzzProcess+Result'
Check-Ordinary ($native.ShutdownSchema -eq 3 -and $null -eq $native.MemberIdentityLedger -and $null -eq $native.RunIdentity -and -not $native.LifetimeDiagnosticOnly) 'LegacyDefaults'
Check-Ordinary (@([IsolatedFuzzProcess].GetMethods()|Where-Object Name -eq Run).Count -eq 2 -and $null -ne [IsolatedFuzzProcess].GetMethod('RunWithRetainedHandleDiagnostics')) 'ModeIsolation'
$legacy=ConvertTo-RecordedProcessReceipt $native 'out' 'err' 'fixture' @() @{}
Check-Ordinary ($null -eq $legacy.PSObject.Properties['MemberIdentityLedger'] -and $null -eq $legacy.PSObject.Properties['RunIdentity']) 'LegacyNoMarkers'
$f=[OrdinaryLedgerFixture]::Create();$f.Ledger.Poll(1000)
[OrdinaryCleanupFixture]::Expected=New-Object InvalidOperationException('expected cleanup error')
$action=[Action][Delegate]::CreateDelegate([Action],[OrdinaryCleanupFixture].GetMethod('Throw'))
$caught=$false;try{[OrdinaryLedgerTeardown]::Run($action,$f.Ledger)}catch{$e=$_.Exception;while($e.InnerException){$e=$e.InnerException};$caught=[object]::ReferenceEquals($e,[OrdinaryCleanupFixture]::Expected)}
Check-Ordinary ($caught -and $f.Api.ReleaseCalls -eq 1 -and $f.Ledger.Report.TeardownAttempted) 'SessionFakeCleanup'
$native.ShutdownSchema=4;$native.ShutdownStatePolicy='ordinary-owned-ledger-final-exit-v1';$native.RunIdentity='run-a';$native.RootAssigned=$true;$native.MemberIdentityLedger=$f.Ledger.Report;$native.LedgerFinalOutcome=New-Object OrdinaryFinalOutcome
$saved=ConvertTo-RecordedProcessReceipt $native 'out' 'err' 'fixture' @() @{}
$refused=$false;try{Assert-ProcessCompletion $saved}catch{$refused=$true}
Check-Ordinary ($refused -and $saved.RunIdentity -ceq 'run-a' -and $saved.RootAssigned -and $saved.ShutdownSchema -eq 4 -and $null -eq $saved.PSObject.Properties['LifetimeDiagnostics']) 'SerializationRefusal'
# Supplemental static wiring checks only; do not imply native behavior proof.
$src=Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'process-native.cs')
Check-Ordinary ($src -match 'ordinaryLedger.Poll\(runtimeDeadline\)' -and $src -match 'ordinaryMode\?10:100' -and $src -match 'OrdinaryLedgerSession' -and $src -match 'OrdinaryLedgerTeardown.Run') 'STATIC OrdinaryHooks'
foreach($script in @('run.ps1','preflight.ps1')){$text=Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot $script);Check-Ordinary ($text.Contains('ordinary-ledger.cs') -and $text.Contains('ordinary-ledger-session.cs') -and $text.Contains('ordinary-ledger-receipt.ps1')) 'STATIC LoaderInventory'}
Write-Output 'LIMIT: fake/reflection/static only; no native entry or adapter invoked.'
