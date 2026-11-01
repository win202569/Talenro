# Definitions-only verifier; never imported by capture or a runtime profile.
function Assert-ODImportedDefinition([object]$Expected,[object]$Loaded,[object]$Comparison){
    $pair=@($Expected,$Loaded);$operands=[Collections.Generic.List[object]]::new()
    foreach($node in $pair){
        $kind=$null;$text=$null;$name=$null;$body=$null;$parameters=$null;$file=$null
        if($null -ne $node){
            $kind=$node.GetType().FullName
            if($node -is [Management.Automation.Language.Ast]){$text=$node.Extent.Text;$file=$node.Extent.File}
            if($node -is [Management.Automation.Language.FunctionDefinitionAst]){
                $name=$node.Name;$body=$node.Body.Extent.Text
                if($null -ne $node.Parameters){$parameters=@($node.Parameters|ForEach-Object {$_.Extent.Text})}
            }
        }
        $operands.Add([ordered]@{NodeType=$kind;Name=$name;Text=$text;BodyText=$body;InlineParameters=$parameters;SourceFile=$file})
    }
    $Comparison.Value=[ordered]@{Expected=$operands[0];Loaded=$operands[1]}
    if($Expected -isnot [Management.Automation.Language.FunctionDefinitionAst] -or $Loaded -isnot [Management.Automation.Language.FunctionDefinitionAst]){throw 'OD: imported definition mismatch'}
    if(@('Invoke-ODProfile','Capture-Value','Capture-Once','Capture-ErrorSummary') -cnotcontains $Expected.Name -or $Loaded.Name -cne ('script:'+$Expected.Name) -or $Expected.Body.Extent.Text -cne $Loaded.Body.Extent.Text){throw 'OD: imported definition mismatch'}
    if(($null -eq $Expected.Parameters) -ne ($null -eq $Loaded.Parameters)){throw 'OD: imported definition mismatch'}
    if($null -ne $Expected.Parameters){
        if($Expected.Parameters.Count -ne $Loaded.Parameters.Count){throw 'OD: imported definition mismatch'}
        for($i=0;$i -lt $Expected.Parameters.Count;$i++){if($Expected.Parameters[$i].Extent.Text -cne $Loaded.Parameters[$i].Extent.Text){throw 'OD: imported definition mismatch'}}
    }
}
function Invoke-ODVerifierCase([string]$Name,[scriptblock]$Body,[object]$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    $State.CurrentGroup=$Name
    try{
        $values=@(& $Body)
        if($values.Count -ne 0){throw ('OD: verifier stray output '+$Name)}
        $State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})
    }catch{
        if($null -eq $State.RootError){$State.RootError=$_}
        $State.Stopped=$true
        $State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})
    }
}
function Convert-ODVerifierError([object]$Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;ErrorId=$Record.FullyQualifiedErrorId;Message=$Record.Exception.Message;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function Convert-ODVerifierDiagnostic([object]$State,[scriptblock]$Projector={param($r)Convert-ODVerifierError $r}){
    $root=$null
    try{$root=& $Projector $State.RootError}catch{
        $State.RecordingError=$_
        $root=Convert-ODVerifierError $State.RootError
    }
    return [ordered]@{FailedGroup=$State.CurrentGroup;CompletedGroups=$State.Groups.ToArray();RootError=$root;Comparison=$State.Comparison.Value;RecordingError=(Convert-ODVerifierError $State.RecordingError)}
}
function ODS-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Owner scope regression: '+$Label)}}
function ODS-Tree([string]$Path){
    $tokens=$null;$issues=$null
    $tree=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$tokens,[ref]$issues)
    ODS-Need ($issues.Count -eq 0) ('source parses '+$Path)
    return $tree
}
function ODS-Function($Tree,[string]$Name){
    $nodes=@($Tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name},$true))
    ODS-Need ($nodes.Count -eq 1) ('unique actual definition '+$Name)
    return $nodes[0]
}
function ODS-Hashes([string]$Repo){
    $hashes=[ordered]@{}
    foreach($name in @('driver.ps1','record-boundary.tests.ps1','capture.ps1','source-model.ps1','observation-contract.tests.ps1')){
        $relative='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/'+$name
        $hashes[$relative]=(Get-FileHash -LiteralPath (Join-Path $Repo $relative)).Hash
    }
    return $hashes
}
function ODS-Equal($Left,$Right){(ConvertTo-Json -InputObject $Left -Depth 25 -Compress) -ceq (ConvertTo-Json -InputObject $Right -Depth 25 -Compress)}
function ODS-Facts {
    [ordered]@{Schema=1;ParentSHA256='7F09A142E228FF16CD7228287DEBBB5EE60DEC9A91FBF1EF691F148C0A911587';HistoricalStopSHA256='4B06ED585A7F63DC38F084D0718E653F17266E2065175E58CAD67EC5AEF6FD14';ApprovalValid=$true;AttemptConsumed=$false;NewStopPresent=$false}
}
function ODS-Reject([scriptblock]$Body){
    $caught=$null;$values=@()
    try{$values=@(& $Body)}catch{$caught=$_}
    ODS-Need ($null -ne $caught -and $values.Count -eq 0) 'actual rejection without output'
    ODS-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq 'OD: error projection recovery mismatch' -and $caught.FullyQualifiedErrorId -ceq 'OD: error projection recovery mismatch') 'exact recovery rejection'
}
function Invoke-ODErrorProjectionScope {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $names=@('FourFunctionsPersist','DefinitionsExact','ModuleIsolation','AttemptBindings','ConsumedAttempt','HistoricalEpoch','RecoveryGate','NewEpochBindings','OriginalContractsUnchanged')
    $state=[pscustomobject]@{CurrentGroup='Setup';Groups=[Collections.Generic.List[object]]::new();RootError=$null;Comparison=[pscustomobject]@{Value=$null};RecordingError=$null;Stopped=$false}
    $testHash=$null;$inputHash=$null;$before=$null;$after=$null
    try{
    $leaf=Split-Path (Get-Command Invoke-ODErrorProjectionScope -CommandType Function).ScriptBlock.File
    $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'))
    $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $testPath=Join-Path $leaf error-projection-scope.tests.ps1
    $inputPath=Join-Path $packet error-projection-scope.input.json
    $testHash=(Get-FileHash -LiteralPath $testPath).Hash
    $inputHash=(Get-FileHash -LiteralPath $inputPath).Hash
    $binding=Get-Content -Raw -LiteralPath $inputPath|ConvertFrom-Json -AsHashtable
    ODS-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen auxiliary identity'
    foreach($name in $binding.HistoricalFiles.Keys){ODS-Need ((Get-FileHash -LiteralPath (Join-Path $packet $name)).Hash -ceq $binding.HistoricalFiles[$name]) ('historical identity '+$name)}
    ODS-Need ((Get-FileHash -LiteralPath (Join-Path $leaf error-projection.tests.ps1)).Hash -ceq $binding.RegressionTestSHA256 -and (Get-FileHash -LiteralPath (Join-Path $packet error-projection.input.json)).Hash -ceq $binding.RegressionInputSHA256) 'frozen verifier regression identities'
    $before=ODS-Hashes $repo
    ODS-Need (ODS-Equal $before $binding.CandidateSources) 'new current sources frozen'
    foreach($relative in $binding.BaselineSources.Keys){
        ODS-Need ((Get-FileHash -LiteralPath (Join-Path (Join-Path $packet snapshots/writer-scope-rebind) $relative)).Hash -ceq $binding.BaselineSources[$relative]) 'baseline snapshot binding'
    }
    $recordTree=ODS-Tree (Join-Path $leaf record-boundary.tests.ps1)
    $owner=ODS-Function $recordTree Invoke-ODOwnerControls
    $initializers=@($owner.Body.FindAll({param($n)
        if($n -isnot [Management.Automation.Language.CommandAst] -or $n.CommandElements.Count -lt 2){return $false}
        $first=$n.CommandElements[0];$second=$n.CommandElements[1]
        if($first -isnot [Management.Automation.Language.VariableExpressionAst] -or $first.VariablePath.UserPath -cne 'controls' -or $second -isnot [Management.Automation.Language.ScriptBlockExpressionAst]){return $false}
        if($null -eq $second.ScriptBlock.ParamBlock){return $false}
        $names=@($second.ScriptBlock.ParamBlock.Parameters|ForEach-Object {$_.Name.VariablePath.UserPath})
        return ($names -join '|') -ceq 'Definitions|ProfileText|Root|Repo|Packet|Leaf|Publisher|FailurePublisher'
    },$true))
    ODS-Need ($initializers.Count -eq 1) 'unique actual initializer'
    $initializer=$initializers[0].CommandElements[1].ScriptBlock.GetScriptBlock()
    $captureTree=ODS-Tree (Join-Path $leaf capture.ps1)
    $profileTree=ODS-Tree (Join-Path $repo scripts/offline-diagnostics-v4/private/profiles.ps1)
    $definitions=@(foreach($name in @('Capture-Value','Capture-Once','Capture-ErrorSummary')){(ODS-Function $captureTree $name).Extent.Text})
    $profileDefinition=ODS-Function $profileTree Invoke-ODProfile
    $actualDefinitions=@($definitions)+@($profileDefinition.Extent.Text)
    $collector=(ODS-Function $captureTree Capture-Failure).Extent.Text
    $caseTree=ODS-Tree (Join-Path $leaf case-module.ps1)
    . ([scriptblock]::Create((ODS-Function $caseTree New-ODCaseModule).Extent.Text))
    $writer=Join-Path $repo scripts/offline-diagnostics-v4/private/write-boundary.ps1
    $outerBefore=@(Get-Command Invoke-ODProfile,Capture-Value,Capture-Once,Capture-ErrorSummary -CommandType Function -ErrorAction SilentlyContinue)
    $module=New-ODCaseModule $writer $collector
    $scopeRoot=Join-Path $packet scope-only-no-io
    $emitted=@(& $module $initializer $definitions $profileDefinition.Extent.Text $scopeRoot $repo $packet $leaf $null $null)
    ODS-Need ($emitted.Count -eq 0) 'initializer emits nothing'
    $commands=@(& $module {foreach($name in @('Invoke-ODProfile','Capture-Value','Capture-Once','Capture-ErrorSummary')){Get-Command $name -CommandType Function -ErrorAction Stop}})
    $driverTree=ODS-Tree (Join-Path $leaf driver.ps1)
    . ([scriptblock]::Create((ODS-Function $driverTree Get-ODAttemptSlot).Extent.Text))
    . ([scriptblock]::Create((ODS-Function $driverTree Get-ODWriterManifestPath).Extent.Text))
    . ([scriptblock]::Create((ODS-Function $driverTree Assert-ODProjectionFacts).Extent.Text))
        Invoke-ODVerifierCase FourFunctionsPersist {
            ODS-Need ($commands.Count -eq 4) 'four actual functions persist'
            $again=@(& $module {foreach($name in @('Invoke-ODProfile','Capture-Value','Capture-Once','Capture-ErrorSummary')){Get-Command $name -CommandType Function -ErrorAction Stop}})
            ODS-Need ($again.Count -eq 4) 'functions persist after another invocation'
            foreach($command in $again){ODS-Need ([object]::ReferenceEquals($command.Module,$module)) 'actual functions module-bound'}
        } $state
        Invoke-ODVerifierCase DefinitionsExact {
            foreach($text in $actualDefinitions){
                $tokens=$null;$issues=$null;$tree=[Management.Automation.Language.Parser]::ParseInput($text,[ref]$tokens,[ref]$issues)
                $function=@($tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$false))[0]
                $command=@($commands|Where-Object {$_.Name -ceq $function.Name})[0]
                Assert-ODImportedDefinition $function $command.ScriptBlock.Ast $state.Comparison
                if($null -ne $function.Parameters){foreach($parameter in $function.Parameters){ODS-Need ($command.Parameters.ContainsKey($parameter.Name.VariablePath.UserPath)) ('inline parameter retained '+$function.Name)}}
                if($null -ne $function.Body.ParamBlock){foreach($parameter in $function.Body.ParamBlock.Parameters){ODS-Need ($command.Parameters.ContainsKey($parameter.Name.VariablePath.UserPath)) ('body parameter retained '+$function.Name)}}
            }
        } $state
        Invoke-ODVerifierCase ModuleIsolation {
            $second=New-ODCaseModule $writer $collector
            $values=@(& $second $initializer $definitions $profileDefinition.Extent.Text $scopeRoot $repo $packet $leaf $null $null)
            ODS-Need ($values.Count -eq 0 -and $module.ExportedFunctions.Count -eq 0 -and $second.ExportedFunctions.Count -eq 0) 'two modules zero exports'
            $null=& $module {$script:ControlRoot='scope-marker-one'}
            ODS-Need ((& $second {$script:ControlRoot}) -ceq $scopeRoot) 'private state does not cross modules'
            $outerAfter=@(Get-Command Invoke-ODProfile,Capture-Value,Capture-Once,Capture-ErrorSummary -CommandType Function -ErrorAction SilentlyContinue)
            ODS-Need ($outerBefore.Count -eq $outerAfter.Count) 'no outer functions added'
            for($i=0;$i -lt $outerBefore.Count;$i++){ODS-Need ([object]::ReferenceEquals($outerBefore[$i].ScriptBlock,$outerAfter[$i].ScriptBlock)) 'outer function unchanged'}
        } $state
        Invoke-ODVerifierCase AttemptBindings {
            ODS-Need ((Get-ODAttemptSlot writer-green) -ceq 'writer-projection-green') 'writer uses separate fixed slot'
            foreach($grant in @('contract-green','core','lifecycle','continuation')){ODS-Need ((Get-ODAttemptSlot $grant) -ceq $grant) 'later slot unchanged'}
            ODS-Reject {Get-ODAttemptSlot writer-projection-green}
            $calls=@($driverTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Get-ODAttemptSlot'},$true))
            ODS-Need ($calls.Count -ge 6) 'actual driver attempt path consumers mapped'
            ODS-Need ((Get-Content -Raw (Join-Path $leaf capture.ps1)).Contains("'writer-projection-green'")) 'capture writer slot'
        } $state
        Invoke-ODVerifierCase ConsumedAttempt {
            $facts=ODS-Facts;$facts.AttemptConsumed=$true;ODS-Reject {Assert-ODProjectionFacts $facts}
        } $state
        Invoke-ODVerifierCase HistoricalEpoch {
            $old=Get-Content -Raw (Join-Path $packet controller-regression-parameter-green.final.json)|ConvertFrom-Json -AsHashtable
            ODS-Need ($old.PASS -eq 13 -and $old.FAIL -eq 0 -and $old.NotRun -eq 0 -and $old.ExitCode -eq 0) 'old GREEN actual result'
            foreach($relative in $old.SourceSHA256.Keys){ODS-Need ((Get-FileHash -LiteralPath (Join-Path (Join-Path $packet snapshots/controller-rebind) $relative)).Hash -ceq $old.SourceSHA256[$relative]) 'old GREEN historical bytes'}
            $history=ODS-Function $driverTree Driver-ControllerRegression
            ODS-Need ($history.Extent.Text.Contains('snapshots/controller-rebind')) 'historical GREEN no current-source regrading'
        } $state
        Invoke-ODVerifierCase RecoveryGate {
            Assert-ODProjectionFacts (ODS-Facts)
            foreach($key in @('ApprovalValid','ParentSHA256','HistoricalStopSHA256','NewStopPresent')){
                $facts=ODS-Facts
                $facts[$key]=switch($key){ApprovalValid{$false};NewStopPresent{$true};default{'0'*64}}
                ODS-Reject {Assert-ODProjectionFacts $facts}
            }
            $facts=ODS-Facts;$facts.Remove('ApprovalValid');ODS-Reject {Assert-ODProjectionFacts $facts}
            foreach($gate in @('Driver-ProjectionGate','Driver-ControllerGate')){ODS-Need (@($driverTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $gate},$true)).Count -gt 0) ('actual gate retained '+$gate)}
        } $state
        Invoke-ODVerifierCase NewEpochBindings {
            ODS-Need ((Get-ODWriterManifestPath) -ceq (Join-Path $packet writer-projection-repair.frozen.json)) 'new fixed manifest selector'
            $record=Get-Content -Raw (Join-Path $leaf record-boundary.tests.ps1)
            $model=Get-Content -Raw (Join-Path $leaf source-model.ps1)
            foreach($source in @($record,$model)){ODS-Need ($source.Contains('writer-projection-repair.frozen.json') -and $source.Contains('unit/writer-projection-green/owner-controls') -and -not $source.Contains('writer-controller-repair.frozen.json')) 'unit/model epoch agrees'}
            $contract=Get-Content -Raw (Join-Path $leaf observation-contract.tests.ps1)
            ODS-Need ($contract.Contains('writer-projection-green.evidence.json') -and $contract.Contains('writer-projection-green.final.json') -and -not $contract.Contains("'writer-green.evidence.json'")) 'contract uses new result only'
            ODS-Need ((ODS-Function $driverTree Driver-RequireFinal).Extent.Text.Contains('Get-ODAttemptSlot')) 'prior final uses fixed new slot'
        } $state
        Invoke-ODVerifierCase OriginalContractsUnchanged {
            $parent=Get-Content -Raw (Join-Path $packet writer-scope-repair.frozen.json)|ConvertFrom-Json -AsHashtable
            ODS-Need ($parent.SourcePaths.Count -eq 31) 'original inventory31'
            foreach($relative in $parent.SourcePaths){if($binding.BaselineSources.Contains($relative)){continue};$snapshot=Join-Path (Join-Path $packet snapshots/writer-scope-rebind) $relative;ODS-Need ((Get-FileHash -LiteralPath (Join-Path $repo $relative)).Hash -ceq (Get-FileHash -LiteralPath $snapshot).Hash) ('outside five-source window unchanged '+$relative)}
            foreach($callee in @('Driver-Verify','Driver-FalseFlags','Driver-Recorder','Driver-UnitEvidence','New-ODRuntimeContext','Assert-ODProfileReport','Assert-ODProfileObservations')){ODS-Need (@($driverTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $callee},$true)).Count -gt 0) ('original actual gate retained '+$callee)}
            ODS-Need ((Get-Content -Raw (Join-Path $leaf capture.ps1)).Contains("'writer-green'")) 'canonical writer group unchanged'
        } $state

        if(-not $state.Stopped){
            $state.CurrentGroup='FinalIdentity'
            $after=ODS-Hashes $repo
            ODS-Need (ODS-Equal $before $after) 'phase source hashes unchanged'
            ODS-Need ((Get-FileHash -LiteralPath $testPath).Hash -ceq $testHash -and (Get-FileHash -LiteralPath $inputPath).Hash -ceq $inputHash) 'frozen test/input unchanged'
            ODS-Need ($state.Groups.Count -eq 9 -and (@($state.Groups|ForEach-Object Name) -join '|') -ceq ($names -join '|')) 'nine ordered groups'
        }
    }catch{
        if($null -eq $state.RootError){$state.RootError=$_}
        $state.Stopped=$true
    }
    if($state.Stopped){
        try{$after=ODS-Hashes $repo}catch{if($null -eq $state.RecordingError){$state.RecordingError=$_}}
        $diagnostic=Convert-ODVerifierDiagnostic $state
        return [pscustomobject][ordered]@{Schema=1;Phase=$Phase;Unexpected=$true;ExitCode=2;FailedGroup=$diagnostic.FailedGroup;CompletedGroups=$diagnostic.CompletedGroups;RootError=$diagnostic.RootError;Comparison=$diagnostic.Comparison;RecordingError=$diagnostic.RecordingError;TestSHA256=$testHash;InputSHA256=$inputHash;SourceSHA256=$before;PostSourceSHA256=$after}
    }
    return [pscustomobject][ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;SourceSHA256=$before;Groups=$state.Groups.ToArray();PASS=9;FAIL=0;NotRun=0;ExitCode=0;RootError=$null}
}
