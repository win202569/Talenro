# Breaks caught: zero-entry stale final queries; retrying/unbounded initial queries.
. (Join-Path $PSScriptRoot 'ordinary-ledger-repair-fixture.ps1')
$good=New-OrdinaryRootEnvelope
Check-Ordinary (Test-OrdinaryLedgerEvidence $good).EvidenceComplete 'RootOnlyValid'
foreach($name in @('FinalAccounting','FinalMemberQuery')){
    $bad=Copy-OrdinaryRepairEnvelope $good;$bad.LedgerFinalOutcome.$name.StartTimestamp=99
    Check-Ordinary (-not (Test-OrdinaryLedgerEvidence $bad).EvidenceComplete) ('StaleRootOnly-'+$name)
}
foreach($scenario in @('error234','truncated','overrun')){
    $f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[UInt64[]]@();$f.Api.Total=1
    switch($scenario){error234{$f.Api.Partial=$true} truncated{$f.Api.Truncated=$true} overrun{$f.Api.AdvanceOnQuery=501}}
    $s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$r=$s.Finish($true,$false,100)
    Check-Ordinary (-not $r.EvidenceComplete -and $f.Ledger.Report.Incomplete -and @($f.Api.Calls|Where-Object {$_ -ceq 'list'}).Count -eq 1 -and $null -ne $r.InitialMemberQuery) ('InitialQuerySticky-'+$scenario)
    $f.Ledger.Dispose()
}
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[UInt64[]]@();$f.Api.Total=1;$f.Api.AccountingError=234
$s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$r=$s.Finish($true,$false,100)
Check-Ordinary (-not $r.EvidenceComplete -and $r.InitialAccounting.Error -eq 234 -and $f.Api.Calls.Count -eq 1) 'InitialAccountingSticky'
$f.Ledger.Dispose()
$f=[OrdinaryLedgerFixture]::Create();$f.Api.Candidates=[UInt64[]]@();$f.Api.Total=1
$s=New-Object OrdinaryLedgerSession($f.Ledger,[IntPtr]1,$f.Api,$f.Clock);$r=$s.Finish($true,$false,100)
Check-Ordinary ($r.EvidenceComplete -and $r.InitialAccounting.TotalProcesses -eq 1 -and $r.InitialMembers.Count -eq 0 -and $r.InitialMemberQuery.Returned -eq 0) 'InitialFactsRecorded'
$src=Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'process-native.cs')
Check-Ordinary ($src -match 'else\s*\{?\s*accounting=ReadInitialJobQueries\(job,result\)' -and $src -match 'InitialAccounting' -and $src -match 'InitialMemberQuery') 'STATIC LegacyInitialQueriesIsolated'
Write-Output 'LIMIT: pure fake and static source wiring only; no native entry.'
