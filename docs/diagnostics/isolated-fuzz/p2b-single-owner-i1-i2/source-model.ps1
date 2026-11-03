function Get-ODEpochManifestName([string]$Epoch='legacy'){
    if($Epoch -ceq 'legacy'){return 'repair-green.frozen.json'}
    if($Epoch -ceq 'extent-contract'){return 'extent-contract.frozen.json'}
    throw 'OD: invalid explicit epoch'
}
function Assert-ODExtentLineage([Collections.IDictionary]$Manifest,[Collections.IDictionary]$Parent,[string]$ParentSHA256,[string]$WriterSHA256){
    if($ParentSHA256 -cne 'A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556' -or $WriterSHA256 -cne '182B85886304BFE3CBFB248DDA70F2EB99DF3E2158F574DBE299519DE7CE0461' -or $Manifest.PreviousFrozenSHA256 -cne $ParentSHA256 -or $Parent.PreviousFrozenSHA256 -cne $WriterSHA256 -or $Manifest.ExtentContractRecovery.ParentSHA256 -cne $ParentSHA256 -or $Manifest.ExtentContractRecovery.WriterSHA256 -cne $WriterSHA256){throw 'OD: extent contract lineage mismatch'}
}
# Read-only context construction, deliberately outside the pure predicates.
# Merely dot-sourcing this file reads no files and executes no stored code.
function Get-ODPinnedSHA256([string]$Path,[Collections.IDictionary]$Manifest){
    $pinRows=@($Manifest.Pins|Where-Object {$_.VerificationPath -ieq $Path})
    if($pinRows.Count -eq 0){throw 'OD: source model input not pinned'}
    foreach($pin in $pinRows){
        if($pin.ExpectedState -cne 'Present' -or $pin.SHA256 -cnotmatch '^[A-F0-9]{64}$' -or $pin.SHA256 -cne $pinRows[0].SHA256){throw 'OD: source model pin conflict'}
    }
    $actual=(Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
    if($actual -cne $pinRows[0].SHA256){throw 'OD: source model input changed'}
    return $actual
}
function Get-ODSourceOwner($Node){
    $parent=$Node.Parent
    while($null -ne $parent){
        if($parent -is [Management.Automation.Language.FunctionDefinitionAst]){return ($parent.Name -creplace '^(script|local|global|private):','')}
        if($parent -is [Management.Automation.Language.ScriptBlockAst]){
            if($parent.Parent -is [Management.Automation.Language.FunctionDefinitionAst]){return ($parent.Parent.Name -creplace '^(script|local|global|private):','')}
            return '<ScriptBlock>'
        }
        $parent=$parent.Parent
    }
    return '<ScriptBlock>'
}
function Get-ODFailureExtent($Node){
    if($Node -is [Management.Automation.Language.CommandAst] -and $Node.GetCommandName() -ceq 'ConvertFrom-Json'){return $Node.Extent}
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
            [ordered]@{Name=($node.Name -creplace '^(script|local|global|private):','');StartLine=$node.Extent.StartLineNumber;EndLine=$node.Extent.EndLineNumber}
        })
        $calls=@(foreach($node in $ast.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst]},$true)){
            $callee=$node.GetCommandName()
            if($null -eq $callee -and $node.CommandElements.Count -gt 0 -and $node.CommandElements[0] -is [Management.Automation.Language.VariableExpressionAst] -and $node.CommandElements[0].VariablePath.UserPath -ceq 'entry'){$callee='@fixed-entry'}
            if($null -eq $callee -and $node.CommandElements.Count -gt 0 -and $node.CommandElements[0] -is [Management.Automation.Language.VariableExpressionAst] -and $node.CommandElements[0].VariablePath.UserPath -ceq 'script:OriginalFactory'){$callee='@fixed-record-factory'}
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
        $dynamicCalls=@(if($relative -ceq 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1'){
            foreach($node in $ast.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst]},$true)){
                if((Get-ODSourceOwner $node) -cne 'Invoke-ODOwnerControls' -or $node.CommandElements.Count -lt 2){continue}
                $variable=$node.CommandElements[0];$literal=$node.CommandElements[1]
                if($variable -isnot [Management.Automation.Language.VariableExpressionAst] -or $variable.VariablePath.UserPath -cne 'controls' -or $literal -isnot [Management.Automation.Language.ScriptBlockExpressionAst]){continue}
                $parameters=$literal.ScriptBlock.ParamBlock
                if($null -eq $parameters -or $parameters.Parameters.Count -ne 1 -or $parameters.Parameters[0].Name.VariablePath.UserPath -cne 'Name'){continue}
                $body=$literal.ScriptBlock.Extent
                [ordered]@{Source=$path;SHA256=$sha;Owner='Invoke-ODOwnerControls';Variable='controls';Line=$node.Extent.StartLineNumber;Column=$node.Extent.StartColumnNumber;BodyStartLine=$body.StartLineNumber;BodyStartColumn=$body.StartColumnNumber;BodyEndLine=$body.EndLineNumber;BodyEndColumn=$body.EndColumnNumber}
            }
        })
        if((Get-ODPinnedSHA256 $path $Manifest) -cne $sha){throw 'OD: source model changed during read'}
        $models[$path]=[ordered]@{SHA256=$sha;Functions=$functions;Calls=$calls;Faults=$faults;DynamicCalls=$dynamicCalls}
    }
    return $models
}
function Get-ODIntentCallBinding([Collections.IDictionary]$SourceModel,[string]$Source){
    if(-not $SourceModel.Contains($Source) -or -not $Source.EndsWith('/docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1'.Replace('/',[IO.Path]::DirectorySeparatorChar),[StringComparison]::Ordinal)){throw 'OD: unit call binding mismatch'}
    $model=$SourceModel[$Source]
    if(-not $model.Contains('DynamicCalls') -or $model.DynamicCalls -isnot [array] -or $model.DynamicCalls.Count -ne 1){throw 'OD: unit call binding mismatch'}
    $call=$model.DynamicCalls[0]
    if($call.Source -cne $Source -or $call.SHA256 -cne $model.SHA256 -or $call.Owner -cne 'Invoke-ODOwnerControls' -or $call.Variable -cne 'controls'){throw 'OD: unit call binding mismatch'}
    return $call
}
function New-ODFixedContext([string]$Repo,[string]$Profile,[Collections.IDictionary]$Manifest,[string]$FrozenSHA256,[bool]$Historical,[string]$Epoch='legacy'){
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
    $manifestPath=Join-Path $packetRoot $(if($Historical){'runtime-green.frozen.json'}else{Get-ODEpochManifestName $Epoch})
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
function New-ODRuntimeContext([string]$Repo,[string]$Profile,[Collections.IDictionary]$Manifest,[string]$FrozenSHA256,[string]$Epoch='legacy'){
    return New-ODFixedContext $Repo $Profile $Manifest $FrozenSHA256 $false $Epoch
}
function New-ODHistoricalContext([string]$Repo,[string]$Profile,[Collections.IDictionary]$Manifest){
    return New-ODFixedContext $Repo $Profile $Manifest '' $true
}

# This explicit data-unit context is never selected by the runtime consumer.
# It describes the already planned writer-green IntentFailureNoWork case only.
function New-ODIntentUnitContext([string]$Repo,[Collections.IDictionary]$Manifest,[string]$RepairSHA256,[string]$Epoch='legacy'){
    $context=New-ODRuntimeContext $Repo 'entry-migration-core' $Manifest $RepairSHA256 $Epoch
    $writerManifest=Join-Path $context.PacketRoot 'writer-projection-repair.frozen.json'
    $writerHash=(Get-FileHash -LiteralPath $writerManifest -Algorithm SHA256).Hash
    if($Epoch -ceq 'extent-contract'){
        $parentPath=Join-Path $context.PacketRoot 'repair-green.frozen.json'
        $parentHash=(Get-FileHash -LiteralPath $parentPath -Algorithm SHA256).Hash
        $parent=Get-Content -LiteralPath $parentPath -Raw|ConvertFrom-Json -AsHashtable
        Assert-ODExtentLineage $Manifest $parent $parentHash $writerHash
        if((Get-FileHash -LiteralPath $parentPath -Algorithm SHA256).Hash -cne $parentHash){throw 'OD: unit parent drift'}
    }elseif($Manifest.PreviousFrozenSHA256 -cne $writerHash){throw 'OD: unit context prior freeze mismatch'}
    $context.Manifest=[ordered]@{Path=$writerManifest;SHA256=$writerHash}
    $context.ProfileRoot=[IO.Path]::GetFullPath((Join-Path $context.PacketRoot 'unit/writer-projection-green/owner-controls/intent-profile/entry-migration-core'))
    $extra=Read-ODSourceModel $Repo @('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/case-module.ps1') $Manifest
    foreach($key in $extra.Keys){$context.SourceModel[$key]=$extra[$key]}
    # The isolated case evaluates only the already frozen profile function AST.
    # Its anonymous source label and relative lines are derived independently,
    # not accepted as arbitrary missing-file locations in runtime contexts.
    $profilePath=[IO.Path]::GetFullPath((Join-Path $Repo 'scripts/offline-diagnostics-v4/private/profiles.ps1'))
    $profileModel=$context.SourceModel[$profilePath]
    $definition=@($profileModel.Functions|Where-Object Name -CEQ Invoke-ODProfile)
    if($definition.Count -ne 1){throw 'OD: unit profile AST identity mismatch'}
    $offset=$definition[0].StartLine-1
    $relativeCalls=@(foreach($call in $profileModel.Calls){if($call.Caller -ceq 'Invoke-ODProfile'){[ordered]@{Caller=$call.Caller;Callee=$call.Callee;Line=($call.Line-$offset);Column=$call.Column}}})
    $relativeFaults=@(foreach($fault in $profileModel.Faults){if($fault.Function -ceq 'Invoke-ODProfile'){[ordered]@{Function=$fault.Function;Operation=$fault.Operation;Line=($fault.Line-$offset);Column=$fault.Column;EndLine=($fault.EndLine-$offset);EndColumn=$fault.EndColumn;Text=$fault.Text}}})
    $context.SourceModel['<No file>']=[ordered]@{SHA256=$profileModel.SHA256;Origin=$profilePath;Functions=@([ordered]@{Name='Invoke-ODProfile';StartLine=1;EndLine=($definition[0].EndLine-$offset)});Calls=$relativeCalls;Faults=$relativeFaults}
    $context.UnitBindings=[ordered]@{ProfileFunction='Invoke-ODProfile';AnonymousSource='<No file>';FactorySource=[IO.Path]::GetFullPath((Join-Path $Repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'));FactoryFunction='New-ODRecordStream'}
    $context.UnitBindings.OwnerControlCall=Get-ODIntentCallBinding $context.SourceModel ([IO.Path]::GetFullPath((Join-Path $Repo 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1')))
    $context.IntentFailure=[ordered]@{
        Type='System.Management.Automation.MethodInvocationException';ErrorId='IOException'
        Message=('Exception calling "Open" with "4" argument(s): "The file ''{0}'' already exists."' -f (Join-Path $context.PacketRoot 'unit/writer-projection-green/owner-controls/intent-existing.bin'))
        Source=[IO.Path]::GetFullPath((Join-Path $Repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'));Function='New-ODRecordStream';Operation='Open'
    }
    return $context
}
