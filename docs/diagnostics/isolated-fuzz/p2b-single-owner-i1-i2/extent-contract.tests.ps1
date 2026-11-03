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
function Invoke-ECCommandInvariants($repo,$leaf,$packet,$binding){
    $Phase='Green'
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();RootError=$null;Stopped=$false}
    $names=@('SavedErrorAccepted','CommandExtent','WrongColumnRejected','WrongLineRejected','WrongOperationRejected','WrongFunctionRejected','WrongSourceRejected','MethodsUnchanged','ThrowsUnchanged','ModelBoundaryUnchanged')
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

    if($state.Stopped){throw $state.RootError}
    EC-Need ($state.Groups.Count -eq 10 -and @($state.Groups|Where-Object Status -CNE PASS).Count -eq 0) 'ten command invariants'
    return ,$state.Groups.ToArray()
}

function EC-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Extent contract regression: '+$Label)}}
function Driver-Need([bool]$Condition,[string]$Message){EC-Need $Condition $Message}
function EC-Reject([scriptblock]$Body){
    $caught=$null;$emitted=@()
    try{$emitted=@(& $Body)}catch{$caught=$_}
    EC-Need ($null -ne $caught -and $emitted.Count -eq 0) 'must reject without output'
}
function EC-Definition($Tree,[string]$Name){
    $defs=@($Tree.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq $Name},$true))
    EC-Need ($defs.Count -eq 1) ('missing '+$Name)
    return $defs[0]
}
function EC-Calls($Tree,[string]$Name){
    return ,@($Tree.FindAll({param($node)$node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq $Name},$true))
}
function Invoke-ODExtentContractRegression {
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $repo=(Get-Location).Path;$leaf=Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
    $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $names=@('BindingAvailable','ManifestAgreement','SlotAgreement','OldRouteBlocked','ApprovalRejected','SpentRejected','StopRejected','HashRejected','ParentRejected','WriterRejected','EpochRejected','PinConflictsRejected','SourceWindow','GuardsUnchanged','CommandExtentRegression','HistoricalContext','RuntimeNoFallback','UnitLineage','ReceiptIdentity','BudgetIdentity','CreateNewIdentity','OriginalSuiteUnchanged')
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();RootError=$null;Stopped=$false;CommandGroups=@()}
    $setup=$null;$recording=$null;$before=[ordered]@{};$after=[ordered]@{};$historyBefore=[ordered]@{};$historyAfter=[ordered]@{}
    $binding=$null;$sources=$null;$code=2;$testHash=$null;$inputHash=$null;$phaseHash=$null
    try{
        $testPath=Join-Path $leaf 'extent-contract.tests.ps1';$inputPath=Join-Path $packet 'extent-contract.input.json'
        $testHash=(Get-FileHash $testPath).Hash;$inputHash=(Get-FileHash $inputPath).Hash
        $binding=Get-Content $inputPath -Raw|ConvertFrom-Json -AsHashtable
        EC-Need ($binding.TestSHA256 -ceq $testHash) 'test identity'
        $sources=$binding.BaselineSources
        if($Phase -ceq 'Green'){
            $phasePath=Join-Path $packet 'extent-contract-candidate.input.json';$phaseHash=(Get-FileHash $phasePath).Hash
            $candidate=Get-Content $phasePath -Raw|ConvertFrom-Json -AsHashtable
            EC-Need ($candidate.InputSHA256 -ceq $inputHash) 'candidate identity';$sources=$candidate.CandidateSources
        }
        $slot='extent-contract-binding-'+$Phase.ToLowerInvariant()
        $intent=Get-Content (Join-Path $packet ($slot+'.intent.json')) -Raw|ConvertFrom-Json -AsHashtable
        EC-Need ($intent.AttemptConsumed -eq $true -and $intent.Computation -eq $(if($Phase -ceq 'Red'){18}else{19}) -and $intent.Ceiling -eq 19 -and $intent.TestSHA256 -ceq $testHash -and $intent.InputSHA256 -ceq $inputHash -and $intent.PhaseInputSHA256 -ceq $phaseHash) 'consumed exact intent'
        foreach($path in $sources.Keys){$before[$path]=(Get-FileHash $path).Hash;EC-Need ($before[$path] -ceq $sources[$path]) ('source '+$path)}
        foreach($path in $binding.HistoricalFiles.Keys){$historyBefore[$path]=(Get-FileHash $path).Hash;EC-Need ($historyBefore[$path] -ceq $binding.HistoricalFiles[$path]) ('history '+$path)}
        $trees=@{};foreach($name in @('driver.ps1','capture.ps1','source-model.ps1','observation-contract.tests.ps1')){$trees[$name]=CE-Tree (Join-Path $leaf $name)}
        CE-Case $names[0] {$null=EC-Definition $trees['source-model.ps1'] 'Get-ODEpochManifestName'} $state
        if(-not $state.Stopped){
            foreach($pair in @(@('driver.ps1','Get-ODAttemptSlot'),@('driver.ps1','Assert-ODExtentContractFacts'),@('driver.ps1','Driver-AddPin'),@('source-model.ps1','Get-ODEpochManifestName'),@('source-model.ps1','Assert-ODExtentLineage'))){
                $d=EC-Definition $trees[$pair[0]] $pair[1];. ([scriptblock]::Create($d.Extent.Text))
            }
            $facts=[ordered]@{Schema=1;ApprovalValid=$true;HistoryValid=$true;CandidateValid=$true;ReceiptsValid=$true;AttemptConsumed=$false;NewStopPresent=$false}
            $parent=Get-Content (Join-Path $packet 'repair-green.frozen.json') -Raw|ConvertFrom-Json -AsHashtable
            $parentHash=$binding.ParentSHA256;$writerHash=$binding.WriterSHA256
            $proposed=[ordered]@{PreviousFrozenSHA256=$parentHash;ExtentContractRecovery=[ordered]@{ParentSHA256=$parentHash;WriterSHA256=$writerHash}}
            CE-Case $names[1] {
                EC-Need ((Get-ODEpochManifestName legacy) -ceq 'repair-green.frozen.json') 'legacy manifest'
                EC-Need ((Get-ODEpochManifestName extent-contract) -ceq 'extent-contract.frozen.json') 'new manifest'
            } $state
            CE-Case $names[2] {
                EC-Need ((Get-ODAttemptSlot contract-green extent-contract) -ceq 'extent-contract-green') 'new host slot'
                EC-Need ((Get-ODAttemptSlot writer-green legacy) -ceq 'writer-projection-green') 'writer slot'
                foreach($name in @('Driver-UnitEvidence','Driver-Recorder','Driver-Evidence','Driver-RequireFinal')){
                    $d=EC-Definition $trees['driver.ps1'] $name
                    foreach($call in (EC-Calls $d 'Get-ODAttemptSlot')){EC-Need ($call.Extent.Text -match '\$Epoch\b') ('epoch propagated '+$name)}
                }
                EC-Need (@(EC-Calls $trees['capture.ps1'] Invoke-ODContractTests|Where-Object {$_.Extent.Text -match '\$Epoch\b'}).Count -eq 1) 'capture suite epoch'
                EC-Need (@(EC-Calls $trees['observation-contract.tests.ps1'] New-ODIntentUnitContext|Where-Object {$_.Extent.Text -match '\$Epoch\b'}).Count -eq 1) 'suite unit epoch'
            } $state
            CE-Case $names[3] {
                EC-Need ((Get-ODAttemptSlot contract-green legacy) -ceq 'contract-green') 'legacy slot unchanged'
                EC-Need (@(EC-Calls $trees['driver.ps1'] Driver-PinBindingGate).Count -eq 1) 'legacy gate retained'
                EC-Need (Test-Path (Join-Path $packet 'pin-binding.stop.json')) 'legacy stopped'
                EC-Reject {Get-ODAttemptSlot core extent-contract}
            } $state
            CE-Case $names[4] {$f=CE-Copy $facts;$f.ApprovalValid=$false;EC-Reject {Assert-ODExtentContractFacts $f}} $state
            CE-Case $names[5] {
                foreach($occupied in @('intent.json','final.json','stdout.bin','stderr.bin','evidence.json','capture','unit','snapshots')){
                    $f=CE-Copy $facts;$f.AttemptConsumed=$true;EC-Reject {Assert-ODExtentContractFacts $f}
                    EC-Need ($binding.SpentKinds -ccontains $occupied) ('spent coverage '+$occupied)
                }
            } $state
            CE-Case $names[6] {$f=CE-Copy $facts;$f.NewStopPresent=$true;EC-Reject {Assert-ODExtentContractFacts $f}} $state
            CE-Case $names[7] {foreach($key in @('HistoryValid','CandidateValid','ReceiptsValid')){$f=CE-Copy $facts;$f[$key]=$false;EC-Reject {Assert-ODExtentContractFacts $f}}} $state
            CE-Case $names[8] {$v=CE-Copy $proposed;$v.PreviousFrozenSHA256=('0'*64);EC-Reject {Assert-ODExtentLineage $v $parent $parentHash $writerHash};EC-Reject {Assert-ODExtentLineage $proposed $parent ('0'*64) $writerHash}} $state
            CE-Case $names[9] {$v=CE-Copy $parent;$v.PreviousFrozenSHA256=('0'*64);EC-Reject {Assert-ODExtentLineage $proposed $v $parentHash $writerHash};EC-Reject {Assert-ODExtentLineage $proposed $parent $parentHash ('0'*64)}} $state
            CE-Case $names[10] {
                foreach($e in @('','unknown','Extent-Contract')){EC-Reject {Get-ODEpochManifestName $e};EC-Reject {Get-ODAttemptSlot contract-green $e}}
                foreach($g in @('writer-green','core','lifecycle','continuation')){EC-Reject {Get-ODAttemptSlot $g extent-contract}}
            } $state
            CE-Case $names[11] {
                $map=[Collections.Generic.Dictionary[string,object]]::new([StringComparer]::OrdinalIgnoreCase);$rows=[Collections.Generic.List[object]]::new()
                $path=Join-Path $repo 'not-created-pin-fixture';Driver-AddPin $map $rows $path ('A'*64) test;Driver-AddPin $map $rows $path ('A'*64) test
                EC-Need ($rows.Count -eq 1) 'identical pin unique';EC-Reject {Driver-AddPin $map $rows $path ('B'*64) test}
            } $state
            CE-Case $names[12] {
                $changed=@($sources.Keys|Where-Object {$sources[$_] -cne $binding.BaselineSources[$_]})
                EC-Need ($sources.Count -eq 31 -and $changed.Count -eq 4 -and @(Compare-Object $changed $binding.AllowedSources).Count -eq 0) 'four-source window'
            } $state
            CE-Case $names[13] {
                foreach($name in $binding.UnchangedFunctions.Keys){
                    $item=$binding.UnchangedFunctions[$name];$d=EC-Definition $trees[$item.File] $name
                    EC-Need ($d.Extent.Text -ceq $item.Text) ('unchanged guard '+$name)
                }
                $driverText=[IO.File]::ReadAllText((Join-Path $leaf 'driver.ps1'))
                $tail=$driverText.Substring($driverText.IndexOf('$process=[Diagnostics.Process]::new()'))
                $tail=$tail.Replace(",'-Epoch',"+'$Epoch','').Replace('Driver-Evidence $actual $profile $oracle $Grant $manifest $FrozenSHA256 $Epoch','Driver-Evidence $actual $profile $oracle $Grant $manifest $FrozenSHA256')
                $end=$tail.LastIndexOf(([string][char]10)+'}catch{')
                EC-Need ($tail.Substring(0,$end) -ceq $binding.ProcessTail) 'owned process tail unchanged'
            } $state
            CE-Case $names[14] {$state.CommandGroups=Invoke-ECCommandInvariants $repo $leaf $packet $binding} $state
            CE-Case $names[15] {
                $context=Get-CEContext (Join-Path $leaf 'source-model.ps1') $repo $parent
                EC-Need ($context.Manifest.Path.EndsWith('2026-10-31-offline-diagnostic-runtime-migration\runtime-green.frozen.json') -and $context.SourceModel.Count -eq 14) 'historical fixed context'
            } $state
            CE-Case $names[16] {
                $d=EC-Definition $trees['source-model.ps1'] New-ODRuntimeContext
                EC-Need (@(EC-Calls $d New-ODFixedContext|Where-Object {$_.Extent.Text -match '\$Epoch\b'}).Count -eq 1) 'runtime explicit epoch'
                EC-Need (@($d.FindAll({param($n)$n -is [Management.Automation.Language.TryStatementAst]},$true)).Count -eq 0) 'runtime no fallback'
                EC-Reject {Get-ODEpochManifestName invalid}
            } $state
            CE-Case $names[17] {Assert-ODExtentLineage $proposed $parent $parentHash $writerHash;$v=CE-Copy $proposed;$v.PreviousFrozenSHA256=$writerHash;EC-Reject {Assert-ODExtentLineage $v $parent $parentHash $writerHash}} $state
            CE-Case $names[18] {
                $redPath=Join-Path $packet 'extent-contract-binding-red.final.json';$red=Get-Content $redPath -Raw|ConvertFrom-Json -AsHashtable
                EC-Need ($candidate.RedFinalSHA256 -ceq (Get-FileHash $redPath).Hash -and $red.PASS -eq 0 -and $red.FAIL -eq 1 -and $red.NotRun -eq 21 -and $red.RootError.Message -ceq 'Extent contract regression: missing Get-ODEpochManifestName') 'exact RED18'
            } $state
            CE-Case $names[19] {
                $a=Get-Content (Join-Path $packet 'extent-contract.approval.json') -Raw|ConvertFrom-Json -AsHashtable
                EC-Need ($a.AddedAuxiliary -eq 2 -and $a.ComputationCeiling -eq 19 -and $a.AddedHosts -eq 1 -and $a.AddedSnapshots -eq 31 -and $a.AddedTargets -eq 0 -and $a.AddedCompilations -eq 0 -and $a.AddedReviewers -eq 0) 'finite new budget'
                EC-Need ($a.ConsumedAuxiliary -eq 17 -and $a.ConsumedHosts -eq 4 -and $a.ConsumedCompilations -eq 3 -and $a.ConsumedSnapshots -eq 186) 'cumulative not reset'
            } $state
            CE-Case $names[20] {
                $slots=@('extent-contract-binding-red','extent-contract-binding-green','extent-contract','extent-contract-green')
                EC-Need (@($slots|Select-Object -Unique).Count -eq 4) 'disjoint namespaces'
                $d=EC-Definition $trees['driver.ps1'] Driver-Once
                EC-Need ($d.Extent.Text -ceq $binding.OnceText -and $d.Extent.Text.Contains('[IO.FileMode]::CreateNew')) 'exclusive writer preserved'
                EC-Need ($binding.AppendPaths.Count -eq 30 -and @($binding.AppendPaths|Select-Object -Unique).Count -eq 30) 'unique provenance'
            } $state
            CE-Case $names[21] {
                $suite=[IO.File]::ReadAllText((Join-Path $leaf 'observation-contract.tests.ps1'))
                $paramLine="    param([string]"+'$Epoch'+"='legacy')"
                $normalized=$suite.Replace($paramLine+[Environment]::NewLine,'').Replace($paramLine+([string][char]10),'').Replace('(Get-ODEpochManifestName $Epoch)',"'repair-green.frozen.json'").Replace('New-ODIntentUnitContext $repo $manifest $FrozenSHA256 $Epoch','New-ODIntentUnitContext $repo $manifest $FrozenSHA256')
                EC-Need ($normalized -ceq $binding.SuiteText) 'original suite unchanged outside binding'
            } $state
        }
        $code=if($state.Stopped){1}else{0}
    }catch{$setup=$_;$code=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    try{
        if($null -ne $sources){foreach($path in $sources.Keys){$after[$path]=(Get-FileHash $path).Hash;EC-Need ($after[$path] -ceq $before[$path]) ('post source '+$path)}}
        if($null -ne $binding){foreach($path in $binding.HistoricalFiles.Keys){$historyAfter[$path]=(Get-FileHash $path).Hash;EC-Need ($historyAfter[$path] -ceq $historyBefore[$path]) ('post history '+$path)}}
        EC-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash) 'post immutable input/test'
    }catch{$recording=$_;$code=2}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;PhaseInputSHA256=$phaseHash;SourceSHA256=$before;PostSourceSHA256=$after;HistoricalSHA256=$historyBefore;PostHistoricalSHA256=$historyAfter;Groups=$state.Groups.ToArray();CommandGroups=$state.CommandGroups;PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$code;RootError=(CE-Error $state.RootError);SetupError=(CE-Error $setup);RecordingError=(CE-Error $recording)}
}
