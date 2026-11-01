# Finite projection auxiliary. Definitions only; no writer/profile/runtime body.
function E-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Projection regression: '+$Label)}}
function E-Tree([string]$Path){
    $t=$null;$e=$null;$ast=[Management.Automation.Language.Parser]::ParseFile($Path,[ref]$t,[ref]$e)
    E-Need ($e.Count -eq 0) ('parse '+$Path);return $ast
}
function E-Definition($Tree,[string]$Name){
    $n=@($Tree.FindAll({param($x)$x -is [Management.Automation.Language.FunctionDefinitionAst] -and $x.Name -ceq $Name},$true))
    E-Need ($n.Count -eq 1) ('unique '+$Name);return $n[0]
}
function E-RealFault {
    try{throw 'OD: literal projection control'}catch{return $_}
}
function E-Error($Record){
    if($null -eq $Record){return $null}
    $position=$null;if($null -ne $Record.InvocationInfo){$position=$Record.InvocationInfo.PositionMessage}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$position;Stack=$Record.ScriptStackTrace}
}
function E-Same($Left,$Right){return (ConvertTo-Json -InputObject $Left -Depth 30 -Compress) -ceq (ConvertTo-Json -InputObject $Right -Depth 30 -Compress)}
function E-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    $State.CurrentGroup=$Name
    try{$v=@(& $Body);E-Need ($v.Count -eq 0) ('no output '+$Name);$State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})}
    catch{if($null -eq $State.RootError){$State.RootError=$_};$State.Stopped=$true;$State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})}
}
function E-Negative($Good,[string]$Member,[bool]$Remove,[string]$Relative){
    $copy=[ordered]@{};foreach($k in $Good.Keys){$copy[$k]=$Good[$k]}
    if($Remove){$copy.Remove($Member)}else{$copy[$Member]=$null}
    $caught=$null;$values=@()
    try{$values=@(Driver-ErrorLocation $copy @($Relative))}catch{$caught=$_}
    E-Need ($null -ne $caught -and $values.Count -eq 0) ('consumer rejects '+$Member)
    $normal=$caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.FullyQualifiedErrorId -ceq 'Actual error source/stack required' -and $caught.Exception.Message -ceq 'Actual error source/stack required'
    $missing=$Remove -and $caught.Exception.GetType().FullName -ceq 'System.Management.Automation.PropertyNotFoundException' -and $caught.FullyQualifiedErrorId -ceq 'PropertyNotFoundStrict' -and $caught.Exception.Message.Contains("'"+$Member+"'")
    E-Need ($normal -or $missing) ('exact consumer rejection '+$Member)
}
function Invoke-ODErrorProjectionRegression {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $state=[pscustomobject]@{CurrentGroup='Setup';Groups=[Collections.Generic.List[object]]::new();RootError=$null;RecordingError=$null;Stopped=$false}
    $testHash=$null;$inputHash=$null;$codeHash=$null;$postHash=$null
    $names=@('NullInvocationInfoProjects','RealCaughtMetadataExact','NullStackProjects','SharedExceptionIdentity','NestedJsonRoundTrip','NullPositionRejected','MissingPositionRejected','NullStackRejected','MissingStackRejected','UnsupportedValueRejected')
    try{
        $leaf=Split-Path (Get-Command Invoke-ODErrorProjectionRegression -CommandType Function).ScriptBlock.File
        $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'))
        $packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
        $testRelative='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/error-projection.tests.ps1'
        $testPath=Join-Path $repo $testRelative;$inputPath=Join-Path $packet error-projection.input.json
        $testHash=(Get-FileHash $testPath).Hash;$inputHash=(Get-FileHash $inputPath).Hash
        $binding=Get-Content -Raw $inputPath|ConvertFrom-Json -AsHashtable
        E-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen regression identity'
        foreach($p in $binding.HistoricalFiles.Keys){E-Need ((Get-FileHash -LiteralPath (Join-Path $packet $p)).Hash -ceq $binding.HistoricalFiles[$p]) ('history '+$p)}
        $expected=$binding.BaselineSources
        if($Phase -ceq 'Green'){
            $scope=Get-Content -Raw (Join-Path $packet error-projection-scope.input.json)|ConvertFrom-Json -AsHashtable
            E-Need ($scope.RegressionTestSHA256 -ceq $testHash -and $scope.RegressionInputSHA256 -ceq $inputHash) 'GREEN binding'
            $expected=$scope.CandidateSources
        }
        foreach($p in $expected.Keys){E-Need ((Get-FileHash -LiteralPath (Join-Path $repo $p)).Hash -ceq $expected[$p]) ('current source '+$p)}
        $captureRelative='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/capture.ps1'
        $capturePath=Join-Path $repo $captureRelative;$codeHash=(Get-FileHash $capturePath).Hash
        if($Phase -ceq 'Red'){E-Need ($codeHash -ceq $binding.OldCaptureSHA256) 'old capture frozen'}
        $driverRelative='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/driver.ps1'
        $oldDriver=Join-Path (Join-Path $packet snapshots/writer-scope-rebind) $driverRelative
        E-Need ((Get-FileHash $oldDriver).Hash -ceq $binding.OldDriverSHA256) 'old consumer frozen'
        $oldTree=E-Tree $oldDriver;$driverTree=E-Tree (Join-Path $repo $driverRelative)
        foreach($n in @('Driver-Need','Driver-ErrorLocation')){
            $fn=E-Definition $driverTree $n
            E-Need ($fn.Extent.Text -ceq (E-Definition $oldTree $n).Extent.Text) ('strict consumer unchanged '+$n)
            . ([scriptblock]::Create($fn.Extent.Text))
        }
        . ([scriptblock]::Create((E-Definition (E-Tree $capturePath) Capture-Value).Extent.Text))
        $real=E-RealFault
        E-Need ($real -is [Management.Automation.ErrorRecord] -and $null -ne $real.InvocationInfo) 'actual caught record'
        $a=[Management.Automation.ErrorRecord]::new($real.Exception,'ContractOnly',[Management.Automation.ErrorCategory]::NotSpecified,$null)
        $b=[Management.Automation.ErrorRecord]::new($real.Exception,'ContractOnly',[Management.Automation.ErrorCategory]::NotSpecified,$null)
        $originalException=$real.Exception;$originalPosition=$real.InvocationInfo.PositionMessage;$originalStack=$real.ScriptStackTrace
        E-Case $names[0] {
            E-Need ($null -eq $a.InvocationInfo) 'literal synthetic lacks InvocationInfo'
            $p=Capture-Value $a
            E-Need ((@($p.Keys) -join '|') -ceq 'Type|Message|ErrorId|Position|Stack' -and $null -eq $p.Position) 'truthful absent Position'
        } $state
        if($Phase -ceq 'Red'){
            E-Need ($state.RootError -is [Management.Automation.ErrorRecord]) 'old projection actually fails'
            $r=$state.RootError
            E-Need ($r.Exception.GetType().FullName -ceq 'System.Management.Automation.PropertyNotFoundException' -and $r.FullyQualifiedErrorId -ceq 'PropertyNotFoundStrict' -and $r.Exception.Message -ceq "The property 'PositionMessage' cannot be found on this object. Verify that the property exists.") 'exact original failure'
            for($i=1;$i -lt $names.Count;$i++){E-Case $names[$i] {} $state}
        }else{
            E-Case $names[1] {
                $p=Capture-Value $real
                E-Need ((@($p.Keys) -join '|') -ceq 'Type|Message|ErrorId|Position|Stack' -and $p.Type -ceq $real.Exception.GetType().FullName -and $p.Message -ceq 'OD: literal projection control' -and $p.ErrorId -ceq $real.FullyQualifiedErrorId -and $p.Position -ceq $originalPosition -and $p.Stack -ceq $originalStack) 'real metadata exact'
                Driver-ErrorLocation $p @($testRelative)
            } $state
            E-Case $names[2] {E-Need ($null -eq $a.ScriptStackTrace) 'literal absent Stack';$p=Capture-Value $a;E-Need ($null -eq $p.Stack) 'null Stack not invented'} $state
            E-Case $names[3] {
                E-Need (-not [object]::ReferenceEquals($a,$b) -and [object]::ReferenceEquals($a.Exception,$b.Exception)) 'two wrappers one exception'
                $pa=Capture-Value $a;$pb=Capture-Value $b
                E-Need ((E-Same $pa $pb) -and $pa.ErrorId -ceq 'ContractOnly' -and [object]::ReferenceEquals($a.Exception,$originalException) -and [object]::ReferenceEquals($b.Exception,$originalException) -and -not [object]::ReferenceEquals($a,$b)) 'projection retains live identity facts'
            } $state
            E-Case $names[4] {
                $literal=[ordered]@{ContractOnly=$true;Nested=[ordered]@{Wrappers=@($a,$b);Actual=$real;Empty=$null;Array=@('literal',$null,7)}}
                $projected=Capture-Value $literal
                $round=ConvertTo-Json -InputObject $projected -Depth 30|ConvertFrom-Json -AsHashtable
                E-Need ((@($round.Keys) -join '|') -ceq 'ContractOnly|Nested' -and $round.ContractOnly -is [bool] -and $round.ContractOnly -and (@($round.Nested.Keys) -join '|') -ceq 'Wrappers|Actual|Empty|Array') 'nested keys and order'
                E-Need ($round.Nested.Wrappers -is [array] -and $round.Nested.Wrappers.Count -eq 2 -and $null -eq $round.Nested.Wrappers[0].Position -and $null -eq $round.Nested.Wrappers[1].Stack -and $null -eq $round.Nested.Empty -and $round.Nested.Array.Count -eq 3 -and $round.Nested.Array[0] -ceq 'literal' -and $null -eq $round.Nested.Array[1] -and $round.Nested.Array[2] -eq 7 -and $round.Nested.Actual.Position -ceq $originalPosition -and $round.Nested.Actual.Stack -ceq $originalStack) 'nested values/null slots exact'
                E-Need ((E-Same $projected $round) -and [object]::ReferenceEquals($literal.Nested.Wrappers[0],$a) -and [object]::ReferenceEquals($literal.Nested.Wrappers[1],$b) -and [object]::ReferenceEquals($literal.Nested.Actual,$real) -and [object]::ReferenceEquals($real.Exception,$originalException) -and $real.InvocationInfo.PositionMessage -ceq $originalPosition -and $real.ScriptStackTrace -ceq $originalStack) 'round trip without mutation'
            } $state
            E-Case $names[5] {E-Negative (Capture-Value $real) Position $false $testRelative} $state
            E-Case $names[6] {E-Negative (Capture-Value $real) Position $true $testRelative} $state
            E-Case $names[7] {E-Negative (Capture-Value $real) Stack $false $testRelative} $state
            E-Case $names[8] {E-Negative (Capture-Value $real) Stack $true $testRelative} $state
            E-Case $names[9] {
                $caught=$null;try{$null=Capture-Value ([object]::new())}catch{$caught=$_}
                E-Need ($null -ne $caught -and $caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq 'Unsupported evidence value' -and $caught.FullyQualifiedErrorId -ceq 'Unsupported evidence value') 'unsupported remains rejected'
            } $state
            E-Need (-not $state.Stopped) 'all GREEN groups pass'
        }
        $postHash=(Get-FileHash $capturePath).Hash
        foreach($p in $expected.Keys){E-Need ((Get-FileHash -LiteralPath (Join-Path $repo $p)).Hash -ceq $expected[$p]) ('post source '+$p)}
        E-Need ($postHash -ceq $codeHash -and (Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash -and $state.Groups.Count -eq 10) 'inputs unchanged/ten groups'
        return [pscustomobject][ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;CodeSHA256=$codeHash;Groups=$state.Groups.ToArray();PASS=$(if($Phase -ceq 'Red'){0}else{10});FAIL=$(if($Phase -ceq 'Red'){1}else{0});NotRun=$(if($Phase -ceq 'Red'){9}else{0});ExitCode=$(if($Phase -ceq 'Red'){1}else{0});RootError=(E-Error $state.RootError)}
    }catch{
        if($null -eq $state.RootError){$state.RootError=$_}else{$state.RecordingError=$_}
        try{if($null -ne $codeHash){$postHash=(Get-FileHash $capturePath).Hash}}catch{if($null -eq $state.RecordingError){$state.RecordingError=$_}}
        return [pscustomobject][ordered]@{Schema=1;Phase=$Phase;Unexpected=$true;ExitCode=2;FailedGroup=$state.CurrentGroup;CompletedGroups=$state.Groups.ToArray();RootError=(E-Error $state.RootError);RecordingError=(E-Error $state.RecordingError);TestSHA256=$testHash;InputSHA256=$inputHash;CodeSHA256=$codeHash;PostCodeSHA256=$postHash}
    }
}
