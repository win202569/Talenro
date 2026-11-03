# Finite local extent regression. Definitions only; no runtime bodies.
function CE-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Command extent regression: '+$Label)}}
function CE-Error($Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function CE-Equal($Left,$Right){return ((ConvertTo-Json -InputObject $Left -Depth 100 -Compress) -ceq (ConvertTo-Json -InputObject $Right -Depth 100 -Compress))}
function CE-Copy($Value){return ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $Value -Depth 80)}
function CE-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{$emitted=@(& $Body);CE-Need ($emitted.Count -eq 0) ('unexpected output '+$Name);$State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})}
    catch{$State.RootError=$_;$State.Stopped=$true;$State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})}
}
function CE-Reject([scriptblock]$Body){
    $caught=$null;$emitted=@()
    try{$emitted=@(& $Body)}catch{$caught=$_}
    CE-Need ($null -ne $caught -and $emitted.Count -eq 0) 'must reject'
    if($caught.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $caught.Exception.Message -cne 'OD: observation contract mismatch' -or $caught.FullyQualifiedErrorId -cne 'OD: observation contract mismatch'){throw $caught}
}
function CE-Tree([string]$Path){
    $tokens=$null;$issues=$null
    $tree=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$tokens,[ref]$issues)
    CE-Need ($issues.Count -eq 0) ('parse '+$Path)
    return $tree
}
function Get-CEContext([string]$ModelPath,[string]$Repo,[Collections.IDictionary]$Manifest){
    $tree=CE-Tree $ModelPath
    foreach($name in @('Get-ODPinnedSHA256','Get-ODSourceOwner','Get-ODFailureExtent','Read-ODSourceModel','New-ODFixedContext','New-ODHistoricalContext')){
        $defs=@($tree.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.FunctionDefinitionAst] -and $_.Name -ceq $name})
        CE-Need ($defs.Count -eq 1) ('unique helper '+$name)
        . ([scriptblock]::Create(([string][char]10 * ($defs[0].Extent.StartLineNumber-1))+$defs[0].Extent.Text))
    }
    return New-ODHistoricalContext $Repo 'entry-migration-core' $Manifest
}
function CE-FaultNodes($Tree){
    return ,@($Tree.FindAll({param($node)
        $node -is [Management.Automation.Language.InvokeMemberExpressionAst] -or
        $node -is [Management.Automation.Language.ThrowStatementAst] -or
        ($node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq 'ConvertFrom-Json')
    },$true))
}
function CE-Owner($Node){
    # Independent raw-AST owner derivation; no production helper invocation.
    for($ancestor=$Node.Parent;$null -ne $ancestor;$ancestor=$ancestor.Parent){
        if($ancestor -is [Management.Automation.Language.FunctionDefinitionAst]){return ($ancestor.Name -creplace '^(script|local|global|private):','')}
        if($ancestor -is [Management.Automation.Language.ScriptBlockAst]){
            if($ancestor.Parent -is [Management.Automation.Language.FunctionDefinitionAst]){return ($ancestor.Parent.Name -creplace '^(script|local|global|private):','')}
            return '<ScriptBlock>'
        }
    }
    return '<ScriptBlock>'
}
function CE-ExtentMatches($Fault,$Node){
    return ($Fault.Line -eq $Node.Extent.StartLineNumber -and $Fault.Column -eq $Node.Extent.StartColumnNumber -and $Fault.EndLine -eq $Node.Extent.EndLineNumber -and $Fault.EndColumn -eq $Node.Extent.EndColumnNumber -and $Fault.Text -ceq $Node.Extent.Text)
}
function Invoke-ODCommandExtentRegression {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();RootError=$null;Stopped=$false}
    $names=@('SavedErrorAccepted','CommandExtent','WrongColumnRejected','WrongLineRejected','WrongOperationRejected','WrongFunctionRejected','WrongSourceRejected','MethodsUnchanged','ThrowsUnchanged','ModelBoundaryUnchanged')
    $setup=$null;$recording=$null;$testHash=$null;$inputHash=$null;$phaseHash=$null
    $before=[ordered]@{};$after=[ordered]@{};$historyBefore=[ordered]@{};$historyAfter=[ordered]@{}
    $binding=$null;$expectedSources=$null;$code=2
    try{
        $repo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
        CE-Need ((Get-Location).Path -ceq $repo) 'original root'
        $leaf=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
        $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
        CE-Need (-not (Test-Path (Join-Path $packet 'command-extent.stop.json'))) 'no new stop'
        $testPath=Join-Path $leaf 'command-extent.tests.ps1';$inputPath=Join-Path $packet 'command-extent.input.json'
        $testHash=(Get-FileHash $testPath).Hash;$inputHash=(Get-FileHash $inputPath).Hash
        $binding=Get-Content $inputPath -Raw|ConvertFrom-Json -AsHashtable
        CE-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'input/test identity'
        $expectedSources=$binding.BaselineSources
        if($Phase -ceq 'Green'){
            $phasePath=Join-Path $packet 'command-extent-candidate.input.json';$phaseHash=(Get-FileHash $phasePath).Hash
            $candidate=Get-Content $phasePath -Raw|ConvertFrom-Json -AsHashtable
            CE-Need ($candidate.Schema -eq 1 -and $candidate.InputSHA256 -ceq $inputHash -and $candidate.RedFinalSHA256 -ceq (Get-FileHash (Join-Path $packet 'command-extent-red.final.json')).Hash) 'candidate binding'
            $expectedSources=$candidate.CandidateSources
        }
        $intent=Get-Content (Join-Path $packet ('command-extent-'+$Phase.ToLowerInvariant()+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
        CE-Need ($intent.Schema -eq 1 -and $intent.Phase -ceq $Phase -and $intent.AttemptConsumed -is [bool] -and $intent.AttemptConsumed -and $intent.Computation -eq $(if($Phase -ceq 'Red'){16}else{17}) -and $intent.Ceiling -eq 17 -and $intent.TestSHA256 -ceq $testHash -and $intent.InputSHA256 -ceq $inputHash -and $intent.PhaseInputSHA256 -ceq $phaseHash -and (CE-Equal $intent.SourceSHA256 $expectedSources)) 'consumed intent'
        CE-Need ($expectedSources.Count -eq 31) '31 source identities'
        foreach($path in $expectedSources.Keys){$before[$path]=(Get-FileHash (Join-Path $repo $path)).Hash;CE-Need ($before[$path] -ceq $expectedSources[$path]) ('source '+$path)}
        foreach($path in $binding.HistoricalFiles.Keys){$historyBefore[$path]=(Get-FileHash (Join-Path $repo $path)).Hash;CE-Need ($historyBefore[$path] -ceq $binding.HistoricalFiles[$path]) ('history '+$path)}
        $manifestPath=Join-Path $packet 'repair-green.frozen.json'
        CE-Need ((Get-FileHash $manifestPath).Hash -ceq $binding.ParentSHA256) 'original manifest'
        $manifest=Get-Content $manifestPath -Raw|ConvertFrom-Json -AsHashtable
        $context=Get-CEContext (Join-Path $leaf 'source-model.ps1') $repo $manifest
        CE-Need ($context.SourceModel.Count -eq 14) 'historical model inventory'
        $tree=CE-Tree (Join-Path $leaf 'observation-contract.ps1')
        foreach($name in @('OC-Need','OC-Keys','OC-Integer','OC-Expand','OC-SourceError')){
            $defs=@($tree.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.FunctionDefinitionAst] -and $_.Name -ceq $name})
            CE-Need ($defs.Count -eq 1) ('pure helper '+$name)
            . ([scriptblock]::Create(([string][char]10 * ($defs[0].Extent.StartLineNumber-1))+$defs[0].Extent.Text))
        }
        $saved=Get-Content (Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration/runtime-green.evidence.json') -Raw|ConvertFrom-Json -AsHashtable
        $returns=@($saved.Observations|Where-Object {$_.Name -ceq 'ProfileReturn'})
        CE-Need ($returns.Count -eq 1 -and $returns[0].Value.Profile -ceq 'entry-migration-core') 'canonical core fixture'
        $rows=ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $returns[0].Value.Observations -Depth 60)
        $actual=$rows[8].Value.View.PrimaryError;$wanted=$context.Rows['entry-migration-core'][8].Value.View.PrimaryError['$error']
        $diagnosis=Get-Content (Join-Path $packet 'fault-location.final.json') -Raw|ConvertFrom-Json -AsHashtable
        CE-Need ((CE-Equal $actual $diagnosis.FirstMismatch.ActualError) -and (CE-Equal $wanted $diagnosis.FirstMismatch.ExpectedError)) 'unchanged diagnosed fixture'
        CE-Case $names[0] {OC-SourceError $actual $wanted $context} $state
        if($Phase -ceq 'Green' -and -not $state.Stopped){
            CE-Case $names[1] {
                $tree=CE-Tree $context.Sources.Sinks
                $nodes=@($tree.FindAll({param($node)$node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq 'ConvertFrom-Json' -and $node.Extent.StartLineNumber -eq 36 -and $node.Extent.StartColumnNumber -eq 50},$true))
                CE-Need ($nodes.Count -eq 1 -and (CE-Owner $nodes[0]) -ceq 'Read-LSLedger') 'independent command node'
                $faults=@($context.SourceModel[$context.Sources.Sinks].Faults|Where-Object {$_.Function -ceq 'Read-LSLedger' -and $_.Operation -ceq 'ConvertFrom-Json'})
                CE-Need ($faults.Count -eq 1 -and (CE-ExtentMatches $faults[0] $nodes[0])) 'command extent not assignment'
            } $state
            CE-Case $names[2] {$bad=CE-Copy $actual;$bad.Position=[regex]::Replace($bad.Position,'\A(At [^\r\n]+:36 char:)50(\r?\n)','${1}9${2}');CE-Need ($bad.Position -cne $actual.Position) 'column mutated';CE-Reject {OC-SourceError $bad $wanted $context}} $state
            CE-Case $names[3] {$bad=CE-Copy $actual;$bad.Position=[regex]::Replace($bad.Position,'\A(At [^\r\n]+:)36( char:)','${1}35${2}');$bad.Stack=[regex]::Replace($bad.Stack,'\A(at [^\r\n]+: line )36(\r?\n)','${1}35${2}');CE-Need ($bad.Position -cne $actual.Position -and $bad.Stack -cne $actual.Stack) 'line mutated';CE-Reject {OC-SourceError $bad $wanted $context}} $state
            CE-Case $names[4] {$bad=CE-Copy $wanted;$bad.Operation='ReadAllText';CE-Reject {OC-SourceError $actual $bad $context}} $state
            CE-Case $names[5] {$bad=CE-Copy $wanted;$bad.Function='Write-LSLedger';CE-Reject {OC-SourceError $actual $bad $context}} $state
            CE-Case $names[6] {$bad=CE-Copy $wanted;$bad.Source='Core';CE-Need ($context.SourceModel.Contains($context.Sources.Core) -and $context.Sources.Core -cne $context.Sources.Sinks) 'different pinned source';CE-Reject {OC-SourceError $actual $bad $context}} $state
            if(-not $state.Stopped){
                $baseline=Get-CEContext (Join-Path $packet 'snapshots/repair-green/docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/source-model.ps1') $repo $manifest
                $nodesByPath=[ordered]@{}
                foreach($path in $context.SourceModel.Keys){$nodesByPath[$path]=CE-FaultNodes (CE-Tree $path)}
            }
            CE-Case $names[7] {
                $seen=0
                foreach($path in $context.SourceModel.Keys){$nodes=$nodesByPath[$path];for($index=0;$index -lt $nodes.Count;$index++){if($nodes[$index] -is [Management.Automation.Language.InvokeMemberExpressionAst]){$seen++;CE-Need (CE-Equal $baseline.SourceModel[$path].Faults[$index] $context.SourceModel[$path].Faults[$index]) 'method unchanged'}}}
                CE-Need ($seen -gt 0) 'nonvacuous methods'
            } $state
            CE-Case $names[8] {
                $seen=0
                foreach($path in $context.SourceModel.Keys){$nodes=$nodesByPath[$path];for($index=0;$index -lt $nodes.Count;$index++){if($nodes[$index] -is [Management.Automation.Language.ThrowStatementAst]){$seen++;CE-Need (CE-Equal $baseline.SourceModel[$path].Faults[$index] $context.SourceModel[$path].Faults[$index]) 'throw unchanged'}}}
                CE-Need ($seen -gt 0) 'nonvacuous throws'
            } $state
            CE-Case $names[9] {
                CE-Need (CE-Equal @($baseline.SourceModel.Keys) @($context.SourceModel.Keys)) 'same model keys'
                $changed=0
                foreach($path in $context.SourceModel.Keys){
                    $old=$baseline.SourceModel[$path];$new=$context.SourceModel[$path];$nodes=$nodesByPath[$path]
                    foreach($key in @('SHA256','Functions','Calls','DynamicCalls')){CE-Need (CE-Equal $old[$key] $new[$key]) ('same '+$key)}
                    CE-Need ($old.Faults.Count -eq $new.Faults.Count -and $nodes.Count -eq $new.Faults.Count) 'fault count/order'
                    for($index=0;$index -lt $nodes.Count;$index++){
                        $node=$nodes[$index];$previous=$old.Faults[$index];$current=$new.Faults[$index]
                        if($node -is [Management.Automation.Language.CommandAst]){
                            CE-Need ($current.Operation -ceq 'ConvertFrom-Json' -and $current.Operation -ceq $previous.Operation -and $current.Function -ceq $previous.Function -and $current.Function -ceq (CE-Owner $node) -and (CE-ExtentMatches $current $node)) 'independent command extent/owner'
                            if(-not (CE-Equal $previous $current)){$changed++}
                        }else{CE-Need (CE-Equal $previous $current) 'all other faults unchanged'}
                    }
                }
                CE-Need ($changed -gt 0) 'real command correction'
            } $state
        }
        $code=if($state.Stopped){1}else{0}
    }catch{$setup=$_;$code=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    try{
        if($null -ne $expectedSources){foreach($path in $expectedSources.Keys){$after[$path]=(Get-FileHash (Join-Path $repo $path)).Hash;CE-Need ($after[$path] -ceq $before[$path]) ('post source '+$path)}}
        if($null -ne $binding){foreach($path in $binding.HistoricalFiles.Keys){$historyAfter[$path]=(Get-FileHash (Join-Path $repo $path)).Hash;CE-Need ($historyAfter[$path] -ceq $historyBefore[$path]) ('post history '+$path)}}
        CE-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash) 'post test/input'
        if($Phase -ceq 'Green'){CE-Need ((Get-FileHash $phasePath).Hash -ceq $phaseHash) 'post candidate input'}
    }catch{$recording=$_;$code=2}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;PhaseInputSHA256=$phaseHash;SourceSHA256=$before;PostSourceSHA256=$after;HistoricalSHA256=$historyBefore;PostHistoricalSHA256=$historyAfter;Groups=$state.Groups.ToArray();PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$code;RootError=(CE-Error $state.RootError);SetupError=(CE-Error $setup);RecordingError=(CE-Error $recording)}
}
