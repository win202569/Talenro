Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$source=Join-Path $PSScriptRoot prepare.ps1
$tokens=$null;$parseErrors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$tokens,[ref]$parseErrors)
if($parseErrors.Count){throw 'Test source must parse'}
$defs=@($ast.FindAll({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Convert-ODInheritedPin'},$true))
if($defs.Count -ne 1){throw 'Exactly one real converter required'}
. ([scriptblock]::Create($defs[0].Extent.Text))
function Test-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw $Message}}
$manifestPath=Join-Path $repo '.superpowers/sdd/2026-10-19-p2b-prestart-cause-repair/classification-green.frozen.json'
Test-Need ((Get-FileHash -LiteralPath $manifestPath -Algorithm SHA256).Hash -ceq '01B99203F29B6073843756E347054F8489F5825C325B6F4084DF9C8018A07123') 'Real inherited manifest identity'
$old=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
$before=ConvertTo-Json -InputObject $old.Pins -Depth 20 -Compress
Test-Need ($old.Pins.Count -eq 1930) 'Real1930pins'
Test-Need (@($old.Pins | Where-Object {-not $_.PSObject.Properties['Role']}).Count -eq 295) 'Real295optional omissions'
$checked=0
foreach($pin in $old.Pins){
    $actual=Convert-ODInheritedPin $pin
    Test-Need (($actual.PSObject.Properties.Name -join ',') -ceq 'VerificationPath,SHA256,Role,ExpectedState') 'Ordered normalized fields'
    Test-Need ($actual.VerificationPath -ceq [IO.Path]::GetFullPath($pin.VerificationPath) -and $actual.SHA256 -ceq $pin.SHA256 -and $actual.ExpectedState -ceq 'Present') 'Path/hash/state preserved'
    $role=$pin.PSObject.Properties['Role']
    if($null -eq $role){Test-Need ($null -eq $actual.Role) 'Missing Role normalized null'}
    else{Test-Need ([object]::Equals($actual.Role,$role.Value)) 'Existing Role preserved'}
    Test-Need ((Get-FileHash -LiteralPath $pin.VerificationPath -Algorithm SHA256).Hash -ceq $pin.SHA256) 'Inherited identity still checked'
    $checked++
}
Test-Need ((ConvertTo-Json -InputObject $old.Pins -Depth 20 -Compress) -ceq $before) 'No inherited input mutation'
$literalPath=Join-Path $repo 'AGENTS.md'
$literalHash='AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
foreach($sample in @(
    [pscustomobject]@{VerificationPath=$literalPath;SHA256=$literalHash;Role='literal-role'},
    [pscustomobject]@{VerificationPath=$literalPath;SHA256=$literalHash;Role=$null},
    [pscustomobject]@{VerificationPath=$literalPath;SHA256=$literalHash}
)){
    $inputBefore=ConvertTo-Json $sample -Compress
    $actual=Convert-ODInheritedPin $sample
    Test-Need ($actual.VerificationPath -ceq $literalPath -and $actual.SHA256 -ceq $literalHash -and $actual.ExpectedState -ceq 'Present') 'Literal required fields'
    $property=$sample.PSObject.Properties['Role']
    if($null -eq $property){Test-Need ($null -eq $actual.Role) 'Literal absent Role'}
    else{Test-Need ([object]::Equals($property.Value,$actual.Role)) 'Literal string/null Role'}
    Test-Need ((ConvertTo-Json $sample -Compress) -ceq $inputBefore) 'Literal no mutation'
}
$rejected=0
foreach($invalid in @(
    [pscustomobject]@{SHA256=$literalHash},
    [pscustomobject]@{VerificationPath='';SHA256=$literalHash},
    [pscustomobject]@{VerificationPath=' ';SHA256=$literalHash},
    [pscustomobject]@{VerificationPath=$null;SHA256=$literalHash},
    [pscustomobject]@{VerificationPath=$literalPath},
    [pscustomobject]@{VerificationPath=$literalPath;SHA256=''},
    [pscustomobject]@{VerificationPath=$literalPath;SHA256=$null},
    [pscustomobject]@{VerificationPath=$literalPath;SHA256='NOT-A-HASH'}
)){
    $caught=$null
    try{$null=Convert-ODInheritedPin $invalid}catch{$caught=$_}
    Test-Need ($null -ne $caught -and $caught.Exception.Message -ceq 'OD: required pin identity invalid') 'Mandatory identity rejected'
    $rejected++
}
[pscustomobject][ordered]@{Status='PASS';InheritedPins=$checked;MissingOptionalRole=295;LiteralRoleCases=3;RequiredIdentityRejected=$rejected;InputMutation=$false;CaptureHosts=0;Children=0;PreparationBodyInvoked=$false}
