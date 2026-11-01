# Definitions-only auxiliary; eight frozen verifier boundary checks.
function V-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Verifier regression: '+$Label)}}
function V-Tree([string]$Path){
    $t=$null;$e=$null;$tree=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$t,[ref]$e)
    V-Need ($e.Count -eq 0) ('parse '+$Path);return $tree
}
function V-Definition($Tree,[string]$Name){
    $items=@($Tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name},$true))
    V-Need ($items.Count -eq 1) ('unique '+$Name);return $items[0]
}
function V-Function([string]$Text,[string]$Name){
    $t=$null;$e=$null;$tree=[Management.Automation.Language.Parser]::ParseInput($Text,[ref]$t,[ref]$e)
    V-Need ($e.Count -eq 0) 'literal AST parses';return V-Definition $tree $Name
}
function V-Error($Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;ErrorId=$Record.FullyQualifiedErrorId;Message=$Record.Exception.Message;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function V-Fixtures([string]$Repo,$Expected){
    $actual=[ordered]@{}
    V-Need ($Expected.Count -eq 6) 'six explicit fixture sources'
    foreach($path in $Expected.Keys){$actual[$path]=(Get-FileHash -LiteralPath (Join-Path $Repo $path)).Hash;V-Need ($actual[$path] -ceq $Expected[$path]) ('fixture unchanged '+$path)}
    return $actual
}
function V-Reject([scriptblock]$Body){
    $caught=$null;$values=@()
    try{$values=@(& $Body)}catch{$caught=$_}
    V-Need ($null -ne $caught -and $values.Count -eq 0) 'negative actually rejects'
    V-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq 'OD: imported definition mismatch' -and $caught.FullyQualifiedErrorId -ceq 'OD: imported definition mismatch') 'exact definition rejection'
}
function V-Case([string]$Name,[scriptblock]$Body,$Groups){
    $values=@(& $Body);V-Need ($values.Count -eq 0) ('no output '+$Name)
    $Groups.Add([ordered]@{Name=$Name;Status='PASS'})
}
function Invoke-ODOwnerVerifierRegression {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $leaf=Split-Path (Get-Command Invoke-ODOwnerVerifierRegression -CommandType Function).ScriptBlock.File
    $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'))
    $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
    $testPath=Join-Path $leaf owner-scope-verifier-regression.tests.ps1
    $inputPath=Join-Path $packet owner-scope-verifier-regression.input.json
    $testHash=(Get-FileHash -LiteralPath $testPath).Hash;$inputHash=(Get-FileHash -LiteralPath $inputPath).Hash
    $names=@('SameDefinitionsAccepted','ChangedBodyRejected','ChangedInlineParameterRejected','UnexpectedAstKindRejected','UnexpectedNameRejected','FailurePreservesOriginal','CompletedGroupsPreserved','LaterCasesNotExecuted')
    $groups=[Collections.Generic.List[object]]::new();$rootError=$null;$comparison=[pscustomobject]@{Value=$null};$active='Setup';$codeHash=$null;$fixtureHashes=$null
    try{
        $binding=Get-Content -Raw -LiteralPath $inputPath|ConvertFrom-Json -AsHashtable
        V-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen regression'
        foreach($file in $binding.HistoricalFiles.Keys){V-Need ((Get-FileHash -LiteralPath (Join-Path $packet $file)).Hash -ceq $binding.HistoricalFiles[$file]) ('history '+$file)}
        $fixtureHashes=V-Fixtures $repo $binding.FixtureSHA256
        $oldPath=Join-Path $leaf owner-scope-signature.tests.ps1
        V-Need ((Get-FileHash -LiteralPath $oldPath).Hash -ceq $binding.OldTestSHA256) 'old frozen checker'
        $oldTree=V-Tree $oldPath
        $record=V-Tree (Join-Path $leaf record-boundary.tests.ps1)
        $owner=V-Definition $record Invoke-ODOwnerControls
        $initializers=@($owner.Body.FindAll({param($n)
            if($n -isnot [Management.Automation.Language.CommandAst] -or $n.CommandElements.Count -lt 2){return $false}
            $a=$n.CommandElements[0];$b=$n.CommandElements[1]
            if($a -isnot [Management.Automation.Language.VariableExpressionAst] -or $a.VariablePath.UserPath -cne 'controls' -or $b -isnot [Management.Automation.Language.ScriptBlockExpressionAst] -or $null -eq $b.ScriptBlock.ParamBlock){return $false}
            return (@($b.ScriptBlock.ParamBlock.Parameters|ForEach-Object {$_.Name.VariablePath.UserPath}) -join '|') -ceq 'Definitions|ProfileText|Root|Repo|Packet|Leaf|Publisher|FailurePublisher'
        },$true))
        V-Need ($initializers.Count -eq 1) 'actual initializer unique'
        $capture=V-Tree (Join-Path $leaf capture.ps1)
        $profile=V-Tree (Join-Path $repo scripts/offline-diagnostics-v4/private/profiles.ps1)
        $definitions=@(foreach($n in @('Capture-Value','Capture-Once','Capture-ErrorSummary')){(V-Definition $capture $n).Extent.Text})
        $profileText=(V-Definition $profile Invoke-ODProfile).Extent.Text
        $actualDefinitions=@($definitions)+@($profileText)
        . ([scriptblock]::Create((V-Definition (V-Tree (Join-Path $leaf case-module.ps1)) New-ODCaseModule).Extent.Text))
        $module=New-ODCaseModule (Join-Path $repo scripts/offline-diagnostics-v4/private/write-boundary.ps1) (V-Definition $capture Capture-Failure).Extent.Text
        $initialValues=@(& $module $initializers[0].CommandElements[1].ScriptBlock.GetScriptBlock() $definitions $profileText (Join-Path $packet verifier-no-io) $repo $packet $leaf $null $null)
        V-Need ($initialValues.Count -eq 0) 'initializer no output'
        $commands=@(& $module {foreach($n in @('Invoke-ODProfile','Capture-Value','Capture-Once','Capture-ErrorSummary')){Get-Command $n -CommandType Function -ErrorAction Stop}})
        V-Need ($commands.Count -eq 4) 'four actual loaded definitions'
        $active=$names[0]
        if($Phase -ceq 'Red'){
            $codeHash=$binding.OldTestSHA256
            . ([scriptblock]::Create((V-Definition $oldTree ODS-Need).Extent.Text))
            $cases=@($oldTree.FindAll({param($n)$n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq 'ODS-Case' -and $n.CommandElements.Count -ge 3 -and $n.CommandElements[1] -is [Management.Automation.Language.StringConstantExpressionAst] -and $n.CommandElements[1].Value -ceq 'DefinitionsExact'},$true))
            V-Need ($cases.Count -eq 1 -and $cases[0].CommandElements[2] -is [Management.Automation.Language.ScriptBlockExpressionAst]) 'unique actual old case body'
            $expected=V-Definition $capture Capture-Value
            $loaded=@($commands|Where-Object Name -CEQ Capture-Value)[0].ScriptBlock.Ast
            $comparison.Value=[ordered]@{ExpectedNodeType=$expected.GetType().FullName;LoadedNodeType=$loaded.GetType().FullName;ExpectedText=$expected.Body.Extent.Text;LoadedText=$loaded.Extent.Text;OldExpectedOperand=$expected.Body.Extent.Text.Trim().TrimStart('{').TrimEnd('}').Trim();OldLoadedOperand=$loaded.Extent.Text.Trim().TrimStart('{').TrimEnd('}').Trim()}
            $caught=$null
            try{$values=@(& $cases[0].CommandElements[2].ScriptBlock.GetScriptBlock())}catch{$caught=$_}
            V-Need ($null -ne $caught) 'old checker must fail'
            $rootError=V-Error $caught
            V-Need ($rootError.Type -ceq 'System.Management.Automation.RuntimeException' -and $rootError.ErrorId -ceq 'Owner scope regression: actual body unchanged Capture-Value' -and $rootError.Message -ceq 'Owner scope regression: actual body unchanged Capture-Value') 'exact old checker failure'
            $rootError.Comparison=$comparison.Value
            foreach($n in $names){$groups.Add([ordered]@{Name=$n;Status=$(if($n -ceq $names[0]){'FAIL'}else{'NotRun'})})}
        }else{
            $fullPath=Join-Path $leaf owner-scope-verifier.tests.ps1
            $fullInput=Get-Content -Raw (Join-Path $packet owner-scope-verifier-scope.input.json)|ConvertFrom-Json -AsHashtable
            $codeHash=(Get-FileHash -LiteralPath $fullPath).Hash
            V-Need ($fullInput.TestSHA256 -ceq $codeHash -and $fullInput.RegressionTestSHA256 -ceq $testHash -and $fullInput.RegressionInputSHA256 -ceq $inputHash) 'new checker frozen identity'
            $fullTree=V-Tree $fullPath
            foreach($n in @('Assert-ODImportedDefinition','Invoke-ODVerifierCase','Convert-ODVerifierError','Convert-ODVerifierDiagnostic')){. ([scriptblock]::Create((V-Definition $fullTree $n).Extent.Text))}
            V-Case $names[0] {
                foreach($text in $actualDefinitions){$t=$null;$e=$null;$tree=[Management.Automation.Language.Parser]::ParseInput($text,[ref]$t,[ref]$e);$fn=@($tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst]},$false))[0];$cmd=@($commands|Where-Object {$_.Name -ceq $fn.Name})[0];Assert-ODImportedDefinition $fn $cmd.ScriptBlock.Ast $comparison}
            } $groups
            $active=$names[1];V-Case $active {$a=V-Function 'function Capture-Value($Value){return 1}' Capture-Value;$b=V-Function 'function script:Capture-Value($Value){return 2}' script:Capture-Value;V-Reject {Assert-ODImportedDefinition $a $b $comparison}} $groups
            $active=$names[2];V-Case $active {$a=V-Function 'function Capture-Value([int]$Value=1){return $Value}' Capture-Value;$b=V-Function 'function script:Capture-Value([string]$Value=2){return $Value}' script:Capture-Value;V-Reject {Assert-ODImportedDefinition $a $b $comparison}} $groups
            $active=$names[3];V-Case $active {$a=V-Function 'function Capture-Value($Value){return $Value}' Capture-Value;V-Reject {Assert-ODImportedDefinition $a $a.Body $comparison}} $groups
            $active=$names[4];V-Case $active {$a=V-Function 'function Capture-Value($Value){return $Value}' Capture-Value;$b=V-Function 'function script:Foreign($Value){return $Value}' script:Foreign;V-Reject {Assert-ODImportedDefinition $a $b $comparison}} $groups
            $state=[pscustomobject]@{CurrentGroup=$null;Groups=[Collections.Generic.List[object]]::new();RootError=$null;Comparison=[pscustomobject]@{Value=$null};RecordingError=$null;Stopped=$false}
            $active=$names[5];V-Case $active {
                Invoke-ODVerifierCase Pass {} $state
                Invoke-ODVerifierCase Fault {throw 'OD: controlled verifier fault'} $state
                V-Need ($state.RootError -is [Management.Automation.ErrorRecord]) 'actual caught first record'
                $first=$state.RootError
                $diagnostic=Convert-ODVerifierDiagnostic $state {param($r)throw 'OD: controlled projection failure'}
                V-Need ([object]::ReferenceEquals($first,$state.RootError) -and $state.RecordingError -is [Management.Automation.ErrorRecord] -and $state.RecordingError.Exception.Message -ceq 'OD: controlled projection failure' -and $diagnostic.RootError.Message -ceq 'OD: controlled verifier fault' -and $diagnostic.RecordingError.Message -ceq 'OD: controlled projection failure') 'first record survives diagnostic failure'
            } $groups
            $active=$names[6];V-Case $active {V-Need ($state.Groups.Count -eq 2 -and $state.Groups[0].Name -ceq 'Pass' -and $state.Groups[0].Status -ceq 'PASS' -and $state.Groups[1].Name -ceq 'Fault' -and $state.Groups[1].Status -ceq 'FAIL' -and $state.CurrentGroup -ceq 'Fault') 'completed group history retained'} $groups
            $active=$names[7];V-Case $active {$counter=[pscustomobject]@{Value=0};$first=$state.RootError;Invoke-ODVerifierCase Later {$counter.Value++} $state;V-Need ($counter.Value -eq 0 -and $state.Groups.Count -eq 3 -and $state.Groups[2].Status -ceq 'NotRun' -and [object]::ReferenceEquals($first,$state.RootError)) 'later body not executed'} $groups
        }
        $after=V-Fixtures $repo $binding.FixtureSHA256
        V-Need ((ConvertTo-Json -InputObject $fixtureHashes -Compress) -ceq (ConvertTo-Json -InputObject $after -Compress) -and (Get-FileHash -LiteralPath $testPath).Hash -ceq $testHash -and (Get-FileHash -LiteralPath $inputPath).Hash -ceq $inputHash) 'phase inputs unchanged'
        V-Need ($groups.Count -eq 8 -and (@($groups|ForEach-Object Name) -join '|') -ceq ($names -join '|')) 'eight ordered cases'
        return [pscustomobject][ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;CodeSHA256=$codeHash;FixtureSHA256=$fixtureHashes;Groups=$groups.ToArray();PASS=$(if($Phase -ceq 'Red'){0}else{8});FAIL=$(if($Phase -ceq 'Red'){1}else{0});NotRun=$(if($Phase -ceq 'Red'){7}else{0});ExitCode=$(if($Phase -ceq 'Red'){1}else{0});RootError=$rootError}
    }catch{
        $actual=$_
        return [pscustomobject][ordered]@{Schema=1;Phase=$Phase;Unexpected=$true;FailedGroup=$active;CompletedGroups=$groups.ToArray();RootError=(V-Error $actual);Comparison=$comparison.Value;TestSHA256=$testHash;InputSHA256=$inputHash;CodeSHA256=$codeHash;FixtureSHA256=$fixtureHashes}
    }
}
