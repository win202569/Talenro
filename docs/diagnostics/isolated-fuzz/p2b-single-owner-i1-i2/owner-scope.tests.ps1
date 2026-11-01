# Finite auxiliary regression. Definitions only; never imported by runtime.
# Breaks caught: lost module-local imports, altered inline parameters/body,
# leaked exports/state, mixed writer attempts/epochs and permissive stop gates.
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
function ODS-Case([string]$Name,[scriptblock]$Body,$Groups){
    $values=@(& $Body)
    ODS-Need ($values.Count -eq 0) ('no stray output '+$Name)
    $Groups.Add([ordered]@{Name=$Name;Status='PASS'})
}
function ODS-Facts {
    [ordered]@{Schema=1;ParentSHA256='D8D31B8BE9F38A35F12CF5951DC28B3723D759B96AF3BC3B3D8B3D4D95FAC57C';HistoricalStopSHA256='ECC2C74E5544CA0EC9A7D2C5A4377DA35F1A352F26CFEADBFEF35124C2A41607';ApprovalValid=$true;AttemptConsumed=$false;NewStopPresent=$false}
}
function ODS-Reject([scriptblock]$Body){
    $caught=$null;$values=@()
    try{$values=@(& $Body)}catch{$caught=$_}
    ODS-Need ($null -ne $caught -and $values.Count -eq 0) 'actual rejection without output'
    ODS-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq 'OD: owner scope recovery mismatch' -and $caught.FullyQualifiedErrorId -ceq 'OD: owner scope recovery mismatch') 'exact recovery rejection'
}
function Invoke-ODOwnerScopeRegression {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest
    $ErrorActionPreference='Stop'
    $leaf=Split-Path (Get-Command Invoke-ODOwnerScopeRegression -CommandType Function).ScriptBlock.File
    $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'))
    $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $testPath=Join-Path $leaf owner-scope.tests.ps1
    $inputPath=Join-Path $packet owner-scope.input.json
    $testHash=(Get-FileHash -LiteralPath $testPath).Hash
    $inputHash=(Get-FileHash -LiteralPath $inputPath).Hash
    $binding=Get-Content -Raw -LiteralPath $inputPath|ConvertFrom-Json -AsHashtable
    ODS-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen auxiliary identity'
    foreach($name in $binding.HistoricalFiles.Keys){ODS-Need ((Get-FileHash -LiteralPath (Join-Path $packet $name)).Hash -ceq $binding.HistoricalFiles[$name]) ('historical identity '+$name)}
    $before=ODS-Hashes $repo
    foreach($relative in $binding.BaselineSources.Keys){
        ODS-Need ((Get-FileHash -LiteralPath (Join-Path (Join-Path $packet snapshots/controller-rebind) $relative)).Hash -ceq $binding.BaselineSources[$relative]) 'baseline snapshot binding'
    }
    if($Phase -ceq 'Red'){ODS-Need (ODS-Equal $before $binding.BaselineSources) 'RED exact baseline'}
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
    $names=@('FourFunctionsPersist','DefinitionsExact','ModuleIsolation','AttemptBindings','ConsumedAttempt','HistoricalEpoch','RecoveryGate','NewEpochBindings','OriginalContractsUnchanged')
    $groups=[Collections.Generic.List[object]]::new()
    $outerBefore=@(Get-Command Invoke-ODProfile,Capture-Value,Capture-Once,Capture-ErrorSummary -CommandType Function -ErrorAction SilentlyContinue)
    $module=New-ODCaseModule $writer $collector
    $scopeRoot=Join-Path $packet scope-only-no-io
    $emitted=@(& $module $initializer $definitions $profileDefinition.Extent.Text $scopeRoot $repo $packet $leaf $null $null)
    ODS-Need ($emitted.Count -eq 0) 'initializer emits nothing'
    $rootError=$null
    try{
        $commands=@(& $module {
            foreach($name in @('Invoke-ODProfile','Capture-Value','Capture-Once','Capture-ErrorSummary')){Get-Command $name -CommandType Function -ErrorAction Stop}
        })
    }catch{
        if($Phase -cne 'Red'){throw}
        ODS-Need ($_.Exception.GetType().FullName -ceq 'System.Management.Automation.CommandNotFoundException' -and $_.FullyQualifiedErrorId -ceq 'CommandNotFoundException' -and $_.Exception.Message.Contains('Invoke-ODProfile')) 'only intended absent actual profile'
        $rootError=[ordered]@{Type=$_.Exception.GetType().FullName;ErrorId=$_.FullyQualifiedErrorId;Message=$_.Exception.Message;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}
    }
    if($Phase -ceq 'Red'){
        ODS-Need ($null -ne $rootError) 'RED must expose scope defect'
        foreach($name in $names){$groups.Add([ordered]@{Name=$name;Status=$(if($name -ceq 'FourFunctionsPersist'){'FAIL'}else{'NotRun'})})}
    }else{
        ODS-Case FourFunctionsPersist {
            ODS-Need ($commands.Count -eq 4) 'four actual functions persist'
            $again=@(& $module {foreach($name in @('Invoke-ODProfile','Capture-Value','Capture-Once','Capture-ErrorSummary')){Get-Command $name -CommandType Function -ErrorAction Stop}})
            ODS-Need ($again.Count -eq 4) 'functions persist after another invocation'
            foreach($command in $again){ODS-Need ([object]::ReferenceEquals($command.Module,$module)) 'actual functions module-bound'}
        } $groups
        ODS-Case DefinitionsExact {
            foreach($text in $actualDefinitions){
                $tokens=$null;$issues=$null;$tree=[Management.Automation.Language.Parser]::ParseInput($text,[ref]$tokens,[ref]$issues)
                $function=@($tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$false))[0]
                $command=@($commands|Where-Object {$_.Name -ceq $function.Name})[0]
                $body=$function.Body.Extent.Text.Trim().TrimStart('{').TrimEnd('}').Trim()
                $loaded=$command.ScriptBlock.Ast.Extent.Text.Trim().TrimStart('{').TrimEnd('}').Trim()
                ODS-Need ($loaded -ceq $body) ('actual body unchanged '+$function.Name)
                foreach($parameter in @($function.Parameters)){ODS-Need ($command.Parameters.ContainsKey($parameter.Name.VariablePath.UserPath)) ('inline parameter retained '+$function.Name)}
                if($null -ne $function.Body.ParamBlock){foreach($parameter in $function.Body.ParamBlock.Parameters){ODS-Need ($command.Parameters.ContainsKey($parameter.Name.VariablePath.UserPath)) ('body parameter retained '+$function.Name)}}
            }
        } $groups
        ODS-Case ModuleIsolation {
            $second=New-ODCaseModule $writer $collector
            $values=@(& $second $initializer $definitions $profileDefinition.Extent.Text $scopeRoot $repo $packet $leaf $null $null)
            ODS-Need ($values.Count -eq 0 -and $module.ExportedFunctions.Count -eq 0 -and $second.ExportedFunctions.Count -eq 0) 'two modules zero exports'
            $null=& $module {$script:ControlRoot='scope-marker-one'}
            ODS-Need ((& $second {$script:ControlRoot}) -ceq $scopeRoot) 'private state does not cross modules'
            $outerAfter=@(Get-Command Invoke-ODProfile,Capture-Value,Capture-Once,Capture-ErrorSummary -CommandType Function -ErrorAction SilentlyContinue)
            ODS-Need ($outerBefore.Count -eq $outerAfter.Count) 'no outer functions added'
            for($i=0;$i -lt $outerBefore.Count;$i++){ODS-Need ([object]::ReferenceEquals($outerBefore[$i].ScriptBlock,$outerAfter[$i].ScriptBlock)) 'outer function unchanged'}
        } $groups
        $driverTree=ODS-Tree (Join-Path $leaf driver.ps1)
        . ([scriptblock]::Create((ODS-Function $driverTree Get-ODAttemptSlot).Extent.Text))
        . ([scriptblock]::Create((ODS-Function $driverTree Get-ODWriterManifestPath).Extent.Text))
        . ([scriptblock]::Create((ODS-Function $driverTree Assert-ODOwnerScopeFacts).Extent.Text))
        ODS-Case AttemptBindings {
            ODS-Need ((Get-ODAttemptSlot writer-green) -ceq 'writer-scope-green') 'writer uses separate fixed slot'
            foreach($grant in @('contract-green','core','lifecycle','continuation')){ODS-Need ((Get-ODAttemptSlot $grant) -ceq $grant) 'later slot unchanged'}
            ODS-Reject {Get-ODAttemptSlot writer-scope-green}
            $calls=@($driverTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Get-ODAttemptSlot'},$true))
            ODS-Need ($calls.Count -ge 6) 'actual driver attempt path consumers mapped'
            ODS-Need ((Get-Content -Raw (Join-Path $leaf capture.ps1)).Contains("'writer-scope-green'")) 'capture writer slot'
        } $groups
        ODS-Case ConsumedAttempt {
            $facts=ODS-Facts;$facts.AttemptConsumed=$true;ODS-Reject {Assert-ODOwnerScopeFacts $facts}
        } $groups
        ODS-Case HistoricalEpoch {
            $old=Get-Content -Raw (Join-Path $packet controller-regression-parameter-green.final.json)|ConvertFrom-Json -AsHashtable
            ODS-Need ($old.PASS -eq 13 -and $old.FAIL -eq 0 -and $old.NotRun -eq 0 -and $old.ExitCode -eq 0) 'old GREEN actual result'
            foreach($relative in $old.SourceSHA256.Keys){ODS-Need ((Get-FileHash -LiteralPath (Join-Path (Join-Path $packet snapshots/controller-rebind) $relative)).Hash -ceq $old.SourceSHA256[$relative]) 'old GREEN historical bytes'}
            $history=ODS-Function $driverTree Driver-ControllerRegression
            ODS-Need ($history.Extent.Text.Contains('snapshots/controller-rebind')) 'historical GREEN no current-source regrading'
        } $groups
        ODS-Case RecoveryGate {
            Assert-ODOwnerScopeFacts (ODS-Facts)
            foreach($key in @('ApprovalValid','ParentSHA256','HistoricalStopSHA256','NewStopPresent')){
                $facts=ODS-Facts
                $facts[$key]=switch($key){ApprovalValid{$false};NewStopPresent{$true};default{'0'*64}}
                ODS-Reject {Assert-ODOwnerScopeFacts $facts}
            }
            $facts=ODS-Facts;$facts.Remove('ApprovalValid');ODS-Reject {Assert-ODOwnerScopeFacts $facts}
            foreach($gate in @('Driver-OwnerScopeGate','Driver-ControllerGate')){ODS-Need (@($driverTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $gate},$true)).Count -gt 0) ('actual gate retained '+$gate)}
        } $groups
        ODS-Case NewEpochBindings {
            ODS-Need ((Get-ODWriterManifestPath) -ceq (Join-Path $packet writer-scope-repair.frozen.json)) 'new fixed manifest selector'
            $record=Get-Content -Raw (Join-Path $leaf record-boundary.tests.ps1)
            $model=Get-Content -Raw (Join-Path $leaf source-model.ps1)
            foreach($source in @($record,$model)){ODS-Need ($source.Contains('writer-scope-repair.frozen.json') -and $source.Contains('unit/writer-scope-green/owner-controls') -and -not $source.Contains('writer-controller-repair.frozen.json')) 'unit/model epoch agrees'}
            $contract=Get-Content -Raw (Join-Path $leaf observation-contract.tests.ps1)
            ODS-Need ($contract.Contains('writer-scope-green.evidence.json') -and $contract.Contains('writer-scope-green.final.json') -and -not $contract.Contains("'writer-green.evidence.json'")) 'contract uses new result only'
            ODS-Need ((ODS-Function $driverTree Driver-RequireFinal).Extent.Text.Contains('Get-ODAttemptSlot')) 'prior final uses fixed new slot'
        } $groups
        ODS-Case OriginalContractsUnchanged {
            $parent=Get-Content -Raw (Join-Path $packet writer-controller-repair.frozen.json)|ConvertFrom-Json -AsHashtable
            ODS-Need ($parent.SourcePaths.Count -eq 31) 'original inventory31'
            foreach($relative in $parent.SourcePaths){if($binding.BaselineSources.Contains($relative)){continue};$snapshot=Join-Path (Join-Path $packet snapshots/controller-rebind) $relative;ODS-Need ((Get-FileHash -LiteralPath (Join-Path $repo $relative)).Hash -ceq (Get-FileHash -LiteralPath $snapshot).Hash) ('outside five-source window unchanged '+$relative)}
            foreach($callee in @('Driver-Verify','Driver-FalseFlags','Driver-Recorder','Driver-UnitEvidence','New-ODRuntimeContext','Assert-ODProfileReport','Assert-ODProfileObservations')){ODS-Need (@($driverTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $callee},$true)).Count -gt 0) ('original actual gate retained '+$callee)}
            ODS-Need ((Get-Content -Raw (Join-Path $leaf capture.ps1)).Contains("'writer-green'")) 'canonical writer group unchanged'
        } $groups
    }
    ODS-Need (ODS-Equal $before (ODS-Hashes $repo)) 'phase source hashes unchanged'
    ODS-Need ((Get-FileHash -LiteralPath $testPath).Hash -ceq $testHash -and (Get-FileHash -LiteralPath $inputPath).Hash -ceq $inputHash) 'frozen test/input unchanged'
    ODS-Need ($groups.Count -eq 9 -and (@($groups|ForEach-Object Name) -join '|') -ceq ($names -join '|')) 'nine ordered groups'
    [pscustomobject][ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;SourceSHA256=$before;Groups=$groups.ToArray();PASS=$(if($Phase -ceq 'Red'){0}else{9});FAIL=$(if($Phase -ceq 'Red'){1}else{0});NotRun=$(if($Phase -ceq 'Red'){8}else{0});ExitCode=$(if($Phase -ceq 'Red'){1}else{0});RootError=$rootError}
}
