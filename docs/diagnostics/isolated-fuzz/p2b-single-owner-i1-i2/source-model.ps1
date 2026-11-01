# Read-only context construction, deliberately outside the pure predicates.
# Merely dot-sourcing this file reads no files and executes no stored code.
function Get-ODPinnedSHA256([string]$Path,[Collections.IDictionary]$Manifest){
    $matches=@($Manifest.Pins|Where-Object {$_.VerificationPath -ieq $Path})
    if($matches.Count -eq 0){throw 'OD: source model input not pinned'}
    foreach($pin in $matches){
        if($pin.ExpectedState -cne 'Present' -or $pin.SHA256 -cnotmatch '^[A-F0-9]{64}$' -or $pin.SHA256 -cne $matches[0].SHA256){throw 'OD: source model pin conflict'}
    }
    $actual=(Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
    if($actual -cne $matches[0].SHA256){throw 'OD: source model input changed'}
    return $actual
}
function Get-ODSourceOwner($Node){
    $parent=$Node.Parent
    while($null -ne $parent){
        if($parent -is [Management.Automation.Language.FunctionDefinitionAst]){return $parent.Name}
        $parent=$parent.Parent
    }
    return '<ScriptBlock>'
}
function Get-ODFailureExtent($Node){
    $candidate=$Node;$parent=$Node.Parent
    while($null -ne $parent -and $parent -isnot [Management.Automation.Language.StatementBlockAst] -and $parent -isnot [Management.Automation.Language.NamedBlockAst]){
        if($parent -is [Management.Automation.Language.StatementAst]){$candidate=$parent}
        $parent=$parent.Parent
    }
    return $candidate.Extent
}
function Read-ODSourceModel([string]$Repo,[string[]]$RelativePaths,[Collections.IDictionary]$Manifest){
    $models=[ordered]@{}
    foreach($relative in $RelativePaths){
        $path=[IO.Path]::GetFullPath((Join-Path $Repo $relative))
        $sha=Get-ODPinnedSHA256 $path $Manifest
        $tokens=$null;$issues=$null
        $ast=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$issues)
        if($issues.Count -ne 0){throw 'OD: source model AST invalid'}
        $functions=@(foreach($node in $ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$true)){
            [ordered]@{Name=$node.Name;StartLine=$node.Extent.StartLineNumber;EndLine=$node.Extent.EndLineNumber}
        })
        $calls=@(foreach($node in $ast.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst]},$true)){
            $callee=$node.GetCommandName()
            if($null -eq $callee -and $node.CommandElements.Count -gt 0 -and $node.CommandElements[0] -is [Management.Automation.Language.VariableExpressionAst] -and $node.CommandElements[0].VariablePath.UserPath -ceq 'entry'){$callee='@fixed-entry'}
            if($null -ne $callee){[ordered]@{Caller=(Get-ODSourceOwner $node);Callee=$callee;Line=$node.Extent.StartLineNumber;Column=$node.Extent.StartColumnNumber}}
        })
        $faults=@(foreach($node in $ast.FindAll({param($n)
            $n -is [Management.Automation.Language.InvokeMemberExpressionAst] -or
            $n -is [Management.Automation.Language.ThrowStatementAst] -or
            ($n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'ConvertFrom-Json')
        },$true)){
            $extent=Get-ODFailureExtent $node
            $operation=if($node -is [Management.Automation.Language.InvokeMemberExpressionAst]){$node.Member.Value}elseif($node -is [Management.Automation.Language.ThrowStatementAst]){'throw'}else{'ConvertFrom-Json'}
            [ordered]@{Function=(Get-ODSourceOwner $node);Operation=$operation;Line=$extent.StartLineNumber;Column=$extent.StartColumnNumber;EndLine=$extent.EndLineNumber;EndColumn=$extent.EndColumnNumber;Text=$extent.Text}
        })
        if((Get-ODPinnedSHA256 $path $Manifest) -cne $sha){throw 'OD: source model changed during read'}
        $models[$path]=[ordered]@{SHA256=$sha;Functions=$functions;Calls=$calls;Faults=$faults}
    }
    return $models
}
function New-ODFixedContext([string]$Repo,[string]$Profile,[Collections.IDictionary]$Manifest,[string]$FrozenSHA256,[bool]$Historical){
    if($Profile -cnotin @('entry-migration-core','entry-migration-owned','entry-migration-live')){throw 'OD: source model unsupported profile'}
    $Repo=[IO.Path]::GetFullPath($Repo)
    # The two callers below choose these fixed namespaces. Runtime never falls
    # back to historical context based on report fields or failed predicates.
    $runtime=if($Historical){'scripts/private/offline-diagnostics'}else{'scripts/offline-diagnostics-v4/private'}
    $leaf=if($Historical){'docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration'}else{'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'}
    $entryRelative=if($Historical){'scripts/diagnose-offline.ps1'}else{'scripts/offline-diagnostics-v4/diagnose-offline.ps1'}
    $packetRelative=if($Historical){'.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration'}else{'.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'}
    $relativePaths=@($entryRelative)
    foreach($name in @('session.psm1','sinks.ps1','owned-step.ps1','sequence.ps1','profiles.ps1','observations.ps1')){$relativePaths+=($runtime+'/'+$name)}
    if(-not $Historical){foreach($name in @('write-boundary.ps1','record-owner.ps1')){$relativePaths+=($runtime+'/'+$name)}}
    foreach($name in @('test-helpers.ps1','core.tests.ps1','owned.tests.ps1','live.tests.ps1','prefix.tests.ps1','entry.tests.ps1','capture.ps1')){$relativePaths+=($leaf+'/'+$name)}
    $sourceModel=Read-ODSourceModel $Repo $relativePaths $Manifest
    $entry=[IO.Path]::GetFullPath((Join-Path $Repo $entryRelative))
    $kernel=[IO.Path]::GetFullPath((Join-Path $Repo ($runtime+'/session.psm1')))
    $packetRoot=[IO.Path]::GetFullPath((Join-Path $Repo $packetRelative))
    $manifestPath=Join-Path $packetRoot $(if($Historical){'runtime-green.frozen.json'}else{'repair-green.frozen.json'})
    if($Historical){$FrozenSHA256=Get-ODPinnedSHA256 $manifestPath $Manifest}
    elseif((Get-FileHash -LiteralPath $manifestPath -Algorithm SHA256).Hash -cne $FrozenSHA256){throw 'OD: source model runtime freeze mismatch'}
    $literalPath=Join-Path $Repo 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/expected-observations.json'
    $null=Get-ODPinnedSHA256 $literalPath $Manifest
    $literal=Get-Content -LiteralPath $literalPath -Raw|ConvertFrom-Json -AsHashtable
    $null=Get-ODPinnedSHA256 $literalPath $Manifest
    $oraclePath=Join-Path $Repo 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/expected-output.json'
    $null=Get-ODPinnedSHA256 $oraclePath $Manifest
    $oracle=Get-Content -LiteralPath $oraclePath -Raw|ConvertFrom-Json -AsHashtable
    $null=Get-ODPinnedSHA256 $oraclePath $Manifest
    return [ordered]@{
        Repo=$Repo;PacketRoot=$packetRoot;ProfileRoot=[IO.Path]::GetFullPath((Join-Path $packetRoot ('profiles/'+$Profile)))
        Oracle=$oracle;Rows=$literal.Rows;Sources=[ordered]@{Sinks=(Join-Path $Repo ($runtime+'/sinks.ps1'));Core=(Join-Path $Repo ($leaf+'/core.tests.ps1'));Capture=(Join-Path $Repo ($leaf+'/capture.ps1'))}
        SourceModel=$sourceModel;Entry=[ordered]@{Path=$entry;SHA256=$sourceModel[$entry].SHA256}
        Kernel=[ordered]@{Path=$kernel;SHA256=$sourceModel[$kernel].SHA256};Manifest=[ordered]@{Path=$manifestPath;SHA256=$FrozenSHA256}
    }
}
function New-ODRuntimeContext([string]$Repo,[string]$Profile,[Collections.IDictionary]$Manifest,[string]$FrozenSHA256){
    return New-ODFixedContext $Repo $Profile $Manifest $FrozenSHA256 $false
}
function New-ODHistoricalContext([string]$Repo,[string]$Profile,[Collections.IDictionary]$Manifest){
    return New-ODFixedContext $Repo $Profile $Manifest '' $true
}
