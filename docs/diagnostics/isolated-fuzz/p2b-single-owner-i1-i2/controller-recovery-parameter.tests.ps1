# Auxiliary finite regression; definitions only. Never loaded by capture/profile.
function ODR-Need([bool]$Condition,[string]$Message){if(-not $Condition){throw ('Controller regression: '+$Message)}}
function ODR-AST([string]$Path){
    $tokens=$null;$issues=$null
    $tree=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$tokens,[ref]$issues)
    ODR-Need ($issues.Count -eq 0) 'source parses'
    return $tree
}
function ODR-Definition($Tree,[string]$Name){
    $nodes=@($Tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name},$true))
    ODR-Need ($nodes.Count -eq 1) ('unique actual definition '+$Name)
    return $nodes[0]
}
function ODR-SourceHashes([string]$Repo){
    $result=[ordered]@{}
    foreach($relative in @('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/driver.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/record-boundary.tests.ps1','docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/source-model.ps1')){
        $result[$relative]=(Get-FileHash -LiteralPath (Join-Path $Repo $relative) -Algorithm SHA256).Hash
    }
    return $result
}
function ODR-Flags {
    return [ordered]@{SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false}
}
function ODR-Facts {
    return [ordered]@{Schema=1;HistoricalWriterSHA256='9AC3EAD775DB2BE7D21DE3B1CD3FE46342B1B176BA6B07C03F082D0F63B0C548';HistoricalStopSHA256='F19687B1628DC1C48CC771A07DA4308D48E89B8D51921DDACDB6C4CBFAFDE9D2';ApprovalValid=$true;ForbiddenOutputsPresent=$false;NewStopPresent=$false}
}
function ODR-Reject([scriptblock]$Body,[string]$Message){
    $caught=$null;$emitted=@()
    try{$emitted=@(& $Body)}catch{$caught=$_}
    ODR-Need ($null -ne $caught -and $emitted.Count -eq 0) 'negative actually rejects without output'
    ODR-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq $Message -and $caught.FullyQualifiedErrorId -ceq $Message) 'exact rejection signature'
}
function ODR-Case([string]$Name,[scriptblock]$Body,$Groups){
    $emitted=@(& $Body)
    ODR-Need ($emitted.Count -eq 0) ('no extra output '+$Name)
    $Groups.Add([ordered]@{Name=$Name;Status='PASS'})
}
function Invoke-ODControllerRecoveryRegression {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Green')][string]$Phase)
    Set-StrictMode -Version Latest
    $ErrorActionPreference='Stop'
    $leaf=Split-Path (Get-Command Invoke-ODControllerRecoveryRegression -CommandType Function).ScriptBlock.File
    $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'))
    $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $testPath=Join-Path $leaf controller-recovery-parameter.tests.ps1
    $testHash=(Get-FileHash -LiteralPath $testPath -Algorithm SHA256).Hash
    $inputPath=Join-Path $packet controller-regression-parameter.input.json
    $inputHash=(Get-FileHash -LiteralPath $inputPath -Algorithm SHA256).Hash
    $binding=Get-Content -LiteralPath $inputPath -Raw|ConvertFrom-Json -AsHashtable
    ODR-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen regression identity'
    foreach($name in @('writer-green.frozen.json','stop.json')){
        $expected=if($name -ceq 'stop.json'){$binding.HistoricalStopSHA256}else{$binding.HistoricalWriterSHA256}
        ODR-Need ((Get-FileHash -LiteralPath (Join-Path $packet $name) -Algorithm SHA256).Hash -ceq $expected) 'original input intact'
    }
    foreach($relative in $binding.BaselineSources.Keys){
        $snapshot=Join-Path (Join-Path $packet snapshots/writer-green) $relative
        ODR-Need ((Get-FileHash -LiteralPath $snapshot -Algorithm SHA256).Hash -ceq $binding.BaselineSources[$relative]) 'baseline snapshot intact'
    }
    $before=ODR-SourceHashes $repo
    if($Phase -ceq 'Red'){ODR-Need ((ConvertTo-Json $before -Compress) -ceq (ConvertTo-Json $binding.BaselineSources -Compress)) 'RED uses exact original sources'}
    $tree=ODR-AST (Join-Path $leaf driver.ps1)
    $needText=(ODR-Definition $tree Driver-Need).Extent.Text
    $flagsText=(ODR-Definition $tree Driver-FalseFlags).Extent.Text
    $assignments=@($tree.FindAll({param($n)
        if($n -isnot [Management.Automation.Language.AssignmentStatementAst] -or $n.Left -isnot [Management.Automation.Language.VariableExpressionAst] -or $n.Left.VariablePath.UserPath -cne 'manifest'){return $false}
        $commands=@($n.Right.FindAll({param($c)$c -is [Management.Automation.Language.CommandAst]},$true))
        $decode=@($commands|Where-Object {$_.GetCommandName() -ceq 'ConvertFrom-Json'})
        $read=@($commands|Where-Object {$_.GetCommandName() -ceq 'Get-Content' -and @($_.CommandElements|Where-Object {$_ -is [Management.Automation.Language.VariableExpressionAst] -and $_.VariablePath.UserPath -ceq 'manifestPath'}).Count -eq 1})
        return $decode.Count -eq 1 -and $read.Count -eq 1
    },$true))
    ODR-Need ($assignments.Count -eq 1) 'one actual manifest decoder AST'
    $rootError=$null
    try{
        $root=& {
            param($NeedText,$FlagsText,$DecoderText,$ManifestFile)
            . ([scriptblock]::Create($NeedText));. ([scriptblock]::Create($FlagsText))
            $manifestPath=$ManifestFile
            . ([scriptblock]::Create($DecoderText))
            Driver-FalseFlags $manifest
            ODR-Need ($manifest -is [Collections.IDictionary]) 'production decoder yields IDictionary'
            return $manifest
        } $needText $flagsText $assignments[0].Extent.Text (Join-Path $packet writer-green.frozen.json)
    }catch{
        if($Phase -cne 'Red'){throw}
        $saved=Get-Content -LiteralPath (Join-Path $packet stop.json) -Raw|ConvertFrom-Json -AsHashtable
        ODR-Need ($_.FullyQualifiedErrorId -ceq 'CannotIndex' -and $_.Exception.Message -ceq $saved.Message) 'RED actual CannotIndex only'
        $rootError=[ordered]@{ErrorId=$_.FullyQualifiedErrorId;Message=$_.Exception.Message;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}
    }
    $names=@('ProductionManifestDictionary','FiveFalseFlags','TrueFlagRejected','StringFlagRejected','MissingFlagRejected','NullFlagRejected','RecoveryExact','MissingApproval','ChangedStop','WrongParent','ExistingAttempt','NewStop','WriterBindings')
    $groups=[Collections.Generic.List[object]]::new()
    if($Phase -ceq 'Red'){
        ODR-Need ($null -ne $rootError) 'root regression must fail before repair'
        foreach($name in $names){$groups.Add([ordered]@{Name=$name;Status=$(if($name -ceq $names[0]){'FAIL'}else{'NotRun'})})}
    }else{
        ODR-Need ($null -eq $rootError) 'GREEN has no root error'
        $model=ODR-AST (Join-Path $leaf source-model.ps1)
        $context=ODR-Definition $model New-ODRuntimeContext
        $parameter=@($context.Parameters|Where-Object {$_.Name.VariablePath.UserPath -ceq 'Manifest'})
        ODR-Need ($parameter.Count -eq 1 -and @($parameter[0].Attributes|Where-Object {$_ -is [Management.Automation.Language.TypeConstraintAst] -and $_.TypeName.FullName -ceq 'Collections.IDictionary'}).Count -eq 1) 'actual context typed boundary retained'
        $groups.Add([ordered]@{Name=$names[0];Status='PASS'})
        . ([scriptblock]::Create($needText));. ([scriptblock]::Create($flagsText))
        . ([scriptblock]::Create((ODR-Definition $tree Assert-ODControllerRecoveryFacts).Extent.Text))
        ODR-Case FiveFalseFlags {Driver-FalseFlags (ODR-Flags)} $groups
        ODR-Case TrueFlagRejected {$v=ODR-Flags;$v.SourceVerified=$true;ODR-Reject {Driver-FalseFlags $v} 'False capability flag required SourceVerified'} $groups
        ODR-Case StringFlagRejected {$v=ODR-Flags;$v.SourceVerified='false';ODR-Reject {Driver-FalseFlags $v} 'False capability flag required SourceVerified'} $groups
        ODR-Case MissingFlagRejected {$v=ODR-Flags;$v.Remove('SourceVerified');ODR-Reject {Driver-FalseFlags $v} 'False capability flag required SourceVerified'} $groups
        ODR-Case NullFlagRejected {$v=ODR-Flags;$v.SourceVerified=$null;ODR-Reject {Driver-FalseFlags $v} 'False capability flag required SourceVerified'} $groups
        ODR-Case RecoveryExact {Assert-ODControllerRecoveryFacts (ODR-Facts)} $groups
        ODR-Case MissingApproval {$v=ODR-Facts;$v.ApprovalValid=$false;ODR-Reject {Assert-ODControllerRecoveryFacts $v} 'OD: controller recovery mismatch'} $groups
        ODR-Case ChangedStop {$v=ODR-Facts;$v.HistoricalStopSHA256=('0'*64);ODR-Reject {Assert-ODControllerRecoveryFacts $v} 'OD: controller recovery mismatch'} $groups
        ODR-Case WrongParent {$v=ODR-Facts;$v.HistoricalWriterSHA256=('0'*64);ODR-Reject {Assert-ODControllerRecoveryFacts $v} 'OD: controller recovery mismatch'} $groups
        ODR-Case ExistingAttempt {$v=ODR-Facts;$v.ForbiddenOutputsPresent=$true;ODR-Reject {Assert-ODControllerRecoveryFacts $v} 'OD: controller recovery mismatch'} $groups
        ODR-Case NewStop {$v=ODR-Facts;$v.NewStopPresent=$true;ODR-Reject {Assert-ODControllerRecoveryFacts $v} 'OD: controller recovery mismatch'} $groups
        ODR-Case WriterBindings {
            $writerName='writer-controller-repair.frozen.json'
            $record=ODR-AST (Join-Path $leaf record-boundary.tests.ps1)
            $intent=ODR-Definition $model New-ODIntentUnitContext
            foreach($scope in @($record,$intent.Body)){
                $old=@($scope.FindAll({param($n)$n -is [Management.Automation.Language.StringConstantExpressionAst] -and $n.Value -ceq 'writer-green.frozen.json'},$true))
                $new=@($scope.FindAll({param($n)$n -is [Management.Automation.Language.StringConstantExpressionAst] -and $n.Value -ceq 'writer-controller-repair.frozen.json'},$true))
                ODR-Need ($old.Count -eq 0 -and $new.Count -eq 1) 'both actual consumers use only new writer'
            }
            $selector=ODR-Definition $tree Get-ODWriterManifestPath
            ODR-Need (@($selector.Body.FindAll({param($n)$n -is [Management.Automation.Language.StringConstantExpressionAst] -and $n.Value -ceq 'writer-controller-repair.frozen.json'},$true)).Count -eq 1) 'fixed new selector'
            $calls=@($tree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Get-ODWriterManifestPath'},$true))
            ODR-Need ($calls.Count -eq 4) 'creation and three actual selector callers'
            $prior=ODR-Definition $tree Driver-RequireFinal
            ODR-Need (@($prior.Body.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'Get-ODWriterManifestPath'},$true)).Count -eq 1) 'prior final uses selector'
            $pathAssignments=@($tree.FindAll({param($n)$n -is [Management.Automation.Language.AssignmentStatementAst] -and $n.Left -is [Management.Automation.Language.VariableExpressionAst] -and $n.Left.VariablePath.UserPath -cin @('manifestPath','previousPath') -and @($n.Right.FindAll({param($c)$c -is [Management.Automation.Language.CommandAst] -and $c.GetCommandName() -ceq 'Get-ODWriterManifestPath'},$true)).Count -eq 1},$true))
            ODR-Need ($pathAssignments.Count -eq 3) 'prepare creation/parent and actual RunOne binding'
            foreach($callee in @('Driver-Verify','Driver-FalseFlags','Driver-Recorder','Driver-UnitEvidence','New-ODRuntimeContext','Assert-ODProfileReport','Assert-ODProfileObservations')){
                ODR-Need (@($tree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $callee},$true)).Count -gt 0) ('actual gate retained '+$callee)
            }
            ODR-Need (@($tree.FindAll({param($n)$n -is [Management.Automation.Language.StringConstantExpressionAst] -and $n.Value -ceq 'repair-green.frozen.json'},$true)).Count -gt 0) 'repair path unchanged'
        } $groups
    }
    $after=ODR-SourceHashes $repo
    ODR-Need ((ConvertTo-Json $before -Compress) -ceq (ConvertTo-Json $after -Compress) -and (Get-FileHash -LiteralPath $testPath).Hash -ceq $testHash -and (Get-FileHash -LiteralPath $inputPath).Hash -ceq $inputHash) 'phase sources/input remain unchanged'
    ODR-Need ($groups.Count -eq 13 -and (@($groups|ForEach-Object Name) -join '|') -ceq ($names -join '|')) 'complete ordered13groups'
    [pscustomobject][ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;SourceSHA256=$before;PASS=$(if($Phase -ceq 'Red'){0}else{13});FAIL=$(if($Phase -ceq 'Red'){1}else{0});NotRun=$(if($Phase -ceq 'Red'){12}else{0});ExitCode=$(if($Phase -ceq 'Red'){1}else{0});Groups=$groups.ToArray();RootError=$rootError}
}
