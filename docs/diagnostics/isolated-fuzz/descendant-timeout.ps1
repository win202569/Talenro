. (Join-Path $PSScriptRoot 'child-readiness.ps1')
function Get-DescendantTimeoutResult([object]$Plan,[object]$Receipt) {
    if($Plan.UseScriptFiles -isnot [bool] -or -not $Plan.UseScriptFiles){throw 'Revised descendant case requires explicit file mode'}
    $evidence=Get-ChildReadinessEvidence $Plan $Receipt
    $total=[UInt64]0
    $countValid=[UInt64]::TryParse([string]$Receipt.TotalProcesses,[ref]$total)
    [pscustomobject]@{RunRoot=$Plan.RunRoot;DiagnosticOnly=$true;Acceptance=$false;OriginalTestPassed=$false;RegressionPassed=($evidence.FixturePassed -and $countValid -and $total -ge 2);TotalProcesses=$total;FixtureEvidence=$evidence}
}
function Invoke-DescendantTimeoutRegression([string]$RunRoot) {
    $root=Get-ChildReadinessRoot $RunRoot $true
    $null=Invoke-ChildReadinessFixture $root $true
    $planPath=Get-IsolatedFullPath (Join-Path $root 'plan.json')
    $receiptPath=Get-IsolatedFullPath (Join-Path $root 'process/receipt.json')
    foreach($path in @($planPath,$receiptPath)){if([IsolatedFuzzFileInfo]::Links($path) -ne 1){throw 'Linked regression input refused'}}
    $plan=Get-Content -LiteralPath $planPath -Raw|ConvertFrom-Json
    $receipt=Get-Content -LiteralPath $receiptPath -Raw|ConvertFrom-Json
    if($plan.RunRoot -ine $root){throw 'Saved regression plan belongs to another root'}
    $result=Get-DescendantTimeoutResult $plan $receipt
    $result|ConvertTo-Json -Depth 6|Set-Content -LiteralPath (Join-Path $root 'regression-result.json')
    return $result
}
