$ErrorActionPreference='Stop'
if(-not(Test-Path -LiteralPath (Join-Path $PSScriptRoot 'ordinary-ledger.cs'))){throw 'ASSERTION FAILED: planned ordinary ledger implementation missing'}
if(-not('OrdinaryMemberLedger' -as [type])){
    $sources=@('process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs','ordinary-ledger.cs')
    if(Test-Path -LiteralPath (Join-Path $PSScriptRoot 'ordinary-ledger-session.cs')){$sources+='ordinary-ledger-session.cs'}
    $sources+='ordinary-ledger-fixture.cs'
    Add-Type -Path @($sources|ForEach-Object{Join-Path $PSScriptRoot $_})
}
function Check-Ordinary([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
