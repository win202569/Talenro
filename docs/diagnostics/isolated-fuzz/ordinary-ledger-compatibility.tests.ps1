# Actual producer copy logic and validators,never a controller/workload invocation.
. (Join-Path $PSScriptRoot 'ordinary-ledger-fixture.ps1')
. (Join-Path $PSScriptRoot 'run.ps1')
function Read-RepairAst([string]$file){$tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot $file),[ref]$tokens,[ref]$errors);if($errors.Count){throw 'Source parse failure'};return $ast}
function Test-LegacyWriter([string]$file,[string]$functionName,$native){
    $ast=Read-RepairAst $file
    $fn=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $functionName},$true))[0]
    $copies=@($fn.FindAll({param($n)$n -is [Management.Automation.Language.ForEachStatementAst] -and $n.Condition.Extent.Text -ceq '$native.PSObject.Properties'},$true))
    $fields=@{}
    if($copies.Count -eq 1){. ([scriptblock]::Create($copies[0].Extent.Text))}
    else{
        $assign=@($fn.FindAll({param($n)$n -is [Management.Automation.Language.AssignmentStatementAst] -and $n.Left.Extent.Text -ceq '$fields' -and $n.Right.Extent.Text -match 'Get-IsolatedNativeReceiptFields'},$true))
        if($assign.Count -ne 1){throw 'Writer copy boundary missing'}
        . ([scriptblock]::Create($assign[0].Extent.Text))
    }
    $result=[pscustomobject]$fields
    $accepted=$true;try{Assert-ProcessCompletion $result}catch{$accepted=$false}
    Check-Ordinary ($accepted -and $null -eq $result.PSObject.Properties['MemberIdentityLedger'] -and $null -eq $result.PSObject.Properties['RunIdentity']) ($file+' LegacyReceipt') | Out-Host
    return $result
}
$native=New-Object 'IsolatedFuzzProcess+Result'
$native.RootId=1;$native.RootExited=$true;$native.CleanupVerified=$true;$native.NaturalShutdownVerified=$true;$native.ShutdownEvidenceComplete=$true
$native.MembersAtRootWait=[UInt64[]]@();$native.MembersBeforeTermination=[UInt64[]]@();$native.MembersAfterCleanup=[UInt64[]]@();$native.ShutdownMemberEvidence=@()
$matrix=Test-LegacyWriter 'matrix-check.ps1' 'Invoke-MatrixRecordedProcess' $native
$leaf=Test-LegacyWriter 'leaf-ordinary-check.ps1' 'Invoke-LeafOrdinaryRecordedProcess' $native
$ast=Read-RepairAst 'leaf-ordinary-check.ps1'
$shape=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Assert-LeafOrdinaryNativeShape'},$true))[0]
. ([scriptblock]::Create($shape.Extent.Text))
Assert-LeafOrdinaryNativeShape $leaf ([IsolatedFuzzProcess+Result]);Check-Ordinary $true 'LegacyShapeWithoutAdditiveFields'
$native.RunIdentity='new-run';$native.ShutdownSchema=4;$native.MemberIdentityLedger=New-Object OrdinaryLedgerReport
$fields=Get-IsolatedNativeReceiptFields $native;$refused=$false;try{Assert-ProcessCompletion ([pscustomobject]$fields)}catch{$refused=$true};Check-Ordinary ($refused -and $fields.ContainsKey('RunIdentity')) 'NewMarkersStillRefused'
$names=@(Get-OrdinaryLedgerSourceNames);Check-Ordinary ($names.Count -eq 3 -and $names -contains 'ordinary-ledger.cs' -and $names -contains 'ordinary-ledger-session.cs' -and $names -contains 'ordinary-ledger-receipt.ps1') 'LoadedDependenciesNamed'
$binding=@(foreach($name in $names){[pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath (Join-Path $PSScriptRoot $name) -Algorithm SHA256).Hash}})
Assert-OrdinaryLedgerSources $binding;Check-Ordinary $true 'CompleteLedgerBinding'
foreach($scenario in @('missing','hash')){$bad=@($binding|ForEach-Object{[pscustomobject]@{Path=$_.Path;SHA256=$_.SHA256}});if($scenario -eq 'missing'){$bad=@($bad|Select-Object -Skip 1)}else{$bad[0].SHA256='0'*64};$refused=$false;try{Assert-OrdinaryLedgerSources $bad}catch{$refused=$true};Check-Ordinary $refused ('BindingDrift-'+$scenario)}
foreach($file in @('leaf-ordinary-check.ps1','retained-child-observation.ps1','leaf-retained-compile.ps1','leaf-stage-retained-compile.ps1','leaf-stage-build-only.ps1')){
    $text=Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot $file)
    foreach($name in $names){Check-Ordinary ($text.Contains("'"+$name+"'")) ('STATIC Inventory-'+$file+'-'+$name)}
    Check-Ordinary ($text.Contains('Assert-OrdinaryLedgerSources')) ('STATIC BindingConsumer-'+$file)
}
Write-Output 'LIMIT: AST/pure fake/filesystem hash checks;no controller/native execution.'
