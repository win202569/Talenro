# Mechanical diagnostic-copy generation/static verification only;never imports
# or invokes the product. Copies are ignored artifacts,not product edits/fixes.
param([switch]$VerifyOnly)
$ErrorActionPreference='Stop'
$projectRoot=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$copyRoot=Join-Path $projectRoot '.superpowers/sdd/2026-10-17-synthetic-capture-context-profile/copies'
$functionNames=@('Get-OfflineCaptureValueType','Get-OfflineEvidenceValueType','Assert-ContractPassiveProperties','Assert-OfflineEvidencePassiveValue','Assert-ContractShape','Get-ContractPacket','Assert-ContractEquivalent')
$sources=[ordered]@{
    'output-channel-classifier.ps1'='F323C9C46190DE06F25BFA55A2B7983707B1FF5877548F9703870F9BB771A0BF'
    'output-evidence-package.ps1'='857543A9DDA9B9C849B2507C6EE419C1F73AA0D0C8599CD9D88738841D021198'
    'output-capture-simulation.ps1'='9DA8A74B984A3111F00B54AA160490DF327BE17A5A07A044B510946EB26A2A13'
    'capture-adapter-contract.ps1'='116FCAAA6506ADC6969D718AC61F8C79908BDED48ADA1204C2E8A60F33A35BB2'
}
$entryTemplate=@'

# CTX-ENTRY-__INDEX__
$__ctxFrame=$null
if($script:ctxActive){
    $__ctxFrame=[pscustomobject]@{Slot=__INDEX__;Start=[long]0;Children=[long]0}
    $__ctxFrame.Start=[Diagnostics.Stopwatch]::GetTimestamp()
    $script:ctxEntries[__INDEX__]++
    $script:ctxFrames.Add($__ctxFrame)
    if($script:ctxFrames.Count -gt $script:ctxMaxDepth){$script:ctxMaxDepth=$script:ctxFrames.Count}
}
try {
# CTX-BODY-__INDEX__

'@
$exitTemplate=@'

# CTX-FINALLY-__INDEX__
}finally{
    if($null -ne $__ctxFrame){
        $__ctxElapsed=[Diagnostics.Stopwatch]::GetTimestamp()-$__ctxFrame.Start
        if($script:ctxFrames.Count -eq 0 -or -not [object]::ReferenceEquals($script:ctxFrames[$script:ctxFrames.Count-1],$__ctxFrame) -or $__ctxElapsed -lt $__ctxFrame.Children){$script:ctxFault=$true;throw 'Diagnostic nesting inconsistency'}
        $script:ctxFrames.RemoveAt($script:ctxFrames.Count-1)
        $script:ctxExits[__INDEX__]++
        $script:ctxInclusive[__INDEX__]+=$__ctxElapsed
        $script:ctxSubtracted[__INDEX__]+=($__ctxElapsed-$__ctxFrame.Children)
        if($script:ctxFrames.Count -gt 0){$script:ctxFrames[$script:ctxFrames.Count-1].Children+=$__ctxElapsed}
    }
}
# CTX-END-__INDEX__

'@
$plans=New-Object 'System.Collections.Generic.List[object]';$seen=New-Object 'System.Collections.Generic.List[string]'
foreach($name in $sources.Keys){
    $path=Join-Path $PSScriptRoot $name
    if((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash -cne $sources[$name]){throw ('Original source hash changed: '+$name)}
    $text=[IO.File]::ReadAllText($path)
    if($text.Contains('__ctx') -or $text.Contains('# CTX-')){throw 'Reserved diagnostic marker collision'}
    $tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseInput($text,[ref]$tokens,[ref]$errors)
    if($errors.Count){throw ('Original parse error: '+$name)}
    $edits=New-Object 'System.Collections.Generic.List[object]';$blocks=New-Object 'System.Collections.Generic.List[string]'
    foreach($definition in $ast.FindAll({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst]},$true)){
        $slot=[Array]::IndexOf($functionNames,$definition.Name)
        if($slot -lt 0){continue}
        if($seen.Contains($definition.Name) -or $null -ne $definition.Body.ParamBlock -or $null -ne $definition.Body.BeginBlock -or $null -ne $definition.Body.ProcessBlock){throw 'Unexpected target definition shape'}
        $seen.Add($definition.Name)
        $entry=$entryTemplate.Replace('__INDEX__',[string]$slot);$exit=$exitTemplate.Replace('__INDEX__',[string]$slot)
        $edits.Add([pscustomobject]@{Offset=$definition.Body.Extent.StartOffset+1;Text=$entry})
        $edits.Add([pscustomobject]@{Offset=$definition.Body.Extent.EndOffset-1;Text=$exit})
        $blocks.Add($entry);$blocks.Add($exit)
    }
    $copy=$text
    foreach($edit in @($edits|Sort-Object Offset -Descending)){$copy=$copy.Insert($edit.Offset,$edit.Text)}
    $restored=$copy;foreach($block in $blocks){$restored=$restored.Replace($block,'')}
    if(-not [string]::Equals($restored,$text,[StringComparison]::Ordinal)){throw 'Insertion removal did not exactly restore original text'}
    $copyTokens=$null;$copyErrors=$null;$copyAst=[Management.Automation.Language.Parser]::ParseInput($copy,[ref]$copyTokens,[ref]$copyErrors)
    if($copyErrors.Count){throw ('Diagnostic copy parse error: '+$name)}
    if($name -ceq 'capture-adapter-contract.ps1'){
        $originalVerifier=@($ast.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Test-SyntheticCaptureRecord'},$true))
        $copyVerifier=@($copyAst.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Test-SyntheticCaptureRecord'},$true))
        if($originalVerifier.Count -ne 1 -or $copyVerifier.Count -ne 1 -or -not [string]::Equals($originalVerifier[0].Extent.Text,$copyVerifier[0].Extent.Text,[StringComparison]::Ordinal)){throw 'Public verifier changed'}
    }
    $plans.Add([pscustomobject]@{Name=$name;OriginalSHA256=$sources[$name];Text=$copy;TargetFunctions=$edits.Count/2})
}
if($seen.Count -ne 7){throw 'Missing instrumentation target'}
# All inputs/transforms validated in memory before creating any copy. This is a
# bulk mechanical insertion,with exact text restoration checked above.
if(-not $VerifyOnly){
    if(Test-Path -LiteralPath $copyRoot){throw 'Copy destination already exists;do not overwrite'}
    $null=New-Item -ItemType Directory -Path $copyRoot
    foreach($plan in $plans){[IO.File]::WriteAllText((Join-Path $copyRoot $plan.Name),$plan.Text,(New-Object Text.UTF8Encoding($true)))}
}
foreach($plan in $plans){
    $copyPath=Join-Path $copyRoot $plan.Name
    if(-not (Test-Path -LiteralPath $copyPath -PathType Leaf) -or -not [string]::Equals([IO.File]::ReadAllText($copyPath),$plan.Text,[StringComparison]::Ordinal)){throw 'Saved copy differs from validated transform'}
    [pscustomobject]@{Name=$plan.Name;OriginalSHA256=$plan.OriginalSHA256;CopySHA256=(Get-FileHash -LiteralPath $copyPath -Algorithm SHA256).Hash;TargetFunctions=$plan.TargetFunctions}
}
